// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http"
	"svnova/models"
	"xorm.io/xorm"
)

func (a *App) deleteLocalUser(w http.ResponseWriter, r *http.Request, actor *models.User, userID int64) error {
	if r.Method != "POST" {
		return &httpError{405, "不支持此操作"}
	}
	if actor == nil || !actor.Admin || !actor.Active {
		return forbidden()
	}
	var in struct{ Name string }
	if err := decode(w, r, &in); err != nil {
		return &httpError{400, "无法读取确认信息"}
	}
	tx := a.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return err
	}
	// Acquire the write lock before reading authorization and ownership state.
	if _, err := tx.Exec("UPDATE user SET id=id WHERE id=?", actor.ID); err != nil {
		return err
	}
	current := new(models.User)
	ok, err := tx.ID(actor.ID).Get(current)
	if err != nil {
		return err
	}
	if !ok || !current.Active || !current.Admin {
		return forbidden()
	}
	target := new(models.User)
	ok, err = tx.ID(userID).Get(target)
	if err != nil {
		return err
	}
	if !ok {
		return &httpError{404, "用户不存在或已被删除"}
	}
	if !target.LDAPAccount && in.Name != target.Name {
		return &httpError{400, "用户名不匹配，请输入完整用户名"}
	}
	reason, err := userDeletionReason(tx, actor, target)
	if err != nil {
		return err
	}
	if reason != "" {
		return &httpError{409, reason}
	}
	for _, query := range []string{"DELETE FROM session WHERE user_id=?", "DELETE FROM team_member WHERE user_id=?", "DELETE FROM membership WHERE user_id=?", "DELETE FROM user WHERE id=?"} {
		if _, err := tx.Exec(query, userID); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return send(w, map[string]bool{"ok": true})
}

// Use the same protection rules for presentation and transactional enforcement.
func userDeletionReason(tx *xorm.Session, actor, target *models.User) (string, error) {
	if target.LDAPAccount {
		return "LDAP用户不能单独删除，请通过 LDAP 配置处理", nil
	}
	if target.ID == actor.ID {
		return "不能删除当前登录的管理员", nil
	}
	if target.Admin && target.Active {
		count, err := tx.Where("admin=1 AND active=1 AND id<>?", target.ID).Count(new(models.User))
		if err != nil {
			return "", err
		}
		if count == 0 {
			return "不能删除最后一位活跃管理员", nil
		}
	}
	var owners []struct{ OrgID int64 }
	if err := tx.SQL("SELECT DISTINCT t.org_id FROM team t JOIN team_member tm ON tm.team_id=t.id JOIN membership m ON m.org_id=t.org_id AND m.user_id=tm.user_id WHERE t.is_owner_team=1 AND tm.user_id=?", target.ID).Find(&owners); err != nil {
		return "", err
	}
	for _, owner := range owners {
		count, err := tx.SQL("SELECT COUNT(*) FROM team t JOIN team_member tm ON tm.team_id=t.id JOIN user u ON u.id=tm.user_id JOIN membership m ON m.org_id=t.org_id AND m.user_id=u.id WHERE t.org_id=? AND t.is_owner_team=1 AND u.active=1 AND u.id<>?", owner.OrgID, target.ID).Count(new(models.User))
		if err != nil {
			return "", err
		}
		if count == 0 {
			return "不能删除组织最后一位活跃Owners，请先指定其他Owners", nil
		}
	}

	return "", nil
}
