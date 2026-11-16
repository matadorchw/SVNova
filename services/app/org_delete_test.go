// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"svnova/models"
	"testing"
)

func TestDeleteOrganizationGuardsAndCleanup(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "delete.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	owner := &models.User{ID: 1, Name: "owner", Active: true}
	member := &models.User{ID: 2, Name: "member", Active: true}
	_, err = db.Insert(owner, member, &models.Organization{ID: 1, Name: "org"}, &models.Membership{OrgID: 1, UserID: 1}, &models.Membership{OrgID: 1, UserID: 2}, &models.Team{ID: 1, OrgID: 1, Name: "Owners", IsOwnerTeam: true}, &models.TeamMember{TeamID: 1, UserID: 1}, &models.Repository{ID: 1, OrgID: 1, Name: "repo"})
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	call := func(u *models.User, name string) error {
		return a.organizations(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(`{"Name":"`+name+`"}`)), u, []string{"1", "delete"})
	}
	if call(member, "org") == nil {
		t.Fatal("member deleted organization")
	}
	if call(owner, "wrong") == nil {
		t.Fatal("wrong name accepted")
	}
	if err := call(owner, "org"); err == nil || !strings.Contains(err.Error(), "仍有仓库") {
		t.Fatal("repository guard missing", err)
	}
	if _, err := db.ID(1).Delete(new(models.Repository)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Where("team_id=?", 1).Delete(new(models.TeamMember)); err != nil {
		t.Fatal(err)
	}
	if call(owner, "org") == nil {
		t.Fatal("revoked owner deleted organization")
	}
	if _, err := db.Insert(&models.TeamMember{TeamID: 1, UserID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := call(owner, "org"); err != nil {
		t.Fatal(err)
	}
	for _, model := range []any{new(models.Organization), new(models.Team), new(models.TeamMember), new(models.Membership)} {
		count, err := db.Count(model)
		if err != nil || count != 0 {
			t.Fatal("associated records retained", count, err)
		}
	}
	if call(owner, "org") == nil {
		t.Fatal("deleted organization accepted")
	}
}

func TestDeletedOrganizationRejectsAdministratorWriter(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "deleted-writer.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Active: true, Admin: true}
	_, err = db.Insert(admin, &models.Organization{ID: 1, Name: "org"})
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	if err := a.deleteOrganization(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(`{"Name":"org"}`)), admin, 1); err != nil {
		t.Fatal(err)
	}
	tx := db.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec("UPDATE organization SET name=name WHERE id=?", 1); err != nil {
		t.Fatal(err)
	}
	if err := checkOrganizationOwner(tx, admin, 1); err == nil || !strings.Contains(err.Error(), "已被删除") {
		t.Fatal("administrator writer accepted deleted organization", err)
	}
}
