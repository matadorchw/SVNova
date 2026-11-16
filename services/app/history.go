// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"svnova/models"
	"svnova/services/svn"
)

const revisionHistoryPageSize = 10

type historyPage struct {
	Entries                  []svn.LogEntry
	Revision                 int64
	OlderCursor, NewerCursor int64
}

func (a *App) repositoryHistory(r *http.Request, repo *models.Repository, path string, head int64) (historyPage, error) {
	page := historyPage{Revision: head, Entries: []svn.LogEntry{}}
	q := r.URL.Query()
	cursor := head
	if raw := q.Get("cursor"); raw != "" {
		var err error
		cursor, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || cursor < 1 || cursor > head {
			return page, &httpError{400, "无效历史游标"}
		}
	}
	direction := q.Get("direction")
	if direction != "" && direction != "older" && direction != "newer" {
		return page, &httpError{400, "无效翻页方向"}
	}
	newer := direction == "newer"
	rows, err := a.SVN.HistoryPage(r.Context(), repo.Path, path, head, cursor, newer, revisionHistoryPageSize+1)
	if err != nil {
		return page, err
	}
	more := len(rows) > revisionHistoryPageSize
	if more {
		rows = rows[:revisionHistoryPageSize]
	}
	if newer {
		slices.Reverse(rows)
	}
	if len(rows) == 0 {
		return page, nil
	}
	page.Entries = rows
	first, last := rows[0].Revision, rows[len(rows)-1].Revision
	if !newer && more {
		page.OlderCursor = last - 1
	}
	if newer && more {
		page.NewerCursor = first + 1
	}
	// Probe the opposite direction instead of assuming revisions are contiguous for a path.
	probe, probeNewer := first+1, true
	if newer {
		probe, probeNewer = last-1, false
	}
	if probe >= 1 && probe <= head {
		other, e := a.SVN.HistoryPage(r.Context(), repo.Path, path, head, probe, probeNewer, 1)
		if e != nil {
			return page, e
		}
		if len(other) > 0 {
			if probeNewer {
				page.NewerCursor = probe
			} else {
				page.OlderCursor = probe
			}
		}
	}
	return page, nil
}

func (p *webPage) HistoryURL(cursor int64, direction string) string {
	q := url.Values{"tab": {"log"}, "path": {p.Path}, "revision": {strconv.FormatInt(p.Revision, 10)}, "cursor": {strconv.FormatInt(cursor, 10)}, "direction": {direction}}
	return repoLink(p.Repo.Organization, p.Repo.Name, p.Repo.ID) + "?" + q.Encode()
}
