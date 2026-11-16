// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCredentialInputPreservesExactPassword(t *testing.T) {
	for _, password := range []string{"!example-password", " spaced password ", "密码例子", "literal-newline\n"} {
		input, err := io.ReadAll((Credentials{Password: password}).Input())
		if err != nil {
			t.Fatal(err)
		}
		if string(input) != password {
			t.Fatal("password input was modified")
		}
	}
	if (Credentials{}).Input() != nil {
		t.Fatal("anonymous credentials provided password input")
	}
}

func TestSafeErrorPrefersInnerCauseAndHidesSecrets(t *testing.T) {
	diagnostic := "svn: E170013: Unable to connect to https://private.example/repo\nsvn: E215004: No more credentials for alice:secret\nAuthentication failed"
	result := safeError(diagnostic)
	if !strings.Contains(result, "E215004") || strings.Contains(result, "secret") || strings.Contains(result, "alice") || strings.Contains(result, "private.example") {
		t.Fatalf("unsafe or unhelpful diagnostic: %s", result)
	}
	if !strings.Contains(safeError("E170013 followed by E230001"), "certificate") {
		t.Fatal("TLS cause hidden by connection wrapper")
	}
	public := (&CommandError{Tool: "svn", Reason: result}).UserMessage()
	if !strings.Contains(public, "认证失败") || strings.Contains(public, "secret") {
		t.Fatalf("incorrect public error: %s", public)
	}
}

func TestRemoteAndPaths(t *testing.T) {
	for _, s := range []string{"file:///repo", "svn://host/repo", "https://user:pass@host/repo", "https://host/repo?password=secret"} {
		if ValidateRemote(s) == nil {
			t.Fatalf("accepted %s", s)
		}
	}
	if ValidateRemote("https://host/repo") != nil {
		t.Fatal("valid remote rejected")
	}
	for _, p := range []string{"../private", "dir/../../private", `dir\..\private`} {
		if _, err := Target(t.TempDir(), p); err == nil {
			t.Fatalf("accepted %s", p)
		}
	}
}
func TestRealRepository(t *testing.T) {
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("set SVN_BIN for real SVN integration")
	}
	r := Runner{Bin: bin}
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if _, err := r.Run(ctx, "svnadmin", nil, "create", source); err != nil {
		t.Fatal(err)
	}
	working := filepath.Join(root, "import")
	_ = os.Mkdir(working, 0700)
	_ = os.WriteFile(filepath.Join(working, "中文.txt"), []byte("hello\n"), 0600)
	if _, err := r.Run(ctx, "svn", nil, "import", working, LocalURL(source), "-m", "initial", "--non-interactive"); err != nil {
		t.Fatal(err)
	}
	uuid, rev, err := r.LocalInfo(ctx, source)
	if err != nil || rev != 1 || uuid == "" {
		t.Fatalf("info: %s %d %v", uuid, rev, err)
	}
	entries, err := r.List(ctx, source, "", 1)
	if err != nil || len(entries) != 1 || entries[0].Name != "中文.txt" {
		t.Fatalf("list: %+v %v", entries, err)
	}
	b, err := r.Cat(ctx, source, "中文.txt", 1)
	if err != nil || string(b) != "hello\n" {
		t.Fatalf("cat: %q %v", b, err)
	}
	logs, err := r.Log(ctx, source, "", 1)
	if err != nil || len(logs) != 1 {
		t.Fatalf("log: %v %v", logs, err)
	}
	// Verify svnsync mechanics using a file source; HTTP transport needs separate acceptance.
	dest := filepath.Join(root, "mirror")
	if _, err = r.Run(ctx, "svnadmin", nil, "create", dest); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(dest, "hooks", hookFilename()), []byte(hookContent()), 0700)
	if _, err = r.Run(ctx, "svnadmin", nil, "setuuid", dest, uuid); err != nil {
		t.Fatal(err)
	}
	if _, err = r.Run(ctx, "svnsync", nil, append([]string{"initialize", LocalURL(dest), LocalURL(source)}, Credentials{}.SyncArgs()...)...); err != nil {
		t.Fatal(err)
	}
	if err = r.Sync(ctx, dest, Credentials{}); err != nil {
		t.Fatal(err)
	}
	if err = r.CopyRevprops(ctx, dest, Credentials{}); err != nil {
		t.Fatal(err)
	}
	_, n, err := r.LocalInfo(ctx, dest)
	if err != nil || n != 1 {
		t.Fatalf("mirror: %d %v", n, err)
	}
	_ = os.WriteFile(filepath.Join(source, "hooks", hookFilename()), []byte(hookContent()), 0700)
	if _, err = r.Run(ctx, "svn", nil, "propset", "--revprop", "-r", "1", "svn:log", "changed", LocalURL(source), "--non-interactive"); err != nil {
		t.Fatal(err)
	}
	if err = r.CopyRevprops(ctx, dest, Credentials{}); err != nil {
		t.Fatal(err)
	}
	logs, err = r.Log(ctx, dest, "", 1)
	if err != nil || logs[0].Message != "changed" {
		t.Fatalf("revprops: %v %v", logs, err)
	}
}
