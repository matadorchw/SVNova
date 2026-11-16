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
	"testing"
)

func TestAuthHistoryPaginationAndAuthorization(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Admin: true, Active: true}
	reader := &models.User{ID: 2, Name: "reader", Active: true}
	if _, err = db.Insert(admin, reader, &models.LDAPConfig{ID: 1}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 27; i++ {
		if _, err = db.Insert(&models.LDAPSyncRun{Trigger: "manual", Started: int64(1000 + i), Status: "failed", FailureReason: "bind"}); err != nil {
			t.Fatal(err)
		}
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Templates: templates}
	w := httptest.NewRecorder()
	if err = a.ldapConfig(w, httptest.NewRequest("GET", "/?page=99&size=15", nil), admin, []string{"sync-history"}); err != nil {
		t.Fatal(err)
	}
	var page authHistoryPage
	if err = json.Unmarshal(w.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Size != 20 || page.Total != 27 || page.Page != 2 || len(page.Items) != 7 || strings.Contains(w.Body.String(), "other-source-only") {
		t.Fatal("history paging or source boundary failed")
	}
	if err = a.ldapConfig(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), reader, []string{"sync-history"}); err == nil {
		t.Fatal("non-admin history access allowed")
	}
	w = httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/-/admin/ldap/sync-history", nil), admin, "csrf")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `aria-label="末页"`) || !strings.Contains(w.Body.String(), "服务账号绑定失败") || !strings.Contains(w.Body.String(), "同步历史分页") {
		t.Fatal("history page missing")
	}
	w = httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/-/admin/ldap", nil), admin, "csrf")
	for _, label := range []string{"LDAP 配置", "保存 LDAP 配置", "同步历史", "同步用户"} {
		if !strings.Contains(w.Body.String(), label) {
			t.Fatal("LDAP action missing", label)
		}
	}
}
