// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"svnova/models"
)

func orgLink(name string) string {
	if reservedOrganizationName(name) {
		return "/org/" + url.PathEscape(name)
	}
	return "/" + url.PathEscape(name)
}
func orgSection(name, section string) string { return "/org/" + url.PathEscape(name) + "/" + section }
func teamLink(org, name string, ids ...int64) string {
	target := orgSection(org, "teams/") + url.PathEscape(name)
	if name == "new" && len(ids) > 0 {
		target += "?team_id=" + strconv.FormatInt(ids[0], 10)
	}
	return target
}
func teamRepositoriesLink(org, name string, ids ...int64) string {
	target, _ := url.Parse(teamLink(org, name, ids...))
	target.Path += "/repositories"
	return target.String()
}
func repoLink(org, name string, ids ...int64) string {
	if reservedOrganizationName(org) && len(ids) > 0 {
		return "/repositories/" + strconv.FormatInt(ids[0], 10)
	}
	return "/" + url.PathEscape(org) + "/" + url.PathEscape(name)
}

var pageRoutes = map[string]string{
	"/tasks":                 "task-center",
	"/activity":              "activity-trend",
	"/-/admin/notifications": "admin/notifications",
	"/-/admin/users/new":     "admin/users/new",
	"/user/settings":         "account",
	"/user/profile":          "user-profile",
	"/user/settings/account": "account",
	"/repo/migrate":          "migrate", "/-/admin": "admin", "/-/admin/users": "admin/users",
	"/-/admin/ldap": "admin/ldap", "/-/admin/config": "admin/site", "/-/admin/orgs": "admin/organizations",
	"/-/admin/ldap/sync-history":  "admin/ldap/history",
	"/user/settings/organization": "settings-organizations", "/org/create": "org/create",
}
var legacyRoutes = map[string]string{"/migrate": "/repo/migrate", "/admin": "/-/admin", "/admin/users": "/-/admin/users", "/admin/site": "/-/admin/config", "/admin/organizations": "/-/admin/orgs"}

