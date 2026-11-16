// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"svnova/models"
	"unicode/utf8"
)

func (a *App) organizationSettings(w http.ResponseWriter, r *http.Request, u *models.User, orgID int64) error {
	if r.Method != "POST" {
		return &httpError{405, "不支持此操作"}
	}
	var in struct{ FullName, Email, Description, Website, Location string }
	if err := decode(w, r, &in); err != nil {
		return err
	}
	fields := []struct {
		value *string
		max   int
	}{{&in.FullName, 100}, {&in.Email, 254}, {&in.Description, 255}, {&in.Website, 255}, {&in.Location, 50}}
	for _, field := range fields {
		*field.value = strings.TrimSpace(*field.value)
		if utf8.RuneCountInString(*field.value) > field.max || strings.ContainsRune(*field.value, 0) {
			return errors.New("组织资料超出长度限制或包含无效字符")
		}
	}
	if in.Email != "" {
		address, err := mail.ParseAddress(in.Email)
		if err != nil || address.Address != in.Email {
			return errors.New("联系邮箱格式无效")
		}
	}
	if in.Website != "" {
		link, err := url.Parse(in.Website)
		if err != nil || (link.Scheme != "http" && link.Scheme != "https") || link.Host == "" || link.User != nil {
			return errors.New("网站必须为 HTTP 或 HTTPS 地址")
		}
	}
	tx := a.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE organization SET id=id WHERE id=?", orgID); err != nil {
		return err
	}
	if err := checkOrganizationOwner(tx, u, orgID); err != nil {
		return err
	}
	rows, err := tx.ID(orgID).Cols("full_name", "email", "description", "website", "location").Update(&models.Organization{FullName: in.FullName, Email: in.Email, Description: in.Description, Website: in.Website, Location: in.Location})
	if err != nil {
		return err
	}
	if rows == 0 {
		return &httpError{404, "组织不存在"}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return send(w, map[string]bool{"ok": true})
}
func (p *webPage) OrgURL() string                   { return orgLink(p.Org.Name) }
func (p *webPage) OrgSection(section string) string { return orgSection(p.Org.Name, section) }

type organizationTeam struct {
	Members            []models.TeamMember
	IsMember, CanLeave bool
	models.Team
	MemberCount, RepoCount int64
}

