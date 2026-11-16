// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"svnova/models"
	"svnova/services/security"
	"testing"
	"time"
)

func TestSourceCleanupNoopExpiryAndActor(t *testing.T) {
	d, source := cleanupFixture(t)
	candidate := *source
	preview, err := d.PreviewCleanup(&candidate, false, "")
	if err != nil || preview.RequiresConfirmation {
		t.Fatal("unchanged source requires cleanup", err)
	}
	if err = d.SaveConfirmed(&candidate, false, "", "", 1); err != nil {
		t.Fatal(err)
	}
	if ok, _ := d.DB.ID(3).Exist(new(models.User)); !ok {
		t.Fatal("unchanged save deleted LDAP user")
	}
	candidate.BaseDN = "dc=changed"
	preview, err = d.PreviewCleanup(&candidate, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err = d.SaveConfirmed(&candidate, false, "", preview.Token, 2); !errors.Is(err, ErrConfigAdmin) {
		t.Fatal("non-administrator cleanup accepted", err)
	}
	raw, err := d.Vault.Open(preview.Token)
	if err != nil {
		t.Fatal(err)
	}
	var token cleanupToken
	if err = json.Unmarshal([]byte(raw), &token); err != nil {
		t.Fatal(err)
	}
	token.Expires = time.Now().Add(-time.Minute).Unix()
	encoded, _ := json.Marshal(token)
	if err = d.SaveConfirmed(&candidate, false, "", d.Vault.Seal(string(encoded)), 1); !errors.Is(err, ErrCleanupConfirmation) {
		t.Fatal("expired confirmation accepted", err)
	}
	candidate.Filter = "(objectClass=changed)"
	if err = d.SaveConfirmed(&candidate, false, "", preview.Token, 1); !errors.Is(err, ErrCleanupConfirmation) {
		t.Fatal("confirmation reused for changed configuration", err)
	}
	if ok, _ := d.DB.ID(3).Exist(new(models.User)); !ok {
		t.Fatal("rejected save removed LDAP user")
	}
}

func cleanupFixture(t *testing.T) (*Directory, *models.LDAPConfig) {
	t.Helper()
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	vault, err := security.NewVault(filepath.Join(root, "key"))
	if err != nil {
		t.Fatal(err)
	}
	d := &Directory{DB: db, Vault: vault}
	source := &models.LDAPConfig{ID: 1, Enabled: true, URL: "ldaps://ldap.example.com", BaseDN: "dc=example", Filter: "(objectClass=person)", UsernameAttribute: "uid", KeyAttribute: "entryUUID"}
	_, err = db.Insert(&models.User{ID: 1, Name: "admin", Admin: true, Active: true, Password: "original", Salt: "salt"}, &models.User{ID: 2, Name: "local", Active: true}, &models.User{ID: 3, Name: "ldap-user", LDAPAccount: true, LDAPKey: "key", Email: "user@example.com", Active: true}, source, &models.Session{Token: "local", UserID: 1}, &models.Session{Token: "ldap", UserID: 3}, &models.Organization{ID: 1, Name: "org"}, &models.Team{ID: 1, OrgID: 1, Name: "Owners", IsOwnerTeam: true}, &models.Membership{OrgID: 1, UserID: 1}, &models.Membership{OrgID: 1, UserID: 3}, &models.TeamMember{TeamID: 1, UserID: 1}, &models.TeamMember{TeamID: 1, UserID: 3}, &models.Repository{ID: 1, OrgID: 1, Name: "repo"}, &models.Task{ID: 1, RepositoryID: 1, State: "succeeded"}, &models.LDAPSyncRun{ID: 1, Status: "success"})
	if err != nil {
		t.Fatal(err)
	}
	return d, source
}

func TestSourceEditCleanupRequiresCurrentConfirmation(t *testing.T) {
	d, source := cleanupFixture(t)
	candidate := *source
	candidate.BaseDN = "dc=renamed"
	var adminBefore models.User
	_, _ = d.DB.ID(1).Get(&adminBefore)
	if err := d.Save(&candidate, false, ""); !errors.Is(err, ErrCleanupConfirmation) {
		t.Fatal("unconfirmed edit accepted", err)
	}
	preview, err := d.PreviewCleanup(&candidate, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !preview.RequiresConfirmation || preview.DeleteUsers != 1 || preview.Sessions != 1 || preview.Memberships != 1 || preview.TeamMemberships != 1 {
		t.Fatalf("wrong cleanup preview: %+v", preview)
	}
	if err = d.SaveConfirmed(&candidate, false, "", "tampered", 1); !errors.Is(err, ErrCleanupConfirmation) {
		t.Fatal("tampered token accepted", err)
	}
	if err = d.SaveConfirmed(&candidate, false, "", preview.Token, 1); err != nil {
		t.Fatal(err)
	}
	var adminAfter models.User
	_, _ = d.DB.ID(1).Get(&adminAfter)
	if !reflect.DeepEqual(adminBefore, adminAfter) {
		t.Fatal("local admin changed")
	}
	if ok, err := d.DB.ID(3).Exist(new(models.User)); err != nil || ok {
		t.Fatal("LDAP account retained", err)
	}
	for _, table := range []any{new(models.Organization), new(models.Team), new(models.Repository), new(models.Task), new(models.LDAPSyncRun), new(models.Membership), new(models.TeamMember), new(models.Session)} {
		if count, err := d.DB.Count(table); err != nil || count != 1 {
			t.Fatal("retained data or local relation changed", err)
		}
	}
	if count, _ := d.DB.Count(new(models.User)); count != 2 {
		t.Fatal("local users were removed")
	}
}

func TestSourceCleanupGuardsAndStalePreview(t *testing.T) {
	d, source := cleanupFixture(t)
	candidate := *source
	candidate.Filter = "(objectClass=posixAccount)"
	preview, err := d.PreviewCleanup(&candidate, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = d.DB.Insert(&models.User{ID: 4, Name: "new-ldap", LDAPAccount: true, LDAPKey: "new", Active: true}); err != nil {
		t.Fatal(err)
	}
	if err = d.SaveConfirmed(&candidate, false, "", preview.Token, 1); !errors.Is(err, ErrCleanupConfirmation) {
		t.Fatal("stale user set accepted", err)
	}
	if _, err = d.DB.Where("team_id=1 AND user_id=1").Delete(new(models.TeamMember)); err != nil {
		t.Fatal(err)
	}
	if _, err = d.PreviewCleanup(&candidate, false, ""); !errors.Is(err, ErrCleanupOwner) {
		t.Fatal("last Owners member removal allowed", err)
	}
	if count, _ := d.DB.Where("ldap_account>0").Count(new(models.User)); count != 2 {
		t.Fatal("guard removed users")
	}
}

func TestLDAPConfigCleanupRollbackKeepsHistory(t *testing.T) {
	d, source := cleanupFixture(t)
	candidate := *source
	candidate.BaseDN = "dc=changed"
	preview, err := d.PreviewCleanup(&candidate, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = d.DB.Exec("CREATE TRIGGER reject_config_update BEFORE UPDATE ON ldap_config BEGIN SELECT RAISE(ABORT,'fixture failure'); END"); err != nil {
		t.Fatal(err)
	}
	if err = d.SaveConfirmed(&candidate, false, "", preview.Token, 1); err == nil {
		t.Fatal("fixture failure accepted")
	}
	if ok, _ := d.DB.ID(3).Exist(new(models.User)); !ok {
		t.Fatal("cleanup did not roll back")
	}
	if count, _ := d.DB.Count(new(models.LDAPSyncRun)); count != 1 {
		t.Fatal("history removed")
	}
}

func TestLDAPCleanupProtectsLastAdministrator(t *testing.T) {
	d, config := cleanupFixture(t)
	if _, err := d.DB.ID(1).Cols("admin").Update(&models.User{Admin: false}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.DB.ID(3).Cols("admin").Update(&models.User{Admin: true}); err != nil {
		t.Fatal(err)
	}
	candidate := *config
	candidate.BaseDN = "dc=changed"
	if _, err := d.PreviewCleanup(&candidate, false, ""); !errors.Is(err, ErrCleanupAdmin) {
		t.Fatal("last LDAP administrator cleanup allowed", err)
	}
	if ok, _ := d.DB.ID(3).Exist(new(models.User)); !ok {
		t.Fatal("administrator removed")
	}
}
