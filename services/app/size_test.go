// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"bytes"
	"os"
	"strings"
	"svnova/models"
	"svnova/services/svn"
	"testing"
)

func TestSizeFormattingAndRepositoryTemplates(t *testing.T) {
	for _, test := range []struct {
		size int64
		want string
	}{{0, "0 B"}, {1, "1 B"}, {1024, "1.0 KiB"}, {1536, "1.5 KiB"}, {1048576, "1.0 MiB"}, {1073741824, "1.0 GiB"}, {-1, "—"}} {
		if got := fileSize(test.size); got != test.want {
			t.Fatalf("size %d: %s", test.size, got)
		}
	}
	if got := repositorySize(models.Repository{}); got != "待统计" {
		t.Fatal(got)
	}
	if got := repositorySize(models.Repository{SizeUpdated: 1}); got != "0 B" {
		t.Fatal(got)
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	repo := repositoryPage{Repository: models.Repository{ID: 1, Name: "repo", Kind: "mirror", Remote: "https://example.com/source", Status: "ready", Size: 1536, SizeUpdated: 1}, Organization: "org"}
	page := &webPage{User: &models.User{Active: true}, Org: &organizationPage{Organization: models.Organization{ID: 1, Name: "org"}}, Team: &models.Team{ID: 1, Name: "team", IsOwnerTeam: true}, DashboardView: "repositories", Repo: &repo, Repositories: []repositoryPage{repo}, Tab: "tasks"}
	for _, name := range []string{"repo/header", "user/dashboard/repolist", "user/profile", "admin/user/details", "org/home", "org/team/repositories"} {
		var html bytes.Buffer
		if err := templates.ExecuteTemplate(&html, name, page); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		wantRemote := name != "repo/header" && name != "user/dashboard/repolist"
		if strings.Contains(html.String(), repo.Remote) != wantRemote {
			t.Fatalf("unexpected remote visibility in %s", name)
		}
		if strings.Contains(html.String(), "大小 1.5 KiB") {
			t.Fatalf("redundant size label in %s", name)
		}
		if !strings.Contains(html.String(), "1.5 KiB") {
			t.Fatalf("size missing in %s", name)
		}
	}
	page.Tab = "tree"
	page.Entries = []svn.Entry{{Kind: "file", Name: "file", Size: 1536}, {Kind: "dir", Name: "directory"}}
	var html bytes.Buffer
	if err := templates.ExecuteTemplate(&html, "repo/view_content", page); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "1.5 KiB") || !strings.Contains(html.String(), "1536 字节") || !strings.Contains(html.String(), "<td>—</td>") {
		t.Fatal("file size or directory marker missing")
	}
}

func TestRepositoryFilesSidebar(t *testing.T) {
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	repo := repositoryPage{Repository: models.Repository{ID: 1, Name: "repo", Kind: "mirror", Status: "ready", Remote: "https://example.com/very/long/repository", Size: 1536, SizeUpdated: 1}, Organization: "org"}
	page := &webPage{Repo: &repo, User: &models.User{Active: true}}
	for _, tab := range []string{"tree", "file"} {
		page.Tab = tab
		var html bytes.Buffer
		if err := templates.ExecuteTemplate(&html, "repo/home", page); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"repo-file-layout", `aria-label="仓库信息"`, "1.5 KiB", repo.Remote, "#database"} {
			if !strings.Contains(html.String(), want) {
				t.Fatalf("%s missing %s", tab, want)
			}
		}
		html.Reset()
		if err := templates.ExecuteTemplate(&html, "repo/header", page); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(html.String(), "repository-size") {
			t.Fatal("file header duplicates sidebar size")
		}
	}
	repo.Kind = "local"
	page.Tab = "tree"
	var html bytes.Buffer
	if err := templates.ExecuteTemplate(&html, "repo/home", page); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html.String(), repo.Remote) || strings.Contains(html.String(), "镜像地址") {
		t.Fatal("local repository shows remote")
	}
	repo.SizeUpdated = 0
	html.Reset()
	if err := templates.ExecuteTemplate(&html, "repo/home", page); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "待统计") {
		t.Fatal("unknown size missing")
	}
	for _, tab := range []string{"tasks", "log", "settings"} {
		page.Tab = tab
		html.Reset()
		if err := templates.ExecuteTemplate(&html, "repo/home", page); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(html.String(), "repo-file-sidebar") {
			t.Fatal("sidebar leaked into " + tab)
		}
	}
}

func TestLocalRepositoryListHidesRemote(t *testing.T) {
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	var html bytes.Buffer
	if err := templates.ExecuteTemplate(&html, "repo/list_remote", models.Repository{Kind: "local", Remote: "https://example.com/stale"}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(html.String()) != "" {
		t.Fatal("local repository shows remote")
	}
}