func (a *App) organizationPageData(r *http.Request, p *webPage) error {
	orgID := id(r.URL.Query().Get("org"))
	if orgID == 0 {
		return nil
	}
	for i := range p.Organizations {
		if p.Organizations[i].ID == orgID {
			p.Org = &p.Organizations[i]
		}
	}
	if p.Org == nil {
		return forbidden()
	}
	p.OrgTab = r.URL.Query().Get("org_view")
	if p.OrgTab == "" {
		p.OrgTab = "repositories"
		if r.URL.Query().Get("team") != "" {
			p.OrgTab = "teams"
		}
	}
	if p.OrgTab == "settings" && !p.Org.CanManage {
		return forbidden()
	}
	filtered := p.Repositories[:0]
	for _, repo := range p.Repositories {
		if repo.OrgID == orgID {
			filtered = append(filtered, repo)
		}
	}
	p.Repositories = filtered
	p.OrgRepoCount = len(filtered)
	p.NewTeam = r.URL.Query().Get("team_new") == "1"
	p.TeamView = r.URL.Query().Get("team_view")
	if p.TeamView == "" {
		p.TeamView = "members"
	}
	if p.NewTeam && !p.Org.CanManage {
		return forbidden()
	}
	p.OwnerIDs = map[int64]bool{}
	if err := a.DB.Where("org_id=?", orgID).Find(&p.Members); err != nil {
		return err
	}
	if err := a.DB.Where("org_id=?", orgID).Find(&p.Teams); err != nil {
		return err
	}
	var orgUsers []models.User
	if err := a.DB.SQL("SELECT u.id,u.name,u.display_name,u.admin,u.active FROM user u JOIN membership m ON m.user_id=u.id WHERE m.org_id=?", orgID).Find(&orgUsers); err != nil {
		return err
	}
	for _, member := range orgUsers {
		found := false
		for _, u := range p.Users {
			if u.ID == member.ID {
				found = true
			}
		}
		if !found {
			p.Users = append(p.Users, member)
		}
	}
	for _, team := range p.Teams {
		view := organizationTeam{Team: team}
		var err error
		view.MemberCount, err = a.DB.Where("team_id=?", team.ID).Count(new(models.TeamMember))
		if err != nil {
			return err
		}
		if team.IsOwnerTeam {
			view.RepoCount = int64(len(p.Repositories))
		} else {
			var grants []models.TeamRepository
			if err = a.DB.Where("team_id=?", team.ID).Find(&grants); err != nil {
				return err
			}
			for _, grant := range grants {
				for _, repo := range p.Repositories {
					if grant.RepositoryID == repo.ID {
						view.RepoCount++
					}
				}
			}
		}
		var members []models.TeamMember
		if err = a.DB.Where("team_id=?", team.ID).Find(&members); err != nil {
			return err
		}
		for _, m := range members {
			if team.IsOwnerTeam {
				p.OwnerIDs[m.UserID] = true
			}
			if m.UserID == p.User.ID {
				view.IsMember = true
			}
		}
		view.CanLeave = view.IsMember && (!team.IsOwnerTeam || len(members) > 1)
		view.Members = members
		if len(view.Members) > 12 {
			view.Members = view.Members[:12]
		}
		p.TeamStats = append(p.TeamStats, view)
	}
	for _, m := range p.Members {
		if p.Search == "" || strings.Contains(strings.ToLower(p.Username(m.UserID)+" "+p.UserDisplayName(m.UserID)), strings.ToLower(p.Search)) {
			p.FilteredMembers = append(p.FilteredMembers, m)
		}
	}
	if p.OrgTab == "members" {
		p.MemberPagination = adminUsersPage{Total: int64(len(p.FilteredMembers)), Page: pageNumber(r.URL.Query().Get("page")), Pages: max(1, (len(p.FilteredMembers)+listPageSize-1)/listPageSize)}
		p.MemberPagination.Page = min(p.MemberPagination.Page, p.MemberPagination.Pages)
		p.MemberPagination.paginateWith(func(page int) string {
			q := url.Values{"q": {p.Search}, "page": {fmt.Sprint(page)}}
			return orgSection(p.Org.Name, "members") + "?" + q.Encode()
		})
		start := min((p.MemberPagination.Page-1)*listPageSize, len(p.FilteredMembers))
		p.FilteredMembers = p.FilteredMembers[start:min(start+listPageSize, len(p.FilteredMembers))]
	}
	for _, team := range p.TeamStats {
		if p.Search == "" || strings.Contains(strings.ToLower(team.Name+" "+team.Description), strings.ToLower(p.Search)) {
			p.FilteredTeams = append(p.FilteredTeams, team)
		}
	}
	if p.OrgTab == "repositories" && p.Search != "" {
		filtered := p.Repositories[:0]
		for _, repo := range p.Repositories {
			if strings.Contains(strings.ToLower(repo.Name), strings.ToLower(p.Search)) {
				filtered = append(filtered, repo)
			}
		}
		p.Repositories = filtered
	}
	teamID := id(r.URL.Query().Get("team"))
	if teamID == 0 {
		if p.OrgTab == "repositories" {
			p.Repositories, p.RepositoryPagination = paginateRepositoryRows(p.Repositories, pageNumber(r.URL.Query().Get("page")), func(page int) string { return repositoryListLink(p.OrgURL(), p.Search, page) })
		}
		return nil
	}
	for i := range p.Teams {
		if p.Teams[i].ID == teamID {
			p.Team = &p.Teams[i]
		}
	}
	if p.Team == nil {
		return forbidden()
	}
	if err := a.DB.Where("team_id=?", teamID).Find(&p.TeamMembers); err != nil {
		return err
	}
	for _, m := range p.TeamMembers {
		if p.Search == "" || strings.Contains(strings.ToLower(p.Username(m.UserID)+" "+p.UserDisplayName(m.UserID)), strings.ToLower(p.Search)) {
			p.FilteredTeamMembers = append(p.FilteredTeamMembers, m)
		}
	}
	var grants []models.TeamRepository
	if err := a.DB.Where("team_id=?", teamID).Find(&grants); err != nil {
		return err
	}
	for _, grant := range grants {
		for _, repo := range p.Repositories {
			if grant.RepositoryID == repo.ID {
				p.Grants = append(p.Grants, grant)
			}
		}
	}
	if p.TeamView == "repositories" {
		_, p.RepositoryPagination = paginateRepositoryRows(p.FilteredTeamRepositories(), pageNumber(r.URL.Query().Get("page")), func(page int) string { return repositoryListLink(p.TeamRepositoriesURL(), p.Search, page) })
	}
	return nil
}

