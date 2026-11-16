// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http"
	"strconv"
	"strings"
	"svnova/models"
	"svnova/services/activity"
	"time"
)

type trendDay struct {
	Date   string           `json:"date"`
	Counts map[string]int64 `json:"counts"`
}
type trendResponse struct {
	Kind    string     `json:"kind"`
	From    string     `json:"from"`
	To      string     `json:"to"`
	Today   string     `json:"today"`
	Pending bool       `json:"pending"`
	Days    []trendDay `json:"days"`
}

func (a *App) activityTrend(w http.ResponseWriter, r *http.Request, u *models.User) error {
	q := r.URL.Query()
	now := time.Now()
	today := activity.Today(now)
	kind := q.Get("kind")
	from, e1 := time.ParseInLocation("2006-01-02", q.Get("from"), activity.Shanghai)
	to, e2 := time.ParseInLocation("2006-01-02", q.Get("to"), activity.Shanghai)
	if (kind != "local" && kind != "mirror" && kind != "sync") || e1 != nil || e2 != nil || from.After(to) || from.Before(activity.Since(now)) || to.After(today) {
		return &httpError{400, "活动类型或日期范围无效"}
	}
	orgID := int64(0)
	if value := q.Get("org"); value != "" {
		var err error
		orgID, err = strconv.ParseInt(value, 10, 64)
		if err != nil || orgID < 0 {
			return &httpError{400, "组织参数无效"}
		}
	}
	if orgID != 0 && !u.Admin {
		ok, err := a.DB.Where("org_id=? AND user_id=?", orgID, u.ID).Exist(new(models.Membership))
		if err != nil {
			return err
		}
		if !ok {
			return forbidden()
		}
	}
	var repos []models.Repository
	if err := a.DB.Find(&repos); err != nil {
		return err
	}
	ids := []string{}
	args := []any{}
	for _, repo := range repos {
		if (orgID == 0 || repo.OrgID == orgID) && models.CanRead(a.DB, u, &repo) {
			ids = append(ids, "?")
			args = append(args, repo.ID)
		}
	}
	out := trendResponse{Kind: kind, From: q.Get("from"), To: q.Get("to"), Today: today.Format("2006-01-02"), Days: []trendDay{}}
	indices := map[string]int{}
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		indices[date] = len(out.Days)
		out.Days = append(out.Days, trendDay{date, map[string]int64{"queued": 0, "running": 0, "succeeded": 0, "failed": 0, "interrupted": 0}})
	}
	if len(ids) > 0 {
		args = append(args, kind, from.Unix(), to.AddDate(0, 0, 1).Unix(), now.Unix())
		var rows []struct {
			Day   string
			State string
			Count int64
		}
		err := a.DB.SQL("SELECT strftime('%Y-%m-%d',timestamp,'unixepoch','+8 hours') AS day,state,COUNT(*) AS count FROM ("+repositoryActivitySQL+") events WHERE repository_id IN ("+strings.Join(ids, ",")+") AND kind=? AND timestamp>=? AND timestamp<? AND timestamp<=? GROUP BY day,state", args...).Find(&rows)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if i, ok := indices[row.Day]; ok {
				out.Days[i].Counts[row.State] = row.Count
				if row.State == "queued" || row.State == "running" {
					out.Pending = true
				}
			}
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	return send(w, out)
}
