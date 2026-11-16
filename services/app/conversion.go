// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"svnova/models"
	"svnova/services/storage"
	"svnova/services/svn"
)

func (a *App) recoverConversions() error {
	var rows []models.MirrorConversion
	if err := a.DB.Find(&rows); err != nil {
		return err
	}
	for _, row := range rows {
		var repo models.Repository
		ok, err := a.DB.ID(row.RepositoryID).Get(&repo)
		if err != nil {
			return err
		}
		if !ok || repo.Kind != "local" {
			return errors.New("conversion recovery repository state mismatch")
		}
		if !storage.Within(a.Config.Repositories, repo.Path) {
			return errors.New("conversion recovery outside managed root")
		}
		if err = storage.RejectLinks(repo.Path); err != nil {
			return err
		}
		var snapshot svn.ConversionSnapshot
		if err = json.Unmarshal([]byte(row.Snapshot), &snapshot); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		err = a.SVN.RestoreConversion(ctx, repo.Path, snapshot)
		cancel()
		if err != nil {
			return err
		}
		if _, err = a.DB.ID(row.RepositoryID).Delete(new(models.MirrorConversion)); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) convertMirror(w http.ResponseWriter, r *http.Request, u *models.User, repo *models.Repository) error {
	if !models.Owner(a.DB, u, repo.OrgID) {
		return forbidden()
	}
	var in conversionInput
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if svn.ValidateRemote(in.Remote) != nil {
		return &httpError{400, "镜像地址须为不含账号密码的 HTTP/HTTPS 仓库根地址"}
	}
	if in.Interval < 0 || (in.Interval > 0 && in.Interval < 600) {
		return &httpError{400, "同步间隔须为 0 或至少 600 秒"}
	}
	if !in.FullAccess {
		return &httpError{400, "请确认远端账号可读取完整仓库及全部修订，且本地仓库不再接受其他写入"}
	}
	if pending, err := a.DB.Exist(new(models.MirrorConversion)); err != nil {
		return err
	} else if pending {
		return &httpError{409, "上次转换尚未恢复，请重启服务完成恢复后再试"}
	}
	if !a.Worker.TryLock() {
		return &httpError{409, "仓库任务正在执行，请稍后重试"}
	}
	defer a.Worker.Unlock()
	if err := a.recoverConversions(); err != nil {
		return &httpError{409, "上次转换尚未恢复，请检查服务日志并重启服务"}
	}
	var current models.Repository
	ok, err := a.DB.ID(repo.ID).Get(&current)
	if err != nil {
		return err
	}
	if !ok || current.Kind != "local" || current.Status != "ready" || current.Halted {
		return &httpError{409, "仅可将已就绪且完整性正常的本地仓库转换为镜像"}
	}
	active, err := a.DB.Where("repository_id=? AND state IN (?,?)", repo.ID, "queued", "running").Exist(new(models.Task))
	if err != nil {
		return err
	}
	if active {
		return &httpError{409, "已有任务正在排队或执行，请完成后再转换"}
	}
	if !storage.Within(a.Config.Repositories, current.Path) {
		return &httpError{409, "仓库不在受管目录内，无法转换"}
	}
	if err = storage.RejectLinks(current.Path); err != nil {
		return &httpError{409, "仓库目录包含链接或无法访问，无法安全转换"}
	}
	var actor models.User
	if ok, err = a.DB.ID(u.ID).Get(&actor); err != nil {
		return err
	}
	if !ok || !models.Owner(a.DB, &actor, current.OrgID) {
		return forbidden()
	}
	in.Password = a.Vault.Seal(in.Password)
	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}
	task := models.Task{RepositoryID: current.ID, Kind: "conversion", Trigger: "manual", State: "queued", Started: time.Now().Unix(), ActorID: actor.ID, ActorName: actor.Name, Source: string(payload)}
	if _, err = a.DB.Insert(&task); err != nil {
		return err
	}
	return send(w, map[string]any{"taskID": task.ID, "queued": true, "taskURL": "/repositories/" + strconv.FormatInt(current.ID, 10) + "?tab=tasks"})
}

type conversionInput struct {
	Remote, Username, Password string
	SkipTLS                    bool
	Interval                   int64
	FullAccess                 bool
}

