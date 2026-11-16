// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"svnova/models"
	"testing"
)

func TestSiteNameValidationAndPersistence(t *testing.T) {
	for _, name := range []string{"", "  ", strings.Repeat("中", 81), "bad\nname"} {
		if _, err := normalizeSiteName(name); err == nil {
			t.Errorf("accepted %q", name)
		}
	}
	path := filepath.Join(t.TempDir(), "app.db")
	db, err := models.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	if a.siteName() != "SVNova" {
		t.Fatal("legacy default changed")
	}
	request := func(admin bool, name string) error {
		return a.siteSettings(httptest.NewRecorder(), httptest.NewRequest("POST", "/api/v1/admin/site", strings.NewReader(`{"SiteName":"`+name+`"}`)), &models.User{Admin: admin})
	}
	if err = request(false, "forbidden"); err == nil {
		t.Fatal("non-admin changed site name")
	}
	if err = request(true, "  中文站点  "); err != nil {
		t.Fatal(err)
	}
	if a.siteName() != "中文站点" {
		t.Fatal("name not updated")
	}
	db.Close()
	db, err = models.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	a.DB = db
	if a.siteName() != "中文站点" {
		t.Fatal("name not persisted")
	}
}
func TestOrganizationCreationDefaultsToCreator(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	u := &models.User{Name: "creator", Active: true, Admin: true}
	if _, err = db.Insert(u); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	w := httptest.NewRecorder()
	if err = a.organizations(w, httptest.NewRequest("POST", "/api/v1/organizations", strings.NewReader(`{"Name":"new-org"}`)), u, nil); err != nil {
		t.Fatal(err)
	}
	var org models.Organization
	ok, err := db.Where("name=?", "new-org").Get(&org)
	if err != nil || !ok || org.Created == 0 {
		t.Fatal("organization timestamp missing")
	}
	if !models.Owner(db, u, org.ID) {
		t.Fatal("creator is not owner")
	}
}

func TestOwnersTeamMembershipProtectsLastOwner(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "owners.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{Name: "admin", Active: true, Admin: true}
	second := &models.User{Name: "second", Active: true}
	_, _ = db.Insert(admin, second)
	a := &App{DB: db}
	w := httptest.NewRecorder()
	if err = a.organizations(w, httptest.NewRequest("POST", "/api/v1/organizations", strings.NewReader(`{"Name":"owned"}`)), admin, nil); err != nil {
		t.Fatal(err)
	}
	var org models.Organization
	_, _ = db.Where("name=?", "owned").Get(&org)
	var team models.Team
	_, _ = db.Where("org_id=? AND is_owner_team=?", org.ID, true).Get(&team)
	call := func(remove bool, user int64) error {
		body := fmt.Sprintf(`{"UserID":%d,"Remove":%t}`, user, remove)
		return a.teams(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(body)), admin, []string{strconv.FormatInt(team.ID, 10), "members"})
	}
	if err = call(true, admin.ID); err == nil {
		t.Fatal("last owner removed")
	}
	_, _ = db.Insert(&models.Membership{OrgID: org.ID, UserID: second.ID})
	if err = call(false, second.ID); err != nil {
		t.Fatal(err)
	}
	if !models.Owner(db, second, org.ID) {
		t.Fatal("Owners membership does not grant authority")
	}
	if err = a.teams(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(fmt.Sprintf(`{"UserID":%d}`, admin.ID))), second, []string{strconv.FormatInt(team.ID, 10), "members"}); err != nil {
		t.Fatalf("non-admin Owners member cannot manage team: %v", err)
	}
	if err = call(true, second.ID); err != nil {
		t.Fatal(err)
	}
	if models.Owner(db, second, org.ID) {
		t.Fatal("removed member retains authority")
	}
	if err = a.organizations(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(`{"Name":"owners"}`)), admin, []string{strconv.FormatInt(org.ID, 10), "teams"}); err == nil {
		t.Fatal("reserved name accepted")
	}
	if err = a.teams(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(`{"RepositoryID":1}`)), admin, []string{strconv.FormatInt(team.ID, 10), "repositories"}); err == nil {
		t.Fatal("Owners repository grants mutated")
	}
}

func TestConcurrentOwnersRemovalPreservesOne(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "owners.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Active: true, Admin: true}
	_, _ = db.Insert(admin, &models.User{ID: 2, Name: "second", Active: true})
	_, _ = db.Insert(&models.Organization{ID: 1, Name: "org"}, &models.Membership{OrgID: 1, UserID: 1}, &models.Membership{OrgID: 1, UserID: 2})
	team := &models.Team{OrgID: 1, Name: models.OwnerTeamName, IsOwnerTeam: true}
	_, _ = db.Insert(team)
	_, _ = db.Insert(&models.TeamMember{TeamID: team.ID, UserID: 1}, &models.TeamMember{TeamID: team.ID, UserID: 2})
	a := &App{DB: db}
	results := make(chan error, 2)
	for _, id := range []int64{1, 2} {
		go func(id int64) {
			results <- a.teams(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(fmt.Sprintf(`{"UserID":%d,"Remove":true}`, id))), admin, []string{strconv.FormatInt(team.ID, 10), "members"})
		}(id)
	}
	successes := 0
	for range 2 {
		if <-results == nil {
			successes++
		}
	}
	count, err := db.Where("team_id=?", team.ID).Count(new(models.TeamMember))
	if err != nil || count != 1 || successes != 1 {
		t.Fatalf("count=%d successes=%d err=%v", count, successes, err)
	}
}
