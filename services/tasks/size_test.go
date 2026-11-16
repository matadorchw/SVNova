// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"os"
	"path/filepath"
	"svnova/models"
	"testing"
)

func TestSizeBackfillAndFailurePreservesSnapshot(t *testing.T) {
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	managed := filepath.Join(root, "managed")
	path := filepath.Join(managed, "repo")
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "file"), []byte("12345"), 0600); err != nil {
		t.Fatal(err)
	}
	bad := &models.Repository{Name: "missing", Path: filepath.Join(managed, "missing"), Status: "ready"}
	good := &models.Repository{Name: "good", Path: path, Status: "ready"}
	if _, err := db.Insert(bad, good); err != nil {
		t.Fatal(err)
	}
	worker := &Worker{DB: db, Root: managed}
	worker.Tick(context.Background())
	worker.Tick(context.Background())
	var measured models.Repository
	if _, err := db.ID(good.ID).Get(&measured); err != nil {
		t.Fatal(err)
	}
	if measured.Size != 5 || measured.SizeUpdated == 0 {
		t.Fatalf("backfill failed: %+v", measured)
	}
	measured.Path = filepath.Join(root, "outside")
	if err := os.Mkdir(measured.Path, 0700); err != nil {
		t.Fatal(err)
	}
	worker.updateSize(context.Background(), &measured)
	var preserved models.Repository
	if _, err := db.ID(good.ID).Get(&preserved); err != nil {
		t.Fatal(err)
	}
	if preserved.Size != 5 || preserved.SizeUpdated != measured.SizeUpdated {
		t.Fatal("measurement failure erased cached size")
	}
	if err := os.WriteFile(filepath.Join(path, "file"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	worker.updateSize(context.Background(), &preserved)
	var empty models.Repository
	if _, err := db.ID(good.ID).Get(&empty); err != nil {
		t.Fatal(err)
	}
	if empty.Size != 0 || empty.SizeUpdated == 0 {
		t.Fatal("zero size not persisted")
	}
	worker.Convert = func(context.Context, *models.Repository, *models.Task) error { return nil }
	task := &models.Task{RepositoryID: bad.ID, Kind: "conversion", State: "queued"}
	if _, err := db.Insert(task); err != nil {
		t.Fatal(err)
	}
	worker.Tick(context.Background())
	var result models.Task
	if _, err := db.ID(task.ID).Get(&result); err != nil {
		t.Fatal(err)
	}
	if result.State != "succeeded" {
		t.Fatal("size failure changed successful task outcome")
	}

}
