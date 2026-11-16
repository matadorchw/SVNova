// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

import (
	"path/filepath"
	"testing"
)

func TestRepositorySizeUpgradeNormalizesNulls(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	repo := &Repository{OrgID: 1, Name: "legacy", Status: "ready"}
	if _, err = db.Insert(repo); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("UPDATE repository SET size=NULL,size_updated=NULL,size_checked=NULL WHERE id=?", repo.ID); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var stored Repository
	if ok, err := db.Where("size_updated=0 AND size_checked=0").ID(repo.ID).Get(&stored); err != nil || !ok {
		t.Fatal("legacy repository cannot be backfilled", err)
	}
	if stored.Size != 0 {
		t.Fatal("unknown size not normalized")
	}
}
