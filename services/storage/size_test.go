// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDirectorySize(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"one": "123", "nested/two": "12345"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if size, err := DirectorySize(context.Background(), root); err != nil || size != 8 {
		t.Fatalf("size %d: %v", size, err)
	}
	if size, err := DirectorySize(context.Background(), t.TempDir()); err != nil || size != 0 {
		t.Fatalf("empty size %d: %v", size, err)
	}
	if _, err := DirectorySize(context.Background(), filepath.Join(root, "missing")); err == nil {
		t.Fatal("missing repository reported empty")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := DirectorySize(ctx, root); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestDirectorySizeRejectsLinks(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Skipf("symlink privilege unavailable: %v", err)
	}
	if _, err := DirectorySize(context.Background(), root); err == nil {
		t.Fatal("linked directory accepted")
	}
	if _, err := DirectorySize(context.Background(), filepath.Join(root, "linked")); err == nil {
		t.Fatal("linked root accepted")
	}
}
