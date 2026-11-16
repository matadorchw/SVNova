// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/security"
	"testing"
	"time"
)

func TestAccountPasswordValidationRevocationAndStaleLogin(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "account.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	hash, salt, err := security.Password("OriginalPassword2026!")
	if err != nil {
		t.Fatal(err)
	}
	user := &models.User{Name: "local", Active: true, Password: hash, Salt: salt}
	if _, err = db.Insert(user); err != nil {
		t.Fatal(err)
	}
	_, err = db.Insert(&models.Session{Token: "first", UserID: user.ID}, &models.Session{Token: "second", UserID: user.ID}, &models.Session{Token: "other", UserID: 999})
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	change := func(current, next, confirm string) error {
		t.Helper()
		body, _ := json.Marshal(map[string]string{"CurrentPassword": current, "NewPassword": next, "ConfirmPassword": confirm})
		return a.changePassword(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(string(body))), user)
	}
	for _, in := range [][3]string{{"wrong", "ReplacementPassword2026!", "ReplacementPassword2026!"}, {"OriginalPassword2026!", "short", "short"}, {"OriginalPassword2026!", "ReplacementPassword2026!", "different"}, {"OriginalPassword2026!", "OriginalPassword2026!", "OriginalPassword2026!"}} {
		if change(in[0], in[1], in[2]) == nil {
			t.Fatal("invalid password change accepted")
		}
	}
	if err = change("OriginalPassword2026!", "ReplacementPassword2026!", "ReplacementPassword2026!"); err != nil {
		t.Fatal(err)
	}
	fresh := new(models.User)
	_, err = db.ID(user.ID).Get(fresh)
	if err != nil {
		t.Fatal(err)
	}
	if security.Verify("OriginalPassword2026!", fresh.Password, fresh.Salt) || !security.Verify("ReplacementPassword2026!", fresh.Password, fresh.Salt) {
		t.Fatal("password not replaced")
	}
	n, _ := db.Where("user_id=?", user.ID).Count(new(models.Session))
	if n != 0 {
		t.Fatal("old sessions survived")
	}
	n, _ = db.Where("user_id=999").Count(new(models.Session))
	if n != 1 {
		t.Fatal("other user's session revoked")
	}
	if a.createSession(user, &models.Session{Token: "stale", UserID: user.ID}) == nil {
		t.Fatal("stale login issued session after change")
	}
	if err = a.createSession(fresh, &models.Session{Token: "new", UserID: user.ID}); err != nil {
		t.Fatal(err)
	}
	fresh.LDAPAccount = true
	if _, err = db.ID(user.ID).Cols("ldap_account").Update(fresh); err != nil {
		t.Fatal(err)
	}
	if change("ReplacementPassword2026!", "AnotherPassword2026!", "AnotherPassword2026!") == nil {
		t.Fatal("LDAP credentials modified")
	}
	a.attempts["password:1"] = attempt{Count: 10, Since: time.Now()}
	if err = change("ReplacementPassword2026!", "AnotherPassword2026!", "AnotherPassword2026!"); err == nil {
		t.Fatal("rate limit missing")
	} else if e, ok := err.(*httpError); !ok || e.Status != 429 {
		t.Fatalf("wrong rate limit response: %v", err)
	}
}
