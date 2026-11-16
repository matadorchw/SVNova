// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"testing"
	"time"
)

func TestRepositoryListsAndFeedUseTenItems(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "pages.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Admin: true, Active: true}
	if _, err := db.Insert(admin, &models.Organization{ID: 1, Name: "org"}, &models.Team{ID: 1, OrgID: 1, Name: "Owners", IsOwnerTeam: true}, &models.Membership{OrgID: 1, UserID: 1}, &models.TeamMember{TeamID: 1, UserID: 1}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 21; i++ {
		repo := &models.Repository{OrgID: 1, Name: fmt.Sprintf("repo%02d", i), Kind: "mirror", Status: "ready"}
		if _, err := db.Insert(repo); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Insert(&models.Task{RepositoryID: repo.ID, Kind: "sync", State: "succeeded", Started: time.Now().Unix() - 60}); err != nil {
			t.Fatal(err)
		}
	}
	a := &App{DB: db}
	for _, kind := range []string{"dashboard", "organization", "team", "profile", "admin"} {
		t.Run(kind, func(t *testing.T) {
			page := &webPage{User: admin}
			target := "/?page=2&repo_page=2"
			switch kind {
			case "dashboard":
				page.Page = "repositories"
			case "organization":
				page.Page = "organizations"
				target = "/organizations?org=1&page=2"
			case "team":
				page.Page = "organizations"
				target = "/organizations?org=1&team=1&team_view=repositories&page=2"
			case "profile":
				page.Page = "user-profile"
				target = "/user/profile?page=2"
			case "admin":
				page.Page = "admin/users/details"
				target = "/-/admin/users/1?user_id=1&page=2"
			}
			if err := a.webData(httptest.NewRequest("GET", target, nil), page); err != nil {
				t.Fatal(err)
			}
			rows := page.Repositories
			pagination := page.RepositoryPagination
			if kind == "profile" {
				pagination = page.ProfilePagination
			}
			if kind == "team" {
				rows = page.FilteredTeamRepositories()
				if len(page.TeamRepositoryList()) != 21 {
					t.Fatal("team count truncated")
				}
			}
			if len(rows) != 10 || pagination.Total != 21 || pagination.Page != 2 || pagination.Pages != 3 {
				t.Fatalf("incorrect %s page: count=%d total=%d pages=%d", kind, len(rows), pagination.Total, pagination.Pages)
			}
			if kind == "organization" && page.OrgRepoCount != 21 {
				t.Fatal("organization total truncated")
			}
			if kind == "dashboard" {
				if len(page.Feed) != 10 || !page.FeedHasNext || page.Feed[0].TaskID != 11 || page.Feed[9].TaskID != 2 {
					t.Fatal("feed page overlaps or wrong size")
				}
				u, _ := url.Parse(page.NextFeedURL())
				if u.Query().Get("repo_page") != "2" {
					t.Fatal("feed navigation reset repository page")
				}
				u, _ = url.Parse(page.RepositoryPagination.Next)
				if u.Query().Get("page") != "2" {
					t.Fatal("repository navigation reset feed page")
				}
			}
		})
	}
	last := &webPage{User: admin, Page: "repositories"}
	if err := a.webData(httptest.NewRequest("GET", "/?page=3&repo_page=9999", nil), last); err != nil {
		t.Fatal(err)
	}
	if len(last.Feed) != 1 || last.FeedHasNext || len(last.Repositories) != 1 || last.RepoPage != 3 || last.RepoHasNext {
		t.Fatal("last page not clamped")
	}
	reader := &models.User{ID: 2, Name: "reader", Active: true}
	if _, err := db.Insert(reader); err != nil {
		t.Fatal(err)
	}
	hidden := &webPage{User: reader, Page: "repositories"}
	if err := a.webData(httptest.NewRequest("GET", "/", nil), hidden); err != nil {
		t.Fatal(err)
	}
	if hidden.RepoMatchTotal != 0 || len(hidden.Feed) != 0 || len(hidden.Repositories) != 0 {
		t.Fatal("pagination leaked unauthorized repositories")
	}
}

func TestRepositoryPaginationLinksAndPagerTemplates(t *testing.T) {
	target := repositoryListLink(teamRepositoriesLink("org", "new", 7), "a&b", 2)
	u, err := url.Parse(target)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("q") != "a&b" || u.Query().Get("team_id") != "7" || u.Query().Get("page") != "2" {
		t.Fatal(target)
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	rows := make([]repositoryPage, 21)
	_, pagination := paginateRepositoryRows(rows, 1, func(page int) string { return repositoryListLink("/org", "", page) })
	p := &webPage{User: &models.User{Active: true}, DashboardView: "repositories", FeedPage: 1, FeedHasNext: true, RepoPage: 1, RepoHasNext: true, RepositoryPagination: pagination}
	for _, name := range []string{"user/dashboard/feeds", "user/dashboard/repolist"} {
		var html bytes.Buffer
		if err := templates.ExecuteTemplate(&html, name, p); err != nil {
			t.Fatal(err)
		}
		text := html.String()
		if !strings.Contains(text, `aria-disabled="true"`) || !strings.Contains(text, `aria-current="page"`) {
			t.Fatal("pager boundary markup missing")
		}
		if name == "user/dashboard/repolist" && (!strings.Contains(text, "octicons.svg#double-chevron-left") || strings.Contains(text, "octicons.svg#skip") || !strings.Contains(text, `aria-label="末页"`) || !strings.Contains(text, `aria-label="首页"`)) {
			t.Fatal("compact pager missing first/last")
		}
	}
}
