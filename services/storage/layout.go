// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"svnova/models"
	"svnova/services/security"
	"svnova/services/svn"
	"time"
	"xorm.io/xorm"
)

const markerFile = ".svnova-layout-move-"

var validToken = regexp.MustCompile(`^[a-f0-9]{64}$`)

type plan struct {
	repo      models.Repository
	move      models.RepositoryPathMove
	persisted bool
}

func directory(path string) (bool, error) {
	if err := CheckParents(path); err != nil {
		return false, err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.IsDir() {
		return false, errors.New("repository path is not a directory")
	}
	return true, nil
}
func markerMatches(dir, token string) bool {
	path := filepath.Join(dir, markerFile+token)
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Size() == 0 && !linked(path, info)
}
func checkTransfer(old, target, token string, resume, committed, required bool) error {
	source, err := directory(old)
	if err != nil {
		return err
	}
	dest, err := directory(target)
	if err != nil {
		return err
	}
	if source && dest {
		return fmt.Errorf("both source and destination exist: %s", target)
	}
	if dest && (!resume || (!committed && !markerMatches(target, token))) {
		return fmt.Errorf("destination is occupied or has no matching move marker: %s", target)
	}
	if !source && !dest && required {
		return fmt.Errorf("repository directory missing: %s", old)
	}
	if source {
		if err := RejectLinks(old); err != nil {
			return err
		}
		_, err := os.Lstat(filepath.Join(old, markerFile+token))
		if err == nil && !markerMatches(old, token) {
			return errors.New("repository contains a conflicting move marker")
		}
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
func verify(ctx context.Context, runner svn.Runner, path, uuid string, revision int64) error {
	if uuid == "" {
		return nil
	}
	timeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	actual, head, err := runner.LocalInfo(timeout, path)
	if err != nil {
		return err
	}
	if actual != uuid || head != revision {
		return errors.New("repository identity changed during layout move")
	}
	return nil
}
func moveDirectory(old, target, token string) error {
	exists, err := directory(old)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	if _, err = os.Lstat(target); err == nil {
		return fmt.Errorf("destination already exists: %s", target)
	} else if !os.IsNotExist(err) {
		return err
	}
	marker := filepath.Join(old, markerFile+token)
	if !markerMatches(old, token) {
		file, err := os.OpenFile(marker, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		err = file.Sync()
		closeErr := file.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	if err = CheckParents(filepath.Dir(target)); err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		return err
	}
	if err = CheckParents(filepath.Dir(target)); err != nil {
		return err
	}
	return os.Rename(old, target)
}

// Relocate runs under the application's instance lock before requests or workers start.
func Relocate(ctx context.Context, db *xorm.Engine, root string, runner svn.Runner) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	if err = CheckParents(root); err != nil {
		return err
	}
	var repos []models.Repository
	if err = db.Asc("id").Find(&repos); err != nil {
		return err
	}
	var records []models.RepositoryPathMove
	if err = db.Find(&records); err != nil {
		return err
	}
	journal := map[int64]models.RepositoryPathMove{}
	for _, record := range records {
		journal[record.RepositoryID] = record
	}
	targets := map[string]int64{}
	sources := map[string]int64{}
	var plans []plan
	for _, repo := range repos {
		sourceKey := strings.ToLower(filepath.Clean(repo.Path))
		if other, ok := sources[sourceKey]; ok {
			return fmt.Errorf("repositories %d and %d share a directory", other, repo.ID)
		}
		sources[sourceKey] = repo.ID
		org := new(models.Organization)
		ok, err := db.ID(repo.OrgID).Get(org)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("repository %d has no organization", repo.ID)
		}
		target, err := RepositoryPath(root, org.Name, repo.Name)
		if err != nil {
			return err
		}
		key := strings.ToLower(filepath.Clean(target))
		if other, ok := targets[key]; ok {
			return fmt.Errorf("repositories %d and %d have conflicting Windows paths", other, repo.ID)
		}
		targets[key] = repo.ID
		record, resume := journal[repo.ID]
		delete(journal, repo.ID)
		if !resume && filepath.Clean(repo.Path) == target {
			exists, err := directory(target)
			if err != nil {
				return err
			}
			if !exists && repo.Status == "ready" {
				return fmt.Errorf("ready repository %d is missing", repo.ID)
			}
			continue
		}
		if !resume {
			record = models.RepositoryPathMove{RepositoryID: repo.ID, OldPath: repo.Path, NewPath: target, Token: security.Token()}
			record.HadRepository, _ = directory(repo.Path)
			record.HadStaging, _ = directory(repo.Path + ".importing")
		}
		if record.NewPath != target || !validToken.MatchString(record.Token) || !filepath.IsAbs(record.OldPath) || (repo.Path != record.OldPath && repo.Path != record.NewPath) {
			return fmt.Errorf("invalid layout journal for repository %d", repo.ID)
		}
		if !Within(root, record.OldPath) || !Within(root, record.NewPath) {
			return errors.New("layout move escapes repository root")
		}
		committed := repo.Path == record.NewPath
		if err = checkTransfer(record.OldPath, record.NewPath, record.Token, resume, committed, repo.Status == "ready" || record.HadRepository); err != nil {
			return err
		}
		if err = checkTransfer(record.OldPath+".importing", StagingPath(root, repo.ID), record.Token, resume, committed, record.HadStaging); err != nil {
			return err
		}
		if !resume {
			if exists, _ := directory(record.OldPath); exists {
				timeout, cancel := context.WithTimeout(ctx, 30*time.Second)
				uuid, rev, e := runner.LocalInfo(timeout, record.OldPath)
				cancel()
				if e != nil && repo.Status == "ready" {
					return fmt.Errorf("cannot verify repository %d: %w", repo.ID, e)
				}
				if e == nil {
					if repo.UUID != "" && repo.UUID != uuid {
						return fmt.Errorf("repository %d UUID differs from database", repo.ID)
					}
					record.UUID = uuid
					record.Revision = rev
				}
			}
		}
		plans = append(plans, plan{repo: repo, move: record, persisted: resume})
	}
	if len(journal) != 0 {
		return errors.New("layout journal refers to a missing repository")
	}
	// An old random directory must never become another repository's organization directory.
	for _, p := range plans {
		for _, repo := range repos {
			if Within(repo.Path, StagingPath(root, p.repo.ID)) {
				return errors.New("repository occupies the staging namespace")
			}
			if repo.ID != p.repo.ID && (Within(repo.Path, p.move.NewPath) || Within(p.move.OldPath, repo.Path) || strings.EqualFold(repo.Path, p.move.NewPath)) {
				return errors.New("repository directories overlap")
			}
			if repo.ID != p.repo.ID && strings.EqualFold(repo.Path, p.move.OldPath+".importing") {
				return errors.New("legacy staging directory is another repository")
			}
		}
	}
	tx := db.NewSession()
	defer tx.Close()
	if err = tx.Begin(); err != nil {
		return err
	}
	version, err := tx.ID(3).Exist(new(models.Migration))
	if err != nil {
		return err
	}
	if !version {
		if _, err = tx.Insert(&models.Migration{ID: 3, Applied: time.Now().Unix()}); err != nil {
			return err
		}
	}
	for _, p := range plans {
		if !p.persisted {
			if _, err = tx.Insert(&p.move); err != nil {
				return err
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	for _, p := range plans {
		if err = finishMove(ctx, db, root, runner, p); err != nil {
			return fmt.Errorf("layout move for repository %d: %w", p.repo.ID, err)
		}
	}
	return nil
}
func finishMove(ctx context.Context, db *xorm.Engine, root string, runner svn.Runner, p plan) error {
	m := p.move
	if err := moveDirectory(m.OldPath, m.NewPath, m.Token); err != nil {
		return err
	}
	if err := moveDirectory(m.OldPath+".importing", StagingPath(root, p.repo.ID), m.Token); err != nil {
		return err
	}
	if exists, err := directory(m.NewPath); err != nil {
		return err
	} else if exists {
		if err = verify(ctx, runner, m.NewPath, m.UUID, m.Revision); err != nil {
			return err
		}
	} else if p.repo.Status == "ready" || m.HadRepository || m.UUID != "" {
		return errors.New("repository directory disappeared during move")
	}
	if m.HadStaging {
		if exists, err := directory(StagingPath(root, p.repo.ID)); err != nil {
			return err
		} else if !exists {
			return errors.New("staging directory disappeared during move")
		}
	}
	affected, err := db.ID(p.repo.ID).Where("path=? OR path=?", m.OldPath, m.NewPath).Cols("path").Update(&models.Repository{Path: m.NewPath})
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.New("repository path changed during layout move")
	}
	for _, target := range []string{m.NewPath, StagingPath(root, p.repo.ID)} {
		if markerMatches(target, m.Token) {
			if err = os.Remove(filepath.Join(target, markerFile+m.Token)); err != nil {
				return err
			}
		}
	}
	_, err = db.ID(p.repo.ID).Delete(new(models.RepositoryPathMove))
	return err
}
