// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

import (
	"path/filepath"
	"testing"
)

func TestRemoveNotificationExclusionsPreservesSettingsAndHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	original := &NotificationConfig{ID: 1, Enabled: true, SVNEnabled: true, LDAPEnabled: true, SVNFailureOnly: true, Webhook: "original-encrypted-webhook", Secret: "original-encrypted-secret", Generation: 4, RetentionDays: 90}
	if _, err := db.Insert(original, &NotificationDelivery{EventKey: "old-cancelled", Kind: "svn", Status: "cancelled"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("ALTER TABLE notification_config ADD COLUMN excluded_repositories TEXT"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE notification_config SET excluded_repositories='[1,2]'"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ID(14).Delete(new(Migration)); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	columns, err := db.QueryString("PRAGMA table_info(notification_config)")
	if err != nil {
		t.Fatal(err)
	}
	for _, column := range columns {
		if column["name"] == "excluded_repositories" {
			t.Fatal("legacy column retained")
		}
	}
	var stored NotificationConfig
	if _, err := db.ID(1).Get(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != *original {
		t.Fatal("migration changed unrelated settings")
	}
	var delivery NotificationDelivery
	if _, err := db.Where("event_key=?", "old-cancelled").Get(&delivery); err != nil {
		t.Fatal(err)
	}
	if delivery.Status != "cancelled" {
		t.Fatal("migration revived cancelled delivery")
	}
	if count, err := db.Count(new(NotificationDelivery)); err != nil || count != 1 {
		t.Fatal("migration backfilled deliveries", err)
	}
	if err := removeNotificationExclusions(db); err != nil {
		t.Fatal("migration not idempotent", err)
	}
}
