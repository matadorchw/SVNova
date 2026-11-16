// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"svnova/models"
	"svnova/services/activity"
	"time"
)

type heatmapResponse struct {
	HeatmapData        [][2]int64 `json:"heatmapData"`
	TotalContributions int64      `json:"totalContributions"`
	Pending            bool       `json:"pending"`
	Failed             bool       `json:"failed"`
	Today              string     `json:"today"`
}
type dashboardEvent struct {
	BeforeRevision, AfterRevision int64
	RevisionKnown                 bool
	ExecutionID                   int64
	TaskID                        int64
	RepositoryID                  int64
	OrganizationID                int64
	Kind, Trigger, State, Error   string
	Timestamp, Finished           int64
	Organization, Repository      string
}

func (e dashboardEvent) RepoURL() string {
	return repoLink(e.Organization, e.Repository, e.RepositoryID)
}
func (e dashboardEvent) TaskURL() string { return e.RepoURL() + "?tab=tasks" }
func (e dashboardEvent) KindText() string {
	return map[string]string{"local": "本地仓库迁移", "mirror": "镜像仓库迁移", "sync": "镜像仓库同步", "conversion": "本地仓库转镜像"}[e.Kind]
}
func (e dashboardEvent) TriggerText() string {
	if value := map[string]string{"migration": "首次迁移", "manual": "手动", "scheduled": "定时"}[e.Trigger]; value != "" {
		return value
	}
	return "历史记录（触发方式未保存）"
}
func (e dashboardEvent) StateText() string {
	return map[string]string{"queued": "排队中", "running": "进行中", "succeeded": "成功", "failed": "失败", "interrupted": "已中断"}[e.State]
}
func (e dashboardEvent) StateIcon() string {
	switch e.State {
	case "succeeded":
		return "check"
	case "failed":
		return "x"
	case "running":
		return "hourglass"
	case "queued":
		return "clock"
	default:
		return "alert"
	}
}
func (e dashboardEvent) Icon() string {
	if e.Kind == "conversion" {
		return "arrow-switch"
	}
	if e.Kind == "sync" {
		return "sync"
	}
	if e.Kind == "mirror" {
		return "mirror"
	}
	return "repo"
}
func (e dashboardEvent) Result() string {
	return executionSummary(models.Task{Kind: e.Kind, State: e.State, Error: e.Error}, &models.TaskExecution{State: e.State, BeforeRevision: e.BeforeRevision, AfterRevision: e.AfterRevision, RevisionKnown: e.RevisionKnown})
}

const repositoryActivitySQL = `SELECT e.id AS execution_id,t.id AS task_id,t.repository_id,t.kind,e.trigger,e.state,e.started/1000 AS timestamp,e.finished/1000 AS finished,e.error,e.before_revision,e.after_revision,e.revision_known FROM task_execution e JOIN task t ON t.id=e.task_id UNION ALL SELECT 0 AS execution_id,t.id AS task_id,t.repository_id,t.kind,t.trigger,t.state,t.started AS timestamp,t.finished,t.error,0 AS before_revision,0 AS after_revision,0 AS revision_known FROM task t WHERE NOT EXISTS (SELECT 1 FROM task_execution e WHERE e.task_id=t.id)`

