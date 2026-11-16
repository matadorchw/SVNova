// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"svnova/models"
	"svnova/services/svn"
	"svnova/services/tasks"
)

func TestConversionValidationAndRecovery(t *testing.T) {
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Admin: true, Active: true}
	reader := &models.User{ID: 2, Name: "reader", Active: true}
	repo := &models.Repository{ID: 1, OrgID: 1, Name: "local", Kind: "local", Status: "ready", Path: filepath.Join(root, "repositories", "local")}
	if _, err = db.Insert(admin, reader, &models.Organization{ID: 1, Name: "org"}, repo); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Worker: &tasks.Worker{DB: db}, Config: Config{Repositories: filepath.Join(root, "repositories")}}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a.Templates = templates
	settingsPage := httptest.NewRecorder()
	a.web(settingsPage, httptest.NewRequest("GET", "/repositories/1?tab=settings&section=mirror", nil), admin, "csrf")
	if settingsPage.Code != 200 || !strings.Contains(settingsPage.Body.String(), "快速转换为镜像") || !strings.Contains(settingsPage.Body.String(), "FullAccess") {
		t.Fatal("local settings conversion form missing")
	}
	call := func(actor *models.User, body string) error {
		return a.convertMirror(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(body)), actor, repo)
	}
	for _, test := range []struct {
		actor       *models.User
		body, match string
	}{
		{reader, `{}`, "没有访问权限"},
		{admin, `{"Remote":"file:///repo","FullAccess":true}`, "HTTP/HTTPS"},
		{admin, `{"Remote":"https://example.com/repo","Interval":1,"FullAccess":true}`, "600"},
		{admin, `{"Remote":"https://example.com/repo"}`, "完整仓库"},
	} {
		err = call(test.actor, test.body)
		if err == nil || !strings.Contains(err.Error(), test.match) {
			t.Fatalf("validation %v", err)
		}
	}
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("set SVN_BIN for real conversion recovery")
	}
	runner := svn.Runner{Bin: bin}
	ctx := context.Background()
	if err = os.MkdirAll(filepath.Dir(repo.Path), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = runner.Run(ctx, "svnadmin", nil, "create", repo.Path); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "source")
	if _, err = runner.Run(ctx, "svnadmin", nil, "hotcopy", repo.Path, source); err != nil {
		t.Fatal(err)
	}
	snapshot, err := runner.SnapshotConversion(ctx, repo.Path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Insert(&models.MirrorConversion{RepositoryID: repo.ID, Snapshot: string(data)}); err != nil {
		t.Fatal(err)
	}
	if a.Worker.TryLock() {
		a.Worker.Unlock()
		t.Fatal("pending recovery allowed repository mutation")
	}
	uuid, rev, err := runner.LocalInfo(ctx, repo.Path)
	if err != nil {
		t.Fatal(err)
	}
	if err = runner.InitializeExistingMirror(ctx, repo.Path, svn.LocalURL(source), svn.Credentials{}, uuid, rev); err != nil {
		t.Fatal(err)
	}
	// Reconstruct the app to simulate restart after filesystem initialization.
	reopened := &App{DB: db, SVN: runner, Config: a.Config}
	if err = reopened.recoverConversions(); err != nil {
		t.Fatal(err)
	}
	if err = reopened.recoverConversions(); err != nil {
		t.Fatal("recovery is not idempotent", err)
	}
	stored := new(models.Repository)
	if _, err = db.ID(repo.ID).Get(stored); err != nil {
		t.Fatal(err)
	}
	if stored.Kind != "local" || stored.Remote != "" || stored.Password != "" || stored.Interval != 0 {
		t.Fatal("failed conversion changed configuration")
	}
	if _, err = runner.SnapshotConversion(ctx, repo.Path); err != nil {
		t.Fatal("recovery did not remove mirror metadata/hook", err)
	}
	pending, err := db.Count(new(models.MirrorConversion))
	if err != nil || pending != 0 {
		t.Fatal("recovery journal remains", err)
	}
}
