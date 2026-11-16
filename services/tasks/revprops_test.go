// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"os"
	"path/filepath"
	"svnova/models"
	"svnova/services/svn"
	"sync/atomic"
	"testing"
)

func TestRealRevisionPropertyProgress(t *testing.T) {
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("set SVN_BIN for real revision property progress")
	}
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	runner := svn.Runner{Bin: bin}
	ctx := context.Background()
	source := filepath.Join(root, "source")
	dest := filepath.Join(root, "mirror")
	if _, err := runner.Run(ctx, "svnadmin", nil, "create", source); err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(ctx, "svn", nil, "mkdir", svn.LocalURL(source)+"/one", "-m", "one", "--non-interactive"); err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(ctx, "svnadmin", nil, "create", dest); err != nil {
		t.Fatal(err)
	}
	uuid, _, err := runner.LocalInfo(ctx, source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(ctx, "svnadmin", nil, "setuuid", dest, uuid); err != nil {
		t.Fatal(err)
	}
	if err := runner.InitializeExistingMirror(ctx, dest, svn.LocalURL(source), svn.Credentials{}, uuid, 0); err != nil {
		t.Fatal(err)
	}
	if err := runner.Sync(ctx, dest, svn.Credentials{}); err != nil {
		t.Fatal(err)
	}
	task := &models.Task{Kind: "sync", State: "running", CurrentRevision: 1, VerificationPosition: 999, ProgressStarted: 1}
	if _, err := db.Insert(task); err != nil {
		t.Fatal(err)
	}
	var sawProgress atomic.Bool
	runner.OnOutput = func() {
		var stored models.Task
		if ok, err := db.ID(task.ID).Get(&stored); err == nil && ok && stored.Phase == "revision-properties" && stored.VerificationPosition >= 0 {
			sawProgress.Store(true)
		}
	}
	worker := &Worker{DB: db, SVN: runner}
	repo := &models.Repository{Path: dest}
	if err := worker.copyRevprops(ctx, repo, task, svn.Credentials{}, 1); err != nil {
		t.Fatal(err)
	}
	var stored models.Task
	if _, err := db.ID(task.ID).Get(&stored); err != nil {
		t.Fatal(err)
	}
	if !sawProgress.Load() || stored.Verification != "revprops" || stored.VerificationPosition != 1 || stored.TotalRevision != 1 || stored.ProgressStarted <= 1 {
		t.Fatalf("missing live progress: %+v", stored)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := worker.copyRevprops(canceled, repo, task, svn.Credentials{}, 1); err == nil {
		t.Fatal("canceled copy succeeded")
	}
	stored = models.Task{}
	if _, err := db.ID(task.ID).Get(&stored); err != nil {
		t.Fatal(err)
	}
	if stored.VerificationPosition != -1 {
		t.Fatal("retry retained completed progress")
	}
}
