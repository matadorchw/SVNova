// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package storage

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"svnova/services/svn"
	"testing"
)

func TestLayoutRejectsWindowsJunctionParent(t *testing.T) {
	db, root, repo := layoutFixture(t)
	outside := t.TempDir()
	link := filepath.Join(root, "wewins")
	output, err := exec.Command("cmd", "/c", "mklink", "/J", link, outside).CombinedOutput()
	if err != nil {
		t.Skipf("junction unavailable: %v %s", err, output)
	}
	t.Cleanup(func() { _ = os.Remove(link) })
	if err = CheckParents(filepath.Join(link, "production")); err == nil {
		t.Fatal("junction ancestor accepted")
	}
	if _, err = DirectorySize(context.Background(), root); err == nil {
		t.Fatal("size traversal accepted junction")
	}
	if _, err = DirectorySize(context.Background(), link); err == nil {
		t.Fatal("size traversal accepted junction root")
	}
	if err = Relocate(context.Background(), db, root, svn.Runner{}); err == nil {
		t.Fatal("layout accepted junction destination")
	}
	if _, err = os.Stat(filepath.Join(repo.Path, "payload")); err != nil {
		t.Fatal("source was modified")
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatal("junction target was modified")
	}
}
