// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http"
	"svnova/models"
)

func (a *App) ldapSchedule(w http.ResponseWriter, r *http.Request, u *models.User) error {
	if !u.Admin {
		return forbidden()
	}
	if r.Method == "GET" {
		value, err := a.Directory.Schedule()
		if err != nil {
			return err
		}
		return send(w, value)
	}
	if r.Method != "POST" {
		return &httpError{405, "不支持此操作"}
	}
	var in struct{ Time string }
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if err := a.Directory.SaveSchedule(in.Time); err != nil {
		return &httpError{400, err.Error()}
	}
	return send(w, map[string]bool{"ok": true})
}
