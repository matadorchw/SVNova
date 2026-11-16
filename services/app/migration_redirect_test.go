// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"testing"
)

func TestMigrationResponseLinksToCreatedRepository(t *testing.T) {
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Admin: true, Active: true}
	_, err = db.Insert(admin, &models.Organization{ID: 1, Name: "wewins"}, &models.Organization{ID: 2, Name: "api"})
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "source")
	if err = os.Mkdir(source, 0700); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Config: Config{Repositories: filepath.Join(root, "repos")}}
	for _, orgID := range []int64{1, 2} {
		body, _ := json.Marshal(map[string]any{"OrgID": orgID, "Name": "production", "Kind": "local", "Source": source, "SourceStopped": true})
		w := httptest.NewRecorder()
		if err = a.migrate(w, httptest.NewRequest("POST", "/api/v1/repositories", strings.NewReader(string(body))), admin); err != nil {
			t.Fatal(err)
		}
		var repo models.Repository
		if err = json.Unmarshal(w.Body.Bytes(), &repo); err != nil {
			t.Fatal(err)
		}
		want := "/wewins/production"
		if orgID == 2 {
			want = repoLink("api", repo.Name, repo.ID)
		}
		if w.Header().Get("Location") != want || repo.Status != "pending" {
			t.Fatalf("wrong creation destination: %s %+v", w.Header().Get("Location"), repo)
		}
		orgName := "wewins"
		if orgID == 2 {
			orgName = "api"
		}
		stored:=new(models.Repository);if _,err=db.ID(repo.ID).Get(stored);err!=nil{t.Fatal(err)}
		if stored.Path != filepath.Join(root, "repos", orgName, "production") {
			t.Fatalf("non-readable repository directory: %s", stored.Path)
		}
		duplicate := httptest.NewRecorder()
		err = a.migrate(duplicate, httptest.NewRequest("POST", "/api/v1/repositories", strings.NewReader(string(body))), admin)
		failure, ok := err.(*httpError)
		if !ok || failure.Status != 409 || !strings.Contains(failure.Message, "已存在仓库“production”") || duplicate.Header().Get("Location") != "" {
			t.Fatalf("duplicate migration did not explain name conflict: %v", err)
		}
		count, err := db.Where("repository_id=? AND state=?", repo.ID, "queued").Count(new(models.Task))
		if err != nil || count != 1 {
			t.Fatal("destination sent without queued migration")
		}
	}
}

func TestMigrationInputErrorsExplainCorrection(t *testing.T) {
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Active: true, Admin: true}
	_, _ = db.Insert(admin, &models.Organization{ID: 1, Name: "wewins"})
	a := &App{DB: db, Config: Config{Repositories: filepath.Join(root, "repos")}}
	for _, test := range []struct {
		kind, source, phrase string
		stopped              bool
		interval             int64
	}{
		{"local", "relative/path", "绝对目录路径", true, 0},
		{"local", filepath.Join(root, "source"), "勾选确认", false, 0},
		{"mirror", "svn://example.com/private", "HTTP/HTTPS", false, 0},
		{"mirror", "https://example.com/repo", "600 秒", false, 60},
	} {
		body, _ := json.Marshal(map[string]any{"OrgID": 1, "Name": "new-repo", "Kind": test.kind, "Source": test.source, "SourceStopped": test.stopped, "Interval": test.interval})
		err = a.migrate(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(string(body))), admin)
		failure, ok := err.(*httpError)
		if !ok || failure.Status != 400 || !strings.Contains(failure.Message, test.phrase) {
			t.Fatalf("unclear migration validation: %v", err)
		}
	}
}
