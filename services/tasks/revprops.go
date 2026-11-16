// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"svnova/models"
	"svnova/services/svn"
	"time"
)

func (w *Worker) copyRevprops(ctx context.Context, r *models.Repository, t *models.Task, c svn.Credentials, total int64) error {
	t.Phase = "revision-properties"
	t.Verification = "revprops"
	t.VerificationPosition = -1
	t.TotalRevision = total
	t.ProgressStarted = time.Now().Unix()
	t.ProgressUpdated = t.ProgressStarted
	save := func() error {
		_, err := w.DB.ID(t.ID).Cols("phase", "verification", "verification_position", "total_revision", "progress_started", "progress_updated").Update(t)
		return err
	}
	if err := save(); err != nil {
		return err
	}
	last := time.Time{}
	err := w.SVN.CopyRevprops(ctx, r.Path, c, func(revision int64) error {
		if revision < 0 || revision > total || revision <= t.VerificationPosition {
			return nil
		}
		t.VerificationPosition = revision
		t.ProgressUpdated = time.Now().Unix()
		if time.Since(last) < 2*time.Second {
			return nil
		}
		last = time.Now()
		return save()
	})
	if err != nil {
		return err
	}
	t.VerificationPosition = total
	t.ProgressUpdated = time.Now().Unix()
	return save()
}
