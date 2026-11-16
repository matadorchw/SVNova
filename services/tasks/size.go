// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"errors"
	"log"
	"svnova/models"
	"svnova/services/storage"
	"time"
)

// Size collection shares the mutation lock but never changes a task's outcome.
func (w *Worker) updateSize(ctx context.Context, r *models.Repository) {
	checked := time.Now().Unix()
	snapshot := models.Repository{SizeChecked: checked}
	columns := []string{"size_checked"}
	collect := func() (int64, error) {
		if !storage.Within(w.Root, r.Path) {
			return 0, errors.New("repository outside managed root")
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		return storage.DirectorySize(ctx, r.Path)
	}
	size, err := collect()
	if err == nil {
		snapshot.Size = size
		snapshot.SizeUpdated = time.Now().Unix()
		columns = append(columns, "size", "size_updated")
	} else if ctx.Err() == nil {
		log.Printf("Cannot collect size for repository %d", r.ID)
	}
	if _, err = w.DB.ID(r.ID).Cols(columns...).Update(&snapshot); err != nil {
		log.Printf("Cannot persist size for repository %d", r.ID)
	}
}

func (w *Worker) backfillSize(ctx context.Context) {
	var repo models.Repository
	// Retry failed measurements after five minutes, allowing other repositories to proceed.
	ok, err := w.DB.Where("status=? AND size_updated=0 AND size_checked<=?", "ready", time.Now().Unix()-300).Asc("size_checked", "id").Get(&repo)
	if err == nil && ok {
		w.updateSize(ctx, &repo)
	}
}
