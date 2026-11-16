// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"bytes"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"testing"
)

func TestUserEditProtectedControls(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "users.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	actor := &models.User{ID: 1, Name: "admin", Admin: true, Active: true}
	_, err = db.Insert(actor, &models.User{ID: 2, Name: "local", Active: true}, &models.User{ID: 3, Name: "ldap", LDAPAccount: true, Active: true}, &models.User{ID: 4, Name: "owner", Active: true}, &models.Organization{ID: 1, Name: "org"}, &models.Team{ID: 1, OrgID: 1, Name: "Owners", IsOwnerTeam: true}, &models.Membership{OrgID: 1, UserID: 4}, &models.TeamMember{TeamID: 1, UserID: 4})
	if err != nil {
		t.Fatal(err)
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	render := func(user string) (*webPage, string) {
		t.Helper()
		page := &webPage{User: actor, Page: "admin/users/edit"}
		if err := a.adminUsersData(httptest.NewRequest("GET", "/?user_id="+user, nil), page); err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		if err := templates.ExecuteTemplate(&out, "admin/user/edit", page); err != nil {
			t.Fatal(err)
		}
		return page, out.String()
	}
	for _, test := range []struct {
		id, reason string
		self       bool
	}{{"1", "不能删除当前登录的管理员", true}, {"2", "", false}, {"3", "LDAP用户不能单独删除", false}, {"4", "不能删除组织最后一位活跃Owners", false}} {
		t.Run(test.id, func(t *testing.T) {
			page, html := render(test.id)
			if !strings.Contains(html, "全站管理员") || !strings.Contains(html, "启用账号") || !strings.Contains(html, `id="open-delete-user"`) {
				t.Fatal("controls missing")
			}
			disabled := strings.Contains(html, `id="open-delete-user" disabled`)
			if disabled != (test.reason != "") {
				t.Fatalf("delete state: %s", html)
			}
			if test.reason != "" && !strings.Contains(html, "/assets/img/octicons.svg#info") {
				t.Fatal("missing restriction icon")
			}
			if test.reason != "" && !strings.Contains(page.UserDeleteDisabledReason, test.reason) {
				t.Fatal(page.UserDeleteDisabledReason)
			}
			if strings.Contains(html, `id="delete-user-dialog"`) != (test.reason == "") {
				t.Fatal("unexpected delete dialog")
			}
			if strings.Contains(html, `disabled aria-describedby="user-self-protection"`) != test.self {
				t.Fatal("self controls state")
			}
			if test.self {
				for _, name := range []string{"Admin", "Active"} {
					if !strings.Contains(html, `type="hidden" name="`+name+`" data-bool value="true"`) {
						t.Fatal("missing submitted invariant")
					}
				}
			}
		})
	}
	session := db.NewSession()
	defer session.Close()
	reason, err := userDeletionReason(session, &models.User{ID: 99}, actor)
	if err != nil || reason != "不能删除最后一位活跃管理员" {
		t.Fatalf("last admin: %q %v", reason, err)
	}
	if _, err = db.Insert(&models.Membership{OrgID: 1, UserID: 1}, &models.TeamMember{TeamID: 1, UserID: 1}); err != nil {
		t.Fatal(err)
	}
	if page, _ := render("4"); page.UserDeleteDisabledReason != "" {
		t.Fatal("additional active owner not recognized")
	}
	// A page-level decision cannot authorize deletion after ownership changes.
	if _, err = db.Insert(&models.Organization{ID: 2, Name: "changed"}, &models.Team{ID: 2, OrgID: 2, Name: "Owners", IsOwnerTeam: true}, &models.Membership{OrgID: 2, UserID: 2}, &models.TeamMember{TeamID: 2, UserID: 2}); err != nil {
		t.Fatal(err)
	}
	err = a.deleteLocalUser(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(`{"Name":"local"}`)), actor, 2)
	if err == nil || !strings.Contains(err.Error(), "最后一位活跃Owners") {
		t.Fatalf("stale page bypassed protection: %v", err)
	}
}
