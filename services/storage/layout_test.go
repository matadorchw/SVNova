// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package storage

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/svn"
	"testing"
	"xorm.io/xorm"
)

func layoutFixture(t *testing.T) (*xorm.Engine, string, *models.Repository) {
	t.Helper()
	base := t.TempDir()
	db, err := models.Open(filepath.Join(base, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := filepath.Join(base, "repos")
	old := filepath.Join(root, "random-legacy")
	if err = os.MkdirAll(old, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(old, "payload"), []byte("verified-copy"), 0600); err != nil {
		t.Fatal(err)
	}
	repo := &models.Repository{ID: 1, OrgID: 1, Name: "production", Kind: "local", Path: old, Status: "failed"}
	if _, err = db.Insert(&models.Organization{ID: 1, Name: "wewins"}, repo); err != nil {
		t.Fatal(err)
	}
	return db, root, repo
}
func TestLayoutRelocatesLegacyAndStagingIdempotently(t *testing.T) {
	db, root, repo := layoutFixture(t)
	old := repo.Path
	if err := os.Mkdir(old+".importing", 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old+".importing", "partial"), []byte("retain"), 0600); err != nil {
		t.Fatal(err)
	}
	runner := svn.Runner{Bin: t.TempDir()}
	for n := 0; n < 2; n++ {
		if err := Relocate(context.Background(), db, root, runner); err != nil {
			t.Fatal(err)
		}
	}
	stored := new(models.Repository)
	_, _ = db.ID(repo.ID).Get(stored)
	want := filepath.Join(root, "wewins", "production")
	if stored.Path != want {
		t.Fatalf("path %s", stored.Path)
	}
	data, err := os.ReadFile(filepath.Join(want, "payload"))
	if err != nil || string(data) != "verified-copy" {
		t.Fatal("repository changed")
	}
	if _, err = os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("old path retained")
	}
	if _, err = os.Stat(filepath.Join(StagingPath(root, repo.ID), "partial")); err != nil {
		t.Fatal("partial staging copy lost")
	}
	count, _ := db.Count(new(models.RepositoryPathMove))
	if count != 0 {
		t.Fatal("completed journal retained")
	}
	exists, _ := db.ID(3).Exist(new(models.Migration))
	if !exists {
		t.Fatal("layout version not persisted")
	}
}
func TestLayoutResumesRenameBeforeDatabaseCommit(t *testing.T) {
	db, root, repo := layoutFixture(t)
	target, _ := RepositoryPath(root, "wewins", repo.Name)
	m := &models.RepositoryPathMove{RepositoryID: repo.ID, OldPath: repo.Path, NewPath: target, Token: strings.Repeat("a", 64)}
	if _, err := db.Insert(m); err != nil {
		t.Fatal(err)
	}
	if err := moveDirectory(m.OldPath, m.NewPath, m.Token); err != nil {
		t.Fatal(err)
	}
	if err := Relocate(context.Background(), db, root, svn.Runner{}); err != nil {
		t.Fatal(err)
	}
	stored := new(models.Repository)
	_, _ = db.ID(repo.ID).Get(stored)
	if stored.Path != target {
		t.Fatal("journal not resumed")
	}
	if markerMatches(target, m.Token) {
		t.Fatal("move marker was not removed")
	}
	// Also recover a crash after the database commit and marker cleanup.
	if _, err := db.Insert(m); err != nil {
		t.Fatal(err)
	}
	if err := Relocate(context.Background(), db, root, svn.Runner{}); err != nil {
		t.Fatal(err)
	}
}
func TestLayoutRefusesOccupiedTargetWithoutMovingSource(t *testing.T) {
	db, root, repo := layoutFixture(t)
	target, _ := RepositoryPath(root, "wewins", repo.Name)
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	if err := Relocate(context.Background(), db, root, svn.Runner{}); err == nil {
		t.Fatal("occupied target accepted")
	}
	if _, err := os.Stat(filepath.Join(repo.Path, "payload")); err != nil {
		t.Fatal("source changed before preflight finished")
	}
	exists, _ := db.ID(3).Exist(new(models.Migration))
	if exists {
		t.Fatal("failed preflight changed layout version")
	}
}
func TestLayoutRejectsCaseCollisionAndUnsafeWindowsNames(t *testing.T) {
	for _, name := range []string{"CON", "nul.txt", "COM1", "Lpt9.repo", "production.", "../escape", "name with space"} {
		if ValidateName(name) == nil {
			t.Fatalf("unsafe name accepted: %s", name)
		}
	}
	for _, name := range []string{"production", "production.importing", "my-repo", "COM10", "repo_1"} {
		if err := ValidateName(name); err != nil {
			t.Fatal(err)
		}
	}
	db, root, repo := layoutFixture(t)
	_, _ = db.Insert(&models.Repository{ID: 2, OrgID: 1, Name: "Production", Path: filepath.Join(root, "another"), Status: "pending"})
	if err := Relocate(context.Background(), db, root, svn.Runner{Bin: t.TempDir()}); err == nil {
		t.Fatal("case-insensitive collision accepted")
	}
	if _, err := os.Stat(repo.Path); err != nil {
		t.Fatal("preflight modified first repository")
	}
}

