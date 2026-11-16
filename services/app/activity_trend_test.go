// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"svnova/models"
	"svnova/services/activity"
	"testing"
	"time"
)

func TestActivityTrend(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "trend.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	user := &models.User{ID: 1, Name: "reader", Active: true}
	day := activity.Today(time.Now())
	_, err = db.Insert(user, &models.Organization{ID: 1, Name: "org"}, &models.Team{ID: 1, OrgID: 1, Name: "readers"}, &models.Membership{OrgID: 1, UserID: 1}, &models.TeamMember{TeamID: 1, UserID: 1}, &models.Repository{ID: 1, OrgID: 1, Name: "repo", Status: "failed"}, &models.TeamRepository{TeamID: 1, RepositoryID: 1}, &models.Task{ID: 1, RepositoryID: 1, Kind: "sync", State: "failed", Started: day.Unix()}, &models.TaskExecution{TaskID: 1, State: "running", Started: day.UnixMilli()}, &models.TaskExecution{TaskID: 1, State: "failed", Started: day.Add(-time.Second).UnixMilli()}, &models.Task{ID: 2, RepositoryID: 1, Kind: "sync", State: "succeeded", Started: day.Unix()}, &models.Task{ID: 3, RepositoryID: 1, Kind: "local", State: "failed", Started: day.Unix()}, &models.Repository{ID: 2, OrgID: 2, Name: "secret"}, &models.Task{ID: 4, RepositoryID: 2, Kind: "sync", State: "failed", Started: day.Unix()})
	if err != nil {
		t.Fatal(err)
	}
	url := "/api/v1/activity/trend?kind=sync&from=" + day.AddDate(0, 0, -2).Format("2006-01-02") + "&to=" + day.Format("2006-01-02")
	read := func() trendResponse {
		t.Helper()
		w := httptest.NewRecorder()
		if err := (&App{DB: db}).activityTrend(w, httptest.NewRequest("GET", url, nil), user); err != nil {
			t.Fatal(err)
		}
		var out trendResponse
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	out := read()
	if len(out.Days) != 3 || out.Days[0].Counts["failed"] != 0 || out.Days[1].Counts["failed"] != 1 || out.Days[2].Counts["running"] != 1 || out.Days[2].Counts["succeeded"] != 1 || out.Days[2].Counts["failed"] != 0 || !out.Pending {
		t.Fatalf("incorrect aggregation: %+v", out)
	}
	if _, err = db.Exec("UPDATE task_execution SET state='succeeded' WHERE state='running'"); err != nil {
		t.Fatal(err)
	}
	out = read()
	if out.Pending || out.Days[2].Counts["succeeded"] != 2 {
		t.Fatal("state update not reflected")
	}
	for _, suffix := range []string{"&org=2", "&org=bad", "&kind=bad"} {
		bad := url + suffix
		if suffix == "&kind=bad" {
			bad = "/api/v1/activity/trend?kind=bad&from=2026-01-01&to=2026-01-02"
		}
		if err := (&App{DB: db}).activityTrend(httptest.NewRecorder(), httptest.NewRequest("GET", bad, nil), user); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
	for _, query := range []string{"kind=sync&from=bad&to=bad", "kind=sync&from=2020-01-01&to=2020-01-02", "kind=sync&from=2099-01-01&to=2099-01-02", "kind=sync&from=" + day.Format("2006-01-02") + "&to=" + day.AddDate(0, 0, -1).Format("2006-01-02")} {
		if err := (&App{DB: db}).activityTrend(httptest.NewRecorder(), httptest.NewRequest("GET", "/?"+query, nil), user); err == nil {
			t.Fatal("accepted invalid dates")
		}
	}
	if _, err = db.Where("team_id=1").Delete(new(models.TeamRepository)); err != nil {
		t.Fatal(err)
	}
	out = read()
	for _, d := range out.Days {
		for _, n := range d.Counts {
			if n != 0 {
				t.Fatal("revoked data exposed")
			}
		}
	}
}
