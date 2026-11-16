// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/directory"
	"testing"
)

func TestAuthSourceSkippedLocalAccountNotice(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Active: true, Admin: true}
	if _, err = db.Insert(admin); err != nil {
		t.Fatal(err)
	}
	report, _ := json.Marshal(directory.SyncReport{MatchedUsers: 3, SyncedUsers: 2, SkippedLocalCount: 1, SkippedLocalUsers: []string{"admin<script>"}})
	source := &models.LDAPConfig{ID: 1, LastSync: 1, LastSyncReport: string(report)}
	if _, err = db.Insert(source); err != nil {
		t.Fatal(err)
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	(&App{DB: db, Templates: templates}).web(w, httptest.NewRequest("GET", "/-/admin/ldap", nil), admin, "csrf")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "已跳过 1 个同名本地账号") || !strings.Contains(w.Body.String(), "已同步 2 个") || !strings.Contains(w.Body.String(), "admin&lt;script&gt;") || strings.Contains(w.Body.String(), "admin<script>") {
		t.Fatalf("missing or unsafe persisted skip notice: %d %s", w.Code, w.Body.String())
	}
}