func TestLayoutRejectsSharedOldDirectory(t *testing.T) {
	db, root, repo := layoutFixture(t)
	_, err := db.Insert(&models.Repository{ID: 2, OrgID: 1, Name: "other", Path: repo.Path, Status: "failed"})
	if err != nil {
		t.Fatal(err)
	}
	if err = Relocate(context.Background(), db, root, svn.Runner{Bin: t.TempDir()}); err == nil {
		t.Fatal("shared old directory accepted")
	}
	if _, err = os.Stat(repo.Path); err != nil {
		t.Fatal("source moved before conflict detection")
	}
}

func TestLayoutMovesMissingQueuedDestinationAndAvoidsStagingNameCollision(t *testing.T) {
	db, root, repo := layoutFixture(t)
	if err := os.Remove(filepath.Join(repo.Path, "payload")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(repo.Path); err != nil {
		t.Fatal(err)
	}
	repo.Status = "pending"
	repo.Name = "production.importing"
	_, _ = db.ID(repo.ID).Cols("name", "status").Update(repo)
	if err := Relocate(context.Background(), db, root, svn.Runner{}); err != nil {
		t.Fatal(err)
	}
	target, _ := RepositoryPath(root, "wewins", repo.Name)
	stored := new(models.Repository)
	_, _ = db.ID(repo.ID).Get(stored)
	if stored.Path != target || strings.EqualFold(target, StagingPath(root, repo.ID)) {
		t.Fatal("queued destination or staging namespace invalid")
	}
}
func TestLayoutRejectsLinkedOrganizationParent(t *testing.T) {
	db, root, repo := layoutFixture(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "wewins")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := Relocate(context.Background(), db, root, svn.Runner{}); err == nil {
		t.Fatal("linked parent accepted")
	}
	if _, err := os.Stat(repo.Path); err != nil {
		t.Fatal("source modified")
	}
}
func TestLayoutPreservesRealSVNHistory(t *testing.T) {
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("set SVN_BIN for real SVN layout verification")
	}
	db, root, repo := layoutFixture(t)
	if err := os.Remove(filepath.Join(repo.Path, "payload")); err != nil {
		t.Fatal(err)
	}
	runner := svn.Runner{Bin: bin}
	ctx := context.Background()
	if _, err := runner.Run(ctx, "svnadmin", nil, "create", repo.Path); err != nil {
		t.Fatal(err)
	}
	content := filepath.Join(t.TempDir(), "content")
	if err := os.Mkdir(content, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(content, "中文.txt"), []byte("history preserved"), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.ToSlash(repo.Path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	sourceURL := (&url.URL{Scheme: "file", Path: path}).String()
	if _, err := runner.Run(ctx, "svn", nil, "import", content, sourceURL, "-m", "initial", "--non-interactive"); err != nil {
		t.Fatal(err)
	}
	uuid, rev, err := runner.LocalInfo(ctx, repo.Path)
	if err != nil {
		t.Fatal(err)
	}
	repo.UUID = uuid
	repo.Revision = rev
	repo.Status = "ready"
	_, _ = db.ID(repo.ID).Cols("uuid", "revision", "status").Update(repo)
	if err = Relocate(ctx, db, root, runner); err != nil {
		t.Fatal(err)
	}
	target, _ := RepositoryPath(root, "wewins", repo.Name)
	actual, head, err := runner.LocalInfo(ctx, target)
	if err != nil || actual != uuid || head != rev {
		t.Fatal("SVN identity or revision changed")
	}
	if _, err = runner.Run(ctx, "svnadmin", nil, "verify", target); err != nil {
		t.Fatal(err)
	}
}
