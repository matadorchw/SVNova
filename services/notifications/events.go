// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"svnova/models"
	"svnova/services/svn"
	"time"
	"xorm.io/xorm"
)

type Event struct {
	Key            string
	Kind           string
	RepositoryID   int64
	Site           string
	Object         string
	Trigger        string
	ExecutionID    int64
	TaskID         int64
	Success        bool
	Changed        bool
	Started        int64
	Finished       int64
	BeforeRevision int64
	AfterRevision  int64
	RevisionKnown  bool
	Created        int
	Updated        int
	Disabled       int
	Skipped        int
	Phase          string
	Reason         string
	Link           string
	HistoryLink    string
}

// Enqueue shares the business result's transaction; it never performs network I/O.
func Enqueue(tx *xorm.Session, e Event) error {
	c, err := config(tx)
	if err != nil {
		return err
	}
	if !eligible(c, e) {
		return nil
	}
	var site models.SiteSetting
	if _, err = tx.ID(1).Get(&site); err != nil {
		return err
	}
	e.Site = site.SiteName
	if e.Site == "" {
		e.Site = "SVNova"
	}
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT OR IGNORE INTO notification_delivery(event_key,kind,repository_id,object,payload,generation,status,attempts,next_attempt,created) VALUES(?,?,?,?,?,?,'pending',0,?,?)", e.Key, e.Kind, e.RepositoryID, e.Object, string(payload), c.Generation, e.Finished, time.Now().UnixMilli())
	return err
}
func eligible(c models.NotificationConfig, e Event) bool {
	if !allowed(c, e.Kind) {
		return false
	}
	if e.Kind == "svn" && e.Success && c.SVNFailureOnly {
		return false
	}
	return !e.Success || e.Changed || e.Kind == "test" || (e.Kind == "svn" && c.SVNNoChange) || (e.Kind == "ldap" && c.LDAPNoChange)
}
func LDAPEvent(run models.LDAPSyncRun) Event {
	return Event{Key: fmt.Sprintf("ldap/%d", run.ID), Kind: "ldap", Object: "LDAP", Trigger: run.Trigger, ExecutionID: run.ID, Success: run.Status == "success", Changed: run.CreatedUsers+run.UpdatedUsers+run.DisabledUsers > 0, Started: run.Started, Finished: run.Finished, Created: run.CreatedUsers, Updated: run.UpdatedUsers, Disabled: run.DisabledUsers, Skipped: run.SkippedLocalCount, Reason: LDAPReason(run.FailureReason), HistoryLink: "/-/admin/ldap/sync-history"}
}
func LDAPReason(code string) string {
	return map[string]string{"connection": "无法连接LDAP，请检查地址、端口和证书", "tls": "TLS协商失败", "bind": "服务账号绑定失败", "query": "用户搜索失败", "empty": "搜索结果为空，已保留用户", "identity": "目录缺少用户名或稳定身份属性", "duplicate_identity": "目录存在重复身份，变更已回滚", "email": "目录邮箱属性无效，变更已回滚", "identity_conflict": "目录身份冲突，变更已回滚", "interrupted": "服务中断，未提交的用户变更已回滚", "internal": "内部错误，用户变更未提交"}[code]
}

func SVNReason(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "任务执行超时，请检查远端连接及仓库大小"
	}
	var command *svn.CommandError
	if errors.As(err, &command) {
		return command.UserMessage()
	}
	for _, cause := range []struct{ match, message string }{{"cannot decrypt remote credential", "无法读取远端凭据，请检查原密钥或重新保存密码"}, {"local mirror identity changed", "本地镜像UUID改变，已停止同步"}, {"remote identity or revision changed", "远端UUID改变或修订回退，已停止同步"}, {"source must be the complete repository root", "来源地址不是完整SVN仓库根目录"}, {"mirror UUID verification failed", "镜像UUID校验失败，本地副本已保留"}, {"outside managed root", "仓库不在受管目录内，已停止操作"}} {
		if strings.Contains(err.Error(), cause.match) {
			return cause.message
		}
	}
	return "SVN操作失败，请查看任务记录并检查仓库设置及服务账号权限"
}
