// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package tasks

import (
	"errors"
	"path/filepath"
	"svnova/models"
	"testing"
	"time"
)

func TestTaskNotificationResultAtomicityAndExecutionIdentity(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Insert(&models.NotificationConfig{ID: 1, Enabled: true, SVNEnabled: true, Generation: 1}, &models.Organization{ID: 1, Name: "org"}, &models.Repository{ID: 1, OrgID: 1, Kind: "mirror", Name: "repo", Status: "pending"}, &models.Task{ID: 1, RepositoryID: 1, Kind: "mirror", Trigger: "migration", State: "queued"})
	if err != nil {
		t.Fatal(err)
	}
	w := &Worker{DB: db}
	task := models.Task{ID: 1, RepositoryID: 1, Kind: "mirror", Trigger: "migration", State: "queued"}
	first := models.TaskExecution{TaskID: 1, Started: time.Now().UnixMilli(), State: "running"}
	if err = w.startExecution(&task, &first); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("CREATE TRIGGER reject_notification BEFORE INSERT ON notification_delivery BEGIN SELECT RAISE(ABORT,'fixture failure'); END"); err != nil {
		t.Fatal(err)
	}
	repo := models.Repository{ID: 1, OrgID: 1, Kind: "mirror", Name: "repo", Status: "ready", Revision: 3}
	task.State = "succeeded"
	first.State = "succeeded"
	first.Finished = first.Started + 1000
	first.AfterRevision = 3
	first.RevisionKnown = true
	if err = w.finish(&repo, &task, &first, nil); err == nil {
		t.Fatal("enqueue failure ignored")
	}
	var saved models.Task
	_, _ = db.ID(1).Get(&saved)
	if saved.State != "running" {
		t.Fatal("task result committed without event")
	}
	var before models.Repository
	_, _ = db.ID(1).Get(&before)
	if before.Status != "pending" {
		t.Fatal("repository result committed without event")
	}
	if _, err = db.Exec("DROP TRIGGER reject_notification"); err != nil {
		t.Fatal(err)
	}
	if err = w.finish(&repo, &task, &first, nil); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ID(1).Cols("state", "trigger").Update(&models.Task{State: "queued", Trigger: "manual"}); err != nil {
		t.Fatal(err)
	}
	task.State = "queued"
	task.Trigger = "manual"
	second := models.TaskExecution{TaskID: 1, Started: time.Now().UnixMilli(), State: "running", BeforeRevision: 3}
	if err = w.startExecution(&task, &second); err != nil {
		t.Fatal(err)
	}
	task.State = "failed"
	second.State = "failed"
	second.Finished = second.Started + 1000
	repo.Status = "ready"
	if err = w.finish(&repo, &task, &second, errors.New("private-token")); err != nil {
		t.Fatal(err)
	}
	var rows []models.NotificationDelivery
	if err = db.Asc("id").Find(&rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].EventKey == rows[1].EventKey || first.ID == second.ID {
		t.Fatal("task retries share event identity")
	}
	if err = w.finish(&repo, &task, &second, errors.New("private-token")); err != nil {
		t.Fatal(err)
	}
	if n, _ := db.Count(new(models.NotificationDelivery)); n != 2 {
		t.Fatal("duplicate completion enqueued")
	}
}
