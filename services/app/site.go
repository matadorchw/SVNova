// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"errors"
	"net/http"
	"strings"
	"svnova/models"
	"unicode"
	"unicode/utf8"
)

func normalizeSiteName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 80 || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return "", errors.New("站点名称须为 1–80 个字符，且不能包含控制字符")
	}
	return name, nil
}
func (a *App) siteName() string {
	a.mu.RLock()
	db := a.DB
	a.mu.RUnlock()
	if db != nil {
		setting := new(models.SiteSetting)
		ok, err := db.ID(1).Get(setting)
		if err == nil && ok && setting.SiteName != "" {
			return setting.SiteName
		}
	}
	return "SVNova"
}
func (a *App) siteSettings(w http.ResponseWriter, r *http.Request, u *models.User) error {
	if !u.Admin {
		return forbidden()
	}
	if r.Method == "GET" {
		return send(w, map[string]string{"SiteName": a.siteName()})
	}
	if r.Method != "POST" {
		return &httpError{405, "不支持此操作"}
	}
	var in struct{ SiteName string }
	if err := decode(w, r, &in); err != nil {
		return err
	}
	name, err := normalizeSiteName(in.SiteName)
	if err != nil {
		return err
	}
	_, err = a.DB.Exec("INSERT INTO site_setting (id, site_name) VALUES (1, ?) ON CONFLICT(id) DO UPDATE SET site_name=excluded.site_name", name)
	if err != nil {
		return err
	}
	return send(w, map[string]string{"SiteName": name})
}
