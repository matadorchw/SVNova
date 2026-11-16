// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"svnova/models"
	"xorm.io/xorm"
)

// Recheck authority after taking the organization writer lock to handle concurrent revocation.
func checkOrganizationOwner(tx *xorm.Session, u *models.User, org int64) error {
	exists, err := tx.ID(org).Exist(new(models.Organization))
	if err != nil {
		return err
	}
	if !exists {
		return &httpError{404, "组织不存在或已被删除，请刷新页面。"}
	}
	if u == nil {
		return forbidden()
	}
	current := new(models.User)
	exists, err = tx.ID(u.ID).Get(current)
	if err != nil {
		return err
	}
	if !exists || !current.Active {
		return forbidden()
	}
	if current.Admin {
		return nil
	}
	var n int64
	_, err = tx.SQL("SELECT COUNT(*) FROM team t JOIN team_member tm ON tm.team_id=t.id JOIN membership m ON m.org_id=t.org_id AND m.user_id=tm.user_id WHERE t.org_id=? AND t.is_owner_team=? AND tm.user_id=?", org, true, u.ID).Get(&n)
	if err != nil {
		return err
	}
	if n == 0 {
		return forbidden()
	}
	return nil
}
