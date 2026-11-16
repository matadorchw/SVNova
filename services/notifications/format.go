// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package notifications

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

func escaped(value string) string {
	value = strings.Map(func(r rune) rune {
		if r < ' ' {
			return ' '
		}
		return r
	}, value)
	runes := []rune(value)
	if len(runes) > 180 {
		value = string(runes[:180]) + "…"
	}
	return strings.NewReplacer("\\", "\\\\", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "`", "\\`", "<", "&lt;", ">", "&gt;", "#", "\\#", "!", "\\!").Replace(value)
}
func Format(e Event, deliveryID int64, origin string) (string, string) {
	title := "SVN 仓库同步结果"
	if e.Kind == "ldap" {
		title = "LDAP 用户同步结果"
	}
	if e.Kind == "test" {
		title = "通知测试"
	}
	result := "❌ 失败"
	if e.Success {
		result = "✅ 成功"
	}
	trigger := map[string]string{"manual": "手动", "scheduled": "定时", "migration": "首次镜像迁移", "test": "管理员测试"}[e.Trigger]
	if trigger == "" {
		trigger = "手动"
	}
	blocks := [][]string{{"### " + title, result}, {"站点：" + escaped(e.Site), "对象：" + escaped(e.Object)}}
	stats := []string{fmt.Sprintf("触发：%s · 执行 #%d · 投递 #%d", trigger, e.ExecutionID, deliveryID)}
	if e.Kind == "svn" {
		revision := "最终修订：未知"
		if e.RevisionKnown {
			revision = fmt.Sprintf("修订：r%d → r%d · 新增 %d", e.BeforeRevision, e.AfterRevision, max(0, e.AfterRevision-e.BeforeRevision))
		}
		stats = append(stats, fmt.Sprintf("任务：#%d", e.TaskID), revision)
	}
	if e.Kind == "ldap" {
		stats = append(stats, fmt.Sprintf("新增 %d · 实际更新 %d · 停用 %d · 跳过 %d", e.Created, e.Updated, e.Disabled, e.Skipped))
	}
	blocks = append(blocks, stats)
	if !e.Success {
		reason := e.Reason
		if reason == "" {
			reason = "操作失败，请查看任务或同步历史"
		}
		lines := []string{"原因：" + escaped(reason)}
		if e.Phase != "" {
			phase := map[string]string{"moving": "移动仓库", "copying": "复制和校验仓库", "published": "发布仓库", "synchronizing": "同步修订", "revision-properties": "同步修订属性", "completed": "完成"}[e.Phase]
			if phase == "" {
				phase = "准备仓库"
			}
			lines = append(lines, "阶段："+phase)
		}
		blocks = append(blocks, lines)
	}
	zone := time.FixedZone("Asia/Shanghai", 8*3600)
	display := func(ms int64) string {
		if ms <= 0 {
			return "未知"
		}
		return time.UnixMilli(ms).In(zone).Format("2006-01-02 15:04:05")
	}
	blocks = append(blocks, []string{"开始：" + display(e.Started), "结束：" + display(e.Finished) + "（北京时间）", fmt.Sprintf("耗时：%.1f 秒", float64(max(0, e.Finished-e.Started))/1000)})
	links := []string{}
	base, err := url.Parse(origin)
	if err == nil && (base.Scheme == "https" || base.Scheme == "http") && base.Host != "" && base.User == nil {
		for _, item := range []struct{ label, path string }{{"查看对象", e.Link}, {"查看记录", e.HistoryLink}} {
			relative, er := url.Parse(item.path)
			if er == nil && strings.HasPrefix(item.path, "/") && !strings.HasPrefix(item.path, "//") && relative.Host == "" {
				target := base.ResolveReference(relative).String()
				if len(target) > 2048 {
					continue
				}
				target = strings.NewReplacer("(", "%28", ")", "%29").Replace(target)
				links = append(links, "["+item.label+"]("+target+")")
			}
		}
	}
	if len(links) > 0 {
		blocks = append(blocks, links)
	}
	rendered := make([]string, 0, len(blocks))
	for _, block := range blocks {
		rendered = append(rendered, strings.Join(block, "  \n"))
	}
	return title, strings.Join(rendered, "  \n\u200b  \n")
}
