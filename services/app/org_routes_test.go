// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"testing"
)

func TestGiteaRoutesAndPrivateOrganizationProfile(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "pages.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	owner := &models.User{ID: 1, Name: "owner", Active: true}
	viewer := &models.User{ID: 2, Name: "viewer", Active: true}
	_, _ = db.Insert(owner, viewer)
	_, _ = db.Insert(&models.Organization{ID: 1, Name: "acs"}, &models.Organization{ID: 2, Name: "private-other"})
	_, _ = db.Insert(&models.Membership{OrgID: 1, UserID: 1}, &models.Membership{OrgID: 1, UserID: 2})
	_, _ = db.Insert(&models.Team{ID: 1, OrgID: 1, Name: "Owners", IsOwnerTeam: true}, &models.Team{ID: 2, OrgID: 1, Name: "readers"})
	_, _ = db.Insert(&models.TeamMember{TeamID: 1, UserID: 1}, &models.TeamMember{TeamID: 2, UserID: 2})
	_, _ = db.Insert(&models.Repository{ID: 1, OrgID: 1, Name: "visible", Status: "pending"}, &models.Repository{ID: 2, OrgID: 1, Name: "hidden-secret", Status: "pending"}, &models.TeamRepository{TeamID: 2, RepositoryID: 1})
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Templates: templates}
	for _, route := range []string{"/acs", "/org/acs/members", "/org/acs/teams", "/org/acs/teams/readers", "/org/acs/dashboard", "/acs/visible"} {
		w := httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", route, nil), viewer, "csrf")
		if w.Code != 200 {
			t.Fatalf("%s: %d %s", route, w.Code, w.Body.String())
		}
		if strings.Contains(w.Body.String(), "hidden-secret") {
			t.Fatalf("repository leaked through %s", route)
		}
	}
	for _, route := range []string{"/org/acs/settings", "/private-other", "/acs/hidden-secret", "/-/admin/orgs"} {
		w := httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", route, nil), viewer, "csrf")
		if w.Code != 403 {
			t.Fatalf("%s: expected 403, got %d", route, w.Code)
		}
	}
	for route, target := range map[string]string{"/migrate": "/repo/migrate", "/organizations?org=1": "/acs", "/repositories/1?tab=settings": "/acs/visible/settings"} {
		w := httptest.NewRecorder()
		if !a.redirectLegacy(w, httptest.NewRequest("GET", route, nil), owner) || w.Code != 303 || w.Header().Get("Location") != target {
			t.Fatalf("bad redirect %s: %d %s", route, w.Code, w.Header().Get("Location"))
		}
	}
	w := httptest.NewRecorder()
	a.redirectLegacy(w, httptest.NewRequest("GET", "/organizations?org=2", nil), viewer)
	if w.Code != 403 || w.Header().Get("Location") != "" {
		t.Fatal("legacy redirect leaks private organization")
	}
	save := func(u *models.User, body string) error {
		return a.organizations(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(body)), u, []string{"1", "settings"})
	}
	if err = save(viewer, `{"FullName":"forbidden"}`); err == nil {
		t.Fatal("non-owner updated profile")
	}
	if err = save(owner, `{"FullName":"AC Service","Description":"<script>unsafe</script>","Website":"https://example.com","Email":"acs@example.com","Location":"上海"}`); err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/acs", nil), owner, "csrf")
	if strings.Contains(w.Body.String(), "<script>unsafe") || !strings.Contains(w.Body.String(), "AC Service") {
		t.Fatal("profile not safely rendered")
	}
	for _, body := range []string{`{"Website":"javascript:alert(1)"}`, `{"Email":"invalid"}`, `{"Name":"renamed"}`} {
		if err = save(owner, body); err == nil {
			t.Fatalf("accepted invalid profile %s", body)
		}
	}
}
