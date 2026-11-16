// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"testing"
)

func TestRepositoryTaskHistoryMetricsPaginationAndAuthorization(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Active: true, Admin: true}
	reader := &models.User{ID: 2, Name: "reader", Active: true}
	repo := &models.Repository{ID: 1, OrgID: 1, Name: "demo", Status: "ready", Kind: "mirror"}
	if _, err = db.Insert(admin, reader, repo, &models.Organization{ID: 1, Name: "acs"}, &models.Membership{OrgID: 1, UserID: 2}, &models.Team{ID: 1, OrgID: 1, Name: "readers"}, &models.TeamMember{TeamID: 1, UserID: 2}, &models.TeamRepository{TeamID: 1, RepositoryID: 1}); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	for i := int64(1); i <= 125; i++ {
		kind, state := "sync", "succeeded"
		before, after := int64(100), int64(100)
		switch i {
		case 1:
			kind = "local"
			before = 0
		case 2:
			after = 105
		case 3:
			kind = "conversion"
			before = 105
			after = 105
		case 4:
			state = "failed"
		case 5:
			state = "interrupted"
		case 6:
			state = "queued"
		case 7:
			state = "running"
		default:
			before = 105
			after = 105
		}
		task := &models.Task{ID: i, RepositoryID: 1, Kind: kind, State: state, Started: 1000 + i, Finished: 1002 + i, Error: "file:///secret private-password"}
		if _, err = db.Insert(task); err != nil {
			t.Fatal(err)
		}
		if i == 8 {
			continue
		} // Legacy task without a snapshot.
		execution := &models.TaskExecution{TaskID: i, Started: (1000 + i) * 1000, Finished: (1002+i)*1000 + 250, State: state, BeforeRevision: before, AfterRevision: after, RevisionKnown: state == "succeeded"}
		if _, err = db.Insert(execution); err != nil {
			t.Fatal(err)
		}
	}
	// A failed attempt before the latest successful retry must not affect counts.
	if _, err = db.Insert(&models.TaskExecution{TaskID: 2, State: "succeeded", Started: 1002000, Finished: 1004500, BeforeRevision: 100, AfterRevision: 105, RevisionKnown: true}); err != nil {
		t.Fatal(err)
	}
	get := func(page string) repositoryTaskHistory {
		t.Helper()
		v, e := a.repositoryTaskHistory(httptest.NewRequest("GET", "/?page="+page, nil), admin, repo)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	result := get("1")
	if result.Summary.Total != 125 || result.Summary.Added != 5 || result.Summary.RevisionSpan != "r100 → r105" || result.Summary.SuccessRate == nil || *result.Summary.SuccessRate != float64(121)*100/123 || !result.Active {
		t.Fatalf("wrong summary: %+v", result.Summary)
	}
	if len(result.Points) != 125 || len(result.Items) != 20 || result.Pages != 7 || result.Items[0].ID != 125 {
		t.Fatal("history was truncated or pagination incorrect")
	}
	if result.Points[1].Duration == nil || *result.Points[1].Duration != 2500 || result.Points[7].After != nil || result.Points[5].Finished != 0 || result.Points[5].Duration != nil {
		t.Fatal("snapshot precision, legacy data or active duration incorrect")
	}
	for _, item := range result.Items {
		if strings.Contains(item.Message, "完成于") {
			t.Fatal("finish time still embedded in result")
		}
	}
	if v := get("999"); v.Page != 7 || len(v.Items) != 5 || v.Items[0].ID != 5 {
		t.Fatal("last page clamp incorrect")
	}
	if v := get("-1"); v.Page != 1 {
		t.Fatal("invalid page not clamped")
	}
	// API reads share repository authorization; retry remains restricted.
	w := httptest.NewRecorder()
	if err = a.repositories(w, httptest.NewRequest("GET", "/api/v1/repositories/1/tasks/history?page=7", nil), reader, []string{"1", "tasks", "history"}); err != nil {
		t.Fatal(err)
	}
	var response repositoryTaskHistory
	if err = json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	for _, item := range response.Items {
		if item.Retry {
			t.Fatal("reader received retry action")
		}
	}
	if strings.Contains(w.Body.String(), "private-password") || strings.Contains(w.Body.String(), "file:///") {
		t.Fatal("API leaked private source or error")
	}
	if _, err = db.Where("team_id=1").Delete(new(models.TeamRepository)); err != nil {
		t.Fatal(err)
	}
	if err = a.repositories(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), reader, []string{"1", "tasks", "history"}); err == nil {
		t.Fatal("revoked reader retained access")
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a.Templates = templates
	w = httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/acs/demo/tasks?task_id=2", nil), admin, "csrf")
	if w.Code != 200 || strings.Contains(w.Body.String(), "data-repository-history") || !strings.Contains(w.Body.String(), "结束时间") {
		t.Fatalf("selected task did not preserve standalone detail: status=%d body=%s", w.Code, w.Body.String())
	}
	if _, err = db.Where("repository_id=1").Delete(new(models.Task)); err != nil {
		t.Fatal(err)
	}
	empty := get("1")
	if empty.Summary.Total != 0 || empty.Summary.SuccessRate != nil || empty.Summary.RevisionSpan != "—" || len(empty.Items) != 0 {
		t.Fatal("empty repository metrics incorrect")
	}
}
