// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"bytes"
	"context"
	"image/png"

	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/svn"
	"testing"
)

func TestServerRenderedPagesAndPermissions(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "pages.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Templates: templates}
	admin := &models.User{Name: "admin", Admin: true, Active: true}
	member := &models.User{Name: "member", Active: true}
	for _, v := range []any{admin, member, &models.Organization{ID: 1, Name: "测试组织"}, &models.Repository{ID: 1, OrgID: 1, Name: "秘密仓库", Kind: "mirror", Status: "pending"}, &models.Team{ID: 1, OrgID: 1, Name: "团队"}, &models.Membership{OrgID: 1, UserID: admin.ID, Owner: true}} {
		if _, err = db.Insert(v); err != nil {
			t.Fatal(err)
		}
	}
	for _, route := range []string{"/activity", "/", "/?org=1", "/migrate", "/organizations", "/organizations?org=1", "/organizations?org=1&team=1", "/admin/site", "/admin/organizations", "/org/create", "/admin/users", "/-/admin/ldap", "/repositories/1", "/repositories/1?tab=tasks", "/repositories/1?tab=settings"} {
		t.Run(route, func(t *testing.T) {
			w := httptest.NewRecorder()
			a.web(w, httptest.NewRequest("GET", route, nil), admin, "token")
			if w.Code != 200 {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			html := w.Body.String()
			if !strings.Contains(html, "</footer>") || strings.Contains(html, "ZgotmplZ") {
				t.Fatalf("incomplete template: %s", html)
			}
			if strings.Contains(html, `id="app"`) {
				t.Fatal("SPA shell remains")
			}
		})
	}
	for _, route := range []string{"/repositories/1?tab=tasks", "/repositories/1?tab=settings", "/organizations?org=1", "/admin/users"} {
		w := httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", route, nil), member, "token")
		if w.Code != 403 {
			t.Fatalf("%s: expected 403, got %d", route, w.Code)
		}
	}
	w := httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/", nil), member, "token")
	if strings.Contains(w.Body.String(), "秘密仓库") {
		t.Fatal("repository leaked")
	}
	for _, route := range []string{"/missing", "/repositories/nope", "/repositories/1?tab=unknown"} {
		w := httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", route, nil), admin, "token")
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s: %d", route, w.Code)
		}
	}
}
func TestRepositoryLinksPreserveUnicodeAndQueryCharacters(t *testing.T) {
	p := &webPage{Repo: &repositoryPage{Repository: models.Repository{ID: 42}}}
	link := p.Link("file", "目录/a&b+#.txt", 3)
	r := httptest.NewRequest("GET", link, nil)
	if r.URL.Query().Get("path") != "目录/a&b+#.txt" || r.URL.Query().Get("revision") != "3" {
		t.Fatal(link)
	}
}

