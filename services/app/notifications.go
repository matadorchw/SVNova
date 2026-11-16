// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"svnova/models"
	"svnova/services/notifications"
)

type notificationPage struct {
	Config             models.NotificationConfig
	Configured         bool
	Deliveries         []models.NotificationDelivery
	Detail             *models.NotificationDelivery
	Attempts           []models.NotificationAttempt
	Preview            string
	Page, Pages        int
	Total              int64
	Previous, Next     int
	Kind, Status, View string
	Keyword, Sort      string
	Pagination         adminUsersPage
}

func (p notificationPage) SortLink(value string) string { p.Sort = value; return p.Link(1) }

func (p notificationPage) Link(page int) string {
	q := url.Values{"tab": {"deliveries"}, "q": {p.Keyword}, "kind": {p.Kind}, "status": {p.Status}, "sort": {p.Sort}, "page": {strconv.Itoa(page)}}
	return "/-/admin/notifications?" + q.Encode()
}

func (a *App) notificationService() *notifications.Service {
	if a.Notifications != nil {
		return a.Notifications
	}
	return &notifications.Service{DB: a.DB, Vault: a.Vault, Origin: a.Config.Origin}
}
func (a *App) notificationData(r *http.Request) (notificationPage, error) {
	s := a.notificationService()
	c, err := s.Config()
	out := notificationPage{Config: c, Configured: c.Webhook != "" && c.Secret != "", Page: 1, Pages: 1, View: r.URL.Query().Get("tab")}
	if err != nil {
		return out, err
	}
	out.Kind = r.URL.Query().Get("kind")
	if out.Kind != "svn" && out.Kind != "ldap" && out.Kind != "test" {
		out.Kind = ""
	}
	out.Status = r.URL.Query().Get("status")
	if !strings.Contains("|pending|sending|retry|success|failed|cancelled|", "|"+out.Status+"|") {
		out.Status = ""
	}
	where := "1=1"
	args := []any{}
	out.Keyword = strings.TrimSpace(r.URL.Query().Get("q"))
	out.Sort = r.URL.Query().Get("sort")
	orders := map[string]string{"newest": "id DESC", "oldest": "id ASC", "created": "created ASC, id ASC", "created-desc": "created DESC, id DESC", "attempts": "attempts ASC, id ASC", "attempts-desc": "attempts DESC, id DESC", "next": "next_attempt ASC, id ASC", "next-desc": "next_attempt DESC, id DESC", "finished": "finished ASC, id ASC", "finished-desc": "finished DESC, id DESC"}
	order, ok := orders[out.Sort]
	if !ok {
		out.Sort = "newest"
		order = orders[out.Sort]
	}
	if out.Keyword != "" {
		where += " AND (instr(lower(object),lower(?))>0 OR instr(lower(last_error),lower(?))>0 OR CAST(id AS TEXT)=?)"
		args = append(args, out.Keyword, out.Keyword, strings.TrimPrefix(out.Keyword, "#"))
	}
	if out.Kind != "" {
		where += " AND kind=?"
		args = append(args, out.Kind)
	}
	if out.Status != "" {
		where += " AND status=?"
		args = append(args, out.Status)
	}
	out.Total, err = a.DB.Where(where, args...).Count(new(models.NotificationDelivery))
	if err != nil {
		return out, err
	}
	out.Pages = max(1, int((out.Total+adminListPageSize-1)/adminListPageSize))
	out.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	out.Page = min(max(out.Page, 1), out.Pages)
	if err = a.DB.Where(where, args...).OrderBy(order).Limit(adminListPageSize, (out.Page-1)*adminListPageSize).Find(&out.Deliveries); err != nil {
		return out, err
	}
	if out.Page > 1 {
		out.Previous = out.Page - 1
	}
	if out.Page < out.Pages {
		out.Next = out.Page + 1
	}
	out.Pagination = adminUsersPage{Page: out.Page, Pages: out.Pages}
	out.Pagination.paginateWith(out.Link)
	if r.URL.Query().Get("delivery") != "" {
		out.View = "deliveries"
		var d models.NotificationDelivery
		ok, e := a.DB.ID(id(r.URL.Query().Get("delivery"))).Get(&d)
		if e != nil {
			return out, e
		}
		if !ok {
			return out, &httpError{404, "投递记录不存在"}
		}
		out.Detail = &d
		if err = a.DB.Where("delivery_id=?", d.ID).Asc("id").Find(&out.Attempts); err != nil {
			return out, err
		}
		var event notifications.Event
		if json.Unmarshal([]byte(d.Payload), &event) == nil {
			_, out.Preview = notifications.Format(event, d.ID, a.Config.Origin)
		}
	}
	return out, nil
}
func (a *App) notificationAPI(w http.ResponseWriter, r *http.Request, u *models.User, p []string) error {
	if !u.Admin {
		return forbidden()
	}
	s := a.notificationService()
	if len(p) == 0 {
		if r.Method == "GET" {
			c, err := s.Config()
			if err != nil {
				return err
			}
			return send(w, map[string]any{"config": c, "configured": c.Webhook != "" && c.Secret != ""})
		}
		if r.Method != "POST" {
			return &httpError{405, "不支持此操作"}
		}
		var in struct {
			Enabled, SVNEnabled, LDAPEnabled, SVNNoChange, LDAPNoChange, SVNFailureOnly bool
			Webhook, Secret                                                             string
			ReplaceWebhook, ReplaceSecret                                               bool
			RetentionDays                                                               int
		}
		if err := decode(w, r, &in); err != nil {
			return err
		}
		if err := s.Save(notifications.Settings{Enabled: in.Enabled, SVNEnabled: in.SVNEnabled, LDAPEnabled: in.LDAPEnabled, SVNNoChange: in.SVNNoChange, SVNFailureOnly: in.SVNFailureOnly, LDAPNoChange: in.LDAPNoChange, Webhook: in.Webhook, Secret: in.Secret, ReplaceWebhook: in.ReplaceWebhook, ReplaceSecret: in.ReplaceSecret, RetentionDays: in.RetentionDays}); err != nil {
			return &httpError{400, err.Error()}
		}
		return send(w, map[string]bool{"ok": true})
	}
	if len(p) == 1 && p[0] == "deliveries" && r.Method == "GET" {
		data, err := a.notificationData(r)
		if err != nil {
			return err
		}
		return send(w, map[string]any{"items": data.Deliveries, "total": data.Total, "page": data.Page, "pages": data.Pages})
	}
	if len(p) == 1 && p[0] == "test" && r.Method == "POST" {
		if err := s.Test(); err != nil {
			return &httpError{400, err.Error()}
		}
		return send(w, map[string]bool{"queued": true})
	}
	if len(p) == 3 && p[0] == "deliveries" && p[2] == "redeliver" && r.Method == "POST" {
		if err := s.Redeliver(id(p[1])); err != nil {
			return &httpError{409, err.Error()}
		}
		return send(w, map[string]bool{"queued": true})
	}
	if len(p) == 2 && p[0] == "deliveries" && r.Method == "GET" {
		q := r.URL.Query()
		q.Set("delivery", p[1])
		r.URL.RawQuery = q.Encode()
		data, err := a.notificationData(r)
		if err != nil {
			return err
		}
		return send(w, map[string]any{"delivery": data.Detail, "attempts": data.Attempts, "preview": data.Preview})
	}
	return &httpError{404, "接口不存在"}
}
func (p *webPage) NotificationState(value string) string {
	return map[string]string{"pending": "待发送", "sending": "发送中", "retry": "等待重试", "success": "成功", "failed": "失败", "cancelled": "已取消", "interrupted": "服务中断"}[value]
}
func (p *webPage) NotificationKind(value string) string {
	return map[string]string{"svn": "SVN同步", "ldap": "LDAP同步", "test": "测试消息"}[value]
}
func (p *webPage) NotificationTime(value int64) string { return p.Time(value / 1000) }