func (a *App) executeConversion(parent context.Context, destination *models.Repository, task *models.Task) (result error) {
	candidate := *destination
	current := &candidate
	defer func() {
		if result == nil {
			*destination = candidate
		}
	}()
	defer func() {
		if result != nil {
			result = conversionError(result)
		}
	}()
	var in conversionInput
	if err := json.Unmarshal([]byte(task.Source), &in); err != nil {
		return err
	}
	password, err := a.Vault.Open(in.Password)
	if err != nil {
		return errors.New("cannot decrypt conversion credential")
	}
	in.Password = password
	var actor models.User
	ok, err := a.DB.ID(task.ActorID).Get(&actor)
	if err != nil {
		return err
	}
	if !ok || !models.Owner(a.DB, &actor, current.OrgID) {
		return forbidden()
	}
	repo := current
	if current.Kind != "local" || current.Status != "ready" || current.Halted {
		return errors.New("repository state changed before conversion")
	}
	if !storage.Within(a.Config.Repositories, current.Path) {
		return errors.New("repository outside managed root")
	}
	if err = storage.RejectLinks(current.Path); err != nil {
		return err
	}
	timeout := a.Config.MirrorConversionTimeout
	if timeout == 0 {
		timeout = 24 * time.Hour
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	runner := a.SVN
	runner.Timeout = timeout
	var outputMu sync.Mutex
	var lastOutput time.Time
	notify := runner.OnOutput
	runner.OnOutput = func() {
		outputMu.Lock()
		defer outputMu.Unlock()
		if notify != nil {
			notify()
		}
		if time.Since(lastOutput) < 2*time.Second {
			return
		}
		lastOutput = time.Now()
		_, _ = a.DB.ID(task.ID).Cols("progress_updated").Update(&models.Task{ProgressUpdated: lastOutput.Unix()})
	}
	phase := func(value string) error {
		task.Phase = value
		task.ProgressStarted = time.Now().Unix()
		task.ProgressUpdated = task.ProgressStarted
		_, err := a.DB.ID(task.ID).Cols("phase", "progress_started", "progress_updated").Update(task)
		return err
	}
	if err = phase("conversion-local"); err != nil {
		return err
	}
	credentials := svn.Credentials{Username: in.Username, Password: in.Password, SkipTLS: in.SkipTLS}
	uuid, revision, err := runner.LocalInfo(ctx, current.Path)
	if err != nil {
		return err
	}
	if uuid != current.UUID || revision != current.Revision {
		return &httpError{409, "本地仓库 UUID 或修订与记录不一致，请核查外部写入"}
	}
	if err = phase("conversion-remote"); err != nil {
		return err
	}
	remoteUUID, root, head, err := runner.Info(ctx, in.Remote, credentials)
	if err != nil {
		return err
	}
	if !svn.SameRepositoryURL(root, in.Remote) {
		return &httpError{400, "远端地址不是完整 SVN 仓库根目录"}
	}
	if err = svn.ValidateConversionIdentity(uuid, revision, remoteUUID, head); err != nil {
		return conversionError(err)
	}
	snapshot, err := runner.SnapshotConversion(ctx, current.Path)
	if err != nil {
		return conversionError(err)
	}
	base, err := runner.RemoteRevisionProperties(ctx, in.Remote, credentials)
	if err != nil {
		return err
	}
	if len(base) != len(snapshot.Properties) {
		return &httpError{409, "本地与远端修订 r0 属性不一致，不能转换镜像"}
	}
	for key, value := range snapshot.Properties {
		if remote, ok := base[key]; !ok || !bytes.Equal(remote, value) {
			return &httpError{409, "本地与远端修订 r0 属性不一致，不能转换镜像"}
		}
	}
	if err = phase("conversion-prepare"); err != nil {
		return err
	}
	b, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	tx := a.DB.NewSession()
	defer tx.Close()
	if err = tx.Begin(); err != nil {
		return err
	}
	if _, err = tx.Exec("INSERT OR IGNORE INTO migration(id,applied) VALUES(4,?)", time.Now().Unix()); err != nil {
		return err
	}
	if _, err = tx.Insert(&models.MirrorConversion{RepositoryID: repo.ID, Snapshot: string(b)}); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = phase("conversion-recovery")
			if err := a.recoverConversions(); err != nil {
				log.Printf("Mirror conversion recovery failed for repository %d; journal retained", repo.ID)
				result = &httpError{500, "转换未完成且恢复尚未完成，原仓库仍保留；请停止操作并检查服务日志后重启"}
			}
		}
	}()
	if err = phase("conversion-initialize"); err != nil {
		return err
	}
	if err = runner.InitializeExistingMirror(ctx, current.Path, in.Remote, credentials, uuid, revision); err != nil {
		return conversionError(err)
	}
	if err = phase("conversion-final"); err != nil {
		return err
	}
	remoteUUID, root, head, err = runner.Info(ctx, in.Remote, credentials)
	if err != nil {
		return err
	}
	if !svn.SameRepositoryURL(root, in.Remote) {
		return errors.New("remote repository root changed")
	}
	if err = svn.ValidateConversionIdentity(uuid, revision, remoteUUID, head); err != nil {
		return conversionError(err)
	}
	finalUUID, finalRevision, err := runner.LocalInfo(ctx, current.Path)
	if err != nil {
		return err
	}
	if finalUUID != uuid || finalRevision != revision {
		return &httpError{409, "校验期间仓库历史发生变化，已取消转换"}
	}
	if err = runner.CheckExistingMirror(ctx, current.Path, in.Remote, uuid, revision); err != nil {
		return conversionError(err)
	}
	current.Kind = "mirror"
	current.Remote = in.Remote
	current.Username = in.Username
	current.Password = a.Vault.Seal(in.Password)
	current.SkipTLS = in.SkipTLS
	current.Interval = in.Interval
	current.NextSync = time.Now().Unix()
	tx2 := a.DB.NewSession()
	defer tx2.Close()
	if err = tx2.Begin(); err != nil {
		return err
	}
	if _, err = tx2.Exec("UPDATE repository SET id=id WHERE id=?", current.ID); err != nil {
		return err
	}
	// Hold SQLite's writer lock while rechecking the current organization role.
	actor = models.User{}
	if ok, err = tx2.ID(task.ActorID).Get(&actor); err != nil {
		return err
	}
	if !ok || !actor.Active {
		return forbidden()
	}
	if !actor.Admin {
		var role struct{ N int64 }
		if _, err = tx2.SQL("SELECT COUNT(*) AS n FROM team t JOIN team_member tm ON tm.team_id=t.id JOIN membership m ON m.org_id=t.org_id AND m.user_id=tm.user_id WHERE t.org_id=? AND t.is_owner_team=? AND tm.user_id=?", current.OrgID, true, actor.ID).Get(&role); err != nil {
			return err
		}
		if role.N == 0 {
			return forbidden()
		}
	}
	changed, err := tx2.ID(current.ID).Where("kind=? AND status=? AND halted=? AND org_id=? AND path=? AND uuid=? AND revision=?", "local", "ready", false, current.OrgID, current.Path, uuid, revision).Cols("kind", "remote", "username", "password", "skip_tls", "interval", "next_sync").Update(current)
	if err != nil {
		return err
	}
	if changed != 1 {
		return &httpError{409, "仓库状态已改变，转换已取消"}
	}
	if _, err = tx2.Insert(&models.Task{RepositoryID: current.ID, Kind: "sync", Trigger: "manual", State: "queued", Started: time.Now().Unix()}); err != nil {
		return err
	}
	if _, err = tx2.ID(current.ID).Delete(new(models.MirrorConversion)); err != nil {
		return err
	}
	finished := time.Now()
	if _, err = tx2.ID(task.ID).Cols("state", "phase", "finished", "current_revision", "total_revision", "source").Update(&models.Task{State: "succeeded", Phase: "completed", Finished: finished.Unix(), CurrentRevision: revision, TotalRevision: revision}); err != nil {
		return err
	}
	if _, err = tx2.Where("task_id=? AND state=?", task.ID, "running").Cols("state", "finished", "after_revision", "revision_known").Update(&models.TaskExecution{State: "succeeded", Finished: finished.UnixMilli(), AfterRevision: revision, RevisionKnown: true}); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = tx2.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func conversionError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return &httpError{504, "镜像转换校验超时，转换未完成，原本地仓库配置保留"}
	}
	if errors.Is(err, context.Canceled) {
		return &httpError{409, "镜像转换校验已中断，原本地仓库配置保留"}
	}
	var failure *svn.ConversionError
	if errors.As(err, &failure) {
		message := map[string]string{"uuid": "远端与本地 UUID 不一致，不能转换镜像", "revision": "远端修订落后于本地，不能丢弃本地历史进行转换", "metadata": "本地仓库已有 svnsync 元数据，请由管理员核查，不能覆盖现有镜像来源", "hook": "本地仓库已有提交或修订属性钩子，不能安全接管；请由管理员核查后再转换", "initialization": "镜像元数据初始化校验失败，已取消转换"}[failure.Code]
		return &httpError{409, message}
	}
	if strings.Contains(err.Error(), "output exceeds limit") {
		return &httpError{409, "命令输出超过安全上限，未转换仓库"}
	}
	return err
}