func TestServerRenderedSVNContent(t *testing.T) {
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("SVN_BIN is not set")
	}
	root := t.TempDir()
	runner := svn.Runner{Bin: bin}
	repoPath := filepath.Join(root, "repo")
	fixture := filepath.Join(root, "fixture")
	if err := os.Mkdir(fixture, 0700); err != nil {
		t.Fatal(err)
	}
	name := "中文 & #.txt"
	text := "<script>alert('unsafe')</script>\n中文内容\n"
	if err := os.WriteFile(filepath.Join(fixture, name), []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(context.Background(), "svnadmin", nil, "create", repoPath); err != nil {
		t.Fatal(err)
	}
	target, err := svn.Target(repoPath, "")
	if err != nil {
		t.Fatal(err)
	}
	target = strings.TrimSuffix(target, "@")
	if _, err = runner.Run(context.Background(), "svn", nil, "import", fixture, target, "-m", "test fixture", "--non-interactive"); err != nil {
		t.Fatal(err)
	}
	db, err := models.Open(filepath.Join(root, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := &models.Repository{ID: 1, OrgID: 1, Name: "fixture", Path: repoPath, Status: "ready", Revision: 1}
	if _, err = db.Insert(repo); err != nil {
		t.Fatal(err)
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Templates: templates, SVN: runner}
	user := &models.User{ID: 1, Admin: true, Active: true}
	for _, tab := range []string{"tree", "file", "log"} {
		t.Run(tab, func(t *testing.T) {
			q := url.Values{"tab": {tab}, "revision": {"1"}}
			if tab == "file" {
				q.Set("path", name)
			}
			w := httptest.NewRecorder()
			a.web(w, httptest.NewRequest("GET", "/repositories/1?"+q.Encode(), nil), user, "csrf")
			if w.Code != 200 || !strings.Contains(w.Body.String(), "</footer>") {
				t.Fatalf("%d: %s", w.Code, w.Body.String())
			}
			if strings.Contains(w.Body.String(), "<script>alert") {
				t.Fatal("unescaped repository content")
			}
			if tab == "file" && !strings.Contains(w.Body.String(), "中文内容") {
				t.Fatal("missing server-rendered file")
			}
		})
	}
}

func TestRemovedRepositoryDiffIsUnavailable(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Insert(&models.Repository{ID: 1, Name: "repo", Status: "ready", Revision: 1}); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	user := &models.User{Admin: true, Active: true}
	w := httptest.NewRecorder()
	a.web(w, httptest.NewRequest("GET", "/repositories/1?tab=diff&revision=1", nil), user, "csrf")
	if w.Code != 404 {
		t.Fatalf("removed page returned %d", w.Code)
	}
	err = a.repositories(httptest.NewRecorder(), httptest.NewRequest("GET", "/api/v1/repositories/1/diff?revision=1", nil), user, []string{"1", "diff"})
	if e, ok := err.(*httpError); !ok || e.Status != 404 {
		t.Fatal("removed API still available", err)
	}
}

func TestTemplateNamesMatchUpstreamDirectories(t *testing.T) {
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"base/head", "base/head_navbar", "base/footer", "user/auth/signin", "user/auth/signin_inner", "repo/home", "repo/header", "repo/view_content", "repo/settings/options", "repo/migrate/svn", "org/home", "org/member/members", "org/team/teams", "admin/layout_head", "admin/user/list", "admin/auth/edit", "install"} {
		if templates.Lookup(name) == nil {
			t.Errorf("missing template %s", name)
		}
	}
}

func TestCurrentAvatarUsesUpstreamIdenticon(t *testing.T) {
	w := httptest.NewRecorder()
	currentAvatar(w, &models.User{Name: "admin"})
	img, err := png.Decode(bytes.NewReader(w.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 512 || img.Bounds().Dy() != 512 {
		t.Fatal("unexpected avatar size")
	}
	other := httptest.NewRecorder()
	currentAvatar(other, &models.User{Name: "admin"})
	if !bytes.Equal(w.Body.Bytes(), other.Body.Bytes()) {
		t.Fatal("avatar must be stable")
	}
	other = httptest.NewRecorder()
	currentAvatar(other, &models.User{Name: "another"})
	if bytes.Equal(w.Body.Bytes(), other.Body.Bytes()) {
		t.Fatal("different users must have distinct avatar seeds")
	}
}

func TestOrganizationAvatarRequiresMembership(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "avatars.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Insert(&models.Organization{ID: 1, Name: "private"}); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	user := &models.User{ID: 2, Name: "member"}
	for _, test := range []struct {
		user   *models.User
		status int
	}{{user, 403}, {&models.User{Admin: true}, 200}} {
		w := httptest.NewRecorder()
		a.organizationAvatar(w, httptest.NewRequest("GET", "/avatars/organizations/1", nil), test.user)
		if w.Code != test.status {
			t.Fatalf("status %d", w.Code)
		}
	}
	if _, err = db.Insert(&models.Membership{OrgID: 1, UserID: 2}); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	a.organizationAvatar(w, httptest.NewRequest("GET", "/avatars/organizations/1", nil), user)
	if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" {
		t.Fatal("member avatar unavailable")
	}
}

func TestActivityPageNavigationAndRemovedLegacyRoute(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "activity.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	user := &models.User{ID: 1, Name: "owner", Active: true}
	if _, err = db.Insert(user, &models.Organization{ID: 1, Name: "activity"}, &models.Membership{OrgID: 1, UserID: 1}, &models.Team{ID: 1, OrgID: 1, Name: "Owners", IsOwnerTeam: true}, &models.TeamMember{TeamID: 1, UserID: 1}); err != nil {
		t.Fatal(err)
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Templates: templates}
	for _, route := range []string{"/activity", "/", "/user/activity", "/org/activity"} {
		w := httptest.NewRecorder()
		a.web(w, httptest.NewRequest("GET", route, nil), user, "csrf")
		if route == "/user/activity" {
			if w.Code != 404 {
				t.Fatal("removed route remains", w.Code)
			}
			continue
		}
		if w.Code != 200 {
			t.Fatal(route, w.Code, w.Body.String())
		}
		html := w.Body.String()
		if route == "/activity" && (!strings.Contains(html, `href="/repo/migrate"`) || !strings.Contains(html, `id="activity-trend"`)) {
			t.Fatal("activity page lost migration navigation or chart")
		}
		if route == "/" && strings.Contains(html, `id="activity-trend"`) {
			t.Fatal("homepage still mounts trend")
		}
	}
}

func TestMemberAvatarProfileSizeAndAuthorization(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "avatar.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	target := &models.User{ID: 1, Name: "target", Active: true}
	if _, err = db.Insert(target); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	for _, test := range []struct {
		query string
		size  int
	}{{"", 64}, {"?size=512", 512}, {"?size=99999", 64}} {
		w := httptest.NewRecorder()
		a.memberAvatar(w, httptest.NewRequest("GET", "/avatars/users/1"+test.query, nil), &models.User{Admin: true})
		img, err := png.Decode(bytes.NewReader(w.Body.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds().Dx() != test.size || img.Bounds().Dy() != test.size {
			t.Fatal("unexpected avatar dimensions", img.Bounds())
		}
	}
	w := httptest.NewRecorder()
	a.memberAvatar(w, httptest.NewRequest("GET", "/avatars/users/1?size=512", nil), &models.User{ID: 2})
	if w.Code != 403 {
		t.Fatal("large avatar bypassed authorization", w.Code)
	}
}

func TestAvatarCacheRevalidationStillChecksMembership(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	user := &models.User{ID: 1, Name: "member", Active: true}
	if _, err = db.Insert(user, &models.Organization{ID: 1, Name: "org"}, &models.Membership{OrgID: 1, UserID: 1}); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	r := httptest.NewRequest("GET", "/avatars/organizations/1?size=200", nil)
	first := httptest.NewRecorder()
	a.organizationAvatar(first, r, user)
	img, err := png.Decode(bytes.NewReader(first.Body.Bytes()))
	if err != nil || img.Bounds().Dx() != 200 {
		t.Fatal("wrong organization avatar", err)
	}
	etag := first.Header().Get("ETag")
	if etag == "" || first.Header().Get("Cache-Control") != "private, no-cache" {
		t.Fatal("missing revalidation headers")
	}
	r.Header.Set("If-None-Match", etag)
	second := httptest.NewRecorder()
	a.organizationAvatar(second, r, user)
	if second.Code != 304 || second.Body.Len() != 0 {
		t.Fatal("unchanged avatar not revalidated", second.Code)
	}
	if _, err = db.Where("org_id=1 AND user_id=1").Delete(new(models.Membership)); err != nil {
		t.Fatal(err)
	}
	denied := httptest.NewRecorder()
	a.organizationAvatar(denied, r, user)
	if denied.Code != 403 {
		t.Fatal("cached avatar bypassed revoked membership", denied.Code)
	}
}

func TestRepositoryBreadcrumbsKeepRevisionAndEscapedPaths(t *testing.T) {
	p := &webPage{Repo: &repositoryPage{Repository: models.Repository{ID: 1, Name: "repo"}, Organization: "org"}, Path: "规范/ASR平台产品/a&b+#.txt", Tab: "file", Revision: 39}
	crumbs := p.RepositoryBreadcrumbs()
	if len(crumbs) != 4 || crumbs[3].Name != "a&b+#.txt" || !crumbs[3].Current || crumbs[3].URL != "" {
		t.Fatal(crumbs)
	}
	for i, want := range []string{"", "规范", "规范/ASR平台产品"} {
		u, err := url.Parse(crumbs[i].URL)
		if err != nil {
			t.Fatal(err)
		}
		if u.Query().Get("path") != want || u.Query().Get("revision") != "39" || u.Query().Get("tab") != "tree" {
			t.Fatal(u)
		}
	}
	p.Tab = "log"
	crumbs = p.RepositoryBreadcrumbs()
	u, _ := url.Parse(crumbs[1].URL)
	if u.Query().Get("tab") != "log" {
		t.Fatal("history breadcrumb changed view")
	}
	p.Path = ""
	crumbs = p.RepositoryBreadcrumbs()
	if len(crumbs) != 1 || !crumbs[0].Current {
		t.Fatal("root breadcrumb incorrect")
	}
	templates, err := ParseTemplates(os.DirFS("../.."))
	if err != nil {
		t.Fatal(err)
	}
	p.Path = "<directory>/a&b"
	p.Tab = "tree"
	var html bytes.Buffer
	if err := templates.ExecuteTemplate(&html, "repo/view_content", p); err != nil {
		t.Fatal(err)
	}
	text := html.String()
	if !strings.Contains(text, `aria-label="仓库路径"`) || !strings.Contains(text, "&lt;directory&gt;") || strings.Contains(text, "上级目录") || strings.Contains(text, "<directory>") {
		t.Fatal("unsafe or obsolete path navigation")
	}
}

func TestRevisionTimeUsesBeijing(t *testing.T) {
	p := &webPage{}
	for _, test := range []struct{ input, want string }{
		{"2021-11-09T01:41:15.835777Z", "2021-11-09 09:41:15"},
		{"1970-01-01T00:00:00Z", "1970-01-01 08:00:00"},
		{"2021-11-09T21:41:15Z", "2021-11-10 05:41:15"},
		{"2021-11-09T09:41:15+08:00", "2021-11-09 09:41:15"},
		{"", "—"}, {"invalid", "—"},
	} {
		if got := p.RevisionTime(test.input); got != test.want {
			t.Fatalf("%s: %s", test.input, got)
		}
	}
}