func (p *webPage) MemberPreview() []models.Membership {
	if len(p.Members) > 12 {
		return p.Members[:12]
	}
	return p.Members
}

func (p *webPage) UserDisplayName(id int64) string {
	for _, u := range p.Users {
		if u.ID == id {
			return u.DisplayName
		}
	}
	return ""
}
func (p *webPage) IsOrgOwner(id int64) bool { return p.OwnerIDs[id] }
func (p *webPage) CanRemoveOrgMember(id int64) bool {
	return p.Org.CanManage && (!p.OwnerIDs[id] || len(p.OwnerIDs) > 1)
}
func (p *webPage) TeamURL() string { return teamLink(p.Org.Name, p.Team.Name, p.Team.ID) }
func (p *webPage) TeamCandidates() []models.User {
	var out []models.User
	for _, u := range p.Users {
		if !u.Active {
			continue
		}
		present := false
		for _, m := range p.TeamMembers {
			if m.UserID == u.ID {
				present = true
			}
		}
		if !present {
			out = append(out, u)
		}
	}
	return out
}
func (p *webPage) AvailableTeamRepositories() []repositoryPage {
	var out []repositoryPage
	for _, repo := range p.Repositories {
		present := false
		for _, g := range p.Grants {
			if g.RepositoryID == repo.ID {
				present = true
			}
		}
		if !present {
			out = append(out, repo)
		}
	}
	return out
}

func (p *webPage) TeamRepositoriesURL() string {
	return teamRepositoriesLink(p.Org.Name, p.Team.Name, p.Team.ID)
}
func (p *webPage) VisibleTeamGrants() []models.TeamRepository {
	var out []models.TeamRepository
	for _, g := range p.Grants {
		if p.Search == "" || strings.Contains(strings.ToLower(p.Reponame(g.RepositoryID)), strings.ToLower(p.Search)) {
			out = append(out, g)
		}
	}
	return out
}

func (p *webPage) IsOrgMember(userID int64) bool {
	for _, m := range p.Members {
		if m.UserID == userID {
			return true
		}
	}
	return false
}

func (p *webPage) CanViewUserProfile(userID int64) bool {
	return p.User != nil && (p.User.Admin || p.User.ID == userID)
}

func (p *webPage) TeamRepositoryList() []repositoryPage {
	var out []repositoryPage
	for _, repo := range p.Repositories {
		if p.Team.IsOwnerTeam {
			out = append(out, repo)
			continue
		}
		for _, grant := range p.Grants {
			if grant.RepositoryID == repo.ID {
				out = append(out, repo)
				break
			}
		}
	}
	return out
}
func (p *webPage) FilteredTeamRepositories() []repositoryPage {
	var out []repositoryPage
	for _, repo := range p.TeamRepositoryList() {
		if p.Search == "" || strings.Contains(strings.ToLower(repo.Organization+"/"+repo.Name), strings.ToLower(p.Search)) {
			out = append(out, repo)
		}
	}
	if p.RepositoryPagination.Page > 0 {
		start := min((p.RepositoryPagination.Page-1)*repositoryListPageSize, len(out))
		out = out[start:min(start+repositoryListPageSize, len(out))]
	}
	return out
}
