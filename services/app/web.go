// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"svnova/models"
	"svnova/services/directory"
	"svnova/services/svn"
	"time"
)

type organizationPage struct {
	TeamCount, MemberCount, RepoCount int64
	models.Organization
	CanManage bool
}
type repositoryPage struct {
	models.Repository
	CanManage    bool
	Organization string
}
type webPage struct {
	SettingsSection                                            string
	SelectedTask                                               bool
	TaskCenter                                                 taskCenterPage
	ProfileUser                                                models.User
	ProfilePagination                                          adminUsersPage
	RepositoryPagination                                       adminUsersPage
	MemberPagination                                           adminUsersPage
	Notifications                                              notificationPage
	AdminUsers                                                 adminUsersPage
	EditUser                                                   models.User
	UserDeleteDisabledReason                                   string
	UserAuthSource                                             string
	RepositoryTaskActive                                       bool
	Migration                                                  migrationStatus
	MigrationPanel                                             bool
	Heatmap                                                    heatmapResponse
	HeatmapURL                                                 string
	Feed                                                       []dashboardEvent
	FeedPage, RepoPage                                         int
	FeedHasNext, RepoHasNext                                   bool
	DashboardView, KindFilter, DateFilter                      string
	RepoTotal, LocalRepoTotal, MirrorRepoTotal, RepoMatchTotal int
	SidebarOrgs                                                []organizationPage
	NewTeam                                                    bool
	TeamView                                                   string
	FilteredMembers                                            []models.Membership
	FilteredTeams                                              []organizationTeam
	FilteredTeamMembers                                        []models.TeamMember
	OwnerIDs                                                   map[int64]bool
	MigrationOrg                                               int64
	OrgTab                                                     string
	OrgRepoCount                                               int
	TeamStats                                                  []organizationTeam
	Sort                                                       string
	SiteName                                                   string
	ContextOrg                                                 int64
	ContextName                                                string
	User                                                       *models.User
	CSRF, Page, Tab, Search, Path, Parent, Error, Content      string
	Repositories                                               []repositoryPage
	Organizations                                              []organizationPage
	Users                                                      []models.User
	Repo                                                       *repositoryPage
	Org                                                        *organizationPage
	Team                                                       *models.Team
	Teams                                                      []models.Team
	Members                                                    []models.Membership
	TeamMembers                                                []models.TeamMember
	Grants                                                     []models.TeamRepository
	Tasks                                                      []models.Task
	Entries                                                    []svn.Entry
	Logs                                                       []svn.LogEntry
	History                                                    historyPage
	LDAP                                                       models.LDAPConfig
	LDAPSchedule                                               models.LDAPSchedule
	AuthHistory                                                authHistoryPage
	Revision                                                   int64
}

func (p *webPage) Username(id int64) string {
	for _, u := range p.Users {
		if u.ID == id {
			return u.Name
		}
	}
	return strconv.FormatInt(id, 10)
}
func (p *webPage) Reponame(id int64) string {
	for _, r := range p.Repositories {
		if r.ID == id {
			return r.Name
		}
	}
	return strconv.FormatInt(id, 10)
}
func (p *webPage) Link(tab, file string, revision int64) string {
	target := repoLink(p.Repo.Organization, p.Repo.Name, p.Repo.ID)
	if !reservedOrganizationName(p.Repo.Organization) && (tab == "settings" || tab == "tasks") {
		return target + "/" + tab
	}
	return fmt.Sprintf("%s?%s", target, url.Values{"tab": {tab}, "path": {file}, "revision": {strconv.FormatInt(revision, 10)}}.Encode())
}

type repositoryBreadcrumb struct {
	Name    string
	URL     string
	Current bool
}

func (p *webPage) RepositoryBreadcrumbs() []repositoryBreadcrumb {
	if p.Repo == nil {
		return nil
	}
	tab := "tree"
	if p.Tab == "log" {
		tab = "log"
	}
	crumbs := []repositoryBreadcrumb{{Name: p.Repo.Name, URL: p.Link(tab, "", p.Revision)}}
	current := ""
	for _, name := range strings.Split(strings.Trim(p.Path, "/"), "/") {
		if name == "" {
			continue
		}
		current = path.Join(current, name)
		crumbs = append(crumbs, repositoryBreadcrumb{Name: name, URL: p.Link(tab, current, p.Revision)})
	}
	crumbs[len(crumbs)-1].Current = true
	crumbs[len(crumbs)-1].URL = ""
	return crumbs
}

