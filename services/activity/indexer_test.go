// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package activity

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/svn"
	"testing"
	"time"
)

func TestRevisionIndexPagesResumeAndRefreshChangedDates(t *testing.T) {
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("SVN_BIN is not set")
	}
	runner := svn.Runner{Bin: bin}
	root := t.TempDir()
	repoPath := filepath.Join(root, "repo")
	fixture := filepath.Join(root, "fixture")
	if err := os.Mkdir(fixture, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture, "中文.txt"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(tool string, args ...string) {
		t.Helper()
		if _, err := runner.Run(context.Background(), tool, nil, args...); err != nil {
			t.Fatal(err)
		}
	}
	run("svnadmin", "create", repoPath)
	target, err := svn.Target(repoPath, "")
	if err != nil {
		t.Fatal(err)
	}
	target = strings.TrimSuffix(target, "@")
	run("svn", "import", fixture, target, "-m", "修订一", "--username", "alpha", "--non-interactive")
	run("svn", "mkdir", strings.TrimRight(target, "/")+"/dir", "-m", "修订二", "--username", "beta", "--non-interactive")
	db, err := models.Open(filepath.Join(root, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := &models.Repository{ID: 1, Name: "fixture", Path: repoPath, Status: "ready", Revision: 2, LastSuccess: 10}
	if _, err = db.Insert(repo); err != nil {
		t.Fatal(err)
	}
	indexer := &Indexer{DB: db, SVN: runner, BatchSize: 1}
	now := time.Now().Add(time.Second)
	if err = indexer.Step(context.Background(), repo.ID, now); err != nil {
		t.Fatal(err)
	}
	var state models.ActivityIndex
	_, _ = db.ID(repo.ID).Get(&state)
	if state.Complete || state.NextRevision != 1 {
		t.Fatalf("bad cursor %+v", state)
	}
	indexer = &Indexer{DB: db, SVN: runner, BatchSize: 1}
	if err = indexer.Step(context.Background(), repo.ID, now); err != nil {
		t.Fatal(err)
	}
	count, _ := db.Count(new(models.RevisionEvent))
	if count != 2 {
		t.Fatalf("indexed %d revisions", count)
	}
	oldDate := filepath.Join(root, "date.txt")
	if err = os.WriteFile(oldDate, []byte(now.AddDate(-2, 0, 0).UTC().Format("2006-01-02T15:04:05.000000Z")), 0600); err != nil {
		t.Fatal(err)
	}
	run("svnadmin", "setrevprop", repoPath, "-r", "1", "svn:date", oldDate)
	repo.LastSuccess++
	_, _ = db.ID(repo.ID).Cols("last_success").Update(repo)
	for range 2 {
		if err = indexer.Step(context.Background(), repo.ID, now); err != nil {
			t.Fatal(err)
		}
	}
	var events []models.RevisionEvent
	_ = db.Find(&events)
	if len(events) != 1 || events[0].Revision != 2 || events[0].Author != "beta" {
		t.Fatalf("stale or incorrect events %+v", events)
	}
	if err = indexer.Step(context.Background(), repo.ID, now); err != nil {
		t.Fatal(err)
	}
	count, _ = db.Count(new(models.RevisionEvent))
	if count != 1 {
		t.Fatal("completed scan duplicated events")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err = indexer.Step(canceled, repo.ID, now); err != context.Canceled {
		t.Fatalf("cancel: %v", err)
	}
}
func TestIndexFailurePersistsAndQueuedTasksTakePriority(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, _ = db.Insert(&models.Repository{ID: 1, Status: "ready", Revision: 1, Path: "missing"}, &models.Task{RepositoryID: 1, State: "queued"})
	indexer := &Indexer{DB: db, SVN: svn.Runner{Bin: filepath.Join(t.TempDir(), "missing")}}
	indexer.tick(context.Background())
	count, _ := db.Count(new(models.ActivityIndex))
	if count != 0 {
		t.Fatal("index ran ahead of queued task")
	}
	if err = indexer.Step(context.Background(), 1, time.Now()); err == nil {
		t.Fatal("missing SVN executable succeeded")
	}
	var state models.ActivityIndex
	_, _ = db.ID(1).Get(&state)
	if !state.Failed || state.Complete || state.NextRevision != 1 {
		t.Fatalf("failure state %+v", state)
	}
}
