// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

import (
	"errors"
	"time"
	"xorm.io/xorm"
)

// Check before schema writes so unsupported LDAP databases remain untouched.
func checkLDAPUpgrade(db *xorm.Engine) error {
	exists, err := db.IsTableExist("ldap_source")
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	for _, q := range []string{"SELECT COUNT(*) FROM ldap_source", "SELECT COUNT(*) FROM user WHERE ldap_id<>0", "SELECT COUNT(*) FROM ldap_sync_run"} {
		var count int64
		if _, err = db.SQL(q).Get(&count); err != nil {
			return err
		}
		if count > 0 {
			return errors.New("legacy LDAP data is not supported; LDAP configuration, users and history must be empty before upgrading")
		}
	}
	return nil
}
func finishLDAPMigration(db *xorm.Engine) error {
	tx := db.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE user SET ldap_account=0 WHERE ldap_account IS NULL"); err != nil {
		return err
	}
	// New configuration IDs are fixed; reject direct inserts with another ID.
	if _, err := tx.Exec("CREATE TRIGGER IF NOT EXISTS ldap_config_singleton_insert BEFORE INSERT ON ldap_config WHEN NEW.id<>1 BEGIN SELECT RAISE(ABORT,'LDAP configuration ID must be 1'); END"); err != nil {
		return err
	}
	if _, err := tx.Exec("CREATE TRIGGER IF NOT EXISTS ldap_config_singleton_update BEFORE UPDATE OF id ON ldap_config WHEN NEW.id<>1 BEGIN SELECT RAISE(ABORT,'LDAP configuration ID must be 1'); END"); err != nil {
		return err
	}
	exists, err := tx.IsTableExist("ldap_source")
	if err != nil {
		return err
	}
	if exists {
		if _, err = tx.Exec("DROP TABLE ldap_source"); err != nil {
			return err
		}
		// Old source columns contain no identities under the supported upgrade boundary.
		if _, err = tx.Exec("ALTER TABLE user DROP COLUMN ldap_id"); err != nil {
			return err
		}
		if _, err = tx.Exec("DROP INDEX IF EXISTS IDX_ldap_sync_run_source_id"); err != nil {
			return err
		}
		if _, err = tx.Exec("ALTER TABLE ldap_sync_run DROP COLUMN source_id"); err != nil {
			return err
		}
		if _, err = tx.Exec("ALTER TABLE ldap_sync_run DROP COLUMN source_name"); err != nil {
			return err
		}
	}
	if _, err = tx.Exec("INSERT OR IGNORE INTO migration(id,applied) VALUES(15,?)", time.Now().Unix()); err != nil {
		return err
	}
	return tx.Commit()
}