func activityScope(ids []int64, now time.Time) (string, []any) {
	args := make([]any, 0, len(ids)+2)
	for _, id := range ids {
		args = append(args, id)
	}
	args = append(args, activity.Since(now).Unix(), now.Unix())
	return "repository_id IN (" + strings.TrimRight(strings.Repeat("?,", len(ids)), ",") + ") AND timestamp>=? AND timestamp<=?", args
}
func heatmapURL(org int64) string {
	return "/api/v1/activity/heatmap?org=" + strconv.FormatInt(org, 10)
}
func (a *App) summarizeActivity(repos []repositoryPage, now time.Time) (heatmapResponse, error) {
	out := heatmapResponse{HeatmapData: [][2]int64{}, Today: activity.Today(now).Format("2006-01-02")}
	ids := []int64{}
	for _, repo := range repos {
		ids = append(ids, repo.ID)
	}
	if len(ids) == 0 {
		return out, nil
	}
	where, args := activityScope(ids, now)
	var rows []struct {
		Day   string
		Count int64
	}
	err := a.DB.SQL("SELECT strftime('%Y-%m-%d',timestamp,'unixepoch','+8 hours') AS day,COUNT(*) AS count FROM ("+repositoryActivitySQL+") events WHERE "+where+" GROUP BY day", args...).Find(&rows)
	if err != nil {
		return out, err
	}
	for _, row := range rows {
		day, err := time.ParseInLocation("2006-01-02", row.Day, activity.Shanghai)
		if err != nil {
			return out, err
		}
		out.HeatmapData = append(out.HeatmapData, [2]int64{day.Unix(), row.Count})
		out.TotalContributions += row.Count
	}
	var pending int64
	_, err = a.DB.SQL("SELECT COUNT(*) FROM ("+repositoryActivitySQL+") events WHERE "+where+" AND state IN ('queued','running')", args...).Get(&pending)
	out.Pending = pending > 0
	return out, err
}
func (a *App) activityHeatmap(w http.ResponseWriter, r *http.Request, u *models.User) error {
	orgID := id(r.URL.Query().Get("org"))
	if orgID != 0 && !u.Admin {
		ok, err := a.DB.Where("org_id=? AND user_id=?", orgID, u.ID).Exist(new(models.Membership))
		if err != nil {
			return err
		}
		if !ok {
			return forbidden()
		}
	}
	var all []models.Repository
	if err := a.DB.Find(&all); err != nil {
		return err
	}
	var repos []repositoryPage
	for _, repo := range all {
		if (orgID == 0 || orgID == repo.OrgID) && models.CanRead(a.DB, u, &repo) {
			repos = append(repos, repositoryPage{Repository: repo})
		}
	}
	summary, err := a.summarizeActivity(repos, time.Now())
	if err != nil {
		return err
	}
	w.Header().Set("Cache-Control", "no-store")
	return send(w, summary)
}
func pageNumber(value string) int {
	n, _ := strconv.Atoi(value)
	if n < 1 || n > 10000 {
		return 1
	}
	return n
}
func (p *webPage) DashboardURL(values ...string) string {
	target := "/"
	if p.ContextOrg != 0 {
		for _, org := range p.Organizations {
			if org.ID == p.ContextOrg {
				target = orgSection(org.Name, "dashboard")
			}
		}
	}
	q := url.Values{}
	if p.FeedPage > 1 {
		q.Set("page", strconv.Itoa(p.FeedPage))
	}
	if p.RepoPage > 1 {
		q.Set("repo_page", strconv.Itoa(p.RepoPage))
	}
	q.Set("view", p.DashboardView)
	if p.Search != "" {
		q.Set("q", p.Search)
	}
	if p.KindFilter != "" {
		q.Set("kind", p.KindFilter)
	}
	if p.DateFilter != "" {
		q.Set("date", p.DateFilter)
	}
	for n := 0; n+1 < len(values); n += 2 {
		if values[n+1] == "" {
			q.Del(values[n])
		} else {
			q.Set(values[n], values[n+1])
		}
	}
	return target + "?" + q.Encode()
}
func (p *webPage) NextFeedURL() string { return p.DashboardURL("page", strconv.Itoa(p.FeedPage+1)) }
func (p *webPage) PreviousFeedURL() string {
	return p.DashboardURL("page", strconv.Itoa(max(1, p.FeedPage-1)))
}
func (p *webPage) NextRepoURL() string {
	return p.DashboardURL("repo_page", strconv.Itoa(p.RepoPage+1))
}
func (p *webPage) PreviousRepoURL() string {
	return p.DashboardURL("repo_page", strconv.Itoa(max(1, p.RepoPage-1)))
}
func (a *App) dashboardData(r *http.Request, p *webPage) error {
	now := time.Now()
	summary, err := a.summarizeActivity(p.Repositories, now)
	if err != nil {
		return err
	}
	p.Heatmap = summary
	p.HeatmapURL = heatmapURL(p.ContextOrg)
	byID := map[int64]repositoryPage{}
	ids := []int64{}
	for _, repo := range p.Repositories {
		byID[repo.ID] = repo
		ids = append(ids, repo.ID)
	}
	p.FeedPage = pageNumber(r.URL.Query().Get("page"))
	p.RepoPage = pageNumber(r.URL.Query().Get("repo_page"))
	p.DateFilter = r.URL.Query().Get("date")
	if len(ids) > 0 {
		where, args := activityScope(ids, now)
		if p.DateFilter != "" {
			day, err := time.ParseInLocation("2006-01-02", p.DateFilter, activity.Shanghai)
			if err != nil {
				return &httpError{400, "日期格式无效"}
			}
			where += " AND timestamp>=? AND timestamp<?"
			args = append(args, day.Unix(), day.AddDate(0, 0, 1).Unix())
		}
		var events []dashboardEvent
		args = append(args, dashboardFeedPageSize+1, (p.FeedPage-1)*dashboardFeedPageSize)
		if err := a.DB.SQL("SELECT * FROM ("+repositoryActivitySQL+") events WHERE "+where+" ORDER BY timestamp DESC,task_id DESC,execution_id DESC LIMIT ? OFFSET ?", args...).Find(&events); err != nil {
			return err
		}
		p.FeedHasNext = len(events) > dashboardFeedPageSize
		if p.FeedHasNext {
			events = events[:dashboardFeedPageSize]
		}
		for _, event := range events {
			repo := byID[event.RepositoryID]
			event.OrganizationID = repo.OrgID
			event.Organization = repo.Organization
			event.Repository = repo.Name
			p.Feed = append(p.Feed, event)
		}
	}
	p.DashboardView = "repositories"
	if r.URL.Query().Get("view") == "organizations" {
		p.DashboardView = "organizations"
	}
	p.KindFilter = r.URL.Query().Get("kind")
	if p.KindFilter != "local" && p.KindFilter != "mirror" {
		p.KindFilter = "all"
	}
	p.RepoTotal = len(p.Repositories)
	for _, repo := range p.Repositories {
		switch repo.Kind {
		case "local":
			p.LocalRepoTotal++
		case "mirror":
			p.MirrorRepoTotal++
		}
	}
	filtered := p.Repositories[:0]
	for _, repo := range p.Repositories {
		if (p.KindFilter == "all" || p.KindFilter == repo.Kind) && (p.Search == "" || strings.Contains(strings.ToLower(repo.Organization+"/"+repo.Name), strings.ToLower(p.Search))) {
			filtered = append(filtered, repo)
		}
	}
	p.RepoMatchTotal = len(filtered)
	p.Repositories, p.RepositoryPagination = paginateRepositoryRows(filtered, p.RepoPage, func(page int) string { return p.DashboardURL("repo_page", strconv.Itoa(page)) })
	p.RepoPage = p.RepositoryPagination.Page
	p.RepoHasNext = p.RepositoryPagination.Next != ""

	for _, org := range p.Organizations {
		if p.Search == "" || strings.Contains(strings.ToLower(org.Name+" "+org.FullName), strings.ToLower(p.Search)) {
			p.SidebarOrgs = append(p.SidebarOrgs, org)
		}
	}
	return nil
}
