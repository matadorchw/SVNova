// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/tasks"
	"testing"
)

func TestMirrorSettingsManualSyncAuthorizationAndActiveTask(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Active: true, Admin: true}
	reader := &models.User{ID: 2, Name: "reader", Active: true}
	_, err = db.Insert(admin, reader, &models.Organization{ID: 1, Name: "acs"}, &models.Repository{ID: 1, OrgID: 1, Name: "mirror", Kind: "mirror", Status: "ready", Remote: "http://127.0.0.1:1/source"})
	if err != nil {
		t.Fatal(err)
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Templates: templates}
	if err = a.repositories(httptest.NewRecorder(), httptest.NewRequest("POST", "/", nil), reader, []string{"1", "sync"}); err == nil {
		t.Fatal("unauthorized manual sync accepted")
	}
	if err = a.repositories(httptest.NewRecorder(), httptest.NewRequest("POST", "/", nil), admin, []string{"1", "sync"}); err != nil {
		t.Fatal(err)
	}
	if err = a.repositories(httptest.NewRecorder(), httptest.NewRequest("POST", "/", nil), admin, []string{"1", "sync"}); err == nil {
		t.Fatal("duplicate task accepted")
	} else if failure, ok := err.(*httpError); !ok || failure.Status != 409 || !strings.Contains(failure.Message, "正在排队") {
		t.Fatalf("unclear active task error: %v", err)
	}
	p := &webPage{User: admin, Page: "repositories/1"}
	if err = a.webData(httptest.NewRequest("GET", "/repositories/1?tab=settings", nil), p); err != nil {
		t.Fatal(err)
	}
	if !p.RepositoryTaskActive {
		t.Fatal("settings did not detect active task")
	}
	w := httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/acs/mirror/settings/mirror", nil), admin, "csrf")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "镜像设置") || !strings.Contains(w.Body.String(), "同步任务正在排队或执行中") || !strings.Contains(w.Body.String(), `data-next="/acs/mirror/tasks"`) {
		t.Fatal("mirror synchronization presentation missing")
	}
	count, _ := db.Where("repository_id=1 AND state='queued'").Count(new(models.Task))
	if count != 1 {
		t.Fatal("duplicate sync created more than one task")
	}
	_, _ = db.Where("repository_id=1").Cols("state", "error").Update(&models.Task{State: "failed", Error: "E215004"})
	_, _ = db.ID(1).Cols("status").Update(&models.Repository{Status: "failed"})
	w = httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/acs/mirror/settings/mirror", nil), admin, "csrf")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `data-api="/acs/mirror/settings/migrate/retry"`) || !strings.Contains(w.Body.String(), `<span>重试迁移</span>`) || !strings.Contains(w.Body.String(), "查看失败原因") {
		t.Fatal("failed mirror settings lack an enabled migration retry")
	}
	_, _ = db.ID(1).Cols("halted").Update(&models.Repository{Halted: true})
	w = httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/acs/mirror/settings/mirror", nil), admin, "csrf")
	if strings.Contains(w.Body.String(), `<span>重试迁移</span>`) || !strings.Contains(w.Body.String(), "完整性检查异常") {
		t.Fatal("integrity halt allowed migration retry")
	}
}

func TestMirrorSettingsConflictMessages(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "conflict.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := &models.Repository{ID: 1, Kind: "mirror", Status: "ready", Remote: "https://example.com/original"}
	if _, err := db.Insert(repo); err != nil {
		t.Fatal(err)
	}
	app := &App{DB: db, Worker: &tasks.Worker{DB: db}}
	actor := &models.User{Active: true, Admin: true}
	check := func(want string) {
		t.Helper()
		err := app.settings(httptest.NewRecorder(), httptest.NewRequest("POST", "/", nil), actor, repo)
		failure, ok := err.(*httpError)
		if !ok || failure.Status != 409 || !strings.Contains(failure.Message, want) {
			t.Fatalf("unclear settings conflict: %v", err)
		}
	}
	if !app.Worker.TryLock() {
		t.Fatal("cannot acquire fixture lock")
	}
	check("其他仓库")
	app.Worker.Unlock()
	if _, err := db.Insert(&models.Task{RepositoryID: repo.ID, Kind: "sync", State: "queued"}); err != nil {
		t.Fatal(err)
	}
	check("查看任务记录")
	var stored models.Repository
	if _, err := db.ID(repo.ID).Get(&stored); err != nil {
		t.Fatal(err)
	}
	if stored.Remote != repo.Remote {
		t.Fatal("rejected settings changed remote")
	}
}
