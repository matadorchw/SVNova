// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"github.com/go-ldap/ldap/v3"
	"path/filepath"
	"svnova/models"
	"testing"
)

func TestLDAPNotificationActualChangesAndRollback(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	source := &models.LDAPConfig{ID: 1, Enabled: true, SyncEnabled: true, UsernameAttribute: "uid", KeyAttribute: "entryUUID", NameAttribute: "cn", EmailAttribute: "mail"}
	_, err = db.Insert(source, &models.NotificationConfig{ID: 1, Enabled: true, LDAPEnabled: true, Generation: 1})
	if err != nil {
		t.Fatal(err)
	}
	name := "Alice"
	d := &Directory{DB: db, readUsers: func(*models.LDAPConfig) ([]*ldap.Entry, error) {
		return []*ldap.Entry{ldap.NewEntry("uid=alice", map[string][]string{"uid": {"alice"}, "entryUUID": {"a"}, "cn": {name}, "mail": {"alice@example.com"}})}, nil
	}}
	if err = d.Sync(); err != nil {
		t.Fatal(err)
	}
	if err = d.Sync(); err != nil {
		t.Fatal(err)
	}
	var run models.LDAPSyncRun
	if n, err := db.Where("ldap_account=1 AND active=0 AND ldap_disabled=0").Count(new(models.User)); err != nil || n != 1 {
		t.Fatal("new LDAP user was enabled by synchronization", err)
	}
	_, _ = db.Desc("id").Get(&run)
	if run.UpdatedUsers != 0 {
		t.Fatal("unchanged attributes counted as updates")
	}
	if n, _ := db.Count(new(models.NotificationDelivery)); n != 1 {
		t.Fatal("unchanged sync notified")
	}
	if _, err = db.Where("ldap_account=1").Cols("active").Update(&models.User{Active: true}); err != nil {
		t.Fatal(err)
	}
	name = "Changed"
	if err = d.SyncAs(0, ""); err != nil {
		t.Fatal(err)
	}
	run = models.LDAPSyncRun{}
	_, _ = db.Desc("id").Get(&run)
	if run.UpdatedUsers != 1 {
		t.Fatal("actual profile change omitted")
	}
	if _, err = db.Exec("CREATE TRIGGER reject_ldap_notification BEFORE INSERT ON notification_delivery BEGIN SELECT RAISE(ABORT,'fixture failure'); END"); err != nil {
		t.Fatal(err)
	}
	name = "Must rollback"
	if err = d.Sync(); err == nil {
		t.Fatal("outbox failure ignored")
	}
	var user models.User
	_, _ = db.Where("name=?", "alice").Get(&user)
	if user.DisplayName != "Changed" {
		t.Fatal("LDAP mutation survived failed outbox commit")
	}
}
