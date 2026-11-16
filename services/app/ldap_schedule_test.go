// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/directory"
	"testing"
)

func TestLDAPScheduleAdminValidationAndUserPreservation(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	a := &App{DB: db, Directory: &directory.Directory{DB: db}}
	user := &models.User{ID: 1, Name: "ldap", LDAPAccount: true, LDAPKey: "key", Active: true}
	if _, err = db.Insert(user); err != nil {
		t.Fatal(err)
	}
	admin := &models.User{Admin: true, Active: true}
	call := func(actor *models.User, body string) error {
		return a.ldapSchedule(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(body)), actor)
	}
	if call(user, `{"Time":"30 2 * * *"}`) == nil {
		t.Fatal("nonadmin changed schedule")
	}
	if call(admin, `{"Time":"25:00"}`) == nil {
		t.Fatal("invalid time accepted")
	}
	if err = call(admin, `{"Time":"30 2 * * *"}`); err != nil {
		t.Fatal(err)
	}
	value, err := a.Directory.Schedule()
	if err != nil || value.Time != "30 2 * * *" {
		t.Fatal("time not persisted", err)
	}
	if n, _ := db.Count(new(models.User)); n != 1 {
		t.Fatal("schedule change cleaned LDAP users")
	}
}
