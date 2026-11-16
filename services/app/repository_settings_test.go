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

func TestRepositorySettingsSections(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Admin: true, Active: true}
	for _, v := range []any{admin, &models.Organization{ID: 1, Name: "acs"}, &models.Repository{ID: 1, OrgID: 1, Name: "production", Kind: "local", Status: "ready"}} {
		if _, err = db.Insert(v); err != nil {
			t.Fatal(err)
		}
	}
	a := &App{DB: db}
	a.Templates, err = ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ path, want, absent string }{{"/acs/production/settings", `id="repo-name"`, `id="repo-confirm"`}, {"/acs/production/settings/mirror", `id="convert-remote"`, `id="repo-name"`}, {"/acs/production/settings/danger", `id="repo-confirm"`, `id="convert-remote"`}, {"/repositories/1?tab=settings&section=mirror", `id="convert-remote"`, `id="repo-confirm"`}} {
		w := httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", tc.path, nil), admin, "csrf")
		if w.Code != 200 || !strings.Contains(w.Body.String(), tc.want) || strings.Contains(w.Body.String(), tc.absent) {
			t.Fatal(tc.path, w.Code, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/acs/production/settings/unknown", nil), admin, "csrf")
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
	reader := &models.User{ID: 2, Active: true}
	for _, path := range []string{"/acs/production/settings", "/acs/production/settings/mirror", "/acs/production/settings/danger"} {
		w = httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", path, nil), reader, "csrf")
		if w.Code != 403 {
			t.Fatal(path, w.Code)
		}
	}
}
