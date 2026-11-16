// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"svnova/models"
	"svnova/services/notifications"
	"svnova/services/security"
	"svnova/services/storage"
	"svnova/services/svn"
	"xorm.io/xorm"
)

type Worker struct {
	Convert func(context.Context, *models.Repository, *models.Task) error
	DB      *xorm.Engine
	SVN     svn.Runner
	Vault   *security.Vault
	Root    string
	mu      sync.Mutex
}

func (w *Worker) TryLock() bool {
	if !w.mu.TryLock() {
		return false
	}
	if pending, err := w.DB.Exist(new(models.MirrorConversion)); err != nil || pending {
		w.mu.Unlock()
		return false
	}
	return true
}
func (w *Worker) Unlock() { w.mu.Unlock() }
func (w *Worker) Credentials(r *models.Repository) (svn.Credentials, error) {
	p, err := w.Vault.Open(r.Password)
	return svn.Credentials{Username: r.Username, Password: p, SkipTLS: r.SkipTLS}, err
}
func (w *Worker) Run(ctx context.Context) {
	_, _ = w.DB.Where("state=?", "running").Cols("state", "finished").Update(&models.TaskExecution{State: "interrupted", Finished: time.Now().UnixMilli()})
	_, _ = w.DB.Where("state=? AND kind=?", "running", "conversion").Cols("state", "error", "finished").Update(&models.Task{State: "failed", Error: "mirror conversion interrupted; retry required", Finished: time.Now().Unix()})
	// There is only one worker and one application instance. Retry interrupted operations.
	_, _ = w.DB.Where("state=?", "running").Cols("state").Update(&models.Task{State: "queued"})
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		w.Tick(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (w *Worker) Tick(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if ctx.Err() != nil {
		return
	}
	if pending, err := w.DB.Exist(new(models.MirrorConversion)); err != nil || pending {
		return
	}
	var due []models.Repository
	if w.DB.Where("kind=? AND status=? AND halted=? AND interval>0 AND next_sync<=?", "mirror", "ready", false, time.Now().Unix()).Find(&due) == nil {
		for _, r := range due {
			_ = models.EnqueueTriggered(w.DB, r.ID, "sync", "", "scheduled")
		}
	}
	var task models.Task
	ok, err := w.DB.Where("state=?", "queued").Asc("id").Get(&task)
	if err != nil || !ok {
		if err == nil {
			w.backfillSize(ctx)
		}
		return
	}
	var r models.Repository
	ok, err = w.DB.ID(task.RepositoryID).Get(&r)
	if err == nil && !ok {
		err = errors.New("repository no longer exists")
	}
	execution := models.TaskExecution{TaskID: task.ID, Trigger: task.Trigger, Started: time.Now().UnixMilli(), State: "running", BeforeRevision: r.Revision}
	if e := w.startExecution(&task, &execution); e != nil {
		log.Print("Cannot persist task execution")
		return
	}
	if err == nil {
		if !storage.Within(w.Root, r.Path) {
			err = errors.New("repository outside managed root")
		} else {
			err = storage.CheckParents(r.Path)
		}
	}
	if err == nil {
		if task.Kind == "conversion" {
			if w.Convert == nil {
				err = errors.New("conversion executor unavailable")
			} else {
				err = w.Convert(ctx, &r, &task)
			}
		} else if task.Kind == "local" {
			err = w.local(ctx, &r, &task)
		} else {
			err = w.mirror(ctx, &r, &task)
		}
	}
	if ctx.Err() != nil && task.Kind != "conversion" {
		_, _ = w.DB.ID(execution.ID).Cols("state", "finished").Update(&models.TaskExecution{State: "interrupted", Finished: time.Now().UnixMilli()})
		task.State = "queued"
		task.Error = ""
		task.Finished = 0
		_, _ = w.DB.ID(r.ID).Cols("uuid", "revision", "halted").Update(&r)
		_, _ = w.DB.ID(task.ID).Cols("state", "phase", "error", "finished").Update(&task)
		return
	}
	task.Finished = time.Now().Unix()
	if err != nil {
		task.State = "failed"
		task.Error = err.Error()
		if r.Status != "ready" {
			r.Status = "failed"
		}
	} else {
		task.State = "succeeded"
		task.Phase = "completed"
		task.CurrentRevision = r.Revision
		task.TotalRevision = r.Revision
		r.Status = "ready"
		if task.Kind != "conversion" {
			r.LastSuccess = task.Finished
		}
	}
	if r.Interval > 0 && task.Kind != "conversion" {
		r.NextSync = time.Now().Unix() + r.Interval
	}
	execution.Finished = time.Now().UnixMilli()
	execution.State = task.State
	execution.Error = task.Error
	execution.AfterRevision = r.Revision
	execution.RevisionKnown = task.State == "succeeded"
	if e := w.finish(&r, &task, &execution, err); e != nil {
		log.Print("Cannot persist task result and notification event")
	} else if err == nil {
		w.updateSize(ctx, &r)
	}
}
func (w *Worker) startExecution(task *models.Task, execution *models.TaskExecution) error {
	tx := w.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return err
	}
	result, err := tx.Exec("UPDATE task SET state='running',started=?,finished=0 WHERE id=? AND state='queued'", execution.Started/1000, task.ID)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return errors.New("task is no longer queued")
	}
	if _, err = tx.Insert(execution); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	task.State = "running"
	task.Started = execution.Started / 1000
	task.Finished = 0
	return nil
}
func (w *Worker) finish(r *models.Repository, task *models.Task, execution *models.TaskExecution, failure error) error {
	tx := w.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return err
	}
	if task.Kind != "conversion" {
		if _, err := tx.ID(r.ID).Cols("status", "uuid", "revision", "halted", "last_success", "next_sync").Update(r); err != nil {
			return err
		}
	}
	cols := []string{"state", "phase", "error", "finished"}
	if task.State == "succeeded" {
		cols = append(cols, "current_revision", "total_revision")
	}
	if _, err := tx.ID(task.ID).Cols(cols...).Update(task); err != nil {
		return err
	}
	if _, err := tx.ID(execution.ID).AllCols().Update(execution); err != nil {
		return err
	}
	if r.Kind == "mirror" && task.Kind != "conversion" {
		var org models.Organization
		if _, err := tx.ID(r.OrgID).Get(&org); err != nil {
			return err
		}
		trigger := task.Trigger
		if trigger == "" {
			trigger = "manual"
			if task.Kind == "mirror" {
				trigger = "migration"
			}
		}
		reason := notifications.SVNReason(failure)
		link := "/" + url.PathEscape(org.Name) + "/" + url.PathEscape(r.Name)
		e := notifications.Event{Key: fmt.Sprintf("svn/%d", execution.ID), Kind: "svn", RepositoryID: r.ID, Object: org.Name + "/" + r.Name, Trigger: trigger, ExecutionID: execution.ID, TaskID: task.ID, Success: task.State == "succeeded", Changed: execution.AfterRevision > execution.BeforeRevision, Started: execution.Started, Finished: execution.Finished, BeforeRevision: execution.BeforeRevision, AfterRevision: execution.AfterRevision, RevisionKnown: execution.RevisionKnown, Reason: reason, Phase: task.Phase, Link: link, HistoryLink: link + "/tasks"}
		if err := notifications.Enqueue(tx, e); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (w *Worker) phase(t *models.Task, p string) error {
	t.Phase = p
	_, err := w.DB.ID(t.ID).Cols("phase").Update(t)
	return err
}
func (w *Worker) mirror(ctx context.Context, r *models.Repository, t *models.Task) error {
	if err := w.phase(t, "checking"); err != nil {
		return err
	}
	if err := storage.CheckParents(filepath.Dir(r.Path)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(r.Path), 0700); err != nil {
		return err
	}
	c, err := w.Credentials(r)
	if err != nil {
		return errors.New("cannot decrypt remote credential")
	}
	uuid, root, head, err := w.SVN.Info(ctx, r.Remote, c)
	if err != nil {
		return err
	}
	if strings.TrimRight(root, "/") != strings.TrimRight(r.Remote, "/") {
		return errors.New("source must be the complete repository root")
	}
	completed := r.Revision
	if _, e := os.Stat(r.Path); e == nil {
		localUUID, localRevision, e := w.SVN.LocalInfo(ctx, r.Path)
		if e != nil {
			return e
		}
		if r.UUID != "" && (r.Status == "ready" || localRevision > 0) && localUUID != r.UUID {
			r.Halted = true
			return errors.New("local mirror identity changed; synchronization halted")
		}
		if localRevision > completed {
			completed = localRevision
		}
	}
	if r.UUID != "" && (uuid != r.UUID || head < completed) {
		r.Halted = true
		return errors.New("remote identity or revision changed; synchronization halted")
	}
	if r.UUID == "" {
		r.UUID = uuid
	}
	t.TotalRevision = head
	if _, err = w.DB.ID(t.ID).Cols("total_revision").Update(t); err != nil {
		return err
	}
	if err = w.phase(t, "synchronizing"); err != nil {
		return err
	}
	if r.Status != "ready" {
		if err = w.SVN.CreateMirror(ctx, r.Path, r.Remote, c); err != nil {
			return err
		}
	}
	progressCtx, stopProgress := context.WithCancel(ctx)
	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-progressCtx.Done():
				return
			case <-ticker.C:
				_, revision, e := w.SVN.LocalInfo(progressCtx, r.Path)
				if e == nil {
					_, _ = w.DB.ID(t.ID).Cols("current_revision").Update(&models.Task{CurrentRevision: revision})
				}
			}
		}
	}()
	err = w.SVN.Sync(ctx, r.Path, c)
	stopProgress()
	<-progressDone
	if err != nil {
		return err
	}
	_, propertyHead, err := w.SVN.LocalInfo(ctx, r.Path)
	if err != nil {
		return err
	}
	if err = w.copyRevprops(ctx, r, t, c, propertyHead); err != nil {
		return err
	}
	localUUID, rev, err := w.SVN.LocalInfo(ctx, r.Path)
	if err != nil {
		return err
	}
	if localUUID != uuid {
		return errors.New("mirror UUID verification failed")
	}
	r.UUID = uuid
	r.Revision = rev
	t.CurrentRevision = rev
	_, _ = w.DB.ID(t.ID).Cols("current_revision").Update(t)
	return nil
}
func (w *Worker) local(ctx context.Context, r *models.Repository, t *models.Task) error {
	source, err := filepath.Abs(t.Source)
	if err != nil {
		return err
	}
	dest, err := filepath.Abs(r.Path)
	if err != nil {
		return err
	}
	if !Within(w.Root, dest) || Within(source, dest) || Within(dest, source) {
		return errors.New("invalid migration source or destination")
	}
	if t.Phase == "published" {
		uuid, rev, e := w.SVN.LocalInfo(ctx, dest)
		if e != nil {
			return e
		}
		if _, e = os.Stat(source); e == nil {
			if e = RejectLinks(source); e != nil {
				return e
			}
			u, n, e := w.SVN.LocalInfo(ctx, source)
			if e != nil {
				return e
			}
			if u != uuid || n != rev {
				return errors.New("source changed after publication; retained for manual review")
			}
			if e = w.verifyLocal(ctx, t, dest, rev, "published"); e != nil {
				return e
			}
			if e = os.RemoveAll(source); e != nil {
				return e
			}
		} else if !os.IsNotExist(e) {
			return e
		}
		r.UUID = uuid
		r.Revision = rev
		return nil
	}
	if _, err = os.Stat(dest); err == nil {
		// A rename may complete before its database phase is committed.
		if _, e := os.Stat(source); os.IsNotExist(e) && t.Phase == "moving" {
			uuid, rev, e := w.SVN.LocalInfo(ctx, dest)
			if e != nil {
				return e
			}
			r.UUID = uuid
			r.Revision = rev
			return w.phase(t, "published")
		}
		return errors.New("migration destination exists; manual recovery required")
	}
	if err = w.phase(t, "checking"); err != nil {
		return err
	}
	if err = RejectLinks(source); err != nil {
		return err
	}
	uuid, rev, err := w.SVN.LocalInfo(ctx, source)
	if err != nil {
		return err
	}
	if err = w.verifyLocal(ctx, t, source, rev, "verify-source"); err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return err
	}
	if err = w.phase(t, "moving"); err != nil {
		return err
	}
	if err = os.Rename(source, dest); err == nil {
		r.UUID = uuid
		r.Revision = rev
		return w.phase(t, "published")
	}
	// Cross-volume moves use a verified hotcopy. Never delete the sole valid copy.
	staging := storage.StagingPath(w.Root, r.ID)
	if err = storage.CheckParents(staging); err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(staging), 0700); err != nil {
		return err
	}
	if _, err = os.Stat(staging); err == nil {
		return errors.New("staging directory exists; retain source and recover manually")
	}
	if err = w.phase(t, "copying"); err != nil {
		return err
	}
	copyRunner := w.SVN
	if copyRunner.Timeout == 0 {
		copyRunner.Timeout = svn.LocalMigrationTimeout
	}
	if _, err = copyRunner.Run(ctx, "svnadmin", nil, "hotcopy", source, staging); err != nil {
		return err
	}
	if err = w.verifyLocal(ctx, t, staging, rev, "verify-target"); err != nil {
		return err
	}
	if err = w.phase(t, "checking-identity"); err != nil {
		return err
	}
	u, n, err := w.SVN.LocalInfo(ctx, staging)
	if err != nil {
		return err
	}
	u2, n2, err := w.SVN.LocalInfo(ctx, source)
	if err != nil {
		return err
	}
	if u != uuid || u2 != uuid || n != rev || n2 != rev {
		return errors.New("source changed during migration; retain both copies")
	}
	if err = os.Rename(staging, dest); err != nil {
		return err
	}
	if err = w.phase(t, "published"); err != nil {
		return err
	}
	// The source has been checked as an absolute non-linked directory; dest is verified.
	if err = os.RemoveAll(source); err != nil {
		return fmt.Errorf("repository published; source cleanup required: %w", err)
	}
	r.UUID = uuid
	r.Revision = rev
	return nil
}
func Within(root, p string) bool {
	return storage.Within(root, p)
}
func RejectLinks(root string) error {
	return storage.RejectLinks(root)
}

