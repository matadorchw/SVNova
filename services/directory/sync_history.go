// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"encoding/json"
	"errors"
	"strings"
	"svnova/models"
	"svnova/services/notifications"
	"time"
)

var ErrSyncBusy = errors.New("LDAP synchronization or configuration change is in progress")

func failureReason(err error) string {
	if err == nil {
		return ""
	}
	for _, pair := range []struct{ match, code string }{{"connection failed", "connection"}, {"StartTLS failed", "tls"}, {"service bind failed", "bind"}, {"query failed", "query"}, {"empty LDAP result", "empty"}, {"stable identity missing", "identity"}, {"duplicate LDAP stable identity", "duplicate_identity"}, {"email attribute is invalid", "email"}, {"another authentication identity", "identity_conflict"}} {
		if strings.Contains(err.Error(), pair.match) {
			return pair.code
		}
	}
	return "internal"
}
func completeRun(run *models.LDAPSyncRun, report SyncReport, err error) {
	run.Finished = time.Now().UnixMilli()
	run.DurationMS = run.Finished - run.Started
	if run.DurationMS < 0 {
		run.DurationMS = 0
	}
	run.Status = "success"
	if err != nil {
		run.Status = "failed"
	}
	run.FailureReason = failureReason(err)
	run.MatchedUsers = report.MatchedUsers
	run.CreatedUsers = report.CreatedUsers
	run.UpdatedUsers = report.UpdatedUsers
	run.DisabledUsers = report.DeactivatedUsers
	run.SkippedLocalCount = report.SkippedLocalCount
	data, _ := json.Marshal(report)
	run.Report = string(data)
}
func (d *Directory) RecoverSyncRuns() error {
	var runs []models.LDAPSyncRun
	if err := d.DB.Where("status=?", "running").Find(&runs); err != nil {
		return err
	}
	for _, run := range runs {
		completeRun(&run, SyncReport{}, errors.New("interrupted"))
		run.FailureReason = "interrupted"
		tx := d.DB.NewSession()
		if err := tx.Begin(); err != nil {
			tx.Close()
			return err
		}
		if _, err := tx.ID(run.ID).AllCols().Update(&run); err != nil {
			tx.Close()
			return err
		}
		if err := notifications.Enqueue(tx, notifications.LDAPEvent(run)); err != nil {
			tx.Close()
			return err
		}
		if err := tx.Commit(); err != nil {
			tx.Close()
			return err
		}
		tx.Close()
	}
	return nil
}
