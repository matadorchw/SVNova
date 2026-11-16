// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"net/http"
	"strings"
	"svnova/models"
	"svnova/services/security"
	"time"
)

func (a *App) changePassword(w http.ResponseWriter, r *http.Request, actor *models.User) error {
	var input struct{ CurrentPassword, NewPassword, ConfirmPassword string }
	if err := decode(w, r, &input); err != nil {
		return err
	}
	key := fmt.Sprintf("password:%d", actor.ID)
	a.mu.Lock()
	if a.attempts == nil {
		a.attempts = map[string]attempt{}
	}
	count := a.attempts[key]
	if time.Since(count.Since) > 15*time.Minute {
		count = attempt{Since: time.Now()}
	}
	count.Count++
	a.attempts[key] = count
	a.mu.Unlock()
	if count.Count > 10 {
		return &httpError{429, "尝试过于频繁，请稍后重试"}
	}
	current := new(models.User)
	ok, err := a.DB.ID(actor.ID).Get(current)
	if err != nil {
		return err
	}
	if !ok || !current.Active {
		return forbidden()
	}
	if current.LDAPAccount {
		return &httpError{400, "LDAP 账号请在目录服务中修改密码"}
	}
	if len(input.NewPassword) < 12 || len(input.NewPassword) > 1024 {
		return &httpError{400, "新密码长度须为 12 至 1024 字节"}
	}
	if input.NewPassword != input.ConfirmPassword {
		return &httpError{400, "两次输入的新密码不一致"}
	}
	if !security.Verify(input.CurrentPassword, current.Password, current.Salt) {
		return &httpError{400, "当前密码不正确"}
	}
	if input.CurrentPassword == input.NewPassword {
		return &httpError{400, "新密码不能与当前密码相同"}
	}
	hash, salt, err := security.Password(input.NewPassword)
	if err != nil {
		return err
	}
	tx := a.DB.NewSession()
	defer tx.Close()
	if err = tx.Begin(); err != nil {
		return err
	}
	result, err := tx.Exec("UPDATE user SET password=?,salt=? WHERE id=? AND password=? AND salt=? AND active=1 AND ldap_account=0", hash, salt, current.ID, current.Password, current.Salt)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return &httpError{409, "账号状态已变化，请重新登录后重试"}
	}
	if _, err = tx.Where("user_id=?", current.ID).Delete(new(models.Session)); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: "svnova_session", Value: "", Path: "/", HttpOnly: true, Secure: strings.HasPrefix(a.Config.Origin, "https://"), MaxAge: -1, SameSite: http.SameSiteLaxMode})
	return send(w, map[string]bool{"changed": true})
}

// Serialize session issuance with password changes and administrator revocation.
func (a *App) createSession(user *models.User, session *models.Session) error {
	tx := a.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return err
	}
	result, err := tx.Exec("UPDATE user SET last_login=? WHERE id=? AND active=1 AND password=? AND salt=? AND ldap_account=?", time.Now().Unix(), user.ID, user.Password, user.Salt, user.LDAPAccount)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("account changed before session issuance")
	}
	if _, err = tx.Insert(session); err != nil {
		return err
	}
	return tx.Commit()
}
