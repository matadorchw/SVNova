// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/activity"
	"testing"
	"time"
)

func TestDashboardActivitiesAuthorizationAndShanghaiDay(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "activity.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	user := &models.User{ID: 1, Name: "reader", Active: true}
	now := time.Now()
	day := activity.Today(now)
	_, err = db.Insert(user, &models.Organization{ID: 1, Name: "visible"}, &models.Organization{ID: 2, Name: "hidden"}, &models.Membership{OrgID: 1, UserID: 1}, &models.Team{ID: 1, OrgID: 1, Name: "readers"}, &models.TeamMember{TeamID: 1, UserID: 1}, &models.Repository{ID: 1, OrgID: 1, Name: "allowed", Status: "failed", Revision: 2}, &models.Repository{ID: 2, OrgID: 2, Name: "secret", Status: "ready", Revision: 1}, &models.TeamRepository{TeamID: 1, RepositoryID: 1}, &models.Task{ID: 1, RepositoryID: 1, Kind: "local", State: "succeeded", Started: day.Add(-time.Second).Unix()}, &models.Task{ID: 2, RepositoryID: 1, Kind: "sync", State: "failed", Started: day.Unix()}, &models.Task{ID: 3, RepositoryID: 2, Kind: "sync", State: "failed", Started: day.Unix()}, &models.TaskExecution{TaskID: 2, State: "failed", Started: day.UnixMilli()}, &models.TaskExecution{TaskID: 2, State: "succeeded", Started: day.UnixMilli()})

	if err != nil {
		t.Fatal(err)
	}
	read := func() heatmapResponse {
		t.Helper()
		w := httptest.NewRecorder()
		if err := (&App{DB: db}).activityHeatmap(w, httptest.NewRequest("GET", "/api/v1/activity/heatmap", nil), user); err != nil {
			t.Fatal(err)
		}
		var out heatmapResponse
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	out := read()
	if out.TotalContributions != 3 || out.Pending || len(out.HeatmapData) != 2 {
		t.Fatalf("wrong scope or day grouping: %+v", out)
	}
	if err := (&App{DB: db}).activityHeatmap(httptest.NewRecorder(), httptest.NewRequest("GET", "/?org=2", nil), user); err == nil {
		t.Fatal("unauthorized organization accepted")
	}
	_, err = db.Where("team_id=1 AND repository_id=1").Delete(new(models.TeamRepository))
	if err != nil {
		t.Fatal(err)
	}
	out = read()
	if out.TotalContributions != 0 || len(out.HeatmapData) != 0 {
		t.Fatal("revoked repository remained in statistics")
	}
}

func TestDashboardFeedOrganizationLayout(t *testing.T) {
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	page := &webPage{FeedPage: 1}
	for i, kind := range []string{"local", "mirror", "sync", "conversion"} {
		event := dashboardEvent{OrganizationID: 9, Organization: "组织", Repository: "仓库", RepositoryID: 8, TaskID: 123, ExecutionID: 456, Kind: kind, Trigger: "manual", State: "succeeded", Timestamp: 1794960000}
		if event.Icon() != []string{"repo", "mirror", "sync", "arrow-switch"}[i] {
			t.Fatal(event.Icon())
		}
		page.Feed = append(page.Feed, event)
	}
	var out bytes.Buffer
	if err := templates.ExecuteTemplate(&out, "user/dashboard/feeds", page); err != nil {
		t.Fatal(err)
	}
	html := out.String()
	if strings.Contains(html, "组织/仓库</a>") || !strings.Contains(html, "本地仓库迁移（手动）") || !strings.Contains(html, "activity-task-link") {
		t.Fatal("incorrect heading or task link")
	}
	if !(strings.Index(html, "activity-task-icon") < strings.Index(html, "activity-status") && strings.Index(html, "activity-status") < strings.Index(html, "activity-task-link")) {
		t.Fatal("incorrect action order")
	}
	for state, icon := range map[string]string{"succeeded": "check", "failed": "x", "running": "hourglass", "queued": "clock", "interrupted": "alert"} {
		if (dashboardEvent{State: state}).StateIcon() != icon {
			t.Fatal(state)
		}
	}

	for _, want := range []string{"/avatars/organizations/9?size=80", "activity-organization-avatar", `width="40"`, "执行了", "手动", "查看任务记录", "2026-", "#arrow-switch"} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(html, "任务 #") || strings.Contains(html, "执行 #") {
		t.Fatal("internal identifiers visible")
	}
	if strings.Index(html, "activity-organization-avatar") > strings.Index(html, "activity-body") || strings.Index(html, "activity-task-icon") < strings.Index(html, "activity-body") {
		t.Fatal("incorrect feed layout order")
	}
	page.Feed = nil
	out.Reset()
	if err := templates.ExecuteTemplate(&out, "user/dashboard/feeds", page); err != nil || !strings.Contains(out.String(), "暂无 SVN 仓库活动") {
		t.Fatalf("empty feed: %v", err)
	}
}
