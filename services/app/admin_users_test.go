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

func TestAdminUsersSearchPaginationAndPages(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "users.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	actor := &models.User{Name: "admin", Admin: true, Active: true}
	if _, err = db.Insert(actor); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 25; i++ {
		if _, err = db.Insert(&models.User{Name: fmt.Sprintf("person%02d", i), Email: fmt.Sprintf("person%02d@example.com", i), Active: i%2 == 0}); err != nil {
			t.Fatal(err)
		}
	}
	a := &App{DB: db}
	page := &webPage{Page: "admin/users", User: actor}
	if err = a.adminUsersData(httptest.NewRequest("GET", "/?q=person&sort=newest&page=99", nil), page); err != nil {
		t.Fatal(err)
	}
	if page.AdminUsers.Total != 25 || page.AdminUsers.Page != 2 || len(page.AdminUsers.Items) != 5 || !strings.Contains(page.AdminUsers.Previous, "sort=newest") {
		t.Fatalf("bad pagination: %+v", page.AdminUsers)
	}
	if err = a.adminUsersData(httptest.NewRequest("GET", "/?q=PERSON00%40example.com&active=1&admin=0", nil), page); err != nil {
		t.Fatal(err)
	}
	if page.AdminUsers.Total != 1 || page.AdminUsers.Items[0].Name != "person00" {
		t.Fatal("email and status filtering failed")
	}
	if err = a.adminUsersData(httptest.NewRequest("GET", "/?q=%25&sort=invalid", nil), page); err != nil {
		t.Fatal(err)
	}
	if page.AdminUsers.Total != 0 || page.AdminUsers.Sort != "alphabetically" {
		t.Fatal("search wildcard or sort allowlist failed")
	}
	a.Templates, err = ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{"/-/admin/users", "/-/admin/users/new", fmt.Sprintf("/-/admin/users/%d/edit", actor.ID), fmt.Sprintf("/-/admin/users/%d", actor.ID)} {
		w := httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", route, nil), actor, "csrf")
		if w.Code != 200 || strings.Contains(w.Body.String(), "页面加载失败") {
			t.Fatalf("page %s failed: %s", route, w.Body.String())
		}
		if route == "/-/admin/users" {
			for _, text := range []string{`id="user-list-search-form"`, `data-user-filter-reset`, `data-user-sort="newest"`, `ui borderless pagination menu`, `aria-current="page"`, `aria-label="末页"`} {
				if !strings.Contains(w.Body.String(), text) {
					t.Fatal("missing Gitea list control", text)
				}
			}
			if strings.Count(w.Body.String(), `class="header item admin-settings-heading">管理设置</div>`) != 1 || strings.Contains(w.Body.String(), "<h1>管理后台</h1>") {
				t.Fatal("admin pages must use the shared settings heading without a page heading")
			}
		}
		w = httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", route, nil), &models.User{ID: 999, Active: true}, "csrf")
		if w.Code != 403 {
			t.Fatalf("nonadmin allowed %s: %d", route, w.Code)
		}
	}
	w := httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/-/admin/users/999/edit", nil), actor, "csrf")
	if w.Code != 404 {
		t.Fatal("missing user not rejected")
	}
}

