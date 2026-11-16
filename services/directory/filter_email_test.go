// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"github.com/go-ldap/ldap/v3"
	"path/filepath"
	"strings"
	"svnova/models"
	"testing"
)

func TestGiteaFilterPlaceholdersAndEscaping(t *testing.T) {
	for _, pattern := range []string{"(&(objectClass=posixAccount)(uid=%[1]s))", "(&(objectClass=posixAccount)(uid=%s))"} {
		source := &models.LDAPConfig{Filter: pattern, UsernameAttribute: "uid"}
		if actual := filter(source); actual != "(&(objectClass=posixAccount)(uid=*))" {
			t.Fatal(actual)
		}
		if actual := loginFilter(source, "alice"); actual != "(&(&(objectClass=posixAccount)(uid=alice))(uid=alice))" {
			t.Fatal(actual)
		}
		actual := loginFilter(source, "*)(uid=*)")
		if strings.Contains(actual, "(uid=*)") || !strings.Contains(actual, `\2a\29\28uid=\2a\29`) {
			t.Fatal("username was not escaped", actual)
		}
		if _, err := ldap.CompileFilter(actual); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLDAPEmailMappingAndOptionalPreservation(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	d := &Directory{DB: db}
	source := &models.LDAPConfig{ID: 1, UsernameAttribute: "uid", KeyAttribute: "entryUUID", EmailAttribute: "mail"}
	entry := func(email string) *ldap.Entry {
		return ldap.NewEntry("uid=test", map[string][]string{"uid": {"test"}, "entryUUID": {"key"}, "mail": {email}})
	}
	user, err := d.upsert(source, entry("one@example.com"))
	if err != nil || user.Email != "one@example.com" {
		t.Fatal("email not provisioned", err)
	}
	if _, err = db.ID(user.ID).Cols("active").Update(&models.User{Active: true}); err != nil {
		t.Fatal(err)
	}
	user, err = d.upsert(source, entry("two@example.com"))
	if err != nil || user.Email != "two@example.com" {
		t.Fatal("email not updated", err)
	}
	source.EmailAttribute = ""
	user, err = d.upsert(source, entry(""))
	if err != nil || user.Email != "two@example.com" {
		t.Fatal("unmapped email changed", err)
	}
	source.EmailAttribute = "mail"
	if _, err = d.upsert(source, entry("invalid-email")); err == nil {
		t.Fatal("invalid email accepted")
	}
}
