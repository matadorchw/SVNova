// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"path/filepath"
	"svnova/models"
	"testing"
)

func TestRestartDoesNotAutomaticallyRetryConversion(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	conversion := &models.Task{Kind: "conversion", State: "running", Started: 1}
	local := &models.Task{Kind: "local", State: "running"}
	if _, err = db.Insert(conversion, local); err != nil {
		t.Fatal(err)
	}
	execution := &models.TaskExecution{TaskID: conversion.ID, State: "running"}
	if _, err = db.Insert(execution); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	(&Worker{DB: db}).Run(ctx)
	var stored models.Task
	if _, err = db.ID(conversion.ID).Get(&stored); err != nil {
		t.Fatal(err)
	}
	if stored.State != "failed" || stored.Finished == 0 {
		t.Fatal("conversion silently requeued")
	}
	stored = models.Task{}
	if _, err = db.ID(local.ID).Get(&stored); err != nil {
		t.Fatal(err)
	}
	if stored.State != "queued" {
		t.Fatal("local migration recovery changed")
	}
	var attempt models.TaskExecution
	if _, err = db.ID(execution.ID).Get(&attempt); err != nil {
		t.Fatal(err)
	}
	if attempt.State != "interrupted" {
		t.Fatal("execution not interrupted")
	}
}
