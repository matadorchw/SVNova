// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/directory"
	"svnova/services/security"
	"testing"
	"time"
)

func TestLDAPConfigSavePagesAndDisabledSession(t *testing.T) {
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	vault, err := security.NewVault(filepath.Join(root, "key"))
	if err != nil {
		t.Fatal(err)
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	admin := &models.User{ID: 1, Name: "admin", Active: true, Admin: true}
	reader := &models.User{ID: 2, Name: "reader", Active: true}
	if _, err = db.Insert(admin, reader); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Vault: vault, Directory: &directory.Directory{DB: db, Vault: vault}, Templates: templates}
	body := func(enabled bool) string {
		b, _ := json.Marshal(map[string]any{"Enabled": enabled, "Protocol": "starttls", "Host": "ldap.example.com", "Port": 389, "BindDN": "cn=service,dc=example", "Password": "bind-secret", "ReplacePassword": true, "BaseDN": "dc=example", "Filter": "(objectClass=person)", "UsernameAttribute": "uid", "KeyAttribute": "entryUUID", "NameAttribute": "cn"})
		return string(b)
	}
	call := func(actor *models.User, method, path, body string) (*httptest.ResponseRecorder, error) {
		w := httptest.NewRecorder()
		err := a.api(w, httptest.NewRequest(method, path, strings.NewReader(body)), actor)
		return w, err
	}
	if _, err = call(reader, "GET", "/api/v1/admin/ldap", ""); err == nil {
		t.Fatal("non-admin read sources")
	}
	configured := strings.TrimSuffix(body(true), "}") + `,"EmailAttribute":"customMail","SkipVerify":true,"UsePagedSearch":true,"SearchPageSize":123,"SyncEnabled":false,"AllowDeactivateAll":true}`
	w, err := call(admin, "POST", "/api/v1/admin/ldap", configured)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(w.Body.String(), "bind-secret") || strings.Contains(w.Body.String(), "Password") {
		t.Fatal("source response leaked password")
	}
	for _, path := range []string{"/-/admin/ldap", "/-/admin/ldap/sync-history"} {
		w = httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", path, nil), admin, "csrf")
		if w.Code != 200 || strings.Contains(w.Body.String(), "bind-secret") {
			t.Fatalf("unsafe or failed auth page %s: %d", path, w.Code)
		}
		w = httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", path, nil), reader, "csrf")
		if w.Code != 403 {
			t.Fatal("non-admin auth page allowed")
		}
	}
	first := new(models.LDAPConfig)
	if _, err = db.ID(1).Get(first); err != nil {
		t.Fatal(err)
	}
	if first.URL != "ldap://ldap.example.com:389" {
		t.Fatal("StartTLS settings not mapped")
	}
	if first.EmailAttribute != "customMail" {
		t.Fatal("email mapping was not saved")
	}
	if !first.SkipVerify || !first.UsePagedSearch || first.SearchPageSize != 123 || first.SyncEnabled || !first.AllowDeactivateAll {
		t.Fatal("source options were not saved")
	}
	ciphertext := first.Password
	edit := strings.ReplaceAll(body(false), `"ReplacePassword":true`, `"ReplacePassword":false`)
	previewResponse, previewErr := call(admin, "POST", "/api/v1/admin/ldap/preview", edit)
	if previewErr != nil {
		t.Fatal(previewErr)
	}
	var preview directory.CleanupPreview
	if err = json.Unmarshal(previewResponse.Body.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	edit = strings.TrimSuffix(edit, "}") + `,"ConfirmCleanup":true,"CleanupToken":"` + preview.Token + `"}`
	if _, err = call(admin, "POST", "/api/v1/admin/ldap", edit); err != nil {
		t.Fatal(err)
	}
	first = new(models.LDAPConfig)
	if _, err = db.ID(1).Get(first); err != nil || first.Password != ciphertext || first.SyncEnabled || !first.SkipVerify || first.SearchPageSize != 123 || first.EmailAttribute != "customMail" {
		t.Fatal("edit replaced saved credential")
	}
	if n, err := db.Count(new(models.LDAPConfig)); err != nil || n != 1 {
		t.Fatal("save created another configuration", err)
	}
	for _, path := range []string{"/api/v1/admin/auths", "/api/v1/admin/ldap/1"} {
		if _, err = call(admin, "GET", path, ""); err == nil {
			t.Fatal("removed API remains accessible", path)
		}
	}
	user := &models.User{ID: 3, Name: "bound-user", LDAPAccount: true, LDAPKey: "stable", Active: true}
	if _, err = db.Insert(user); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte("old-source-session"))
	if _, err = db.Insert(&models.Session{Token: hex.EncodeToString(sum[:]), UserID: user.ID, Expires: time.Now().Add(time.Hour).Unix()}); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("GET", "/", nil)
	request.AddCookie(&http.Cookie{Name: "svnova_session", Value: "old-source-session"})
	if actor, _ := a.current(request); actor != nil {
		t.Fatal("disabled LDAP session accepted")
	}
}
