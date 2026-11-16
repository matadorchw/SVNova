// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"encoding/json"
	"errors"
	"github.com/go-ldap/ldap/v3"
	"path/filepath"
	"reflect"
	"svnova/models"
	"testing"
)

func TestSyncSkipsLocalAccountWithoutChangingAdmin(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", DisplayName: "Local Admin", Email: "local@example.com", Password: "existing-hash", Salt: "existing-salt", Admin: true, Active: true}
	session := &models.Session{Token: "existing-session", UserID: 1, CSRF: "existing-csrf", Expires: 1234}
	if _, err = db.Insert(admin, session, &models.Organization{ID: 1, Name: "org"}, &models.Membership{OrgID: 1, UserID: 1}, &models.Team{ID: 1, OrgID: 1, Name: "Owners", IsOwnerTeam: true}, &models.TeamMember{TeamID: 1, UserID: 1}); err != nil {
		t.Fatal(err)
	}
	source := &models.LDAPConfig{ID: 1, UsernameAttribute: "uid", KeyAttribute: "entryUUID", NameAttribute: "cn", EmailAttribute: "mail"}
	entry := func(name, key, email string) *ldap.Entry {
		return ldap.NewEntry("uid="+name, map[string][]string{"uid": {name}, "entryUUID": {key}, "mail": {email}, "cn": {"Directory " + name}})
	}
	collision := entry("admin", "", "invalid-email")
	d := &Directory{DB: db}
	entries := []*ldap.Entry{entry("alice", "alice-key", "alice@example.com"), collision, entry("bob", "bob-key", "bob@example.com")}
	if err = d.syncEntries(source, entries); err != nil {
		t.Fatal(err)
	}
	var report SyncReport
	if err = json.Unmarshal([]byte(source.LastSyncReport), &report); err != nil {
		t.Fatal(err)
	}
	if report.MatchedUsers != 3 || report.SyncedUsers != 2 || report.SkippedLocalCount != 1 || len(report.SkippedLocalUsers) != 1 || report.SkippedLocalUsers[0] != "admin" {
		t.Fatalf("unexpected report: %+v", report)
	}
	var after models.User
	if _, err = db.ID(1).Get(&after); err != nil || !reflect.DeepEqual(*admin, after) {
		t.Fatal("local administrator was modified")
	}
	var afterSession models.Session
	if _, err = db.ID(session.Token).Get(&afterSession); err != nil || !reflect.DeepEqual(*session, afterSession) {
		t.Fatal("local administrator session changed")
	}
	count, err := db.Where("ldap_account=1 AND active=0 AND admin=0 AND email<>''").Count(new(models.User))
	if err != nil || count != 2 {
		t.Fatal("other users did not synchronize", err)
	}
	if count, err = db.Count(new(models.TeamMember)); err != nil || count != 1 {
		t.Fatal("sync changed team permissions", err)
	}
	_, err = d.upsert(source, entry("admin", "ldap-admin", "remote@example.com"))
	var conflict *LocalAccountConflict
	if !errors.As(err, &conflict) {
		t.Fatal("login could take over local account", err)
	}
	if err = d.syncEntries(source, entries); err != nil {
		t.Fatal("repeated sync failed", err)
	}
	if count, err = db.Count(new(models.User)); err != nil || count != 3 {
		t.Fatal("repeated sync duplicated users", err)
	}
}
