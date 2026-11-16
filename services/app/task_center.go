// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"svnova/models"
	"time"
)

type taskCenterOption struct {
	ID   int64
	Name string
}
type taskCenterRow struct {
	OrgName, RepoName, OrgURL, RepoURL                                             string
	Progress                                                                       int
	ID                                                                             int64
	Repository, Kind, State, Phase, Trigger, Actor, Started, Duration, Result, URL string
	Active                                                                         bool
}
type taskCenterPage struct {
	Items                        []taskCenterRow
	Organizations, Repositories  []taskCenterOption
	Org, Repository, Kind, State string
	Keyword                      string
	Pagination                   adminUsersPage
	Active                       bool
}

// Keep visibility in SQL so counts and page boundaries share the same authorization.
func taskVisibility(u *models.User) (string, []any) {
	if u == nil || !u.Active {
		return "0=1", nil
	}
	if u.Admin {
		return "1=1", nil
	}
	return "EXISTS (SELECT 1 FROM membership m WHERE m.org_id=r.org_id AND m.user_id=?) AND EXISTS (SELECT 1 FROM team t JOIN team_member tm ON tm.team_id=t.id WHERE t.org_id=r.org_id AND tm.user_id=? AND (t.is_owner_team=1 OR EXISTS (SELECT 1 FROM team_repository tr WHERE tr.team_id=t.id AND tr.repository_id=r.id)))", []any{u.ID, u.ID}
}
func (p taskCenterPage) Link(page int) string {
	q := url.Values{"q": {p.Keyword}, "org": {p.Org}, "repository": {p.Repository}, "kind": {p.Kind}, "state": {p.State}, "page": {strconv.Itoa(page)}}
	return "/tasks?" + q.Encode()
}
func (a *App) taskCenterData(r *http.Request, u *models.User) (taskCenterPage, error) {
	q := r.URL.Query()
	out := taskCenterPage{Keyword: strings.TrimSpace(q.Get("q")), Items: []taskCenterRow{}, Organizations: []taskCenterOption{}, Repositories: []taskCenterOption{}, Org: q.Get("org"), Repository: q.Get("repository"), Kind: q.Get("kind"), State: q.Get("state")}
	for _, v := range []string{out.Org, out.Repository} {
		if v != "" {
			n, e := strconv.ParseInt(v, 10, 64)
			if e != nil || n < 1 {
				return out, &httpError{400, "筛选参数无效"}
			}
		}
	}
	if out.Kind != "" && out.Kind != "local" && out.Kind != "mirror" && out.Kind != "sync" && out.Kind != "conversion" {
		return out, &httpError{400, "任务类型无效"}
	}
	if out.State != "" && out.State != "queued" && out.State != "running" && out.State != "succeeded" && out.State != "failed" {
		return out, &httpError{400, "任务状态无效"}
	}
	visible, args := taskVisibility(u)
	var repos []struct {
		ID, OrgID     int64
		Name, OrgName string
	}
	err := a.DB.SQL("SELECT r.id,r.org_id,r.name,o.name AS org_name FROM repository r JOIN organization o ON o.id=r.org_id WHERE "+visible+" ORDER BY o.name,r.name,r.id", args...).Find(&repos)
	if err != nil {
		return out, err
	}
	seen := map[int64]bool{}
	for _, repo := range repos {
		if !seen[repo.OrgID] {
			out.Organizations = append(out.Organizations, taskCenterOption{repo.OrgID, repo.OrgName})
			seen[repo.OrgID] = true
		}
		if out.Org == "" || out.Org == strconv.FormatInt(repo.OrgID, 10) {
			out.Repositories = append(out.Repositories, taskCenterOption{repo.ID, repo.OrgName + "/" + repo.Name})
		}
	}
	where := visible
	for _, f := range []struct{ Column, Value string }{{"r.org_id", out.Org}, {"r.id", out.Repository}, {"t.kind", out.Kind}, {"t.state", out.State}} {
		if f.Value != "" {
			where += " AND " + f.Column + "=?"
			args = append(args, f.Value)
		}
	}
	if out.Keyword != "" {
		where += " AND (instr(lower(o.name), ?) > 0 OR instr(lower(r.name), ?) > 0 OR t.id=?)"
		taskID, parseErr := strconv.ParseInt(strings.TrimPrefix(out.Keyword, "#"), 10, 64)
		if parseErr != nil || taskID < 1 {
			taskID = 0
		}
		keyword := strings.ToLower(out.Keyword)
		args = append(args, keyword, keyword, taskID)
	}
	from := " FROM task t JOIN repository r ON r.id=t.repository_id JOIN organization o ON o.id=r.org_id WHERE " + where
	var total int64
	if _, err = a.DB.SQL("SELECT COUNT(*)"+from, args...).Get(&total); err != nil {
		return out, err
	}
	out.Pagination = adminUsersPage{Total: total, Pages: max(1, int((total+19)/20))}
	out.Pagination.Page = min(pageNumber(q.Get("page")), out.Pagination.Pages)
	out.Pagination.paginateWith(out.Link)
	var rows []struct {
		models.Task       `xorm:"extends"`
		RepoName, OrgName string
	}
	pageArgs := append(append([]any{}, args...), 20, (out.Pagination.Page-1)*20)
	if err = a.DB.SQL("SELECT t.*,r.name AS repo_name,o.name AS org_name"+from+" ORDER BY t.id DESC LIMIT ? OFFSET ?", pageArgs...).Find(&rows); err != nil {
		return out, err
	}
	ids := make([]int64, 0, len(rows))
	for _, v := range rows {
		ids = append(ids, v.ID)
	}
	executions, err := a.latestTaskExecutions(ids)
	if err != nil {
		return out, err
	}
	p := &webPage{}
	for _, v := range rows {
		active := v.State == "queued" || v.State == "running"
		actor := v.ActorName
		if actor == "" {
			actor = "未记录"
		}
		started, duration := "—", "—"
		if v.State != "queued" && v.Started > 0 {
			started = p.Time(v.Started)
			end := v.Finished
			if active {
				end = time.Now().Unix()
			}
			if end >= v.Started {
				duration = fmt.Sprintf("%d 秒", end-v.Started)
			}
		}
		trigger := map[string]string{"manual": "手动", "scheduled": "定时", "migration": "首次迁移"}[v.Trigger]
		if trigger == "" {
			trigger = "未记录"
		}
		target := repoLink(v.OrgName, v.RepoName, v.RepositoryID) + "?tab=tasks&task_id=" + strconv.FormatInt(v.ID, 10)
		out.Items = append(out.Items, taskCenterRow{OrgName: v.OrgName, RepoName: v.RepoName, OrgURL: orgLink(v.OrgName), RepoURL: repoLink(v.OrgName, v.RepoName, v.RepositoryID), Progress: taskProgress(v.Task), ID: v.ID, Repository: v.OrgName + "/" + v.RepoName, Kind: map[string]string{"local": "本地迁移", "mirror": "镜像迁移", "sync": "镜像同步", "conversion": "转换镜像"}[v.Kind], State: p.State(v.State), Phase: taskPhase(v.Task), Trigger: trigger, Actor: actor, Started: started, Duration: duration, Result: executionSummary(v.Task, executions[v.ID]), URL: target, Active: active})
		out.Active = out.Active || active
	}
	return out, nil
}
func (a *App) taskCenterAPI(w http.ResponseWriter, r *http.Request, u *models.User) error {
	data, err := a.taskCenterData(r, u)
	if err != nil {
		return err
	}
	var html bytes.Buffer
	if err = a.Templates.ExecuteTemplate(&html, "tasks/results", data); err != nil {
		return err
	}
	w.Header().Set("Cache-Control", "no-store")
	return send(w, map[string]any{"items": data.Items, "total": data.Pagination.Total, "page": data.Pagination.Page, "pages": data.Pagination.Pages, "active": data.Active, "html": html.String()})
}
