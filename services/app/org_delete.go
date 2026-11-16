// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http"
	"svnova/models"
)

func (a *App) deleteOrganization(w http.ResponseWriter, r *http.Request, u *models.User, orgID int64) error {
	if r.Method != "POST" {
		return &httpError{405, "不支持此操作"}
	}
	var in struct{ Name string }
	if err := decode(w, r, &in); err != nil {
		return &httpError{400, "无法读取确认信息，请重新输入组织名称。"}
	}
	tx := a.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return err
	}
	// Serialize deletion with repository creation and membership changes.
	if _, err := tx.Exec("UPDATE organization SET id=id WHERE id=?", orgID); err != nil {
		return err
	}
	org := new(models.Organization)
	ok, err := tx.ID(orgID).Get(org)
	if err != nil {
		return err
	}
	if !ok {
		return &httpError{404, "组织不存在或已被删除。"}
	}
	if err := checkOrganizationOwner(tx, u, orgID); err != nil {
		return &httpError{403, "只有组织 Owners 或站点管理员可以删除组织。"}
	}
	if in.Name != org.Name {
		return &httpError{400, "组织名称不匹配，请输入完整的组织名称。"}
	}
	count, err := tx.Where("org_id=?", orgID).Count(new(models.Repository))
	if err != nil {
		return err
	}
	if count > 0 {
		return &httpError{409, "无法删除组织：组织内仍有仓库，请先处理全部仓库后重试。"}
	}
	for _, query := range []string{
		"DELETE FROM team_repository WHERE team_id IN (SELECT id FROM team WHERE org_id=?)",
		"DELETE FROM team_member WHERE team_id IN (SELECT id FROM team WHERE org_id=?)",
		"DELETE FROM team WHERE org_id=?", "DELETE FROM membership WHERE org_id=?", "DELETE FROM organization WHERE id=?",
	} {
		if _, err := tx.Exec(query, orgID); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return send(w, map[string]bool{"ok": true})
}
