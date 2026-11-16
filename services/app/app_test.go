// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"

	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"svnova/models"
	"svnova/services/security"
)

func TestInstallLoginAndAuthorization(t *testing.T) {
	data := t.TempDir()
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a, err := New(data, "", templates, http.NotFoundHandler())
	if err != nil {
		t.Fatal(err)
	}
	form := url.Values{"site_name": {"测试站点"}, "name": {"admin"}, "password": {"administrator-password"}, "origin": {"http://example.com"}, "listen": {"127.0.0.1:3000"}, "database": {filepath.Join(data, "app.db")}, "repositories": {filepath.Join(data, "repos")}}
	req := httptest.NewRequest("POST", "http://example.com/install", strings.NewReader(form.Encode()))
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	a.Handler().ServeHTTP(response, req)
	if response.Code != 303 {
		t.Fatalf("install %d %s", response.Code, response.Body.String())
	}
	defer a.DB.Close()
	if a.siteName() != "测试站点" {
		t.Fatal("installation site name was not saved")
	}
	response = httptest.NewRecorder()
	a.Handler().ServeHTTP(response, req)
	if response.Code != 403 {
		t.Fatal("installation was not locked")
	}
	login := url.Values{"name": {"admin"}, "password": {"administrator-password"}, "source": {"local"}}
	req = httptest.NewRequest("POST", "http://example.com/login", strings.NewReader(login.Encode()))
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	a.Handler().ServeHTTP(response, req)
	if response.Code != 303 || len(response.Result().Cookies()) != 1 {
		t.Fatalf("login %d %s", response.Code, response.Body.String())
	}
	cookie := response.Result().Cookies()[0]
	req = httptest.NewRequest("GET", "http://example.com/", nil)
	req.AddCookie(cookie)
	user, session := a.current(req)
	if user == nil || !user.Admin {
		t.Fatal("session not established")
	}
	call := func(method, path, body string, c *http.Cookie, csrf string) *httptest.ResponseRecorder {
		t.Helper()
		rq := httptest.NewRequest(method, "http://example.com"+path, strings.NewReader(body))
		rq.Header.Set("Origin", "http://example.com")
		rq.Header.Set("X-CSRF-Token", csrf)
		if c != nil {
			rq.AddCookie(c)
		}
		rr := httptest.NewRecorder()
		a.Handler().ServeHTTP(rr, rq)
		return rr
	}
	if rr := call("POST", "/api/v1/admin/site", `{"SiteName":"changed"}`, cookie, ""); rr.Code != 403 {
		t.Fatal("site settings accepted missing CSRF")
	}
	if rr := call("POST", "/api/v1/user/password", `{}`, cookie, ""); rr.Code != 403 {
		t.Fatal("password change accepted missing CSRF")
	}
	if rr := call("POST", "/api/v1/user/password", `{}`, nil, ""); rr.Code != 401 {
		t.Fatal("anonymous password change accepted")
	}
	if rr := call("POST", "/api/v1/organizations", `{"Name":"one","OwnerID":1}`, cookie, ""); rr.Code != 403 {
		t.Fatal("missing CSRF accepted")
	}
	rr := call("POST", "/api/v1/organizations", `{"Name":"one","OwnerID":1}`, cookie, session.CSRF)
	if rr.Code != 200 {
		t.Fatalf("create organization %d %s", rr.Code, rr.Body.String())
	}
	var org models.Organization
	_ = json.Unmarshal(rr.Body.Bytes(), &org)
	reader := &models.User{Name: "reader", Active: true}
	_, _ = a.DB.Insert(reader)
	token := security.Token()
	sum := sha256.Sum256([]byte(token))
	_, _ = a.DB.Insert(&models.Session{Token: hex.EncodeToString(sum[:]), UserID: reader.ID, CSRF: "reader-csrf", Expires: time.Now().Add(time.Hour).Unix()})
	readerCookie := &http.Cookie{Name: "svnova_session", Value: token}
	repo := &models.Repository{OrgID: org.ID, Name: "private", Status: "ready", Kind: "local"}
	_, _ = a.DB.Insert(repo)
	for _, path := range []string{"/admin", "/api/v1/admin/users", "/api/v1/repositories/1", "/api/v1/repositories/1/tasks", "/api/v1/repositories/1/download?path=secret.txt"} {
		if rr = call("GET", path, "", readerCookie, ""); rr.Code != 403 {
			t.Fatalf("unauthorized %s returned %d", path, rr.Code)
		}
	}
	if rr = call("POST", "/api/v1/repositories", `{"OrgID":1,"Name":"hack","Kind":"local","Source":"C:\\repo","SourceStopped":true}`, readerCookie, "reader-csrf"); rr.Code != 403 {
		t.Fatalf("local migration unauthorized: %d", rr.Code)
	}
	reader.Active = false
	_, _ = a.DB.ID(reader.ID).Cols("active").Update(reader)
	if rr = call("GET", "/api/v1/me", "", readerCookie, ""); rr.Code != 401 {
		t.Fatal("disabled user session remained valid")
	}
}
