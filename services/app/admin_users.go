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

const listPageSize = 10
const adminListPageSize = 20

type adminUsersPage struct {
	Compact                            bool
	Label                              string
	Items                              []models.User
	Total                              int64
	Page, Pages                        int
	Keyword, Active, Admin, Sort, Auth string
	Previous, Next                     string
	First, Last                        string
	Numbers                            []adminUserPageNumber
}

type adminUserPageNumber struct {
	Number  int
	Current bool
	URL     string
}

func (v adminUsersPage) SortLink(value string) string { v.Sort = value; return v.Link(1) }

func (v adminUsersPage) Link(page int) string {
	values := url.Values{"q": {v.Keyword}, "active": {v.Active}, "admin": {v.Admin}, "auth": {v.Auth}, "sort": {v.Sort}, "page": {strconv.Itoa(page)}}
	return "/-/admin/users?" + values.Encode()
}

func (v *adminUsersPage) paginate() {
	v.paginateWith(v.Link)
}

func (v *adminUsersPage) paginateWith(link func(int) string) {
	if v.Page > 1 {
		v.First = link(1)
		v.Previous = link(v.Page - 1)
	}
	if v.Page < v.Pages {
		v.Last = link(v.Pages)
		v.Next = link(v.Page + 1)
	}
	// Keep a five-page window, with ellipses for omitted pages.
	start := max(1, min(v.Page-2, v.Pages-4))
	end := min(v.Pages, start+4)
	if start > 1 {
		v.Numbers = append(v.Numbers, adminUserPageNumber{Number: -1})
	}
	for n := start; n <= end; n++ {
		v.Numbers = append(v.Numbers, adminUserPageNumber{Number: n, Current: n == v.Page, URL: link(n)})
	}
	if end < v.Pages {
		v.Numbers = append(v.Numbers, adminUserPageNumber{Number: -1})
	}
}

func (p *webPage) IsAdminUsers() bool { return strings.HasPrefix(p.Page, "admin/users") }
func (p *webPage) IsAdminPage() bool  { return strings.HasPrefix(p.Page, "admin/") }

func (a *App) adminUsersData(r *http.Request, p *webPage) error {
	if p.Page == "admin/users/new" {
		return nil
	}
	if p.Page == "admin/users/edit" || p.Page == "admin/users/details" {
		ok, err := a.DB.ID(id(r.URL.Query().Get("user_id"))).Get(&p.EditUser)
		if err != nil {
			return err
		}
		if !ok {
			return &httpError{404, "用户不存在"}
		}
		if p.Page == "admin/users/edit" {
			session := a.DB.NewSession()
			defer session.Close()
			p.UserDeleteDisabledReason, err = userDeletionReason(session, p.User, &p.EditUser)
			if err != nil {
				return err
			}
		}
		if p.Page == "admin/users/details" {
			p.UserAuthSource = "本地"
			if p.EditUser.LDAPAccount {
				p.UserAuthSource = "LDAP"
			}
			filtered := p.Repositories[:0]
			for _, repo := range p.Repositories {
				if models.CanRead(a.DB, &p.EditUser, &repo.Repository) {
					filtered = append(filtered, repo)
				}
			}
			p.Repositories, p.RepositoryPagination = paginateRepositoryRows(filtered, pageNumber(r.URL.Query().Get("page")), func(page int) string {
				return repositoryListLink("/-/admin/users/"+strconv.FormatInt(p.EditUser.ID, 10), "", page)
			})
			var orgs []models.Organization
			if err := a.DB.SQL("SELECT o.* FROM organization o JOIN membership m ON m.org_id=o.id WHERE m.user_id=? ORDER BY o.name", p.EditUser.ID).Find(&orgs); err != nil {
				return err
			}
			p.Organizations = nil
			for _, org := range orgs {
				p.Organizations = append(p.Organizations, organizationPage{Organization: org})
			}
		}
		return nil
	}
	q := r.URL.Query()
	v := adminUsersPage{Keyword: strings.TrimSpace(q.Get("q")), Active: q.Get("active"), Admin: q.Get("admin"), Sort: q.Get("sort"), Auth: q.Get("auth")}
	orders := map[string]string{"created": "created ASC, id ASC", "created-desc": "created DESC, id DESC", "oldest": "id ASC", "newest": "id DESC", "alphabetically": "name COLLATE NOCASE ASC, id ASC", "reversealphabetically": "name COLLATE NOCASE DESC, id DESC", "lastlogin": "last_login DESC, id DESC", "reverselastlogin": "last_login ASC, id ASC"}
	order, ok := orders[v.Sort]
	if !ok {
		v.Sort = "alphabetically"
		order = orders[v.Sort]
	}
	where, args := []string{"1=1"}, []any{}
	if v.Keyword != "" {
		where = append(where, "(instr(lower(name),lower(?))>0 OR instr(lower(display_name),lower(?))>0 OR instr(lower(email),lower(?))>0)")
		args = append(args, v.Keyword, v.Keyword, v.Keyword)
	}
	for _, f := range []struct {
		column string
		value  *string
	}{{"active", &v.Active}, {"admin", &v.Admin}} {
		if *f.value == "1" || *f.value == "0" {
			where = append(where, f.column+"=?")
			args = append(args, *f.value)
		} else {
			*f.value = ""
		}
	}
	switch v.Auth {
	case "local":
		where = append(where, "ldap_account=0")
	case "ldap":
		where = append(where, "ldap_account<>0")
	default:
		v.Auth = ""
	}
	condition := strings.Join(where, " AND ")
	var err error
	v.Total, err = a.DB.Where(condition, args...).Count(new(models.User))
	if err != nil {
		return err
	}
	v.Pages = max(1, int((v.Total+adminListPageSize-1)/adminListPageSize))
	v.Page, _ = strconv.Atoi(q.Get("page"))
	v.Page = min(max(1, v.Page), v.Pages)
	if err = a.DB.Where(condition, args...).OrderBy(order).Limit(adminListPageSize, (v.Page-1)*adminListPageSize).Find(&v.Items); err != nil {
		return err
	}
	v.paginate()
	p.AdminUsers = v
	return nil
}

func (p *webPage) UserTime(value int64) string {
	if value <= 0 {
		return "未知"
	}
	return p.Time(value)
}

func (v adminUsersPage) CompactView() adminUsersPage {
	v.Compact = true
	v.Label = "仓库分页"
	return v
}