func (p *webPage) EntryLink(e svn.Entry) string {
	tab := "file"
	if e.Kind == "dir" {
		tab = "tree"
	}
	return p.Link(tab, path.Join(p.Path, e.Name), p.Revision)
}
func (p *webPage) Time(t int64) string {
	if t == 0 {
		return "—"
	}
	return time.Unix(t, 0).In(time.FixedZone("Asia/Shanghai", 8*3600)).Format("2006-01-02 15:04:05")
}
func (p *webPage) RevisionTime(value string) string {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "—"
	}
	return parsed.In(time.FixedZone("Asia/Shanghai", 8*3600)).Format("2006-01-02 15:04:05")
}
func (p *webPage) State(s string) string {
	if v, ok := map[string]string{"ready": "可浏览", "pending": "等待迁移", "failed": "失败", "queued": "排队中", "running": "进行中", "succeeded": "成功"}[s]; ok {
		return v
	}
	return s
}
func (p *webPage) TaskError(s string) string {
	if s == "" {
		return "任务失败，但未记录具体原因；请检查仓库设置后重试"
	}
	for _, cause := range []struct{ match, message string }{
		{"没有访问权限", "没有访问权限，转换未完成，请检查后重试"},
		{"本地仓库 UUID 或修订与记录不一致", "本地仓库 UUID 或修订与记录不一致，转换未完成，请检查后重试"},
		{"远端地址不是完整 SVN 仓库根目录", "远端地址不是完整 SVN 仓库根目录，转换未完成，请检查后重试"},
		{"本地与远端修订 r0 属性不一致", "本地与远端修订 r0 属性不一致，转换未完成，请检查后重试"},
		{"远端与本地 UUID 不一致", "远端与本地 UUID 不一致，转换未完成，请检查后重试"},
		{"远端修订落后于本地", "远端修订落后于本地，转换未完成，请检查后重试"},
		{"本地仓库已有 svnsync 元数据", "本地仓库已有 svnsync 元数据，转换未完成，请检查后重试"},
		{"本地仓库已有提交或修订属性钩子", "本地仓库已有提交或修订属性钩子，转换未完成，请检查后重试"},
		{"镜像元数据初始化校验失败", "镜像元数据初始化校验失败，转换未完成，请检查后重试"},
		{"校验期间仓库历史发生变化", "校验期间仓库历史发生变化，转换未完成，请检查后重试"},
		{"仓库状态已改变", "仓库状态已改变，转换未完成，请检查后重试"},
		{"转换未完成且恢复尚未完成", "转换未完成且恢复尚未完成，转换未完成，请检查后重试"},
		{"镜像转换校验超时", "镜像转换校验超时，转换未完成，请检查后重试"},
		{"镜像转换校验已中断", "镜像转换校验已中断，转换未完成，请检查后重试"},
		{"mirror conversion interrupted", "镜像转换已中断，原配置已恢复，请手动重试"},
		{"context canceled", "任务已中断，原配置保留，请手动重试"},
		{"context deadline exceeded", "任务执行超时，请检查远端连接及仓库大小后重试"},
		{"cannot decrypt remote credential", "无法解密已保存的远端凭据，请恢复原密钥或重新保存镜像密码"},
		{"source must be the complete repository root", "来源地址不是完整 SVN 仓库根目录，请修改镜像地址"},
		{"local mirror identity changed", "本地镜像 UUID 已改变，同步已停止，请核查仓库目录"},
		{"remote identity or revision changed", "远端 UUID 改变或修订回退，同步已停止，本地副本已保留"},
		{"mirror UUID verification failed", "镜像 UUID 校验失败，请检查来源仓库，保留现有副本"},
		{"migration destination exists", "迁移目标目录已存在，已停止以防覆盖，请由管理员核查后恢复"},
		{"staging directory exists", "迁移临时目录已存在，源仓库已保留，请由管理员核查后恢复"},
		{"source changed during migration", "迁移期间源仓库发生变化，已保留两份副本，请停止源写入后核查"},
		{"source changed after publication", "发布后源仓库发生变化，源副本已保留，需要管理员核查"},
		{"source cleanup required", "仓库已发布，但源目录清理失败，源副本已保留，需要管理员核查"},
	} {
		if strings.Contains(s, cause.match) {
			return cause.message
		}
	}
	return (&svn.CommandError{Reason: s}).UserMessage()
}
func (a *App) web(w http.ResponseWriter, r *http.Request, u *models.User, csrf string) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "不支持此操作", 405)
		return
	}
	resolved, err := a.resolvePage(r)
	if err != nil {
		var e *httpError
		if errors.As(err, &e) {
			http.Error(w, e.Message, e.Status)
		} else {
			http.Error(w, "页面加载失败", 500)
		}
		return
	}
	r = resolved
	p := &webPage{User: u, CSRF: csrf, Page: strings.Trim(r.URL.Path, "/"), Search: r.URL.Query().Get("q")}
	if p.Page == "" {
		p.Page = "repositories"
	}
	if err := a.webData(r, p); err != nil {
		var e *httpError
		var command *svn.CommandError
		switch {
		case errors.As(err, &e):
			http.Error(w, e.Message, e.Status)
			return
		case errors.As(err, &command):
			p.Error = command.UserMessage()
		default:
			p.Error = "页面加载失败，请刷新重试"
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	a.page(w, "app", p)
}
func (a *App) webData(r *http.Request, p *webPage) error {
	if p.Page == "task-center" {
		var err error
		p.TaskCenter, err = a.taskCenterData(r, p.User)
		return err
	}
	var orgs []models.Organization
	if p.User.Admin {
		if err := a.DB.Find(&orgs); err != nil {
			return err
		}
	} else {
		if err := a.DB.SQL("SELECT o.* FROM organization o JOIN membership m ON m.org_id=o.id WHERE m.user_id=?", p.User.ID).Find(&orgs); err != nil {
			return err
		}
	}
	for _, o := range orgs {
		p.Organizations = append(p.Organizations, organizationPage{Organization: o, CanManage: models.Owner(a.DB, p.User, o.ID)})
	}
	if p.Page == "activity-trend" {
		return nil
	}
	var repos []models.Repository
	if err := a.DB.Asc("id").Find(&repos); err != nil {
		return err
	}
	for _, repo := range repos {
		if models.CanRead(a.DB, p.User, &repo) {
			name := ""
			for _, o := range orgs {
				if o.ID == repo.OrgID {
					name = o.Name
				}
			}
			p.Repositories = append(p.Repositories, repositoryPage{repo, models.Owner(a.DB, p.User, repo.OrgID), name})
		}
	}
	if p.Page == "repositories" {
		p.ContextName = p.User.Name
		p.ContextOrg = id(r.URL.Query().Get("org"))
		if p.ContextOrg != 0 {
			found := false
			for _, o := range p.Organizations {
				if o.ID == p.ContextOrg {
					p.ContextName = o.Name
					found = true
				}
			}
			if !found {
				return forbidden()
			}
			filtered := p.Repositories[:0]
			for _, repo := range p.Repositories {
				if repo.OrgID == p.ContextOrg {
					filtered = append(filtered, repo)
				}
			}
			p.Repositories = filtered
		}
		return a.dashboardData(r, p)
	}
	if strings.HasPrefix(p.Page, "repositories/") {
		parts := strings.Split(p.Page, "/")
		if len(parts) != 2 {
			return &httpError{404, "页面不存在"}
		}
		n, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || n <= 0 {
			return &httpError{404, "仓库不存在"}
		}
		for i := range p.Repositories {
			if p.Repositories[i].ID == n {
				p.Repo = &p.Repositories[i]
				break
			}
		}
		if p.Repo == nil {
			return forbidden()
		}
		p.Page = "repository"
		p.Tab = r.URL.Query().Get("tab")
		if p.Tab == "" {
			p.Tab = "tree"
			if p.Repo.Status != "ready" {
				p.Tab = "tasks"
				p.MigrationPanel = true
			}
		}
		p.Path = r.URL.Query().Get("path")
		p.Parent = path.Dir(p.Path)
		if p.Parent == "." {
			p.Parent = ""
		}
		p.Revision = p.Repo.Revision
		if rev := r.URL.Query().Get("revision"); rev != "" {
			p.Revision, err = strconv.ParseInt(rev, 10, 64)
			if err != nil || p.Revision < 0 || p.Revision > p.Repo.Revision {
				return &httpError{400, "无效修订"}
			}
		}
		switch p.Tab {
		case "settings":
			p.SettingsSection = r.URL.Query().Get("section")
			if p.SettingsSection == "" {
				p.SettingsSection = "basic"
			}
			if p.SettingsSection != "basic" && p.SettingsSection != "mirror" && p.SettingsSection != "danger" {
				return &httpError{404, "设置页面不存在"}
			}
			if !p.Repo.CanManage {
				return forbidden()
			}
			p.RepositoryTaskActive, err = a.DB.Where("repository_id=? AND state IN (?,?)", p.Repo.ID, "queued", "running").Exist(new(models.Task))
			if err == nil {
				p.Migration, err = a.migrationSnapshot(p.User, &p.Repo.Repository)
			}
			return err
		case "tasks":
			if raw := r.URL.Query().Get("task_id"); raw != "" {
				selected, e := strconv.ParseInt(raw, 10, 64)
				if e != nil || selected < 1 {
					return &httpError{400, "任务编号无效"}
				}
				p.SelectedTask = true
				p.Migration, err = a.migrationSnapshot(p.User, &p.Repo.Repository, selected)
				return err
			}
			if p.MigrationPanel {
				p.Migration, err = a.migrationSnapshot(p.User, &p.Repo.Repository)
			}
			return err
		case "tree", "file", "log":
			if p.Repo.Status != "ready" {
				return &httpError{409, "迁移完成后可浏览内容"}
			}
		default:
			return &httpError{404, "页面不存在"}
		}
		switch p.Tab {
		case "tree":
			p.Entries, err = a.SVN.List(r.Context(), p.Repo.Path, p.Path, p.Revision)
		case "log":
			p.History, err = a.repositoryHistory(r, &p.Repo.Repository, p.Path, p.Revision)
			p.Logs = p.History.Entries
		case "file":
			var b []byte
			b, err = a.SVN.Cat(r.Context(), p.Repo.Path, p.Path, p.Revision)
			p.Content = svn.DecodeText(b)
			if p.Tab == "file" && svn.IsBinary(b) {
				p.Content = "二进制文件，请下载查看"
			}
		}
		return err
	}
	query := a.DB.Cols("id", "name", "display_name", "admin", "active")
	if !p.User.Admin {
		query = query.Where("active=?", true)
	}
	if err := query.Find(&p.Users); err != nil {
		return err
	}
	switch p.Page {
	case "user-profile":
		p.ProfileUser = *p.User
		if targetID := id(r.URL.Query().Get("user_id")); targetID != 0 && targetID != p.User.ID {
			if !p.User.Admin {
				return forbidden()
			}
			p.ProfileUser = models.User{}
			ok, err := a.DB.ID(targetID).Get(&p.ProfileUser)
			if err != nil {
				return err
			}
			if !ok {
				return &httpError{404, "用户不存在"}
			}
		}
		var orgs []models.Organization
		if err := a.DB.SQL("SELECT o.* FROM organization o JOIN membership m ON m.org_id=o.id WHERE m.user_id=? ORDER BY o.name", p.ProfileUser.ID).Find(&orgs); err != nil {
			return err
		}
		p.Organizations = nil
		for _, org := range orgs {
			p.Organizations = append(p.Organizations, organizationPage{Organization: org})
		}

		filtered := p.Repositories[:0]
		for _, repo := range p.Repositories {
			if models.CanRead(a.DB, &p.ProfileUser, &repo.Repository) && strings.Contains(strings.ToLower(repo.Organization+"/"+repo.Name), strings.ToLower(p.Search)) {
				filtered = append(filtered, repo)
			}
		}
		p.Sort = r.URL.Query().Get("sort")
		switch p.Sort {
		case "name-desc":
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].Organization+"/"+filtered[i].Name > filtered[j].Organization+"/"+filtered[j].Name
			})
		default:
			p.Sort = "name"
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].Organization+"/"+filtered[i].Name < filtered[j].Organization+"/"+filtered[j].Name
			})
		}
		p.Repositories, p.ProfilePagination = paginateRepositoryRows(filtered, pageNumber(r.URL.Query().Get("page")), func(page int) string { return p.ProfileListURL(p.Sort, page) })

		return nil
	case "settings-organizations":
		var memberships []models.Organization
		if err := a.DB.SQL("SELECT o.* FROM organization o JOIN membership m ON m.org_id=o.id WHERE m.user_id=? ORDER BY o.name", p.User.ID).Find(&memberships); err != nil {
			return err
		}
		p.Organizations = nil
		for _, org := range memberships {
			p.Organizations = append(p.Organizations, organizationPage{Organization: org})
		}
		return nil
	case "account", "profile":
		return nil
	case "org/create":
		if !p.User.Admin {
			return forbidden()
		}
		return nil
	case "admin/organizations":
		if !p.User.Admin {
			return forbidden()
		}
		p.Sort = r.URL.Query().Get("sort")
		switch p.Sort {
		case "name", "name-desc", "newest", "oldest", "created", "created-desc":
		default:
			p.Sort = "oldest"
		}
		filtered := p.Organizations[:0]
		for _, o := range p.Organizations {
			if strings.Contains(strings.ToLower(o.Name), strings.ToLower(p.Search)) {
				filtered = append(filtered, o)
			}
		}
		p.Organizations = filtered
		sort.Slice(p.Organizations, func(i, j int) bool {
			a, b := p.Organizations[i], p.Organizations[j]
			switch p.Sort {
			case "name":
				return a.Name < b.Name
			case "name-desc":
				return a.Name > b.Name
			case "created":
				if a.Created == b.Created {
					return a.ID < b.ID
				}
				return a.Created < b.Created
			case "created-desc":
				if a.Created == b.Created {
					return a.ID > b.ID
				}
				return a.Created > b.Created
			case "newest":
				return a.ID > b.ID
			default:
				return a.ID < b.ID
			}
		})
		for i := range p.Organizations {
			o := &p.Organizations[i]
			var err error
			o.TeamCount, err = a.DB.Where("org_id=?", o.ID).Count(new(models.Team))
			if err != nil {
				return err
			}
			o.MemberCount, err = a.DB.Where("org_id=?", o.ID).Count(new(models.Membership))
			if err != nil {
				return err
			}
			o.RepoCount, err = a.DB.Where("org_id=?", o.ID).Count(new(models.Repository))
			if err != nil {
				return err
			}
		}
		return nil
	case "migrate":
		p.MigrationOrg = id(r.URL.Query().Get("org"))
		if p.MigrationOrg != 0 {
			allowed := false
			for _, org := range p.Organizations {
				if org.ID == p.MigrationOrg && org.CanManage {
					allowed = true
				}
			}
			if !allowed {
				return forbidden()
			}
		}
		return nil
	case "admin/users/details", "admin", "admin/users", "admin/users/new", "admin/users/edit", "admin/ldap", "admin/ldap/history", "admin/site", "admin/notifications":
		if !p.User.Admin {
			return forbidden()
		}
		if p.Page == "admin/notifications" {
			var err error
			p.Notifications, err = a.notificationData(r)
			return err
		}
		if strings.HasPrefix(p.Page, "admin/ldap") {
			d := a.Directory
			if d == nil {
				d = &directory.Directory{DB: a.DB}
			}
			var err error
			p.LDAPSchedule, err = d.Schedule()
			if err != nil {
				return err
			}
		}
		if p.Page == "admin" {
			p.Page = "admin/users"
		}
		if strings.HasPrefix(p.Page, "admin/users") {
			return a.adminUsersData(r, p)
		}
		if strings.HasPrefix(p.Page, "admin/ldap") {
			p.LDAP = models.LDAPConfig{}
			if exists, err := a.DB.ID(1).Get(&p.LDAP); err != nil {
				return err
			} else if !exists {
				p.LDAP = defaultLDAPConfig()
			}
			if p.Page == "admin/ldap/history" {
				var err error
				p.AuthHistory, err = a.authHistory(r)
				return err
			}
		}
		return nil
	case "organizations":
		return a.organizationPageData(r, p)
	}
	return &httpError{404, "页面不存在"}
}

func (p *webPage) CanMigrate() bool {
	for _, o := range p.Organizations {
		if o.CanManage {
			return true
		}
	}
	return p.User.Admin
}

func (p *webPage) ProfileListURL(order string, page int) string {
	return "/user/profile?" + url.Values{"user_id": {strconv.FormatInt(p.ProfileUser.ID, 10)}, "q": {p.Search}, "sort": {order}, "page": {strconv.Itoa(page)}}.Encode()
}

func (p *webPage) SettingsURL(section string) string {
	target := p.Link("settings", "", p.Revision)
	if section == "basic" {
		return target
	}
	if !reservedOrganizationName(p.Repo.Organization) {
		return target + "/" + section
	}
	return target + "&section=" + url.QueryEscape(section)
}

func defaultLDAPConfig() models.LDAPConfig {
	return models.LDAPConfig{ID: 1, UsePagedSearch: true, SearchPageSize: 500, SyncEnabled: true, URL: "ldaps://", Filter: "(objectClass=person)", UsernameAttribute: "sAMAccountName", KeyAttribute: "objectGUID", NameAttribute: "displayName", EmailAttribute: "mail"}
}
