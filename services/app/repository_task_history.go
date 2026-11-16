// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"svnova/models"
)

type repositoryTaskPoint struct {
	ID       int64  `json:"id"`
	Kind     string `json:"kind"`
	State    string `json:"state"`
	Started  int64  `json:"started"`
	Finished int64  `json:"finished"`
	Duration *int64 `json:"duration"`
	Before   *int64 `json:"before"`
	After    *int64 `json:"after"`
	Added    int64  `json:"added"`
}
type repositoryTaskSummary struct {
	Total        int      `json:"total"`
	SuccessRate  *float64 `json:"successRate"`
	RevisionSpan string   `json:"revisionSpan"`
	Added        int64    `json:"added"`
}
type repositoryTaskHistory struct {
	Summary         repositoryTaskSummary `json:"summary"`
	Points          []repositoryTaskPoint `json:"points"`
	Items           []migrationTask       `json:"items"`
	Page            int                   `json:"page"`
	Pages           int                   `json:"pages"`
	Active          bool                  `json:"active"`
	RepositoryState string                `json:"repositoryState"`
}

func finishedTaskTime(task models.Task) string {
	if task.Finished <= 0 || task.State == "queued" || task.State == "running" {
		return ""
	}
	return (&webPage{}).Time(task.Finished)
}

func (a *App) repositoryTaskHistory(r *http.Request, user *models.User, repo *models.Repository) (repositoryTaskHistory, error) {
	out := repositoryTaskHistory{Points: []repositoryTaskPoint{}, Items: []migrationTask{}}
	var tasks []models.Task
	if err := a.DB.Cols("id", "kind", "state", "phase", "started", "finished", "error", "actor_name", "verification", "verification_position", "progress_started", "progress_updated", "current_revision", "total_revision").Where("repository_id=?", repo.ID).Asc("id").Find(&tasks); err != nil {
		return out, err
	}
	ids := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	// Chunk IDs to keep the query below SQLite's bound parameter limit.
	executions := map[int64]*models.TaskExecution{}
	for offset := 0; offset < len(ids); offset += 500 {
		batch, err := a.latestTaskExecutions(ids[offset:min(offset+500, len(ids))])
		if err != nil {
			return out, err
		}
		for id, execution := range batch {
			executions[id] = execution
		}
	}
	ended, succeeded := 0, 0
	for _, task := range tasks {
		point := repositoryTaskPoint{ID: task.ID, Kind: task.Kind, State: task.State, Started: task.Started * 1000}
		active := task.State == "queued" || task.State == "running"
		out.Active = out.Active || active
		if !active && task.Finished > 0 {
			point.Finished = task.Finished * 1000
		}
		execution := executions[task.ID]
		if execution != nil && execution.State == task.State {
			if execution.Started > 0 {
				point.Started = execution.Started
			}
			if !active && execution.Finished > 0 {
				point.Finished = execution.Finished
			}
			if task.State == "succeeded" && execution.RevisionKnown && execution.BeforeRevision >= 0 && execution.AfterRevision >= execution.BeforeRevision {
				before, after := execution.BeforeRevision, execution.AfterRevision
				point.Before, point.After = &before, &after
				if task.Kind == "sync" {
					point.Added = after - before
				}
			}
		}
		if point.Started > 0 && point.Finished >= point.Started && point.Finished > 0 {
			duration := point.Finished - point.Started
			point.Duration = &duration
		}
		switch task.State {
		case "succeeded":
			succeeded++
			ended++
		case "failed", "interrupted":
			ended++
		}
		out.Summary.Added += point.Added
		out.Points = append(out.Points, point)
	}
	sort.SliceStable(out.Points, func(i, j int) bool {
		if out.Points[i].Started == out.Points[j].Started {
			return out.Points[i].ID < out.Points[j].ID
		}
		return out.Points[i].Started < out.Points[j].Started
	})
	out.Summary.Total = len(tasks)
	out.Summary.RevisionSpan = "—"
	var first, last *int64
	for _, point := range out.Points {
		if point.After != nil {
			if first == nil {
				first = point.Before
				if point.Kind != "sync" {
					first = point.After
				}
			}
			last = point.After
		}
	}
	if first != nil && last != nil {
		out.Summary.RevisionSpan = fmt.Sprintf("r%d → r%d", *first, *last)
	}
	if ended > 0 {
		rate := float64(succeeded) * 100 / float64(ended)
		out.Summary.SuccessRate = &rate
	}
	out.Pages = max(1, (len(tasks)+19)/20)
	out.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	out.Page = max(1, min(out.Page, out.Pages))
	manage := models.Owner(a.DB, user, repo.OrgID)
	out.RepositoryState = repositoryState(repo, manage)
	for i := len(tasks) - (out.Page-1)*20 - 1; i >= max(0, len(tasks)-out.Page*20); i-- {
		task := tasks[i]
		out.Items = append(out.Items, repositoryTaskRow(task, executions[task.ID], user, repo, manage, true))
	}
	return out, nil
}

func repositoryTaskRow(task models.Task, execution *models.TaskExecution, user *models.User, repo *models.Repository, manage, allowRetry bool) migrationTask {
	retry := allowRetry && task.State == "failed" && manage && !repo.Halted && (task.Kind != "local" || user.Admin)
	if task.Kind == "conversion" && (repo.Kind != "local" || repo.Status != "ready") {
		retry = false
	}
	kind := map[string]string{"local": "本地迁移", "mirror": "镜像迁移", "conversion": "转换镜像", "sync": "镜像同步"}[task.Kind]
	if kind == "" {
		kind = "迁移"
	}
	return migrationTask{ID: task.ID, Kind: kind + conversionActor(task), State: (&webPage{}).State(task.State), Phase: taskPhase(task), Started: (&webPage{}).Time(task.Started), Finished: finishedTaskTime(task), Message: taskResult(task, execution), Retry: retry, Indeterminate: (task.Kind == "conversion" || (task.Phase == "revision-properties" && (task.Verification != "revprops" || task.VerificationPosition < 0))) && task.State == "running"}
}
