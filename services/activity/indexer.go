// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package activity

import (
	"context"
	"errors"
	"log"
	"svnova/models"
	"svnova/services/svn"
	"time"
	"xorm.io/xorm"
)

var Shanghai = time.FixedZone("Asia/Shanghai", 8*3600)

func Today(now time.Time) time.Time {
	now = now.In(Shanghai)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Shanghai)
}
func Since(now time.Time) time.Time { return Today(now).AddDate(0, 0, -365) }

type Locker interface {
	TryLock() bool
	Unlock()
}
type Indexer struct {
	DB        *xorm.Engine
	SVN       svn.Runner
	Locker    Locker
	BatchSize int
	next      int
}

func (i *Indexer) Run(ctx context.Context) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			i.tick(ctx)
		}
	}
}
func (i *Indexer) tick(ctx context.Context) {
	busy, err := i.DB.Where("state IN (?,?)", "queued", "running").Exist(new(models.Task))
	if err != nil || busy {
		return
	}
	var repos []models.Repository
	if i.DB.Where("status=?", "ready").Asc("id").Find(&repos) != nil || len(repos) == 0 {
		return
	}
	var states []models.ActivityIndex
	if i.DB.Find(&states) != nil {
		return
	}
	byID := map[int64]models.ActivityIndex{}
	for _, state := range states {
		byID[state.RepositoryID] = state
	}
	now := time.Now()
	day := Today(now).Format("2006-01-02")
	for n := 0; n < len(repos); n++ {
		pos := (i.next + n) % len(repos)
		repo := repos[pos]
		state, ok := byID[repo.ID]
		if ok && state.Failed && now.Unix()-state.Updated < 15 {
			continue
		}
		if ok && state.Complete && state.HeadRevision == repo.Revision && state.LastSuccess == repo.LastSuccess && state.WindowDay == day {
			continue
		}
		i.next = (pos + 1) % len(repos)
		if err := i.Step(ctx, repo.ID, now); err != nil && ctx.Err() == nil {
			log.Printf("SVN activity indexing failed for repository %d", repo.ID)
		}
		return
	}
}
func shorten(text string, max int) string {
	runes := []rune(text)
	if len(runes) > max {
		return string(runes[:max]) + "…"
	}
	return text
}
func saveState(tx *xorm.Session, state *models.ActivityIndex) error {
	_, err := tx.Exec(`INSERT INTO activity_index (repository_id,head_revision,next_revision,last_success,generation,updated,window_day,page_size,complete,failed) VALUES (?,?,?,?,?,?,?,?,?,?) ON CONFLICT(repository_id) DO UPDATE SET head_revision=excluded.head_revision,next_revision=excluded.next_revision,last_success=excluded.last_success,generation=excluded.generation,updated=excluded.updated,window_day=excluded.window_day,page_size=excluded.page_size,complete=excluded.complete,failed=excluded.failed`, state.RepositoryID, state.HeadRevision, state.NextRevision, state.LastSuccess, state.Generation, state.Updated, state.WindowDay, state.PageSize, state.Complete, state.Failed)
	return err
}
func (i *Indexer) Step(ctx context.Context, repoID int64, now time.Time) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if i.Locker != nil {
		if !i.Locker.TryLock() {
			return nil
		}
		defer i.Locker.Unlock()
	}
	repo := new(models.Repository)
	ok, err := i.DB.ID(repoID).Get(repo)
	if err != nil {
		return err
	}
	if !ok || repo.Status != "ready" {
		return nil
	}
	state := &models.ActivityIndex{RepositoryID: repoID}
	found, err := i.DB.ID(repoID).Get(state)
	if err != nil {
		return err
	}
	day := Today(now).Format("2006-01-02")
	batch := i.BatchSize
	if batch < 1 || batch > 500 {
		batch = 500
	}
	if !found || state.HeadRevision > repo.Revision || (state.Complete && (state.HeadRevision != repo.Revision || state.LastSuccess != repo.LastSuccess || state.WindowDay != day)) {
		state.HeadRevision = repo.Revision
		state.NextRevision = repo.Revision
		state.LastSuccess = repo.LastSuccess
		state.WindowDay = day
		state.Generation++
		state.PageSize = batch
		state.Complete = false
		state.Failed = false
	}
	if state.Complete {
		return nil
	}
	if state.PageSize < 1 {
		state.PageSize = batch
	}
	var entries []svn.LogEntry
	if state.NextRevision > 0 {
		pageCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		entries, err = i.SVN.LogPage(pageCtx, repo.Path, state.HeadRevision, state.NextRevision, state.PageSize)
		cancel()
	}
	events := []models.RevisionEvent{}
	next := state.NextRevision
	if err == nil {
		for _, entry := range entries {
			if entry.Revision < 1 || entry.Revision > state.NextRevision {
				err = errors.New("invalid revision page")
				break
			}
			date, e := time.Parse(time.RFC3339Nano, entry.Date)
			if e != nil {
				err = e
				break
			}
			if date.Unix() >= Since(now).Unix() && date.Before(Today(now).AddDate(0, 0, 1)) {
				events = append(events, models.RevisionEvent{RepositoryID: repoID, Revision: entry.Revision, Author: shorten(entry.Author, 256), Message: shorten(entry.Message, 2000), Timestamp: date.Unix(), Generation: state.Generation})
			}
			if entry.Revision-1 < next {
				next = entry.Revision - 1
			}
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	tx := i.DB.NewSession()
	defer tx.Close()
	if e := tx.Begin(); e != nil {
		return e
	}
	state.Updated = now.Unix()
	if err != nil {
		state.Failed = true
		if state.PageSize > 1 {
			state.PageSize = max(1, state.PageSize/10)
		}
		if e := saveState(tx, state); e != nil {
			return e
		}
		if e := tx.Commit(); e != nil {
			return e
		}
		return err
	}
	for _, event := range events {
		if _, e := tx.Exec(`INSERT INTO revision_event (repository_id,revision,author,message,timestamp,generation) VALUES (?,?,?,?,?,?) ON CONFLICT(repository_id,revision) DO UPDATE SET author=excluded.author,message=excluded.message,timestamp=excluded.timestamp,generation=excluded.generation`, event.RepositoryID, event.Revision, event.Author, event.Message, event.Timestamp, event.Generation); e != nil {
			return e
		}
	}
	state.NextRevision = next
	state.Complete = next < 1 || len(entries) < state.PageSize
	state.Failed = false
	if state.Complete {
		if _, e := tx.Where("repository_id=? AND generation<>?", repoID, state.Generation).Delete(new(models.RevisionEvent)); e != nil {
			return e
		}
	}
	if _, e := tx.Where("repository_id=? AND timestamp<?", repoID, Since(now).Unix()).Delete(new(models.RevisionEvent)); e != nil {
		return e
	}
	if e := saveState(tx, state); e != nil {
		return e
	}
	return tx.Commit()
}
