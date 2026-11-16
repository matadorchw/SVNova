// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/url"
	"strconv"
)

const repositoryListPageSize = 10
const dashboardFeedPageSize = 10

func paginateRepositoryRows(rows []repositoryPage, page int, link func(int) string) ([]repositoryPage, adminUsersPage) {
	pagination := adminUsersPage{Total: int64(len(rows)), Pages: max(1, (len(rows)+repositoryListPageSize-1)/repositoryListPageSize)}
	pagination.Page = min(max(1, page), pagination.Pages)
	pagination.paginateWith(link)
	start := (pagination.Page - 1) * repositoryListPageSize
	return rows[start:min(start+repositoryListPageSize, len(rows))], pagination
}

func repositoryListLink(base, keyword string, page int) string {
	target, err := url.Parse(base)
	if err != nil {
		return base
	}
	q := target.Query()
	q.Set("page", strconv.Itoa(page))
	if keyword != "" {
		q.Set("q", keyword)
	} else {
		q.Del("q")
	}
	target.RawQuery = q.Encode()
	return target.String()
}
