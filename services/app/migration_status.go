// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"svnova/models"
	"time"
)

type migrationTask struct {
	Indeterminate bool   `json:"indeterminate"`
	ID            int64  `json:"id"`
	Kind          string `json:"kind"`
	State         string `json:"state"`
	Phase         string `json:"phase"`
	Started       string `json:"started"`
	Finished      string `json:"finished"`
	Message       string `json:"message"`
	Retry         bool   `json:"retry"`
}
type migrationStatus struct {
	Progress        int             `json:"progress"`
	Status          int             `json:"status"`
	Message         string          `json:"message"`
	RepositoryState string          `json:"repositoryState"`
	CanRetry        bool            `json:"canRetry"`
	Tasks           []migrationTask `json:"tasks"`
}

func taskPhase(t models.Task) string {
	if t.State == "succeeded" {
		return "已完成"
	}
	if t.State == "failed" {
		return "失败阶段：" + mapTaskPhase(t.Phase)
	}
	phase := map[string]string{"checking": "正在检查仓库", "verify-source": "正在校验源仓库", "verify-target": "正在校验目标副本", "checking-identity": "正在核对仓库身份", "cleaning": "正在清理源目录", "moving": "正在移动仓库", "copying": "正在复制仓库", "published": "正在清理源目录", "synchronizing": "正在同步修订", "revision-properties": "正在同步修订属性"}[t.Phase]
	if t.State == "queued" {
		return "正在排队，等待迁移任务执行"
	}
	if t.Kind == "conversion" {
		phase = mapTaskPhase(t.Phase)
	}
	if phase == "" {
		phase = "正在准备 SVN 仓库"
	}
	if (t.Phase == "verify-source" || t.Phase == "verify-target") && t.Verification == "metadata" {
		phase += "元数据"
		if t.VerificationPosition >= 0 {
			phase += fmt.Sprintf(" · 当前检查位置 r%d", t.VerificationPosition)
		} else {
			phase += " · 正在检查仓库级元数据"
		}
		if t.ProgressStarted > 0 {
			phase += fmt.Sprintf(" · 本阶段已用时 %d秒", max(int64(0), time.Now().Unix()-t.ProgressStarted))
		}
		if t.ProgressUpdated > 0 {
			phase += fmt.Sprintf(" · 距上次进度更新 %d秒", max(int64(0), time.Now().Unix()-t.ProgressUpdated))
		}
	} else if t.Phase == "verify-source" || t.Phase == "verify-target" {
		completed := max(int64(0), t.CurrentRevision+1)
		total := t.TotalRevision + 1
		phase += fmt.Sprintf(" · 已校验修订 %d / %d（%.1f%%）", completed, total, float64(completed)*100/float64(max(int64(1), total)))
	} else if t.Phase == "revision-properties" {
		if t.Verification == "revprops" && t.VerificationPosition >= 0 {
			completed := min(t.VerificationPosition, t.TotalRevision) + 1
			total := max(int64(1), t.TotalRevision+1)
			phase += fmt.Sprintf(" · 已处理修订属性 %d / %d（%.1f%%） · 当前 r%d", completed, total, float64(completed)*100/float64(total), t.VerificationPosition)
		} else {
			phase += " · 等待修订属性进度"
		}
		if t.ProgressStarted > 0 {
			phase += fmt.Sprintf(" · 本阶段已用时 %d秒", max(int64(0), time.Now().Unix()-t.ProgressStarted))
		}
		if t.ProgressUpdated > 0 {
			phase += fmt.Sprintf(" · 距上次进度更新 %d秒", max(int64(0), time.Now().Unix()-t.ProgressUpdated))
		}
	} else if t.TotalRevision > 0 && t.Kind != "local" {
		phase += " · r" + strconv.FormatInt(t.CurrentRevision, 10) + " / " + strconv.FormatInt(t.TotalRevision, 10)
	}
	if t.Kind == "conversion" && t.State == "running" && t.ProgressStarted > 0 {
		phase += fmt.Sprintf(" · 本阶段已用时 %d秒", max(int64(0), time.Now().Unix()-t.ProgressStarted))
		if t.ProgressUpdated > 0 {
			phase += fmt.Sprintf(" · 距上次进度更新 %d秒", max(int64(0), time.Now().Unix()-t.ProgressUpdated))
		}
	}
	if t.State == "running" && t.Started > 0 {
		label := "已用时"
		if t.Phase == "revision-properties" {
			label = "累计已用时"
		}
		phase += fmt.Sprintf(" · %s %d秒", label, max(int64(0), time.Now().Unix()-t.Started))
	}
	return phase
}
func mapTaskPhase(phase string) string {
	value := map[string]string{"conversion-prepare": "准备镜像配置", "conversion-local": "检查本地仓库", "conversion-remote": "检查远端仓库", "conversion-initialize": "初始化镜像", "conversion-final": "核验镜像配置", "conversion-recovery": "恢复原配置", "checking": "正在检查仓库", "verify-source": "正在校验源仓库", "verify-target": "正在校验目标副本", "checking-identity": "正在核对仓库身份", "cleaning": "正在清理源目录", "moving": "移动仓库", "copying": "复制和校验仓库", "published": "发布仓库", "synchronizing": "同步修订", "revision-properties": "同步修订属性", "completed": "完成"}[phase]
	if value == "" {
		return "准备仓库"
	}
	return value
}
func taskResult(task models.Task, execution ...*models.TaskExecution) string {
	var snapshot *models.TaskExecution
	if len(execution) > 0 {
		snapshot = execution[0]
	}
	message := executionSummary(task, snapshot)

	return message
}
func repositoryState(repo *models.Repository, manage bool) string {
	return repo.Status + ":" + boolText(repo.Halted) + ":" + boolText(manage)
}
func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
func (a *App) migrationSnapshot(u *models.User, repo *models.Repository, selected ...int64) (migrationStatus, error) {
	out := migrationStatus{Progress: -1, Status: 1, Message: "正在准备迁移 SVN 仓库", Tasks: []migrationTask{}}
	manage := models.Owner(a.DB, u, repo.OrgID)
	out.RepositoryState = repositoryState(repo, manage)
	var tasks []models.Task
	query := a.DB.Where("repository_id=?", repo.ID).Desc("id").Limit(100)
	if len(selected) > 0 {
		query.And("id=?", selected[0])
	}
	if err := query.Find(&tasks); err != nil {
		return out, err
	}
	if len(selected) > 0 && len(tasks) == 0 {
		return out, &httpError{404, "任务不存在"}
	}
	ids := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	executions, err := a.latestTaskExecutions(ids)
	if err != nil {
		return out, err
	}
	p := &webPage{}
	for _, task := range tasks {
		out.Tasks = append(out.Tasks, repositoryTaskRow(task, executions[task.ID], u, repo, manage, len(selected) == 0))
	}
	if repo.Status == "ready" {
		out.Status = 4
		out.Message = "迁移完成"
	}
	if len(tasks) > 0 {
		latest := tasks[0]
		switch latest.State {
		case "queued", "running":
			out.Status = 2
			out.Message = taskPhase(latest)
			out.Progress = taskProgress(latest)
		case "failed":
			out.Status = 3
			out.Message = p.TaskError(latest.Error)
			out.CanRetry = out.Tasks[0].Retry
		}
	} else if repo.Status == "failed" {
		out.Status = 3
		out.Message = "迁移失败，请检查仓库设置"
	}
	return out, nil
}

