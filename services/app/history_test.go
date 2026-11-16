// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/svn"
	"testing"
)

func TestHistoryCursorPagination(t *testing.T) {
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("set SVN_BIN for real SVN integration")
	}
	runner := svn.Runner{Bin: bin}
	root := filepath.Join(t.TempDir(), "repo")
	ctx := context.Background()
	if _, err := runner.Run(ctx, "svnadmin", nil, "create", root); err != nil {
		t.Fatal(err)
	}
	// A sparse path history also verifies that cursors are not offsets or page numbers.
	var dump strings.Builder
	dump.WriteString("SVN-fs-dump-format-version: 2\n\n")
	for rev := 1; rev <= 66; rev++ {
		fmt.Fprintf(&dump, "Revision-number: %d\nProp-content-length: 10\nContent-length: 10\n\nPROPS-END\n\n", rev)
		name := fmt.Sprintf("other%d", rev)
		if rev == 1 {
			name = "中文"
		} else if rev%2 == 1 {
			name = fmt.Sprintf("中文/dir%d", rev)
		}
		fmt.Fprintf(&dump, "Node-path: %s\nNode-kind: dir\nNode-action: add\n\n\n", name)
	}
	if _, err := runner.Run(ctx, "svnadmin", strings.NewReader(dump.String()), "load", root); err != nil {
		t.Fatal(err)
	}
	a := &App{SVN: runner}
	repo := &models.Repository{Path: root}
	request := func(query string) historyPage {
		t.Helper()
		page, err := a.repositoryHistory(httptest.NewRequest("GET", "/history?"+query, nil), repo, "中文", 65)
		if err != nil {
			t.Fatal(err)
		}
		return page
	}
	first := request("")
	if len(first.Entries) != 10 || first.Entries[0].Revision != 65 || first.Entries[9].Revision != 47 || first.OlderCursor != 46 || first.NewerCursor != 0 {
		t.Fatalf("first: %+v", first)
	}
	middle := request("cursor=46&direction=older")
	if len(middle.Entries) != 10 || middle.Entries[0].Revision != 45 || middle.Entries[9].Revision != 27 || middle.OlderCursor != 26 || middle.NewerCursor != 46 {
		t.Fatalf("middle: %+v", middle)
	}
	last := request("cursor=6&direction=older")
	if len(last.Entries) != 3 || last.Entries[0].Revision != 5 || last.Entries[2].Revision != 1 || last.OlderCursor != 0 || last.NewerCursor != 6 {
		t.Fatalf("last: %+v", last)
	}
	back := request("cursor=46&direction=newer")
	if len(back.Entries) != 10 || back.Entries[0].Revision != 65 || back.Entries[9].Revision != 47 || back.NewerCursor != 0 || back.OlderCursor != 46 {
		t.Fatalf("back: %+v", back)
	}
	for _, query := range []string{"cursor=0", "cursor=66", "cursor=bad", "direction=sideways"} {
		if _, err := a.repositoryHistory(httptest.NewRequest("GET", "/?"+query, nil), repo, "", 65); err == nil {
			t.Fatalf("accepted %s", query)
		}
	}
	page := &webPage{Repo: &repositoryPage{Organization: "组织", Repository: models.Repository{Name: "repo"}}, Path: "中文/file.txt", Revision: 65}
	u := page.HistoryURL(6, "older")
	if !strings.Contains(u, "revision=65") || !strings.Contains(u, "cursor=6") || !strings.Contains(u, "path=%E4%B8%AD%E6%96%87%2Ffile.txt") {
		t.Fatal(u)
	}
	db, err := models.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo.ID, repo.Name, repo.Status, repo.Revision = 1, "fixture", "ready", 66
	if _, err = db.Insert(repo); err != nil {
		t.Fatal(err)
	}
	a.DB = db
	w := httptest.NewRecorder()
	if err = a.repositories(w, httptest.NewRequest("GET", "/api/v1/repositories/1/log?revision=65&path=%E4%B8%AD%E6%96%87", nil), &models.User{Admin: true, Active: true}, []string{"1", "log"}); err != nil {
		t.Fatal(err)
	}
	var entries []svn.LogEntry
	if err = json.Unmarshal(w.Body.Bytes(), &entries); err != nil || len(entries) != 10 || w.Header().Get("X-Older-Cursor") != "46" || w.Header().Get("X-Revision") != "65" {
		t.Fatalf("API: %s %v", w.Body.String(), err)
	}
	if err = a.repositories(httptest.NewRecorder(), httptest.NewRequest("GET", "/api/v1/repositories/1/log", nil), &models.User{ID: 99}, []string{"1", "log"}); err == nil {
		t.Fatal("unauthorized history exposed")
	}
	empty, err := a.repositoryHistory(httptest.NewRequest("GET", "/", nil), repo, "", 0)
	if err != nil || len(empty.Entries) != 0 || empty.Entries == nil {
		t.Fatalf("empty: %+v %v", empty, err)
	}
}
