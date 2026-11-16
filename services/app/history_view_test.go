// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"bytes"
	"encoding/xml"
	"os"
	"strings"
	"svnova/models"
	"svnova/services/svn"
	"testing"
)

func TestHistoryTable(t *testing.T) {
	var entry svn.LogEntry
	err := xml.Unmarshal([]byte(`<logentry revision="7"><author>作者</author><msg>&lt;script&gt;第一行
第二行</msg><paths><path action="M">/中文</path><path action="A" copyfrom-path="/旧目录" copyfrom-rev="3">/新目录</path><path action="M">/重复</path><path action="D">/删除</path><path action="R">/替换</path><path action="?">/未知</path></paths></logentry>`), &entry)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Paths[1].CopyFromPath != "/旧目录" || entry.Paths[1].CopyFromRevision != 3 || entry.Paths[0].CopyFromRevision != 0 {
		t.Fatalf("paths: %+v", entry.Paths)
	}
	actions := historyActions(entry)
	if len(actions) != 5 {
		t.Fatalf("actions: %+v", actions)
	}
	for i, code := range []string{"A", "M", "D", "R", "?"} {
		if actions[i].Code != code {
			t.Fatal(actions)
		}
	}
	for i, label := range []string{"新增", "修改", "删除", "替换", "未知动作"} {
		if actions[i].Label != label {
			t.Fatal(actions)
		}
	}
	if historySummary("a\r\nb\rc\nd") != "a b c d" {
		t.Fatal("summary line breaks")
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	page := &webPage{Repo: &repositoryPage{Repository: models.Repository{Name: "repo"}, Organization: "org"}, Logs: []svn.LogEntry{entry, {Revision: 6}}}
	var out bytes.Buffer
	if err = templates.ExecuteTemplate(&out, "repo/history", page); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`aria-expanded="false"`, `id="revision-detail-7"`, `hidden`, `&lt;script&gt;第一行 第二行`, `&lt;script&gt;第一行` + "\n第二行", `/旧目录`, `r3`, `<td class="history-summary"></td>`, `暂无修订条目`, `复制来源路径`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(out.String(), "<h4>") || !strings.Contains(out.String(), `aria-label="新增"`) {
		t.Fatal("detail headings or missing action code")
	}
	if strings.Contains(out.String(), "<script>") {
		t.Fatal("unescaped message")
	}
	page.Logs = nil
	out.Reset()
	if err = templates.ExecuteTemplate(&out, "repo/history", page); err != nil || !strings.Contains(out.String(), "暂无修订记录") {
		t.Fatalf("empty: %v %s", err, out.String())
	}
}
