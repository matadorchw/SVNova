// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyOutputHandlesChunksAndLimits(t *testing.T) {
	var revisions []int64
	output := &verifyOutput{progress: func(n int64) error { revisions = append(revisions, n); return nil }}
	for _, chunk := range []string{"* Verified rev", "ision 0.\r\n* Verified revision 1.", "\n"} {
		if _, err := output.Write([]byte(chunk)); err != nil {
			t.Fatal(err)
		}
	}
	if len(revisions) != 2 || revisions[0] != 0 || revisions[1] != 1 {
		t.Fatal("bad verification progress", revisions)
	}
	if _, err := output.Write([]byte(strings.Repeat("x", 4097))); err == nil {
		t.Fatal("unbounded line accepted")
	}
	sentinel := errors.New("fixture")
	failed := &verifyOutput{progress: func(int64) error { return sentinel }}
	if _, err := failed.Write([]byte("* Verified revision 2.\n")); !errors.Is(err, sentinel) {
		t.Fatal("callback failure ignored")
	}
}

func TestVerifyRealRepositoryProgress(t *testing.T) {
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("set SVN_BIN for real SVN integration")
	}
	r := Runner{Bin: bin}
	repo := filepath.Join(t.TempDir(), "repo")
	if _, err := r.Run(context.Background(), "svnadmin", nil, "create", repo); err != nil {
		t.Fatal(err)
	}
	var seen []int64
	if err := r.Verify(context.Background(), repo, func(n int64) error { seen = append(seen, n); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1 || seen[0] != 0 {
		t.Fatal("real svnadmin progress not parsed", seen)
	}
}

func TestVerifyMetadataNotifications(t *testing.T) {
	var metadata []int64
	var completed []int64
	output := &verifyOutput{metadata: func(n int64) error { metadata = append(metadata, n); return nil }, progress: func(n int64) error { completed = append(completed, n); return nil }}
	for _, chunk := range []string{"* Verifying meta", "data at revision 1000 ...\r\n", "* Verifying repository metadata ...\n", "* Verified revision 0.\n"} {
		if _, err := output.Write([]byte(chunk)); err != nil {
			t.Fatal(err)
		}
	}
	if len(metadata) != 2 || metadata[0] != 1000 || metadata[1] != -1 {
		t.Fatal("metadata locations not parsed", metadata)
	}
	if len(completed) != 1 || completed[0] != 0 {
		t.Fatal("metadata counted as completed revisions", completed)
	}
}

func TestVerifyCancellation(t *testing.T) {
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("set SVN_BIN for real SVN integration")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (Runner{Bin: bin}).Verify(ctx, "unused", func(int64) error { return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal("verification ignored cancellation", err)
	}
}
