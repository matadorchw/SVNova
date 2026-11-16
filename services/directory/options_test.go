// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"errors"
	"github.com/go-ldap/ldap/v3"
	"path/filepath"
	"svnova/models"
	"testing"
)

type recordedSearch struct {
	paged bool
	size  uint32
}

func (s *recordedSearch) Search(*ldap.SearchRequest) (*ldap.SearchResult, error) {
	s.paged = false
	return &ldap.SearchResult{}, nil
}
func (s *recordedSearch) SearchWithPaging(_ *ldap.SearchRequest, size uint32) (*ldap.SearchResult, error) {
	s.paged = true
	s.size = size
	return &ldap.SearchResult{}, nil
}

func TestLDAPSearchOptions(t *testing.T) {
	client := &recordedSearch{}
	source := &models.LDAPConfig{UsePagedSearch: true, SearchPageSize: 123}
	if _, err := searchUsers(client, source); err != nil || !client.paged || client.size != 123 {
		t.Fatal("configured page size not used", err)
	}
	source.UsePagedSearch = false
	if _, err := searchUsers(client, source); err != nil || client.paged {
		t.Fatal("paging was not disabled", err)
	}
}

func TestEmptyLDAPResultProtectionAndSyncFlag(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	source := &models.LDAPConfig{ID: 1, Enabled: true, SyncEnabled: false}
	if _, err = db.Insert(source, &models.User{ID: 1, Name: "bound", LDAPAccount: true, LDAPKey: "one", Active: true}, &models.User{ID: 3, Name: "local", Active: true}); err != nil {
		t.Fatal(err)
	}
	for _, s := range []models.Session{{Token: "one", UserID: 1}, {Token: "three", UserID: 3}} {
		if _, err = db.Insert(&s); err != nil {
			t.Fatal(err)
		}
	}
	d := &Directory{DB: db}
	if err = d.Sync(); err != nil {
		t.Fatal("disabled automatic sync attempted directory access", err)
	}
	if err = d.SyncAs(0, ""); !errors.Is(err, ErrSyncDisabled) {
		t.Fatal("manual sync ignored flag", err)
	}
	if err = d.syncEntries(source, nil); err == nil {
		t.Fatal("empty result did not retain users")
	}
	before := new(models.User)
	if _, err = db.ID(1).Get(before); err != nil || !before.Active {
		t.Fatal("protected user disabled")
	}
	source.AllowDeactivateAll = true
	if err = d.syncEntries(source, nil); err != nil {
		t.Fatal(err)
	}
	after := new(models.User)
	if _, err = db.ID(1).Get(after); err != nil || after.Active || !after.LDAPDisabled {
		t.Fatal("explicit empty-result policy not applied")
	}
	for _, userID := range []int64{3} {
		var user models.User
		if _, err = db.ID(userID).Get(&user); err != nil || !user.Active {
			t.Fatal("unrelated user was disabled")
		}
	}
	if sessions, err := db.Count(new(models.Session)); err != nil || sessions != 1 {
		t.Fatal("unrelated sessions removed", err)
	}
}
