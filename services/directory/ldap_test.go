// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"github.com/go-ldap/ldap/v3"
	"path/filepath"
	"svnova/models"
	"testing"
)

func TestStableIdentityAndManualDisable(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	d := &Directory{DB: db}
	s := &models.LDAPConfig{ID: 1, UsernameAttribute: "uid", KeyAttribute: "entryUUID", NameAttribute: "cn"}
	entry := ldap.NewEntry("uid=user,dc=example", map[string][]string{"uid": {"user"}, "entryUUID": {"stable"}, "cn": {"User"}})
	u, err := d.upsert(s, entry)
	if err != nil {
		t.Fatal(err)
	}
	if u.Admin || u.Active || !u.LDAPAccount {
		t.Fatalf("unexpected provision: %+v", u)
	}
	if _, err = d.upsert(s, entry); err == nil {
		t.Fatal("pending LDAP account was automatically enabled")
	}
	if _, err = db.ID(u.ID).Cols("active").Update(&models.User{Active: true}); err != nil {
		t.Fatal(err)
	}
	if enabled, err := d.upsert(s, entry); err != nil || !enabled.Active {
		t.Fatal("administrator-enabled account rejected", err)
	}
	u.Active = false
	_, _ = db.ID(u.ID).Cols("active").Update(u)
	if _, err = d.upsert(s, entry); err == nil {
		t.Fatal("manually disabled account restored")
	}
	u.LDAPDisabled = true
	_, _ = db.ID(u.ID).Cols("ldap_disabled").Update(u)
	restored, err := d.upsert(s, entry)
	if err != nil || !restored.Active || restored.LDAPDisabled || restored.ID != u.ID {
		t.Fatalf("sync-disabled account not restored: %+v %v", restored, err)
	}
	changed := ldap.NewEntry("uid=renamed,dc=example", map[string][]string{"uid": {"renamed"}, "entryUUID": {"stable"}, "cn": {"Renamed"}})
	renamed, err := d.upsert(s, changed)
	if err != nil || renamed.ID != u.ID {
		t.Fatal("stable identity not preserved")
	}
	if !disabled(ldap.NewEntry("x", map[string][]string{"userAccountControl": {"546"}})) {
		t.Fatal("disabled bit ignored")
	}
}
