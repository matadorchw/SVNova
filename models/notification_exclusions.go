// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

import "xorm.io/xorm"

func removeNotificationExclusions(db *xorm.Engine) error {
	columns, err := db.QueryString("PRAGMA table_info(notification_config)")
	if err != nil {
		return err
	}
	for _, column := range columns {
		if column["name"] == "excluded_repositories" {
			_, err = db.Exec("ALTER TABLE notification_config DROP COLUMN excluded_repositories")
			return err
		}
	}
	return nil
}
