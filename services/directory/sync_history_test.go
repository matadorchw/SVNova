// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"errors"
	"github.com/go-ldap/ldap/v3"
	"path/filepath"
	"svnova/models"
	"testing"
	"time"
)

func TestSyncHistoryLifecycleAndAtomicFailure(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	source := &models.LDAPConfig{ID: 1, Enabled: true, SyncEnabled: true, UsernameAttribute: "uid", KeyAttribute: "entryUUID"}
	if _, err = db.Insert(source, &models.User{Name: "admin", Admin: true, Active: true}); err != nil {
		t.Fatal(err)
	}
	entry := func(name, key string) *ldap.Entry {
		return ldap.NewEntry("uid="+name, map[string][]string{"uid": {name}, "entryUUID": {key}})
	}
	d := &Directory{DB: db, readUsers: func(*models.LDAPConfig) ([]*ldap.Entry, error) {
		return []*ldap.Entry{entry("admin", "admin-key"), entry("alice", "alice")}, nil
	}}
	if err = d.SyncAs(1, "admin"); err != nil {
		t.Fatal(err)
	}
	var run models.LDAPSyncRun
	_, _ = db.Desc("id").Get(&run)
	if run.Status != "success" || run.Trigger != "manual" || run.CreatedUsers != 1 || run.UpdatedUsers != 0 || run.SkippedLocalCount != 1 || run.ActorName != "admin" || run.DurationMS < 0 {
		t.Fatalf("incorrect success history: %+v", run)
	}
	if _, err = db.Where("ldap_account=1").Cols("active").Update(&models.User{Active: true}); err != nil {
		t.Fatal(err)
	}
	d.readUsers = func(*models.LDAPConfig) ([]*ldap.Entry, error) {
		return []*ldap.Entry{entry("bob", "bob"), entry("broken", "")}, nil
	}
	if err = d.Sync(); err == nil {
		t.Fatal("invalid identity accepted")
	}
	var failed models.LDAPSyncRun
	_, _ = db.Desc("id").Get(&failed)
	if failed.Status != "failed" || failed.Trigger != "scheduled" || failed.FailureReason != "identity" || failed.CreatedUsers != 0 || failed.UpdatedUsers != 0 {
		t.Fatalf("incorrect failed history: %+v", failed)
	}
	if exists, _ := db.Where("name=?", "bob").Exist(new(models.User)); exists {
		t.Fatal("partial synchronization was committed")
	}
	if count, _ := db.Where("ldap_account=1 AND active=1").Count(new(models.User)); count != 1 {
		t.Fatal("failed sync disabled existing user")
	}
	d.readUsers = func(*models.LDAPConfig) ([]*ldap.Entry, error) {
		return []*ldap.Entry{entry("duplicate", "same"), entry("again", "same")}, nil
	}
	if err = d.SyncAs(0, ""); err == nil {
		t.Fatal("duplicate stable identity accepted")
	}
	failed = models.LDAPSyncRun{}
	_, _ = db.Desc("id").Get(&failed)
	if failed.FailureReason != "duplicate_identity" || failed.CreatedUsers != 0 {
		t.Fatal("duplicate failure not safe")
	}
}

func TestSyncRunningBusyDisabledAndRestartHistory(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Insert(&models.LDAPConfig{ID: 1, Enabled: true, SyncEnabled: true, AllowDeactivateAll: true}); err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	d := &Directory{DB: db, readUsers: func(*models.LDAPConfig) ([]*ldap.Entry, error) { close(entered); <-release; return nil, nil }}
	go func() { done <- d.SyncAs(0, "") }()
	<-entered
	var running models.LDAPSyncRun
	_, _ = db.Desc("id").Get(&running)
	if running.Status != "running" {
		t.Fatal("running state not visible")
	}
	if err = d.SyncAs(0, ""); !errors.Is(err, ErrSyncBusy) {
		t.Fatal("parallel sync not rejected")
	}
	close(release)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	if count, _ := db.Count(new(models.LDAPSyncRun)); count != 1 {
		t.Fatal("busy request created history")
	}
	if _, err = db.ID(1).Cols("sync_enabled").Update(&models.LDAPConfig{SyncEnabled: false}); err != nil {
		t.Fatal(err)
	}
	if err = d.SyncAs(0, ""); !errors.Is(err, ErrSyncDisabled) {
		t.Fatal("disabled sync accepted")
	}
	if count, _ := db.Count(new(models.LDAPSyncRun)); count != 1 {
		t.Fatal("disabled request created history")
	}
	if _, err = db.Insert(&models.LDAPSyncRun{Status: "running", Started: time.Now().Add(-time.Second).UnixMilli()}); err != nil {
		t.Fatal(err)
	}
	if err = d.RecoverSyncRuns(); err != nil {
		t.Fatal(err)
	}
	var recovered models.LDAPSyncRun
	_, _ = db.Desc("id").Get(&recovered)
	if recovered.Status != "failed" || recovered.FailureReason != "interrupted" || recovered.DurationMS < 1000 {
		t.Fatal("interrupted run not recovered")
	}
	if err = d.RecoverSyncRuns(); err != nil {
		t.Fatal(err)
	}
}
