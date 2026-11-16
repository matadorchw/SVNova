// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

import (
	"path/filepath"
	"testing"
)

func TestNotificationAndScheduleUpgradePreservesExistingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Insert(&User{ID: 1, Name: "admin", Password: "original-hash", Salt: "original-salt", Admin: true, Active: true}, &LDAPConfig{ID: 1, Enabled: true, Password: "encrypted-original", SyncEnabled: true}, &Task{ID: 1, Kind: "mirror", State: "queued"}); err != nil {
		t.Fatal(err)
	}
	if err = db.DropTables(new(NotificationConfig), new(NotificationDelivery), new(NotificationAttempt), new(TaskExecution), new(LDAPSchedule)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("ALTER TABLE task DROP COLUMN trigger"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Where("id>=10").Delete(new(Migration)); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var user User
	_, _ = db.ID(1).Get(&user)
	if user.Password != "original-hash" || user.Salt != "original-salt" || !user.Admin || !user.Active {
		t.Fatal("upgrade changed local administrator")
	}
	var source LDAPConfig
	_, _ = db.ID(1).Get(&source)
	if source.Password != "encrypted-original" || !source.Enabled {
		t.Fatal("upgrade changed LDAP credential or activation")
	}
	var task Task
	_, _ = db.ID(1).Get(&task)
	if task.State != "queued" || task.Trigger != "" {
		t.Fatal("upgrade changed pending task")
	}
	if n, _ := db.Count(new(NotificationConfig)); n != 0 {
		t.Fatal("upgrade enabled or configured notifications")
	}
	if n, _ := db.Count(new(NotificationDelivery)); n != 0 {
		t.Fatal("upgrade backfilled notifications")
	}
	if _, err = db.Insert(&Migration{ID: 16}); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if future, err := Open(path); err == nil {
		future.Close()
		t.Fatal("newer schema accepted")
	}
}