// Resolve Gitea-shaped status/retry routes while retaining SVN authorization.
func (a *App) migrationAction(w http.ResponseWriter, r *http.Request, u *models.User) bool {
	suffix := "/-/migrate/status"
	retry := false
	if strings.HasSuffix(r.URL.Path, "/settings/migrate/retry") {
		suffix = "/settings/migrate/retry"
		retry = true
	} else if !strings.HasSuffix(r.URL.Path, suffix) {
		return false
	}
	if (!retry && r.Method != "GET") || (retry && r.Method != "POST") {
		http.Error(w, "不支持此操作", 405)
		return true
	}
	clone := r.Clone(r.Context())
	target := *r.URL
	clone.URL = &target
	target.Path = strings.TrimSuffix(target.Path, suffix)
	resolved, err := a.resolvePage(clone)
	if err != nil || !strings.HasPrefix(resolved.URL.Path, "/repositories/") {
		http.NotFound(w, r)
		return true
	}
	raw := strings.TrimPrefix(resolved.URL.Path, "/repositories/")
	if strings.Contains(raw, "/") || id(raw) == 0 {
		http.NotFound(w, r)
		return true
	}
	repo := new(models.Repository)
	ok, err := a.DB.ID(id(raw)).Get(repo)
	if err != nil {
		http.Error(w, "状态暂不可用", 500)
		return true
	}
	if !ok {
		http.NotFound(w, r)
		return true
	}
	if !models.CanRead(a.DB, u, repo) {
		http.Error(w, "没有访问权限", 403)
		return true
	}
	if retry {
		if err = a.repositories(w, r, u, []string{raw, "retry"}); err != nil {
			var failure *httpError
			if errors.As(err, &failure) {
				http.Error(w, failure.Message, failure.Status)
			} else {
				http.Error(w, "无法重试，请检查权限、仓库状态及现有任务", 400)
			}
		}
		return true
	}
	snapshot, err := a.migrationSnapshot(u, repo)
	if err != nil {
		http.Error(w, "状态暂不可用", 500)
		return true
	}
	_ = send(w, snapshot)
	return true
}

func conversionActor(t models.Task) string {
	if t.Kind == "conversion" {
		return " · 操作者：" + t.ActorName
	}
	return ""
}

func taskProgress(task models.Task) int {
	if task.State != "running" {
		return -1
	}
	if task.Phase == "revision-properties" && task.Verification == "revprops" && task.VerificationPosition >= 0 {
		return int((min(task.VerificationPosition, task.TotalRevision) + 1) * 100 / max(int64(1), task.TotalRevision+1))
	}
	if task.Verification != "metadata" && (task.Phase == "verify-source" || task.Phase == "verify-target") {
		return int(min(int64(100), max(int64(0), task.CurrentRevision+1)*100/max(int64(1), task.TotalRevision+1)))
	}
	return -1
}
