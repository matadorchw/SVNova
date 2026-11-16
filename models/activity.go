// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

type RevisionEvent struct {
	ID              int64 `xorm:"pk autoincr"`
	RepositoryID    int64 `xorm:"unique(repo_revision) index"`
	Revision        int64 `xorm:"unique(repo_revision)"`
	Author, Message string
	Timestamp       int64 `xorm:"index"`
	Generation      int64
}
type ActivityIndex struct {
	RepositoryID                                                 int64 `xorm:"pk"`
	HeadRevision, NextRevision, LastSuccess, Generation, Updated int64
	WindowDay                                                    string
	PageSize                                                     int
	Complete, Failed                                             bool
}
