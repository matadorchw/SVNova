// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestConversionIdentityRejectsMismatches(t *testing.T) {
	for _, test := range []struct {
		local, remote       string
		localRev, remoteRev int64
		valid               bool
	}{
		{"uuid", "uuid", 2, 2, true}, {"uuid", "uuid", 2, 3, true}, {"uuid", "other", 2, 2, false}, {"uuid", "uuid", 2, 1, false}, {"", "", 0, 0, false},
	} {
		err := ValidateConversionIdentity(test.local, test.localRev, test.remote, test.remoteRev)
		if (err == nil) != test.valid {
			t.Fatalf("unexpected identity result: %v", err)
		}
	}
}

func TestRealExistingMirrorConversionAndRollback(t *testing.T) {
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("set SVN_BIN for real SVN conversion integration")
	}
	r := Runner{Bin: bin}
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	dest := filepath.Join(root, "dest")
	run := func(tool string, args ...string) []byte {
		t.Helper()
		b, err := r.Run(ctx, tool, nil, args...)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	run("svnadmin", "create", source)
	fixture := filepath.Join(root, "fixture")
	if err := os.Mkdir(fixture, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture, "中文.txt"), []byte("first\n"), 0600); err != nil {
		t.Fatal(err)
	}
	run("svn", "import", fixture, LocalURL(source), "-m", "first", "--non-interactive")
	run("svnadmin", "hotcopy", source, dest)
	uuid, revision, err := r.LocalInfo(ctx, dest)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := r.SnapshotConversion(ctx, dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Properties) == 0 {
		t.Fatal("revision-zero properties were not captured")
	}
	before := run("svn", "cat", "-r", "1", LocalURL(dest)+"/"+url.PathEscape("中文.txt"), "--non-interactive")
	if err = r.InitializeExistingMirror(ctx, dest, LocalURL(source), Credentials{}, uuid, revision); err != nil {
		props, _ := r.RevisionProperties(ctx, dest)
		for key, value := range props {
			t.Logf("fixture %s=%s", key, value)
		}
		t.Fatal(err)
	}
	if _, err = r.SnapshotConversion(ctx, dest); err == nil {
		t.Fatal("existing sync metadata accepted")
	}
	if err = r.RestoreConversion(ctx, dest, snapshot); err != nil {
		t.Fatal(err)
	}
	// Failed initialization must also undo its hook and partial r0 changes.
	if err = r.InitializeExistingMirror(ctx, dest, LocalURL(filepath.Join(root, "missing-source")), Credentials{}, uuid, revision); err == nil {
		t.Fatal("missing source initialized")
	}
	if err = r.RestoreConversion(ctx, dest, snapshot); err != nil {
		t.Fatal(err)
	}
	restored, err := r.SnapshotConversion(ctx, dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(restored.Properties) != len(snapshot.Properties) {
		t.Fatal("revision-zero properties changed")
	}
	customHook := filepath.Join(dest, "hooks", "post-commit.bat")
	if err = os.WriteFile(customHook, []byte("fixture hook"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = r.SnapshotConversion(ctx, dest); err == nil {
		t.Fatal("custom commit hook accepted")
	}
	if data, err := os.ReadFile(customHook); err != nil || string(data) != "fixture hook" {
		t.Fatal("custom hook modified")
	}
	if err = os.Remove(customHook); err != nil {
		t.Fatal(err)
	}
	for key, value := range snapshot.Properties {
		if string(restored.Properties[key]) != string(value) {
			t.Fatal("property rollback mismatch")
		}
	}
	after := run("svn", "cat", "-r", "1", LocalURL(dest)+"/"+url.PathEscape("中文.txt"), "--non-interactive")
	if string(before) != string(after) {
		t.Fatal("history changed during rollback")
	}
	if err = r.InitializeExistingMirror(ctx, dest, LocalURL(source), Credentials{}, uuid, revision); err != nil {
		t.Fatal(err)
	}
	// Append a source revision and exercise normal incremental synchronization.
	run("svn", "mkdir", LocalURL(source)+"/next", "-m", "next", "--non-interactive")
	if err = r.Sync(ctx, dest, Credentials{}); err != nil {
		t.Fatal(err)
	}
	if err = r.CopyRevprops(ctx, dest, Credentials{}); err != nil {
		t.Fatal(err)
	}
	_, final, err := r.LocalInfo(ctx, dest)
	if err != nil || final != 2 {
		t.Fatalf("sync head %d %v", final, err)
	}
}
