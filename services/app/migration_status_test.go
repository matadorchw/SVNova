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
	"time"
)

func TestMigrationStatusPresentationAndAuthorization(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "migration.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Active: true, Admin: true}
	reader := &models.User{ID: 2, Name: "reader", Active: true}
	repo := &models.Repository{ID: 1, OrgID: 1, Name: "demo", Kind: "local", Status: "pending"}
	task := &models.Task{ID: 1, RepositoryID: 1, State: "running", Kind: "local", Phase: "copying"}
	_, err = db.Insert(admin, reader, &models.Organization{ID: 1, Name: "acs"}, &models.Membership{OrgID: 1, UserID: 2}, &models.Team{ID: 1, OrgID: 1, Name: "readers"}, &models.TeamMember{TeamID: 1, UserID: 2}, &models.TeamRepository{TeamID: 1, RepositoryID: 1}, repo, task)
	if err != nil {
		t.Fatal(err)
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Templates: templates}
	get := func(actor *models.User, path string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		if !a.migrationAction(w, httptest.NewRequest("GET", path, nil), actor) {
			t.Fatal("route not handled")
		}
		return w
	}
	decode := func(w *httptest.ResponseRecorder) migrationStatus {
		t.Helper()
		var v migrationStatus
		if w.Code != 200 {
			t.Fatalf("status %d: %s", w.Code, w.Body.String())
		}
		if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	result := decode(get(reader, "/acs/demo/-/migrate/status"))
	if result.Status != 2 || !strings.Contains(result.Message, "复制") || result.CanRetry {
		t.Fatalf("unexpected running status: %+v", result)
	}
	w := httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/acs/demo", nil), reader, "csrf")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "data-migration-panel") || strings.Contains(w.Body.String(), "data-poll") {
		t.Fatal("pending repository did not render migration panel")
	}
	task.State = "failed"
	task.Error = "file:///private-source credential-secret"
	_, err = db.ID(task.ID).Cols("state", "error").Update(task)
	if err != nil {
		t.Fatal(err)
	}
	repo.Status = "failed"
	_, _ = db.ID(repo.ID).Cols("status").Update(repo)
	result = decode(get(reader, "/repositories/1/-/migrate/status"))
	if result.Status != 3 || result.CanRetry || strings.Contains(result.Message, "credential-secret") {
		t.Fatalf("unsafe failure response: %+v", result)
	}
	result = decode(get(admin, "/acs/demo/-/migrate/status"))
	if !result.CanRetry {
		t.Fatal("administrator lost local retry")
	}
	w = httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/acs/demo", nil), reader, "csrf")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "查看任务记录") || strings.Contains(w.Body.String(), "credential-secret") {
		t.Fatal("failed migration has no safe history entry")
	}
	w = httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/acs/demo/tasks", nil), reader, "csrf")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "data-task-status") || strings.Contains(w.Body.String(), "data-migration-panel") || strings.Contains(w.Body.String(), "credential-secret") {
		t.Fatal("failed repository task history is inaccessible or unsafe")
	}
	denied := httptest.NewRecorder()
	a.migrationAction(denied, httptest.NewRequest("POST", "/acs/demo/settings/migrate/retry", nil), reader)
	if denied.Code != 403 {
		t.Fatal("reader retry accepted")
	}
	task.State = "queued"
	_, _ = db.ID(task.ID).Cols("state").Update(task)
	if result = decode(get(reader, "/acs/demo/-/migrate/status")); result.Status != 2 || result.CanRetry {
		t.Fatal("queued retry remained failed")
	}
	task.State = "succeeded"
	_, _ = db.ID(task.ID).Cols("state").Update(task)
	repo.Status = "ready"
	_, _ = db.ID(repo.ID).Cols("status").Update(repo)
	if result = decode(get(reader, "/acs/demo/-/migrate/status")); result.Status != 4 {
		t.Fatal("successful migration not finished")
	}
	w = httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/acs/demo/tasks", nil), reader, "csrf")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "data-task-status") || strings.Contains(w.Body.String(), "data-migration-panel") {
		t.Fatal("ready task page did not render history")
	}
	if !strings.Contains(w.Body.String(), `aria-label="本地仓库"`) {
		t.Fatal("local repository icon label missing")
	}
	repo.Kind = "mirror"
	_, _ = db.ID(repo.ID).Cols("kind").Update(repo)
	w = httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/acs/demo/tasks", nil), reader, "csrf")
	if !strings.Contains(w.Body.String(), `octicons.svg#mirror`) || strings.Contains(w.Body.String(), `<span class="ui basic label">镜像</span>`) {
		t.Fatal("mirror type was not represented by upstream icon")
	}
	_, _ = db.Where("team_id=1").Delete(new(models.TeamRepository))
	if get(reader, "/acs/demo/-/migrate/status").Code != 403 {
		t.Fatal("revoked user retained migration status")
	}
}

