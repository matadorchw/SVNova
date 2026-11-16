// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"svnova/models"
	"svnova/services/svn"
	"testing"
	"time"
)

func TestExecutionSummary(t *testing.T) {
	for _, tc := range []struct {
		kind, state   string
		before, after int64
		known         bool
		want          string
	}{
		{"sync", "succeeded", 100, 112, true, "新增 12 个修订（r100 → r112）"},
		{"sync", "succeeded", 112, 112, true, "无新增修订（r112）"},
		{"sync", "succeeded", 0, 0, true, "无新增修订（r0）"},
		{"sync", "succeeded", 0, 1, false, "同步完成"},
		{"sync", "succeeded", 4, 3, true, "同步完成"},
		{"sync", "succeeded", -1, 3, true, "同步完成"},
		{"mirror", "succeeded", 0, 112, true, "已迁移至 r112"},
		{"local", "succeeded", 0, 0, true, "已迁移至 r0"},
		{"conversion", "succeeded", 112, 112, true, "已转换为镜像（r112），未逐修订比对历史内容"},
		{"sync", "queued", 100, 112, true, "—"},
		{"sync", "running", 100, 112, true, "—"},
		{"sync", "interrupted", 100, 112, true, "执行已中断"},
	} {
		t.Run(tc.kind+tc.state+tc.want, func(t *testing.T) {
			got := executionSummary(models.Task{Kind: tc.kind, State: tc.state}, &models.TaskExecution{State: tc.state, BeforeRevision: tc.before, AfterRevision: tc.after, RevisionKnown: tc.known})
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
	if got := executionSummary(models.Task{Kind: "sync", State: "failed", Error: "private internal failure"}, nil); got == "" || strings.Contains(got, "新增") {
		t.Fatal("failure presented as success", got)
	}
}

func TestResultSnapshotsAcrossActivityAndTasks(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Active: true, Admin: true}
	now := time.Now().UnixMilli()
	if _, err = db.Insert(admin, &models.Organization{ID: 1, Name: "org"}, &models.Repository{ID: 1, OrgID: 1, Name: "repo", Status: "ready", Revision: 999}, &models.Task{ID: 1, RepositoryID: 1, Kind: "sync", State: "succeeded"}, &models.TaskExecution{ID: 1, TaskID: 1, State: "succeeded", Started: now - 2000, BeforeRevision: 100, AfterRevision: 112, RevisionKnown: true}, &models.TaskExecution{ID: 2, TaskID: 1, State: "succeeded", Started: now - 1000, BeforeRevision: 112, AfterRevision: 115, RevisionKnown: true}); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	a.Templates, err = ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	page, err := a.taskCenterData(httptest.NewRequest("GET", "/tasks", nil), admin)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].Result != "新增 3 个修订（r112 → r115）" {
		t.Fatal("latest result missing", page.Items)
	}
	w := httptest.NewRecorder()
	if err = a.taskCenterAPI(w, httptest.NewRequest("GET", "/api/v1/tasks", nil), admin); err != nil {
		t.Fatal(err)
	}
	var response struct {
		HTML  string
		Items []taskCenterRow
	}
	if err = json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 1 || response.Items[0].Result != page.Items[0].Result || !strings.Contains(response.HTML, "结果概要") || !strings.Contains(response.HTML, page.Items[0].Result) {
		t.Fatal("polling result differs", response)
	}
	snapshot, err := a.migrationSnapshot(admin, &models.Repository{ID: 1, OrgID: 1, Status: "ready"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(snapshot.Tasks[0].Message, page.Items[0].Result) {
		t.Fatal("repository result differs", snapshot)
	}
	var events []dashboardEvent
	if err = db.SQL("SELECT * FROM (" + repositoryActivitySQL + ") ORDER BY execution_id").Find(&events); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Result() != "新增 12 个修订（r100 → r112）" || events[1].Result() != page.Items[0].Result {
		t.Fatal("historical snapshot changed", events)
	}
	if _, err = db.ID(1).Cols("state").Update(&models.Task{State: "queued"}); err != nil {
		t.Fatal(err)
	}
	page, err = a.taskCenterData(httptest.NewRequest("GET", "/tasks", nil), admin)
	if err != nil || page.Items[0].Result != "—" {
		t.Fatal("retry reused previous success", err, page)
	}
}

func TestRealSVNResultSummaries(t *testing.T) {
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("set SVN_BIN for integration")
	}
	root := t.TempDir()
	ctx := context.Background()
	runner := svn.Runner{Bin: bin}
	remote := filepath.Join(root, "remote")
	if _, err := runner.Run(ctx, "svnadmin", nil, "create", remote); err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(ctx, "svn", nil, "mkdir", svn.LocalURL(remote)+"/initial", "-m", "initial", "--non-interactive"); err != nil {
		t.Fatal(err)
	}
	db, err := models.Open(filepath.Join(root, "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dest := filepath.Join(root, "mirror")
	if _, err = runner.Run(ctx, "svnadmin", nil, "create", dest); err != nil {
		t.Fatal(err)
	}
	hook, content := "pre-revprop-change", "#!/bin/sh\nexit 0\n"
	if runtime.GOOS == "windows" {
		hook, content = "pre-revprop-change.bat", "@exit /b 0\r\n"
	}
	if err = os.WriteFile(filepath.Join(dest, "hooks", hook), []byte(content), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = runner.Run(ctx, "svnsync", nil, "initialize", svn.LocalURL(dest), svn.LocalURL(remote), "--non-interactive"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Insert(&models.Organization{ID: 1, Name: "org"}, &models.Repository{ID: 1, OrgID: 1, Name: "mirror", Kind: "mirror", Status: "ready", Path: dest}); err != nil {
		t.Fatal(err)
	}
	run := func(kind, want string) {
		t.Helper()
		_, before, err := runner.LocalInfo(ctx, dest)
		if err != nil {
			t.Fatal(err)
		}
		if err = runner.Sync(ctx, dest, svn.Credentials{}); err != nil {
			t.Fatal(err)
		}
		_, after, err := runner.LocalInfo(ctx, dest)
		if err != nil {
			t.Fatal(err)
		}
		task := models.Task{RepositoryID: 1, Kind: kind, State: "succeeded"}
		if _, err = db.Insert(&task); err != nil {
			t.Fatal(err)
		}
		if _, err = db.Insert(&models.TaskExecution{TaskID: task.ID, State: "succeeded", BeforeRevision: before, AfterRevision: after, RevisionKnown: true}); err != nil {
			t.Fatal(err)
		}
		executions, err := (&App{DB: db}).latestTaskExecutions([]int64{task.ID})
		if err != nil {
			t.Fatal(err)
		}
		if got := executionSummary(task, executions[task.ID]); got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
	run("mirror", "已迁移至 r1")
	if _, err = runner.Run(ctx, "svn", nil, "mkdir", svn.LocalURL(remote)+"/next", "-m", "next", "--non-interactive"); err != nil {
		t.Fatal(err)
	}
	run("sync", "新增 1 个修订（r1 → r2）")
	run("sync", "无新增修订（r2）")
}
