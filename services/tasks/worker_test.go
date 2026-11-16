// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"os"
	"path/filepath"
	"svnova/models"
	"svnova/services/security"
	"svnova/services/storage"
	"svnova/services/svn"
	"testing"
)

func TestLocalMigrationAndRecovery(t *testing.T) {
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("set SVN_BIN for integration")
	}
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v, err := security.NewVault(filepath.Join(root, "key"))
	if err != nil {
		t.Fatal(err)
	}
	r := svn.Runner{Bin: bin}
	source := filepath.Join(root, "source")
	if _, err = r.Run(context.Background(), "svnadmin", nil, "create", source); err != nil {
		t.Fatal(err)
	}
	managed := filepath.Join(root, "managed")
	repo := &models.Repository{Name: "local", OrgID: 1, Path: filepath.Join(managed, "repo"), Kind: "local", Status: "pending"}
	_, _ = db.Insert(repo)
	task := &models.Task{RepositoryID: repo.ID, Kind: "local", Source: source, State: "queued"}
	_, _ = db.Insert(task)
	w := &Worker{DB: db, SVN: r, Vault: v, Root: managed}
	w.Tick(context.Background())
	repoID, taskID := repo.ID, task.ID
	repo = new(models.Repository)
	task = new(models.Task)
	_, _ = db.ID(repoID).Get(repo)
	_, _ = db.ID(taskID).Get(task)
	if repo.Status != "ready" || task.State != "succeeded" {
		t.Fatalf("migration %+v %+v", repo, task)
	}
	size, sizeErr := storage.DirectorySize(context.Background(), repo.Path)
	if sizeErr != nil || repo.SizeUpdated == 0 || repo.Size != size {
		t.Fatalf("migration size %d/%d: %v", repo.Size, size, sizeErr)
	}
	if task.Phase != "completed" || task.Finished == 0 || task.CurrentRevision != repo.Revision || task.TotalRevision != repo.Revision {
		t.Fatal("successful task did not persist final progress")
	}
	if _, err = os.Stat(source); !os.IsNotExist(err) {
		t.Fatal("source not moved")
	}
	if _, err = os.Stat(repo.Path); err != nil {
		t.Fatal(err)
	}
	task.Phase = "moving"
	if err = w.local(context.Background(), repo, task); err != nil {
		t.Fatalf("rename recovery: %v", err)
	}
}
func TestPathIsolation(t *testing.T) {
	root := t.TempDir()
	if Within(root, root) || Within(root, filepath.Dir(root)) || !Within(root, filepath.Join(root, "repo")) {
		t.Fatal("invalid root boundary")
	}
}

func TestRetryExecutionResetsTaskStartTime(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "retry.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	task := &models.Task{State: "queued", Started: 100, Finished: 200}
	if _, err = db.Insert(task); err != nil {
		t.Fatal(err)
	}
	worker := &Worker{DB: db}
	execution := &models.TaskExecution{TaskID: task.ID, State: "running", Started: 1794700000123}
	if err = worker.startExecution(task, execution); err != nil {
		t.Fatal(err)
	}
	var stored models.Task
	if _, err = db.ID(task.ID).Get(&stored); err != nil {
		t.Fatal(err)
	}
	if stored.Started != 1794700000 || task.Started != stored.Started || stored.Finished != 0 || task.Finished != 0 {
		t.Fatal("retry retained old timestamps", stored.Started, stored.Finished)
	}
}
