// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"svnova/models"
	"svnova/services/directory"
	"svnova/services/tasks"
)

func TestBackgroundWaitsForWorkerOnCancellation(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	task := &models.Task{State: "running"}
	if _, err = db.Insert(task); err != nil {
		t.Fatal(err)
	}
	worker := &tasks.Worker{DB: db}
	if !worker.TryLock() {
		t.Fatal("cannot hold worker lock")
	}
	locked := true
	defer func() {
		if locked {
			worker.Unlock()
		}
	}()
	a := &App{DB: db, Worker: worker, Directory: &directory.Directory{DB: db}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { defer close(done); a.Background(ctx) }()
	deadline := time.After(5 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
waiting:
	for {
		select {
		case <-deadline:
			t.Fatal("worker did not start")
		case <-ticker.C:
			var current models.Task
			ok, e := db.ID(task.ID).Get(&current)
			if e != nil {
				t.Fatal(e)
			}
			if ok && current.State == "queued" {
				break waiting
			}
		}
	}
	cancel()
	select {
	case <-done:
		t.Fatal("background returned while worker still held its task lock")
	case <-time.After(50 * time.Millisecond):
	}
	worker.Unlock()
	locked = false
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("background did not join canceled worker")
	}
}