func (a *App) resolvePage(r *http.Request) (*http.Request, error) {
	clone := r.Clone(r.Context())
	target := *r.URL
	clone.URL = &target
	if page, ok := pageRoutes[r.URL.Path]; ok {
		target.Path = "/" + page
		return clone, nil
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 4 && parts[0] == "-" && parts[1] == "admin" && parts[2] == "users" {
		if id(parts[3]) == 0 {
			return nil, &httpError{404, "用户不存在"}
		}
		q := target.Query()
		q.Set("user_id", parts[3])
		target.Path = "/admin/users/details"
		target.RawQuery = q.Encode()
		return clone, nil
	}
	if len(parts) == 5 && parts[0] == "-" && parts[1] == "admin" && parts[2] == "users" && parts[4] == "edit" {
		if id(parts[3]) == 0 {
			return nil, &httpError{404, "用户不存在"}
		}
		q := target.Query()
		q.Set("user_id", parts[3])
		target.Path = "/admin/users/edit"
		target.RawQuery = q.Encode()
		return clone, nil
	}
	if len(parts) == 2 && parts[0] == "org" {
		org := new(models.Organization)
		ok, err := a.DB.Where("name=?", parts[1]).Get(org)
		if err != nil {
			return nil, err
		}
		if ok {
			q := target.Query()
			q.Set("org", strconv.FormatInt(org.ID, 10))
			q.Set("org_view", "repositories")
			target.Path = "/organizations"
			target.RawQuery = q.Encode()
			return clone, nil
		}
	}
	if len(parts) >= 3 && parts[0] == "org" {
		org := new(models.Organization)
		ok, err := a.DB.Where("name=?", parts[1]).Get(org)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, &httpError{404, "组织不存在"}
		}
		q := target.Query()
		q.Set("org", strconv.FormatInt(org.ID, 10))
		q.Set("org_view", parts[2])
		if parts[2] == "dashboard" && len(parts) == 3 {
			target.Path = "/"
			target.RawQuery = q.Encode()
			return clone, nil
		}
		q.Del("team")
		q.Del("team_new")
		q.Del("team_view")
		if parts[2] == "teams" && len(parts) == 4 && parts[3] == "new" && id(r.URL.Query().Get("team_id")) == 0 {
			q.Set("team_new", "1")
		} else if parts[2] == "teams" && (len(parts) == 4 || (len(parts) == 5 && parts[4] == "repositories")) {
			var teams []models.Team
			query := a.DB.Where("org_id=? AND name=?", org.ID, parts[3])
			if parts[3] == "new" && id(r.URL.Query().Get("team_id")) != 0 {
				query = query.And("id=?", id(r.URL.Query().Get("team_id")))
			}
			if err := query.Find(&teams); err != nil {
				return nil, err
			}
			if len(teams) != 1 {
				return nil, &httpError{404, "团队不存在或名称不唯一"}
			}
			q.Set("team", strconv.FormatInt(teams[0].ID, 10))
			if len(parts) == 5 {
				q.Set("team_view", "repositories")
			}
		} else if len(parts) != 3 || (parts[2] != "members" && parts[2] != "teams" && parts[2] != "settings") {
			return nil, &httpError{404, "页面不存在"}
		}
		target.Path = "/organizations"
		target.RawQuery = q.Encode()
		return clone, nil
	}
	if len(parts) == 1 && parts[0] != "" && parts[0] != "organizations" && parts[0] != "migrate" && parts[0] != "admin" {
		org := new(models.Organization)
		ok, err := a.DB.Where("name=?", parts[0]).Get(org)
		if err != nil {
			return nil, err
		}
		if ok {
			q := target.Query()
			q.Set("org", strconv.FormatInt(org.ID, 10))
			q.Set("org_view", "repositories")
			target.Path = "/organizations"
			target.RawQuery = q.Encode()
			return clone, nil
		}
	}
	if len(parts) >= 2 && parts[0] != "repositories" && parts[0] != "admin" && parts[0] != "org" {
		org := new(models.Organization)
		ok, err := a.DB.Where("name=?", parts[0]).Get(org)
		if err != nil {
			return nil, err
		}
		if ok {
			repo := new(models.Repository)
			ok, err = a.DB.Where("org_id=? AND name=?", org.ID, parts[1]).Get(repo)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, &httpError{404, "仓库不存在"}
			}
			q := target.Query()
			if len(parts) == 4 && parts[2] == "settings" && (parts[3] == "mirror" || parts[3] == "danger") {
				q.Set("tab", "settings")
				q.Set("section", parts[3])
			} else if len(parts) == 3 && (parts[2] == "settings" || parts[2] == "tasks") {
				q.Set("tab", parts[2])
			} else if len(parts) != 2 {
				return nil, &httpError{404, "页面不存在"}
			}
			target.Path = "/repositories/" + strconv.FormatInt(repo.ID, 10)
			target.RawQuery = q.Encode()
			return clone, nil
		}
	}
	return clone, nil
}
func (a *App) redirectLegacy(w http.ResponseWriter, r *http.Request, u *models.User) bool {
	if r.Method != "GET" && r.Method != "HEAD" {
		return false
	}
	if target, ok := legacyRoutes[r.URL.Path]; ok {
		if strings.HasPrefix(r.URL.Path, "/admin") && !u.Admin {
			http.Error(w, "仅管理员可访问", 403)
			return true
		}
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, 303)
		return true
	}
	if r.URL.Path == "/organizations" {
		orgID := id(r.URL.Query().Get("org"))
		if orgID == 0 {
			http.Redirect(w, r, "/user/settings/organization", 303)
			return true
		}
		org := new(models.Organization)
		ok, _ := a.DB.ID(orgID).Get(org)
		if ok {
			if !u.Admin {
				member, err := a.DB.Where("org_id=? AND user_id=?", orgID, u.ID).Exist(new(models.Membership))
				if err != nil || !member {
					http.Error(w, "没有访问权限", 403)
					return true
				}
			}
			target := orgLink(org.Name)
			if teamID := id(r.URL.Query().Get("team")); teamID != 0 {
				team := new(models.Team)
				ok, _ := a.DB.Where("id=? AND org_id=?", teamID, orgID).Get(team)
				if ok {
					count, err := a.DB.Where("org_id=? AND name=?", orgID, team.Name).Count(new(models.Team))
					if err != nil || count != 1 {
						return false
					}
					target = teamLink(org.Name, team.Name, team.ID)
				}
			}
			http.Redirect(w, r, target, 303)
			return true
		}
	}
	if strings.HasPrefix(r.URL.Path, "/repositories/") {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) == 2 {
			repo := new(models.Repository)
			ok, _ := a.DB.ID(id(parts[1])).Get(repo)
			if ok && models.CanRead(a.DB, u, repo) {
				org := new(models.Organization)
				ok, _ := a.DB.ID(repo.OrgID).Get(org)
				if ok {
					if reservedOrganizationName(org.Name) {
						return false
					}
					target := repoLink(org.Name, repo.Name, repo.ID)
					q := r.URL.Query()
					if tab := q.Get("tab"); tab == "settings" || tab == "tasks" {
						target += "/" + tab
						q.Del("tab")
					}
					if len(q) > 0 {
						target += "?" + q.Encode()
					}
					http.Redirect(w, r, target, 303)
					return true
				}
			}
		}
	}
	return false
}

func reservedOrganizationName(name string) bool {
	switch strings.ToLower(name) {
	case "tasks", "activity", "user", "org", "repo", "api", "assets", "avatars", "install", "admin", "organizations", "repositories", "migrate", "login", "logout":
		return true
	}
	return false
}