func TestAdminTargetProfileKeepsViewerIdentity(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "profile.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	actor := &models.User{Name: "viewer", Admin: true, Active: true}
	target := &models.User{Name: "target", Email: "target@example.com", Active: true}
	if _, err = db.Insert(actor, target); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	a.Templates, err = ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	route := fmt.Sprintf("/user/profile?user_id=%d", target.ID)
	page := &webPage{Page: "user-profile", User: actor}
	if err = a.webData(httptest.NewRequest("GET", route, nil), page); err != nil {
		t.Fatal(err)
	}
	if page.User.ID != actor.ID || page.ProfileUser.ID != target.ID {
		t.Fatal("viewer and profile subject confused")
	}
	w := httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", route, nil), actor, "csrf")
	if w.Code != 200 || !strings.Contains(w.Body.String(), target.Email) {
		t.Fatalf("target profile failed: %d", w.Code)
	}
	w = httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", fmt.Sprintf("/user/profile?user_id=%d", actor.ID), nil), target, "csrf")
	if w.Code != 403 {
		t.Fatal("nonadmin viewed another private profile", w.Code)
	}
}
func TestUserTimestampsLegacyAndSessionAtomicity(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "legacy.db")
	db, err := models.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("ALTER TABLE user DROP COLUMN created"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("ALTER TABLE user DROP COLUMN last_login"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO user (name,active,password,salt,ldap_account) VALUES ('legacy',1,'hash','salt',0)"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err = models.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var legacy models.User
	if _, err = db.Where("name=?", "legacy").Get(&legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.Created != 0 || legacy.LastLogin != 0 || (&webPage{}).UserTime(legacy.Created) != "未知" {
		t.Fatal("legacy times fabricated")
	}
	for _, u := range []*models.User{{Name: "new-local", Active: true}, {Name: "new-ldap", LDAPAccount: true, LDAPKey: "key", Active: true}} {
		if _, err = db.Insert(u); err != nil {
			t.Fatal(err)
		}
		if u.Created <= 0 || u.LastLogin != 0 {
			t.Fatal("creation timestamp missing")
		}
	}
	a := &App{DB: db}
	if err = a.createSession(&legacy, &models.Session{Token: "one", UserID: legacy.ID}); err != nil {
		t.Fatal(err)
	}
	var fresh models.User
	if _, err = db.ID(legacy.ID).Get(&fresh); err != nil {
		t.Fatal(err)
	}
	if fresh.LastLogin <= 0 || fresh.Created != 0 {
		t.Fatal("successful login timestamp missing")
	}
	if _, err = db.Exec("UPDATE user SET last_login=0 WHERE id=?", legacy.ID); err != nil {
		t.Fatal(err)
	}
	if err = a.createSession(&legacy, &models.Session{Token: "one", UserID: legacy.ID}); err == nil {
		t.Fatal("duplicate session accepted")
	}
	fresh = models.User{}
	if _, err = db.ID(legacy.ID).Get(&fresh); err != nil {
		t.Fatal(err)
	}
	if fresh.LastLogin != 0 {
		t.Fatal("failed session changed last login")
	}
}

func TestAdminUserPageWindow(t *testing.T) {
	for _, tc := range []struct {
		page, pages int
		numbers     string
		first, last bool
	}{
		{1, 1, "1", false, false},
		{1, 6, "1,2,3,4,5,-1", false, true},
		{5, 10, "-1,3,4,5,6,7,-1", true, true},
		{10, 10, "-1,6,7,8,9,10", true, false},
	} {
		v := adminUsersPage{Page: tc.page, Pages: tc.pages, Keyword: "a & b", Admin: "1", Active: "0", Sort: "newest"}
		v.paginate()
		var values []string
		for _, n := range v.Numbers {
			values = append(values, fmt.Sprint(n.Number))
			if n.Number > 0 && (!strings.Contains(n.URL, "q=a+%26+b") || !strings.Contains(n.URL, "admin=1") || !strings.Contains(n.URL, "active=0") || !strings.Contains(n.URL, "sort=newest")) {
				t.Fatal("page link lost filters", n.URL)
			}
			if n.Current != (n.Number == tc.page) {
				t.Fatal("current page marker incorrect")
			}
		}
		if strings.Join(values, ",") != tc.numbers || (v.First != "") != tc.first || (v.Last != "") != tc.last {
			t.Fatalf("incorrect page window: %+v", v)
		}
	}
}

func TestAdminUsersAuthFilter(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, u := range []*models.User{{Name: "local", Active: true}, {Name: "ldap-one", LDAPAccount: true, Active: true}, {Name: "ldap-two", LDAPAccount: true}} {
		if _, err := db.Insert(u); err != nil {
			t.Fatal(err)
		}
	}
	a := &App{DB: db}
	for _, tc := range []struct {
		query, auth string
		total       int64
	}{
		{"auth=local", "local", 1}, {"auth=ldap", "ldap", 2},
		{"auth=ldap&active=1&admin=0&q=one", "ldap", 1}, {"auth=invalid", "", 3}, {"", "", 3},
	} {
		p := &webPage{Page: "admin/users"}
		if err := a.adminUsersData(httptest.NewRequest("GET", "/?"+tc.query, nil), p); err != nil {
			t.Fatal(err)
		}
		if p.AdminUsers.Total != tc.total || p.AdminUsers.Auth != tc.auth {
			t.Fatalf("%s: %+v", tc.query, p.AdminUsers)
		}
		if !strings.Contains(p.AdminUsers.Link(2), "auth="+tc.auth) {
			t.Fatal("pagination lost auth filter")
		}
	}
}

func TestProfileRepositoryPaginationAndSort(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "repos.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	user := &models.User{ID: 1, Name: "admin", Admin: true, Active: true}
	if _, err = db.Insert(user, &models.Organization{ID: 1, Name: "org"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 31; i++ {
		if _, err = db.Insert(&models.Repository{OrgID: 1, Name: fmt.Sprintf("repo%02d", i), Status: "ready"}); err != nil {
			t.Fatal(err)
		}
	}
	a := &App{DB: db}
	page := &webPage{User: user, Page: "user-profile"}
	if err = a.webData(httptest.NewRequest("GET", "/user/profile?sort=name-desc&page=4", nil), page); err != nil {
		t.Fatal(err)
	}
	if page.ProfilePagination.Total != 31 || len(page.Repositories) != 1 || page.Repositories[0].Name != "repo00" {
		t.Fatal("incorrect profile pagination")
	}
	if !strings.Contains(page.ProfilePagination.Previous, "sort=name-desc") || !strings.Contains(page.ProfilePagination.Previous, "user_id=1") {
		t.Fatal("pagination lost context")
	}
}

func TestMemberProfileLinksFollowExistingAccess(t *testing.T) {
	page := &webPage{User: &models.User{ID: 1}}
	if !page.CanViewUserProfile(1) || page.CanViewUserProfile(2) {
		t.Fatal("ordinary member profile access changed")
	}
	page.User.Admin = true
	if !page.CanViewUserProfile(2) {
		t.Fatal("administrator profile link missing")
	}
}
