// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"svnova/models"
	"svnova/services/directory"
)

type authSyncRow struct {
	models.LDAPSyncRun
	StartedText  string   `json:"-"`
	TriggerText  string   `json:"-"`
	StatusText   string   `json:"-"`
	FailureText  string   `json:"-"`
	DurationText string   `json:"-"`
	SkippedNames []string `json:"skippedLocalUsers"`
}
type authHistoryPage struct {
	Pagination adminUsersPage `json:"-"`
	Items      []authSyncRow  `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Size       int            `json:"size"`
	Pages      int            `json:"pages"`
	Previous   int            `json:"-"`
	Next       int            `json:"-"`
}

func (a *App) authHistory(r *http.Request) (authHistoryPage, error) {
	out := authHistoryPage{Items: []authSyncRow{}, Page: 1, Size: adminListPageSize, Pages: 1}
	for key, target := range map[string]*int{"page": &out.Page, "size": new(int)} {
		if value := r.URL.Query().Get(key); value != "" {
			number, err := strconv.Atoi(value)
			if err != nil || number < 1 {
				return out, &httpError{400, "分页参数无效"}
			}
			*target = number
		}
	}
	var err error
	out.Total, err = a.DB.Count(new(models.LDAPSyncRun))
	if err != nil {
		return out, err
	}
	out.Pages = int((out.Total + int64(out.Size) - 1) / int64(out.Size))
	if out.Pages < 1 {
		out.Pages = 1
	}
	if out.Page > out.Pages {
		out.Page = out.Pages
	}
	var runs []models.LDAPSyncRun
	if err = a.DB.Desc("started", "id").Limit(out.Size, (out.Page-1)*out.Size).Find(&runs); err != nil {
		return out, err
	}
	for _, run := range runs {
		row := authSyncRow{LDAPSyncRun: run, StartedText: (&webPage{}).Time(run.Started / 1000), TriggerText: map[string]string{"manual": "手动", "scheduled": "定时"}[run.Trigger], StatusText: map[string]string{"running": "进行中", "success": "成功", "failed": "失败"}[run.Status], DurationText: fmt.Sprintf("%d 毫秒", run.DurationMS)}
		if run.Status == "running" {
			row.DurationText = "—"
		}
		if run.FailureReason != "" {
			row.FailureText = map[string]string{"connection": "无法连接 LDAP，请检查地址、端口及证书", "tls": "StartTLS 协商失败，请检查证书和服务器配置", "bind": "服务账号绑定失败，请检查绑定 DN 和密码", "query": "用户搜索失败，请检查搜索基准、过滤器及读取权限", "empty": "搜索结果为空，已保留用户", "identity": "目录记录缺少用户名或稳定身份属性", "duplicate_identity": "目录返回重复稳定身份，用户变更已回滚", "email": "目录邮箱属性无效，用户变更已回滚", "identity_conflict": "用户名属于其他 LDAP 身份，用户变更已回滚", "interrupted": "服务重启导致同步中断，未提交的用户变更已回滚", "internal": "内部错误，用户变更未提交，请查看服务日志"}[run.FailureReason]
		}
		var report directory.SyncReport
		_ = json.Unmarshal([]byte(run.Report), &report)
		row.SkippedNames = report.SkippedLocalUsers
		out.Items = append(out.Items, row)
	}
	if out.Page > 1 {
		out.Previous = out.Page - 1
	}
	if out.Page < out.Pages {
		out.Next = out.Page + 1
	}
	out.Pagination = adminUsersPage{Total: out.Total, Page: out.Page, Pages: out.Pages, Label: "同步历史分页"}
	out.Pagination.paginateWith(func(page int) string { return fmt.Sprintf("/-/admin/ldap/sync-history?page=%d", page) })
	return out, nil
}
