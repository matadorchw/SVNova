// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"svnova/models"
	"svnova/services/security"
	"svnova/services/storage"
	"svnova/services/svn"
	"svnova/services/tasks"
)

func (a *App) api(w http.ResponseWriter, r *http.Request, u *models.User) error {
	p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(p) < 3 || p[1] != "v1" {
		return &httpError{404, "接口不存在"}
	}
	resource := p[2]
	if resource == "tasks" && len(p) == 3 && r.Method == "GET" {
		return a.taskCenterAPI(w, r, u)
	}
	if resource == "activity" && len(p) == 4 && p[3] == "trend" && r.Method == "GET" {
		return a.activityTrend(w, r, u)
	}
	if resource == "user" && len(p) == 4 && p[3] == "password" && r.Method == "POST" {
		return a.changePassword(w, r, u)
	}
	if resource == "activity" && len(p) == 4 && p[3] == "heatmap" && r.Method == "GET" {
		return a.activityHeatmap(w, r, u)
	}
	if resource == "me" && r.Method == "GET" {
		return send(w, u)
	}
	if resource == "repositories" {
		return a.repositories(w, r, u, p[3:])
	}
	if resource == "organizations" {
		return a.organizations(w, r, u, p[3:])
	}
	if resource == "teams" {
		return a.teams(w, r, u, p[3:])
	}
	if resource == "admin" {
		if !u.Admin {
			return forbidden()
		}
		return a.admin(w, r, u, p[3:])
	}
	if resource == "users" && r.Method == "GET" {
		var users []models.User
		if err := a.DB.Cols("id", "name", "display_name").Where("active=?", true).Find(&users); err != nil {
			return err
		}
		return send(w, users)
	}
	return &httpError{404, "接口不存在"}
}
func (a *App) organizations(w http.ResponseWriter, r *http.Request, u *models.User, p []string) error {
	if len(p) == 0 {
		switch r.Method {
		case "GET":
			var orgs []models.Organization
			var err error
			if u.Admin {
				err = a.DB.Find(&orgs)
			} else {
				err = a.DB.SQL("SELECT o.* FROM organization o JOIN membership m ON m.org_id=o.id WHERE m.user_id=?", u.ID).Find(&orgs)
			}
			if err != nil {
				return err
			}
			type orgView struct {
				models.Organization
				CanManage bool
			}
			out := []orgView{}
			for _, o := range orgs {
				out = append(out, orgView{o, models.Owner(a.DB, u, o.ID)})
			}
			return send(w, out)
		case "POST":
			if !u.Admin {
				return forbidden()
			}
			var in struct {
				Name    string
				OwnerID int64
			}
			if err := decode(w, r, &in); err != nil {
				return err
			}
			if !validName(in.Name) {
				return errors.New("invalid organization name")
			}
			if err := storage.ValidateName(in.Name); err != nil {
				return &httpError{400, "组织名称不能用作 Windows 目录名：请避免设备保留名和末尾的点"}
			}
			if in.OwnerID != 0 && in.OwnerID != u.ID {
				return errors.New("organization creator must be the initial owner")
			}
			if in.OwnerID == 0 {
				in.OwnerID = u.ID
			}
			owner := new(models.User)
			ok, err := a.DB.ID(in.OwnerID).Get(owner)
			if err != nil || !ok || !owner.Active {
				return errors.New("invalid owner")
			}
			tx := a.DB.NewSession()
			defer tx.Close()
			if err = tx.Begin(); err != nil {
				return err
			}
			if _, err = tx.Exec("UPDATE migration SET id=id"); err != nil {
				return err
			}
			duplicate, err := tx.Where("lower(name)=?", strings.ToLower(in.Name)).Exist(new(models.Organization))
			if err != nil {
				return err
			}
			if duplicate {
				return &httpError{409, "已存在同名组织（名称不区分大小写）"}
			}
			o := &models.Organization{Name: in.Name, Created: time.Now().Unix()}
			if _, err = tx.Insert(o); err != nil {
				return err
			}
			if _, err = tx.Insert(&models.Membership{OrgID: o.ID, UserID: u.ID}); err != nil {
				return err
			}
			ownerTeam := &models.Team{OrgID: o.ID, Name: models.OwnerTeamName, IsOwnerTeam: true}
			if _, err = tx.Insert(ownerTeam); err != nil {
				return err
			}
			if _, err = tx.Insert(&models.TeamMember{TeamID: ownerTeam.ID, UserID: u.ID}); err != nil {
				return err
			}
			if err = tx.Commit(); err != nil {
				return err
			}
			return send(w, o)
		}
	}
	if len(p) != 2 {
		return &httpError{404, "接口不存在"}
	}
	orgID := id(p[0])
	selfLeave := p[1] == "leave" && r.Method == "POST"
	if selfLeave {
		p[1] = "members"
	}
	if !selfLeave && !models.Owner(a.DB, u, orgID) {
		return forbidden()
	}
	if p[1] == "delete" {
		return a.deleteOrganization(w, r, u, orgID)
	}
	if p[1] == "settings" {
		return a.organizationSettings(w, r, u, orgID)
	}
	switch p[1] {
	case "members":
		if r.Method == "GET" {
			var members []models.Membership
			if err := a.DB.Where("org_id=?", orgID).Find(&members); err != nil {
				return err
			}
			return send(w, members)
		}
		var in struct {
			UserID int64
			Remove bool
		}
		if err := decode(w, r, &in); err != nil {
			return err
		}
		if selfLeave {
			in.UserID = u.ID
			in.Remove = true
		}
		target := new(models.User)
		ok, err := a.DB.ID(in.UserID).Get(target)
		if err != nil || !ok {
			return errors.New("user not found")
		}
		tx := a.DB.NewSession()
		defer tx.Close()
		if err = tx.Begin(); err != nil {
			return err
		}
		if _, err = tx.Exec("UPDATE organization SET id=id WHERE id=?", orgID); err != nil {
			return err
		}
		if !selfLeave {
			if err = checkOrganizationOwner(tx, u, orgID); err != nil {
				return err
			}
		}
		if !in.Remove {
			exists, err := tx.ID(in.UserID).Exist(new(models.User))
			if err != nil {
				return err
			}
			if !exists {
				return &httpError{404, "用户不存在或已被删除"}
			}
		}
		old := new(models.Membership)
		ok, err = tx.Where("org_id=? AND user_id=?", orgID, in.UserID).Get(old)
		if err != nil {
			return err
		}
		ownerTeam := new(models.Team)
		if _, err = tx.Where("org_id=? AND is_owner_team=?", orgID, true).Get(ownerTeam); err != nil {
			return err
		}
		isOwner, err := tx.Where("team_id=? AND user_id=?", ownerTeam.ID, in.UserID).Exist(new(models.TeamMember))
		if err != nil {
			return err
		}
		if selfLeave && !ok {
			return forbidden()
		}
		if in.Remove && isOwner {
			count, e := tx.Where("team_id=?", ownerTeam.ID).Count(new(models.TeamMember))
			if e != nil {
				return e
			}
			if count <= 1 {
				return &httpError{409, "无法离开组织：您是最后一位 Owners 成员，请先将其他成员加入 Owners 团队。"}
			}
		}
		if in.Remove {
			_, err = tx.Where("org_id=? AND user_id=?", orgID, in.UserID).Delete(new(models.Membership))
			if err == nil {
				_, err = tx.Exec("DELETE FROM team_member WHERE user_id=? AND team_id IN (SELECT id FROM team WHERE org_id=?)", in.UserID, orgID)
			}
		} else if ok {
			err = nil
		} else {
			_, err = tx.Insert(&models.Membership{OrgID: orgID, UserID: in.UserID})
		}
		if err != nil {
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
		return send(w, map[string]bool{"ok": true})
	case "teams":
		if r.Method == "GET" {
			var teams []models.Team
			if err := a.DB.Where("org_id=?", orgID).Find(&teams); err != nil {
				return err
			}
			return send(w, teams)
		}
		var in struct{ Name, Description string }
		if err := decode(w, r, &in); err != nil {
			return err
		}
		if !validName(in.Name) {
			return &httpError{400, "团队名称不能包含空格；须以字母或数字开头，只能包含字母、数字、点、下划线和连字符，最长 80 个字符"}
		}
		if strings.EqualFold(in.Name, models.OwnerTeamName) || strings.EqualFold(in.Name, "new") {
			return &httpError{400, "Owners 和 new 是保留团队名称，请使用其他名称"}
		}
		if len([]rune(in.Description)) > 255 {
			return &httpError{400, "团队描述不能超过 255 个字符"}
		}
		tx := a.DB.NewSession()
		defer tx.Close()
		if err := tx.Begin(); err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE organization SET id=id WHERE id=?", orgID); err != nil {
			return err
		}
		exists, err := tx.ID(orgID).Exist(new(models.Organization))
		if err != nil {
			return err
		}
		if !exists {
			return &httpError{404, "组织不存在"}
		}
		if err = checkOrganizationOwner(tx, u, orgID); err != nil {
			return err
		}
		exists, err = tx.Where("org_id=? AND lower(name)=?", orgID, strings.ToLower(in.Name)).Exist(new(models.Team))
		if err != nil {
			return err
		}
		if exists {
			return &httpError{409, "此组织中已存在同名团队，请更换团队名称（名称不区分大小写）"}
		}
		team := &models.Team{OrgID: orgID, Name: in.Name, Description: strings.TrimSpace(in.Description)}
		if _, err = tx.Insert(team); err != nil {
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
		return send(w, team)
	}
	return &httpError{405, "不支持此操作"}
}
func (a *App) teams(w http.ResponseWriter, r *http.Request, u *models.User, p []string) error {
	if len(p) != 2 {
		return &httpError{404, "接口不存在"}
	}
	team := new(models.Team)
	ok, err := a.DB.ID(id(p[0])).Get(team)
	if err != nil {
		return err
	}
	if !ok {
		return &httpError{404, "团队不存在"}
	}
	if p[1] == "leave" {
		if r.Method != "POST" {
			return &httpError{405, "不支持此操作"}
		}
		return a.changeTeamMember(w, u, team, u.ID, true, true)
	}
	if !models.Owner(a.DB, u, team.OrgID) {
		return forbidden()
	}
	if r.Method == "GET" {
		if p[1] == "members" {
			var x []models.TeamMember
			err = a.DB.Where("team_id=?", team.ID).Find(&x)
			if err != nil {
				return err
			}
			return send(w, x)
		}
		if p[1] == "repositories" {
			var x []models.TeamRepository
			err = a.DB.Where("team_id=?", team.ID).Find(&x)
			if err != nil {
				return err
			}
			return send(w, x)
		}
	}
	if r.Method != "POST" {
		return &httpError{405, "不支持此操作"}
	}
	var in struct {
		UserID       int64
		Username     string
		RepositoryID int64
		Remove       bool
	}
	if err = decode(w, r, &in); err != nil {
		return err
	}
	switch p[1] {
	case "members":
		if in.UserID == 0 && !in.Remove {
			target := new(models.User)
			ok, err := a.DB.Where("name=? AND active=?", strings.TrimSpace(in.Username), true).Get(target)
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("找不到启用的用户")
			}
			in.UserID = target.ID
		}
		return a.changeTeamMember(w, u, team, in.UserID, in.Remove, false)
	case "repositories":
		if team.IsOwnerTeam {
			return errors.New("Owners team automatically manages all repositories")
		}
		ok, err = a.DB.Where("id=? AND org_id=?", in.RepositoryID, team.OrgID).Exist(new(models.Repository))
		if err != nil {
			return err
		}
		if !ok {
			return forbidden()
		}
		if in.Remove {
			_, err = a.DB.Where("team_id=? AND repository_id=?", team.ID, in.RepositoryID).Delete(new(models.TeamRepository))
		} else {
			_, err = a.DB.Insert(&models.TeamRepository{TeamID: team.ID, RepositoryID: in.RepositoryID})
		}
	default:
		return &httpError{404, "接口不存在"}
	}
	if err != nil {
		return err
	}
	return send(w, map[string]bool{"ok": true})
}
func (a *App) repositories(w http.ResponseWriter, r *http.Request, u *models.User, p []string) error {
	if len(p) == 0 {
		if r.Method == "GET" {
			var all []models.Repository
			if err := a.DB.Find(&all); err != nil {
				return err
			}
			type view struct {
				models.Repository
				CanManage bool
			}
			out := []view{}
			for _, repo := range all {
				if models.CanRead(a.DB, u, &repo) {
					out = append(out, view{repo, models.Owner(a.DB, u, repo.OrgID)})
				}
			}
			return send(w, out)
		}
		if r.Method == "POST" {
			return a.migrate(w, r, u)
		}
		return &httpError{405, "不支持此操作"}
	}
	repo := new(models.Repository)
	ok, err := a.DB.ID(id(p[0])).Get(repo)
	if err != nil {
		return err
	}
	if !ok {
		return &httpError{404, "仓库不存在"}
	}
	if !models.CanRead(a.DB, u, repo) {
		return forbidden()
	}
	if len(p) == 3 && p[1] == "tasks" && p[2] == "history" {
		if r.Method != "GET" {
			return &httpError{405, "不支持此操作"}
		}
		history, err := a.repositoryTaskHistory(r, u, repo)
		if err != nil {
			return err
		}
		return send(w, history)
	}
	if len(p) == 1 && r.Method == "GET" {
		return send(w, repo)
	}
	if len(p) == 1 && r.Method == "DELETE" {
		if !models.Owner(a.DB, u, repo.OrgID) {
			return forbidden()
		}
		if !a.Worker.TryLock() {
			return errors.New("task running; retry later")
		}
		defer a.Worker.Unlock()
		var in struct {
			DeleteFiles bool
			Confirm     string
		}
		if err = decode(w, r, &in); err != nil {
			return err
		}
		if in.Confirm != repo.Name {
			return errors.New("repository name confirmation required")
		}
		active, e := a.DB.Where("repository_id=? AND state IN (?,?)", repo.ID, "queued", "running").Exist(new(models.Task))
		if e != nil {
			return e
		}
		if active {
			return errors.New("active task exists")
		}
		if in.DeleteFiles {
			if !tasks.Within(a.Config.Repositories, repo.Path) {
				return errors.New("unsafe repository path")
			}
			if err = tasks.RejectLinks(repo.Path); err != nil {
				return err
			}
			if err = os.RemoveAll(repo.Path); err != nil {
				return err
			}
		}
		tx := a.DB.NewSession()
		defer tx.Close()
		if err = tx.Begin(); err != nil {
			return err
		}
		for _, q := range []string{"DELETE FROM revision_event WHERE repository_id=?", "DELETE FROM activity_index WHERE repository_id=?", "DELETE FROM task WHERE repository_id=?", "DELETE FROM team_repository WHERE repository_id=?", "DELETE FROM repository WHERE id=?"} {
			if _, err = tx.Exec(q, repo.ID); err != nil {
				return err
			}
		}
		if err = tx.Commit(); err != nil {
			return err
		}
		return send(w, map[string]bool{"ok": true})
	}
	if len(p) != 2 {
		return &httpError{404, "接口不存在"}
	}
	action := p[1]
	if action == "tasks" && r.Method == "GET" {
		var rows []models.Task
		if err = a.DB.Where("repository_id=?", repo.ID).Desc("id").Limit(100).Find(&rows); err != nil {
			return err
		}
		return send(w, rows)
	}
	if action == "convert-mirror" && r.Method == "POST" {
		return a.convertMirror(w, r, u, repo)
	}
	if action == "sync" && r.Method == "POST" {
		if !models.Owner(a.DB, u, repo.OrgID) {
			return forbidden()
		}
		if repo.Kind != "mirror" || repo.Status != "ready" || repo.Halted {
			return &httpError{409, "当前仓库尚未就绪或同步已停止，请先检查任务记录及仓库状态"}
		}
		if err = models.Enqueue(a.DB, repo.ID, "sync", ""); err != nil {
			if errors.Is(err, models.ErrActiveTask) {
				return &httpError{409, "已有同步任务正在排队或执行中，请查看任务记录"}
			}
			return err
		}
		return send(w, map[string]bool{"queued": true})
	}
	if action == "retry" && r.Method == "POST" {
		if !models.Owner(a.DB, u, repo.OrgID) {
			return forbidden()
		}
		if repo.Halted {
			return errors.New("integrity failure requires manual review")
		}
		if !a.Worker.TryLock() {
			return errors.New("task running; retry later")
		}
		defer a.Worker.Unlock()
		active, e := a.DB.Where("repository_id=? AND state IN (?,?)", repo.ID, "queued", "running").Exist(new(models.Task))
		if e != nil {
			return e
		}
		if active {
			return errors.New("active task exists")
		}
		task := new(models.Task)
		ok, e = a.DB.Where("repository_id=? AND state=?", repo.ID, "failed").Desc("id").Get(task)
		if e != nil {
			return e
		}
		if !ok {
			return errors.New("no failed task")
		}
		if task.Kind == "local" && !u.Admin {
			return forbidden()
		}
		if task.Kind == "conversion" {
			if repo.Kind != "local" || repo.Status != "ready" {
				return &httpError{409, "当前仓库状态不支持重试转换"}
			}
			task.ActorID = u.ID
			task.ActorName = u.Name
		}
		task.State = "queued"
		task.Trigger = "manual"
		task.Error = ""
		task.Finished = 0
		if _, e = a.DB.ID(task.ID).Cols("state", "error", "finished", "trigger", "actor_id", "actor_name").Update(task); e != nil {
			return e
		}
		return send(w, map[string]bool{"queued": true})
	}
	if action == "settings" && r.Method == "POST" {
		return a.settings(w, r, u, repo)
	}
	if r.Method != "GET" {
		return &httpError{405, "不支持此操作"}
	}
	if repo.Status != "ready" {
		return errors.New("repository migration has not completed")
	}
	rev := repo.Revision
	if s := r.URL.Query().Get("revision"); s != "" {
		rev, err = strconv.ParseInt(s, 10, 64)
		if err != nil || rev < 0 || rev > repo.Revision {
			return errors.New("invalid revision")
		}
	}
	path := r.URL.Query().Get("path")
	switch action {
	case "tree":
		rows, e := a.SVN.List(r.Context(), repo.Path, path, rev)
		if e != nil {
			return e
		}
		return send(w, rows)
	case "log":
		page, e := a.repositoryHistory(r, repo, path, rev)
		if e != nil {
			return e
		}
		w.Header().Set("X-Revision", strconv.FormatInt(page.Revision, 10))
		w.Header().Set("X-Older-Cursor", strconv.FormatInt(page.OlderCursor, 10))
		w.Header().Set("X-Newer-Cursor", strconv.FormatInt(page.NewerCursor, 10))
		return send(w, page.Entries)
	case "file", "download":
		b, e := a.SVN.Cat(r.Context(), repo.Path, path, rev)
		if e != nil {
			return e
		}
		if action == "download" {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filepath.Base(path)}))
			_, err = w.Write(b)
			return err
		}
		return send(w, map[string]any{"Content": svn.DecodeText(b), "Size": len(b), "Binary": svn.IsBinary(b)})
	}
	return &httpError{404, "接口不存在"}
}
func (a *App) migrate(w http.ResponseWriter, r *http.Request, u *models.User) error {
	var in struct {
		OrgID                                  int64
		Name, Kind, Source, Username, Password string
		SkipTLS                                bool
		Interval                               int64
		SourceStopped                          bool
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if !models.Owner(a.DB, u, in.OrgID) {
		return forbidden()
	}
	org := new(models.Organization)
	orgExists, err := a.DB.ID(in.OrgID).Get(org)
	if err != nil {
		return err
	}
	if !orgExists {
		return &httpError{404, "所选组织不存在，请重新选择"}
	}
	if !validName(in.Name) {
		return &httpError{400, "仓库名称须以字母或数字开头，且只能包含字母、数字、点、下划线和连字符，最长 80 个字符"}
	}
	repoPath, err := storage.RepositoryPath(a.Config.Repositories, org.Name, in.Name)
	if err != nil {
		return &httpError{400, "组织或仓库名称不能用作 Windows 目录名：请避免设备保留名和末尾的点"}
	}
	conflict := func() error {
		return &httpError{409, fmt.Sprintf("组织“%s”中已存在仓库“%s”，请更换仓库名称。", org.Name, in.Name)}
	}
	duplicate, err := a.DB.Where("org_id=? AND lower(name)=?", org.ID, strings.ToLower(in.Name)).Exist(new(models.Repository))
	if err != nil {
		return err
	}
	if duplicate {
		return conflict()
	}
	switch in.Kind {
	case "local":
		if !u.Admin {
			return &httpError{403, "仅管理员可以迁移服务器本地 SVN 仓库"}
		}
		if !in.SourceStopped {
			return &httpError{400, "请先停止原服务访问，并勾选确认移动仓库"}
		}
		if !filepath.IsAbs(in.Source) || filepath.VolumeName(in.Source) == in.Source || strings.HasPrefix(in.Source, `\\`) {
			return &httpError{400, "本地来源须为服务器上的绝对目录路径，不支持磁盘根目录或网络共享路径"}
		}
		if tasks.Within(a.Config.Repositories, in.Source) || strings.EqualFold(filepath.Clean(in.Source), filepath.Clean(a.Config.Repositories)) {
			return &httpError{400, "来源目录已位于 SVNova 仓库根目录内，不能重复迁移"}
		}
		if err = tasks.RejectLinks(in.Source); err != nil {
			return &httpError{400, "来源目录不存在、无法访问或包含不允许的链接，请检查来源路径"}
		}
	case "mirror":
		if err = svn.ValidateRemote(in.Source); err != nil {
			return &httpError{400, "远端来源须为 HTTP/HTTPS 仓库地址，不能包含账号密码、查询参数或片段；请在下方填写认证信息"}
		}
		if in.Interval != 0 && in.Interval < 600 {
			return &httpError{400, "同步间隔须为 0（关闭自动同步）或至少 600 秒"}
		}
	default:
		return &httpError{400, "请选择远端 SVN 镜像或本地 SVN 仓库"}
	}
	repo := &models.Repository{OrgID: in.OrgID, Name: in.Name, Kind: in.Kind, Status: "pending", Path: repoPath, Interval: in.Interval, Username: in.Username, SkipTLS: in.SkipTLS}
	if in.Kind == "mirror" {
		repo.Remote = in.Source
		repo.Password = a.Vault.Seal(in.Password)
	}
	tx := a.DB.NewSession()
	defer tx.Close()
	if err = tx.Begin(); err != nil {
		return err
	}
	// Serialize duplicate checks with repository creation and authority changes.
	if _, err = tx.Exec("UPDATE organization SET name=name WHERE id=?", org.ID); err != nil {
		return err
	}
	if err = checkOrganizationOwner(tx, u, org.ID); err != nil {
		return err
	}
	duplicate, err = tx.Where("org_id=? AND lower(name)=?", org.ID, strings.ToLower(in.Name)).Exist(new(models.Repository))
	if err != nil {
		return err
	}
	if duplicate {
		return conflict()
	}
	if err = storage.CheckParents(repoPath); err != nil {
		return &httpError{409, "仓库目标目录包含链接或无法访问，请联系管理员检查仓库根目录"}
	}
	if _, e := os.Lstat(repoPath); e == nil {
		return &httpError{409, "目标仓库目录已存在，不能覆盖，请更换仓库名称或联系管理员处理保留目录"}
	} else if !os.IsNotExist(e) {
		return e
	}
	if _, err = tx.Insert(repo); err != nil {
		return err
	}
	if _, err = tx.Insert(&models.Task{RepositoryID: repo.ID, Trigger: "migration", Kind: in.Kind, Source: in.Source, State: "queued", Started: time.Now().Unix()}); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	w.Header().Set("Location", repoLink(org.Name, repo.Name, repo.ID))
	return send(w, repo)
}
func (a *App) settings(w http.ResponseWriter, r *http.Request, u *models.User, repo *models.Repository) error {
	if !models.Owner(a.DB, u, repo.OrgID) {
		return forbidden()
	}
	if repo.Kind != "mirror" {
		return errors.New("not a mirror")
	}
	if !a.Worker.TryLock() {
		return &httpError{409, "当前有仓库任务或维护操作正在执行，暂时无法保存镜像设置。即使是其他仓库的操作，也需等待结束后重试。"}
	}
	defer a.Worker.Unlock()
	active, err := a.DB.Where("repository_id=? AND state IN (?,?)", repo.ID, "queued", "running").Exist(new(models.Task))
	if err != nil {
		return err
	}
	if active {
		return &httpError{409, "此仓库已有排队或执行中的任务，暂时无法保存镜像设置。请先查看任务记录，等待任务结束后重试。"}
	}
	var in struct {
		Remote, Username, Password string
		ReplacePassword, SkipTLS   bool
		Interval                   int64
	}
	if err = decode(w, r, &in); err != nil {
		return err
	}
	if err = svn.ValidateRemote(in.Remote); err != nil {
		return err
	}
	if in.Interval < 0 || (in.Interval > 0 && in.Interval < 600) {
		return errors.New("invalid mirror interval")
	}
	candidate := *repo
	candidate.Remote = in.Remote
	candidate.Username = in.Username
	candidate.SkipTLS = in.SkipTLS
	if in.ReplacePassword {
		candidate.Password = a.Vault.Seal(in.Password)
	}
	// Credential rotation must remain possible while the remote is unavailable.
	// Changing the repository address still requires an online identity check.
	if candidate.Remote != repo.Remote {
		c, err := a.Worker.Credentials(&candidate)
		if err != nil {
			return err
		}
		uuid, root, head, err := a.SVN.Info(r.Context(), candidate.Remote, c)
		if err != nil {
			return err
		}
		if (repo.UUID != "" && uuid != repo.UUID) || head < repo.Revision || strings.TrimRight(root, "/") != strings.TrimRight(candidate.Remote, "/") {
			return errors.New("remote repository does not match")
		}
		if _, e := os.Stat(repo.Path); e == nil {
			if err = a.SVN.SetRemote(r.Context(), repo.Path, candidate.Remote); err != nil {
				return err
			}
		}
	}
	candidate.Interval = in.Interval
	candidate.NextSync = time.Now().Unix() + in.Interval
	if _, err = a.DB.ID(repo.ID).Cols("remote", "username", "password", "skip_tls", "interval", "next_sync").Update(&candidate); err != nil {
		return err
	}
	return send(w, &candidate)
}
func (a *App) admin(w http.ResponseWriter, r *http.Request, u *models.User, p []string) error {
	if len(p) == 0 {
		return &httpError{404, "接口不存在"}
	}
	if p[0] == "notifications" {
		return a.notificationAPI(w, r, u, p[1:])
	}
	if p[0] == "ldap-schedule" && len(p) == 1 {
		return a.ldapSchedule(w, r, u)
	}
	if p[0] == "site" && len(p) == 1 {
		return a.siteSettings(w, r, u)
	}
	if p[0] == "users" {
		if len(p) == 3 && p[2] == "delete" {
			return a.deleteLocalUser(w, r, u, id(p[1]))
		}
		if len(p) == 1 && r.Method == "GET" {
			var users []models.User
			if err := a.DB.Find(&users); err != nil {
				return err
			}
			return send(w, users)
		}
		var in struct {
			Name, Password string
			Admin, Active  bool
		}
		if err := decode(w, r, &in); err != nil {
			return err
		}
		if len(p) == 1 && r.Method == "POST" {
			if !validName(in.Name) || len(in.Password) < 12 {
				return errors.New("invalid user or password")
			}
			h, salt, err := security.Password(in.Password)
			if err != nil {
				return err
			}
			user := &models.User{Name: in.Name, Password: h, Salt: salt, Admin: in.Admin, Active: true}
			if _, err = a.DB.Insert(user); err != nil {
				return err
			}
			return send(w, user)
		}
		if len(p) == 2 && r.Method == "POST" {
			target := new(models.User)
			ok, err := a.DB.ID(id(p[1])).Get(target)
			if err != nil || !ok {
				return errors.New("user not found")
			}
			if target.ID == u.ID && (!in.Admin || !in.Active) {
				return errors.New("cannot disable or demote current administrator")
			}
			target.Admin = in.Admin
			target.Active = in.Active
			target.LDAPDisabled = false
			if in.Password != "" {
				if target.LDAPAccount || len(in.Password) < 12 {
					return errors.New("invalid password reset")
				}
				target.Password, target.Salt, err = security.Password(in.Password)
				if err != nil {
					return err
				}
			}
			tx := a.DB.NewSession()
			defer tx.Close()
			if err = tx.Begin(); err != nil {
				return err
			}
			if _, err = tx.ID(target.ID).Cols("admin", "active", "password", "salt", "ldap_disabled").Update(target); err != nil {
				return err
			}
			if _, err = tx.Where("user_id=?", target.ID).Delete(new(models.Session)); err != nil {
				return err
			}
			if err = tx.Commit(); err != nil {
				return err
			}
			return send(w, target)
		}
	}
	if p[0] == "ldap" {
		return a.ldapConfig(w, r, u, p[1:])
	}
	return &httpError{405, fmt.Sprintf("不支持此操作：%s", p[0])}
}
