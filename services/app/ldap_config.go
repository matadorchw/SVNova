// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"svnova/models"
	"svnova/services/directory"
)

func (a *App) ldapConfig(w http.ResponseWriter, r *http.Request, u *models.User, parts []string) error {
	if !u.Admin {
		return forbidden()
	}
	if len(parts) > 1 {
		return &httpError{404, "接口不存在"}
	}
	preview := len(parts) == 1 && parts[0] == "preview"
	if len(parts) == 1 && parts[0] == "sync-history" && r.Method == "GET" {
		history, err := a.authHistory(r)
		if err != nil {
			return err
		}
		return send(w, history)
	}
	if len(parts) == 1 && parts[0] == "sync" && r.Method == "POST" {
		if err := a.Directory.SyncAs(u.ID, u.Name); err != nil {
			if errors.Is(err, directory.ErrSyncBusy) {
				return &httpError{409, "LDAP同步或配置修改正在进行，请稍后刷新历史查看"}
			}
			if errors.Is(err, directory.ErrSyncDisabled) {
				return &httpError{409, "LDAP未启用用户同步"}
			}
			if errors.Is(err, directory.ErrEmptyResult) {
				return &httpError{400, "LDAP搜索结果为空，已保留原有用户"}
			}
			return &httpError{400, "LDAP同步失败，请检查配置、网络及服务账号"}
		}
		var config models.LDAPConfig
		if _, err := a.DB.ID(1).Get(&config); err != nil {
			return err
		}
		var report directory.SyncReport
		_ = json.Unmarshal([]byte(config.LastSyncReport), &report)
		return send(w, map[string]any{"ok": true, "report": report})
	}
	if len(parts) == 1 && !preview {
		return &httpError{404, "接口不存在"}
	}
	previous := models.LDAPConfig{}
	exists, err := a.DB.ID(1).Get(&previous)
	if err != nil {
		return err
	}
	if !exists {
		previous = defaultLDAPConfig()
	}
	if r.Method == "GET" && !preview {
		return send(w, previous)
	}
	if r.Method != "POST" {
		return &httpError{405, "不支持此操作"}
	}
	var in struct {
		CleanupToken                                                                                          string
		ConfirmCleanup                                                                                        bool
		EmailAttribute                                                                                        *string
		URL, Protocol, Host, BindDN, Password, BaseDN, Filter, UsernameAttribute, KeyAttribute, NameAttribute string
		Port                                                                                                  int
		Enabled, ReplacePassword                                                                              bool
		SkipVerify                                                                                            *bool
		UsePagedSearch, SyncEnabled, AllowDeactivateAll                                                       *bool
		SearchPageSize                                                                                        *uint32
	}
	if err := decode(w, r, &in); err != nil {
		return err
	}
	if in.Host != "" || in.Protocol != "" {
		if (in.Protocol != "ldaps" && in.Protocol != "starttls") || in.Port < 1 || in.Port > 65535 || strings.ContainsAny(in.Host, "/\\@?# ") {
			return &httpError{400, "请选择 LDAPS 或 StartTLS，并填写有效主机和端口"}
		}
		scheme := "ldaps"
		if in.Protocol == "starttls" {
			scheme = "ldap"
		}
		in.URL = (&url.URL{Scheme: scheme, Host: net.JoinHostPort(strings.Trim(in.Host, "[]"), strconv.Itoa(in.Port))}).String()
	}
	candidate := &models.LDAPConfig{ID: 1, Enabled: in.Enabled, URL: in.URL, BindDN: in.BindDN, BaseDN: in.BaseDN, Filter: in.Filter, UsernameAttribute: in.UsernameAttribute, KeyAttribute: in.KeyAttribute, NameAttribute: in.NameAttribute}
	candidate.UsePagedSearch = previous.UsePagedSearch
	candidate.EmailAttribute = previous.EmailAttribute
	candidate.SearchPageSize = previous.SearchPageSize
	candidate.SyncEnabled = previous.SyncEnabled
	candidate.AllowDeactivateAll = previous.AllowDeactivateAll
	candidate.SkipVerify = previous.SkipVerify
	if in.UsePagedSearch != nil {
		candidate.UsePagedSearch = *in.UsePagedSearch
	}
	if in.EmailAttribute != nil {
		candidate.EmailAttribute = *in.EmailAttribute
	}
	if in.SearchPageSize != nil {
		candidate.SearchPageSize = *in.SearchPageSize
	}
	if in.SyncEnabled != nil {
		candidate.SyncEnabled = *in.SyncEnabled
	}
	if in.AllowDeactivateAll != nil {
		candidate.AllowDeactivateAll = *in.AllowDeactivateAll
	}
	if in.SkipVerify != nil {
		candidate.SkipVerify = *in.SkipVerify
	}
	if err := directory.Validate(candidate); err != nil {
		return &httpError{400, "LDAP 配置无效：检查安全地址、搜索基准、过滤器及属性映射"}
	}
	var saveErr error
	if preview {
		plan, err := a.Directory.PreviewCleanup(candidate, in.ReplacePassword || !exists, in.Password)
		if err == nil {
			return send(w, plan)
		}
		saveErr = err
	} else {
		token := ""
		if in.ConfirmCleanup {
			token = in.CleanupToken
		}
		saveErr = a.Directory.SaveConfirmed(candidate, in.ReplacePassword || !exists, in.Password, token, u.ID)
	}
	if err := saveErr; err != nil {
		var owner *directory.CleanupOwnerError
		if errors.As(err, &owner) {
			return &httpError{409, fmt.Sprintf("清理会移除组织“%s”的最后一名 Owners 成员，请先补充所有者", owner.Organization)}
		}
		switch {
		case errors.Is(err, directory.ErrCleanupConfirmation):
			return &httpError{409, "更改LDAP配置将清理相关LDAP用户，请重新预览并确认；确认已过期或数据已变化时需再次确认"}
		case errors.Is(err, directory.ErrCleanupOwner):
			return &httpError{409, "清理会移除组织最后一名 Owners 成员，请先补充所有者"}
		case errors.Is(err, directory.ErrCleanupAdmin):
			return &httpError{409, "清理会移除最后一名启用的管理员，请先保留本地管理员"}
		case errors.Is(err, directory.ErrConfigAdmin):
			return forbidden()
		case errors.Is(err, directory.ErrBindPassword):
			return &httpError{400, "启用LDAP认证需要有效绑定密码，请填写密码并勾选替换"}
		}
		return err
	}
	return send(w, candidate)
}
