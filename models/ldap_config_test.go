// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

import (
	"path/filepath"
	"strings"
	"testing"

	"xorm.io/xorm"
)

func legacyLDAPFixture(t *testing.T) (string, *xorm.Engine) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		"DROP TABLE ldap_config",
		"CREATE TABLE ldap_source(id INTEGER PRIMARY KEY, name TEXT, password TEXT)",
		"ALTER TABLE user RENAME COLUMN ldap_account TO ldap_id",
		"ALTER TABLE ldap_sync_run ADD COLUMN source_id INTEGER",
		"ALTER TABLE ldap_sync_run ADD COLUMN source_name TEXT",
		"CREATE INDEX IDX_ldap_sync_run_source_id ON ldap_sync_run(source_id)",
		"DELETE FROM migration WHERE id=15",
	} {
		if _, err = db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	return path, db
}

func TestEmptyLDAPUpgradePreservesBusinessData(t *testing.T) {
	path, db := legacyLDAPFixture(t)
	if _, err := db.Exec("INSERT INTO user(name,password,salt,admin,active,ldap_id) VALUES('admin','hash','salt',1,1,0)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Insert(&Organization{Name: "org"}, &Repository{Name: "repo", OrgID: 1}, &Task{RepositoryID: 1, State: "queued"}); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	var user User
	if ok, err := db.Where("name=?", "admin").Get(&user); err != nil || !ok || user.Password != "hash" || user.LDAPAccount {
		t.Fatal("local account changed", err)
	}
	if ok, err := db.Where("name=? AND ldap_account=0", "admin").Exist(new(User)); err != nil || !ok {
		t.Fatal("upgraded local login cannot find account", err)
	}
	for _, table := range []any{new(Organization), new(Repository), new(Task)} {
		if n, err := db.Count(table); err != nil || n != 1 {
			t.Fatal("business data changed", err)
		}
	}
	if exists, err := db.IsTableExist("ldap_source"); err != nil || exists {
		t.Fatal("legacy table retained", err)
	}
	if _, err := db.Insert(&LDAPConfig{ID: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Insert(&LDAPConfig{ID: 2}); err == nil {
		t.Fatal("multiple LDAP configurations accepted")
	}
	if _, err := db.Exec("UPDATE ldap_config SET id=2"); err == nil {
		t.Fatal("singleton key changed")
	}
	db.Close()
	db, err = Open(path)
	if err != nil {
		t.Fatal("upgraded database cannot reopen", err)
	}
}

func TestLegacyLDAPDataRejectedBeforeSchemaChanges(t *testing.T) {
	for _, q := range []string{
		"INSERT INTO ldap_source(id,name,password) VALUES(1,'LDAP','ciphertext')",
		"INSERT INTO user(name,ldap_id) VALUES('ldap-user',1)",
		"INSERT INTO ldap_sync_run(status,source_id) VALUES('success',1)",
	} {
		t.Run(q, func(t *testing.T) {
			path, db := legacyLDAPFixture(t)
			if _, err := db.Exec(q); err != nil {
				t.Fatal(err)
			}
			db.Close()
			if upgraded, err := Open(path); err == nil {
				upgraded.Close()
				t.Fatal("legacy LDAP data accepted")
			} else if !strings.Contains(err.Error(), "legacy LDAP data") {
				t.Fatal(err)
			}
			db, err := xorm.NewEngine("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if exists, err := db.IsTableExist("ldap_config"); err != nil || exists {
				t.Fatal("schema changed before rejection", err)
			}
			var count int64
			table := "ldap_source"
			if strings.Contains(q, "INTO user") {
				table = "user"
			}
			if strings.Contains(q, "INTO ldap_sync_run") {
				table = "ldap_sync_run"
			}
			if _, err := db.SQL("SELECT COUNT(*) FROM " + table).Get(&count); err != nil || count != 1 {
				t.Fatal("legacy data changed", err)
			}
		})
	}
}