func TestCompletedTaskHasTerminalPhaseAndExplicitResult(t *testing.T) {
	task := models.Task{State: "succeeded", Kind: "sync", Phase: "revision-properties", CurrentRevision: 39, TotalRevision: 39, Finished: 1}
	if taskPhase(task) != "已完成" {
		t.Fatal("successful task retained running phase")
	}
	result := taskResult(task)
	if !strings.Contains(result, "同步完成") || strings.Contains(result, "r39") || strings.Contains(result, "完成于") {
		t.Fatalf("unclear task result: %s", result)
	}
	task.Kind = "mirror"
	if !strings.Contains(taskResult(task), "迁移完成") {
		t.Fatal("migration result missing")
	}
	task.CurrentRevision = 3
	if strings.Contains(taskResult(task), "r3") {
		t.Fatal("legacy intermediate progress was presented as final revision")
	}
	task.State = "failed"
	if strings.Contains(taskPhase(task), "正在") || !strings.Contains(taskPhase(task), "失败阶段") {
		t.Fatal("failed task retained running wording")
	}
}

func TestLocalVerificationProgressText(t *testing.T) {
	task := models.Task{Kind: "local", State: "running", Phase: "verify-source", CurrentRevision: 4, TotalRevision: 9}
	if text := taskPhase(task); !strings.Contains(text, "5 / 10（50.0%）") {
		t.Fatal("incorrect verification progress", text)
	}
	task.TotalRevision = 0
	task.CurrentRevision = -1
	if text := taskPhase(task); !strings.Contains(text, "0 / 1（0.0%）") {
		t.Fatal("empty repository progress incorrect", text)
	}
}

func TestMetadataProgressDoesNotInventPercentage(t *testing.T) {
	task := models.Task{Kind: "local", State: "running", Phase: "verify-source", Verification: "metadata", VerificationPosition: 1000, TotalRevision: 2000}
	text := taskPhase(task)
	if !strings.Contains(text, "当前检查位置 r1000") || strings.Contains(text, "%") || strings.Contains(text, "已校验修订") {
		t.Fatal("incorrect metadata progress", text)
	}
	task.VerificationPosition = -1
	if text = taskPhase(task); !strings.Contains(text, "仓库级元数据") {
		t.Fatal("missing repository metadata state", text)
	}
	task.Verification = "revisions"
	task.CurrentRevision = 0
	if text = taskPhase(task); !strings.Contains(text, "已校验修订 1 / 2001") {
		t.Fatal("revision transition failed", text)
	}
}

func TestRevisionPropertyStatusUsesIndependentProgress(t *testing.T) {
	now := time.Now().Unix()
	task := models.Task{RepositoryID: 1, Kind: "sync", State: "running", Phase: "revision-properties", CurrentRevision: 6191, TotalRevision: 6191, Started: now - 484, ProgressStarted: now - 10, ProgressUpdated: now - 2}
	text := taskPhase(task)
	if strings.Contains(text, "r6191 / 6191") || !strings.Contains(text, "等待修订属性进度") {
		t.Fatal(text)
	}
	task.Verification = "revprops"
	task.VerificationPosition = 0
	text = taskPhase(task)
	if !strings.Contains(text, "1 / 6192") || !strings.Contains(text, "当前 r0") || !strings.Contains(text, "本阶段已用时") || !strings.Contains(text, "累计已用时") {
		t.Fatal(text)
	}
	db, err := models.Open(filepath.Join(t.TempDir(), "progress.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	task.VerificationPosition = 3095
	if _, err := db.Insert(&task); err != nil {
		t.Fatal(err)
	}
	app := &App{DB: db}
	repo := &models.Repository{ID: 1, Kind: "mirror", Status: "ready"}
	snapshot, err := app.migrationSnapshot(&models.User{Active: true, Admin: true}, repo)
	if err != nil || snapshot.Progress != 50 || snapshot.Tasks[0].Indeterminate {
		t.Fatalf("unexpected property progress: %+v %v", snapshot, err)
	}
}
