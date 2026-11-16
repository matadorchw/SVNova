// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestAuthorizationAndTaskExclusion(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	owner := &User{Name: "owner", Active: true}
	reader := &User{Name: "reader", Active: true}
	outsider := &User{Name: "outsider", Active: true}
	if _, err = db.Insert(owner, reader, outsider); err != nil {
		t.Fatal(err)
	}
	org := &Organization{Name: "test"}
	_, _ = db.Insert(org)
	_, _ = db.Insert(&Membership{OrgID: org.ID, UserID: owner.ID, Owner: true}, &Membership{OrgID: org.ID, UserID: reader.ID})
	owners := &Team{OrgID: org.ID, Name: OwnerTeamName, IsOwnerTeam: true}
	_, _ = db.Insert(owners)
	_, _ = db.Insert(&TeamMember{TeamID: owners.ID, UserID: owner.ID})
	repo := &Repository{OrgID: org.ID, Name: "project", Status: "ready", Kind: "mirror"}
	_, _ = db.Insert(repo)
	if !CanRead(db, owner, repo) || CanRead(db, reader, repo) || CanRead(db, outsider, repo) {
		t.Fatal("unexpected initial permissions")
	}
	team := &Team{OrgID: org.ID, Name: "readers"}
	_, _ = db.Insert(team)
	_, _ = db.Insert(&TeamMember{TeamID: team.ID, UserID: reader.ID})
	_, _ = db.Insert(&TeamRepository{TeamID: team.ID, RepositoryID: repo.ID})
	if !CanRead(db, reader, repo) {
		t.Fatal("team grant not applied")
	}
	reader.Active = false
	if CanRead(db, reader, repo) {
		t.Fatal("inactive user allowed")
	}
	reader.Active = true
	_, _ = db.Where("team_id=?", team.ID).Delete(new(TeamRepository))
	if CanRead(db, reader, repo) {
		t.Fatal("revoked grant still active")
	}
	var wg sync.WaitGroup
	success := make(chan bool, 8)
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); success <- Enqueue(db, repo.ID, "sync", "") == nil }()
	}
	wg.Wait()
	close(success)
	n := 0
	for ok := range success {
		if ok {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("expected one queued task, got %d", n)
	}
}
func TestLDAPColumnNames(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	u := &User{Name: "ldap", LDAPAccount: true, LDAPKey: "stable", Active: true}
	_, err = db.Insert(u)
	if err != nil {
		t.Fatal(err)
	}
	var found User
	ok, err := db.Where("ldap_account=1 AND ldap_key=?", "stable").Get(&found)
	if err != nil || !ok {
		t.Fatalf("LDAP lookup failed: %v", err)
	}
}

func TestOwnersMigrationPreservesAuthority(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	owner := &User{Name: "owner", Active: true}
	reader := &User{Name: "reader", Active: true}
	_, _ = db.Insert(owner, reader)
	org := &Organization{Name: "legacy"}
	_, _ = db.Insert(org)
	_, _ = db.Insert(&Membership{OrgID: org.ID, UserID: owner.ID, Owner: true}, &Membership{OrgID: org.ID, UserID: reader.ID})
	ordinary := &Team{OrgID: org.ID, Name: "Owners"}
	_, _ = db.Insert(ordinary)
	_, _ = db.Insert(&TeamMember{TeamID: ordinary.ID, UserID: reader.ID})
	_, _ = db.Where("id>=2").Delete(new(Migration))
	db.Close()
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if !Owner(db, owner, org.ID) || Owner(db, reader, org.ID) {
		t.Fatal("migration changed authority")
	}
	var migrated Team
	_, _ = db.ID(ordinary.ID).Get(&migrated)
	if migrated.IsOwnerTeam || migrated.Name == "Owners" {
		t.Fatal("ordinary team was elevated")
	}
	db.Close()
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	count, err := db.Where("org_id=? AND is_owner_team=?", org.ID, true).Count(new(Team))
	if err != nil || count != 1 {
		t.Fatal("migration not idempotent")
	}
}
