// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"testing"
)

func TestTaskCenterPermissionsPagingAndSelection(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "tasks.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Admin: true, Active: true}
	reader := &models.User{ID: 2, Name: "reader", Active: true}
	for _, v := range []any{admin, reader, &models.Organization{ID: 1, Name: "visible"}, &models.Organization{ID: 2, Name: "secret"}, &models.Repository{ID: 1, OrgID: 1, Name: "repo", Status: "ready"}, &models.Repository{ID: 2, OrgID: 2, Name: "hidden", Status: "ready"}, &models.Membership{OrgID: 1, UserID: 2}, &models.Team{ID: 1, OrgID: 1, Name: "readers"}, &models.TeamMember{TeamID: 1, UserID: 2}, &models.TeamRepository{TeamID: 1, RepositoryID: 1}} {
		if _, err = db.Insert(v); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 125; i++ {
		if _, err = db.Insert(&models.Task{RepositoryID: 1, Kind: "sync", State: "succeeded", Started: 1, Finished: 2}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.Insert(&models.Task{RepositoryID: 2, Kind: "local", State: "failed", Error: "secret"}); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	a.Templates, err = ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	page, err := a.taskCenterData(httptest.NewRequest("GET", "/tasks?page=2", nil), reader)
	if err != nil || page.Pagination.Total != 125 || page.Pagination.Pages != 7 || len(page.Items) != 20 || page.Items[0].ID != 105 || len(page.Organizations) != 1 || len(page.Repositories) != 1 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	filtered, err := a.taskCenterData(httptest.NewRequest("GET", "/tasks?repository=2", nil), reader)
	if err != nil || filtered.Pagination.Total != 0 {
		t.Fatal(filtered, err)
	}
	all, err := a.taskCenterData(httptest.NewRequest("GET", "/tasks", nil), admin)
	if err != nil || all.Pagination.Total != 126 {
		t.Fatal(all, err)
	}
	snapshot, err := a.migrationSnapshot(reader, &models.Repository{ID: 1, OrgID: 1}, 1)
	if err != nil || len(snapshot.Tasks) != 1 || snapshot.Tasks[0].ID != 1 || snapshot.Tasks[0].Retry {
		t.Fatal(snapshot, err)
	}
	if _, err = a.migrationSnapshot(reader, &models.Repository{ID: 1, OrgID: 1}, 126); err == nil {
		t.Fatal("foreign task selected")
	}
	for _, path := range []string{"/tasks", "/visible/repo/tasks?task_id=1"} {
		w := httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", path, nil), reader, "csrf")
		if w.Code != 200 || strings.Contains(w.Body.String(), "页面加载失败") {
			t.Fatal(path, w.Code, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	if err = a.taskCenterAPI(w, httptest.NewRequest("GET", "/api/v1/tasks", nil), reader); err != nil || strings.Contains(w.Body.String(), "secret") {
		t.Fatal(err, w.Body.String())
	}
	if _, err = db.Where("team_id=?", 1).Delete(new(models.TeamRepository)); err != nil {
		t.Fatal(err)
	}
	denied, err := a.taskCenterData(httptest.NewRequest("GET", "/tasks", nil), reader)
	if err != nil || denied.Pagination.Total != 0 || len(denied.Repositories) != 0 || len(denied.Organizations) != 0 {
		t.Fatal(denied, err)
	}
	for _, query := range []string{"?kind=invalid", "?state=interrupted", "?org=-1", "?repository=abc"} {
		if _, err = a.taskCenterData(httptest.NewRequest("GET", "/tasks"+query, nil), admin); err == nil {
			t.Fatal("invalid filter accepted", query)
		}
	}
	if _, err = db.ID(1).Cols("is_owner_team").Update(&models.Team{IsOwnerTeam: true}); err != nil {
		t.Fatal(err)
	}
	owner, err := a.taskCenterData(httptest.NewRequest("GET", "/tasks?page=999&state=succeeded&kind=sync&org=1", nil), reader)
	if err != nil || owner.Pagination.Total != 125 || owner.Pagination.Page != 7 || len(owner.Items) != 5 {
		t.Fatal(owner, err)
	}
	if !strings.Contains(owner.Pagination.Previous, "state=succeeded") || !strings.Contains(owner.Pagination.Previous, "org=1") {
		t.Fatal("lost filters")
	}
	if _, err = db.Insert(&models.Task{RepositoryID: 1, Kind: "sync", State: "running", Phase: "verify-target", CurrentRevision: 4, TotalRevision: 9, Started: 1}); err != nil {
		t.Fatal(err)
	}
	running, err := a.taskCenterData(httptest.NewRequest("GET", "/tasks?state=running", nil), reader)
	if err != nil || !running.Active || len(running.Items) != 1 || running.Items[0].Progress != 50 {
		t.Fatal(running, err)
	}
	reader.Active = false
	denied, err = a.taskCenterData(httptest.NewRequest("GET", "/tasks", nil), reader)
	if err != nil || denied.Pagination.Total != 0 {
		t.Fatal(denied, err)
	}
}
