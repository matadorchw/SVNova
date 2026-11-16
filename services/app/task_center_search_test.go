// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"testing"
)

func TestTaskCenterSearchAndCard(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "tasks.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Active: true, Admin: true}
	reader := &models.User{ID: 2, Name: "reader", Active: true}
	_, err = db.Insert(admin, reader, &models.Organization{ID: 1, Name: "Alpha"}, &models.Organization{ID: 2, Name: "Secret"}, &models.Repository{ID: 1, OrgID: 1, Name: "Tools"}, &models.Repository{ID: 2, OrgID: 2, Name: "Hidden"}, &models.Repository{ID: 3, OrgID: 1, Name: "100%_archive"}, &models.Membership{OrgID: 1, UserID: 2}, &models.Team{ID: 1, OrgID: 1, Name: "Owners", IsOwnerTeam: true}, &models.TeamMember{TeamID: 1, UserID: 2}, &models.Task{ID: 10, RepositoryID: 1, Kind: "local", State: "failed"}, &models.Task{ID: 11, RepositoryID: 1, Kind: "sync", State: "succeeded"}, &models.Task{ID: 20, RepositoryID: 2, Kind: "sync", State: "failed"}, &models.Task{ID: 30, RepositoryID: 3, Kind: "local", State: "failed"})
	if err != nil {
		t.Fatal(err)
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Templates: templates}
	for _, tc := range []struct {
		q, filter string
		total     int64
	}{{"", "", 3}, {"alpha", "", 3}, {"  TOOLS  ", "", 2}, {"100%_", "", 1}, {"%", "", 1}, {"_", "", 1}, {"#10", "", 1}, {"10", "", 2}, {"hidden", "", 0}, {"Secret", "", 0}, {"Tools", "&kind=sync&state=succeeded", 1}, {"Tools", "&kind=local&state=succeeded", 0}, {"Tools", "&org=1&repository=1", 2}, {"' OR 1=1 --", "", 0}, {"999999999999999999999999999", "", 0}} {
		page, err := a.taskCenterData(httptest.NewRequest("GET", "/tasks?q="+url.QueryEscape(tc.q)+tc.filter, nil), reader)
		if err != nil || page.Pagination.Total != tc.total {
			t.Fatalf("search %q%s: total=%d error=%v", tc.q, tc.filter, page.Pagination.Total, err)
		}
	}
	page, err := a.taskCenterData(httptest.NewRequest("GET", "/tasks?q=Tools&kind=sync", nil), reader)
	if err != nil {
		t.Fatal(err)
	}
	var html bytes.Buffer
	if err = templates.ExecuteTemplate(&html, "tasks/index", page); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`data-task-center-total>1</span>`, `list-filter-toolbar`, `name="q" value="Tools"`, `name="kind"`, `name="state"`, `ui attached table segment table-scroll`, `已筛选任务类型`} {
		if !strings.Contains(html.String(), want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, unwanted := range []string{`name="org"`, `name="repository"`, `>重置<`, `>筛选<`, `>排序<`, `共 1 个任务`} {
		if strings.Contains(html.String(), unwanted) {
			t.Errorf("unexpected %q", unwanted)
		}
	}
	w := httptest.NewRecorder()
	if err = a.taskCenterAPI(w, httptest.NewRequest("GET", "/api/v1/tasks?q=Hidden", nil), admin); err != nil {
		t.Fatal(err)
	}
	var data struct {
		Total int64
		HTML  string
	}
	if err = json.Unmarshal(w.Body.Bytes(), &data); err != nil || data.Total != 1 || !strings.Contains(data.HTML, "Hidden") {
		t.Fatalf("API: %s %v", w.Body.String(), err)
	}
	for i := 0; i < 25; i++ {
		if _, err = db.Insert(&models.Task{RepositoryID: 1, Kind: "sync", State: "succeeded"}); err != nil {
			t.Fatal(err)
		}
	}
	page, err = a.taskCenterData(httptest.NewRequest("GET", "/tasks?q=Tools&page=2", nil), reader)
	if err != nil || page.Pagination.Total != 27 || len(page.Items) != 7 {
		t.Fatalf("paging: %+v %v", page, err)
	}
	link, err := url.Parse(page.Link(1))
	if err != nil || link.Query().Get("q") != "Tools" {
		t.Fatal("search lost in pagination", err)
	}
	if _, err = db.Where("team_id=1 AND user_id=2").Delete(new(models.TeamMember)); err != nil {
		t.Fatal(err)
	}
	page, err = a.taskCenterData(httptest.NewRequest("GET", "/tasks?q=Tools", nil), reader)
	if err != nil || page.Pagination.Total != 0 || len(page.Items) != 0 {
		t.Fatal("revoked visibility leaked", err)
	}
}