func (w *Worker) verifyLocal(ctx context.Context, t *models.Task, target string, total int64, phase string) error {
	t.Phase = phase
	t.CurrentRevision = -1
	t.TotalRevision = total
	t.Verification = "metadata"
	t.VerificationPosition = -1
	t.ProgressStarted = time.Now().Unix()
	t.ProgressUpdated = t.ProgressStarted
	cols := []string{"phase", "current_revision", "total_revision", "verification", "verification_position", "progress_started", "progress_updated"}
	save := func() error { _, err := w.DB.ID(t.ID).Cols(cols...).Update(t); return err }
	if err := save(); err != nil {
		return err
	}
	last := time.Time{}
	update := func(stage string, position int64) error {
		changed := stage != t.Verification
		if changed {
			t.Verification = stage
			t.ProgressStarted = time.Now().Unix()
		}
		t.ProgressUpdated = time.Now().Unix()
		if stage == "metadata" {
			t.VerificationPosition = position
		} else {
			t.CurrentRevision = position
		}
		if !changed && time.Since(last) < 2*time.Second {
			return nil
		}
		last = time.Now()
		return save()
	}
	err := w.SVN.Verify(ctx, target, func(revision int64) error {
		if revision < 0 || revision > total {
			return nil
		}
		return update("revisions", revision)
	}, func(position int64) error { return update("metadata", position) })
	if err != nil {
		return err
	}
	t.CurrentRevision = total
	t.Verification = "revisions"
	t.ProgressUpdated = time.Now().Unix()
	return save()
}
