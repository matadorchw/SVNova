// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"testing"
)

func TestTeamMembershipEnrollmentAndLastTeamRevocation(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "teams.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Admin: true, Active: true}
	reader := &models.User{ID: 2, Name: "reader", DisplayName: "读者", Active: true}
	inactive := &models.User{ID: 3, Name: "disabled"}
	_, _ = db.Insert(admin, reader, inactive)
	_, _ = db.Insert(&models.Organization{ID: 1, Name: "acs"}, &models.Membership{OrgID: 1, UserID: 1}, &models.Team{ID: 1, OrgID: 1, Name: "Owners", IsOwnerTeam: true}, &models.TeamMember{TeamID: 1, UserID: 1}, &models.Team{ID: 2, OrgID: 1, Name: "readers", Description: "只读"}, &models.Team{ID: 3, OrgID: 1, Name: "another"})
	repo := &models.Repository{ID: 1, OrgID: 1, Name: "visible", Status: "pending"}
	_, _ = db.Insert(repo, &models.Repository{ID: 2, OrgID: 1, Name: "hidden-secret", Status: "pending"}, &models.TeamRepository{TeamID: 2, RepositoryID: 1})
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Templates: templates}
	call := func(actor *models.User, team, action, body string) error {
		return a.teams(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(body)), actor, []string{team, action})
	}
	if err = call(admin, "2", "members", `{"Username":"reader"}`); err != nil {
		t.Fatal(err)
	}
	if !models.CanRead(db, reader, repo) {
		t.Fatal("team enrollment did not enroll organization and grant access")
	}
	if err = call(reader, "2", "members", `{"UserID":3}`); err == nil {
		t.Fatal("ordinary member changed team membership")
	}
	for route, status := range map[string]int{"/org/acs/teams/new": 403, "/org/acs/teams/readers": 200, "/org/acs/teams/readers/repositories": 200} {
		w := httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", route, nil), reader, "csrf")
		if w.Code != status {
			t.Fatalf("%s: %d", route, w.Code)
		}
		if strings.Contains(w.Body.String(), "hidden-secret") {
			t.Fatal("team page leaked unauthorized repository")
		}
	}
	p := &webPage{User: admin, Page: "organizations", Search: "读者"}
	if err = a.webData(httptest.NewRequest("GET", "/organizations?org=1&org_view=members", nil), p); err != nil {
		t.Fatal(err)
	}
	if len(p.FilteredMembers) != 1 || p.FilteredMembers[0].UserID != reader.ID || len(p.Members) != 2 {
		t.Fatal("member search changed totals or did not match display name")
	}
	if err = call(admin, "3", "members", `{"UserID":2}`); err != nil {
		t.Fatal(err)
	}
	if err = call(reader, "2", "leave", `{}`); err != nil {
		t.Fatal(err)
	}
	member, _ := db.Where("org_id=1 AND user_id=2").Exist(new(models.Membership))
	if !member {
		t.Fatal("leaving one of multiple teams removed organization membership")
	}
	if models.CanRead(db, reader, repo) {
		t.Fatal("team access was not revoked")
	}
	if err = call(reader, "3", "leave", `{}`); err != nil {
		t.Fatal(err)
	}
	member, _ = db.Where("org_id=1 AND user_id=2").Exist(new(models.Membership))
	if member {
		t.Fatal("leaving the last team retained organization membership")
	}
	if err = call(admin, "1", "leave", `{}`); err == nil {
		t.Fatal("last owner left the team")
	}
	if err = call(admin, "2", "members", `{"UserID":3}`); err == nil {
		t.Fatal("inactive user enrolled")
	}
	member, _ = db.Where("org_id=1 AND user_id=3").Exist(new(models.Membership))
	if member {
		t.Fatal("failed enrollment left orphan membership")
	}
	for _, body := range []string{`{"Name":"sub admin"}`, `{"Name":"new"}`, `{"Name":"READERS"}`, fmt.Sprintf(`{"Name":"oversized","Description":"%s"}`, strings.Repeat("中", 256))} {
		if err = a.organizations(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(body)), admin, []string{"1", "teams"}); err == nil {
			t.Fatalf("invalid team accepted: %s", body)
		}
		if failure, ok := err.(*httpError); !ok || failure.Message == "" || (failure.Status != 400 && failure.Status != 409) {
			t.Fatalf("team validation did not explain failure: %v", err)
		}
	}
	w := httptest.NewRecorder()
	if err = a.organizations(w, httptest.NewRequest("POST", "/", strings.NewReader(`{"Name":"ops","Description":"运维组"}`)), admin, []string{"1", "teams"}); err != nil {
		t.Fatal(err)
	}
	p = &webPage{User: admin, Page: "organizations", Search: "运维"}
	if err = a.webData(httptest.NewRequest("GET", "/organizations?org=1&org_view=teams", nil), p); err != nil {
		t.Fatal(err)
	}
	if len(p.FilteredTeams) != 1 || p.FilteredTeams[0].Name != "ops" {
		t.Fatal("team description search failed")
	}
}

func TestOrganizationSelfLeaveProtectsOwnerAndRevokesTeams(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "leave.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	owner := &models.User{ID: 1, Name: "owner", Active: true}
	member := &models.User{ID: 2, Name: "member", Active: true}
	_, err = db.Insert(owner, member, &models.Organization{ID: 1, Name: "org"}, &models.Membership{OrgID: 1, UserID: 1}, &models.Membership{OrgID: 1, UserID: 2}, &models.Team{ID: 1, OrgID: 1, Name: "Owners", IsOwnerTeam: true}, &models.Team{ID: 2, OrgID: 1, Name: "readers"}, &models.TeamMember{TeamID: 1, UserID: 1}, &models.TeamMember{TeamID: 2, UserID: 2})
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	call := func(u *models.User, body string) error {
		return a.organizations(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(body)), u, []string{"1", "leave"})
	}
	if err := call(owner, `{}`); err == nil || !strings.Contains(err.Error(), "最后一位 Owners") {
		t.Fatal("last owner left")
	}
	if err := call(member, `{"UserID":1,"Remove":false}`); err != nil {
		t.Fatal(err)
	}
	exists, err := db.Where("org_id=? AND user_id=?", 1, 2).Exist(new(models.Membership))
	if err != nil || exists {
		t.Fatal("membership retained", err)
	}
	exists, err = db.Where("team_id=? AND user_id=?", 2, 2).Exist(new(models.TeamMember))
	if err != nil || exists {
		t.Fatal("team membership retained", err)
	}
	exists, err = db.Where("team_id=? AND user_id=?", 1, 1).Exist(new(models.TeamMember))
	if err != nil || !exists {
		t.Fatal("forged target affected owner", err)
	}
	if call(member, `{}`) == nil {
		t.Fatal("non-member leave accepted")
	}
}
