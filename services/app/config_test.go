// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"svnova/services/security"
)

func TestINIPathsAndKeyRoundTrip(t *testing.T) {
	root := t.TempDir()
	a, err := New(root, "", nil, http.NotFoundHandler())
	if err != nil {
		t.Fatal(err)
	}
	if a.Data != filepath.Join(root, "data") {
		t.Fatal(a.Data)
	}
	for _, path := range []string{"custom/conf", "data", "log"} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatal(err)
		}
	}
	a.Config = Config{Listen: "127.0.0.1:33003", Origin: "http://localhost:33003", Database: filepath.Join(a.Data, "test.db"), Repositories: filepath.Join(a.Data, "repositories"), SecretKey: filepath.Join(a.Data, "密钥 #1.key")}
	vault, err := security.NewVault(a.Config.SecretKey)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext := vault.Seal("isolated-service-password")
	key, err := os.ReadFile(a.Config.SecretKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.saveConfig(); err != nil {
		t.Fatal(err)
	}
	b := &App{WorkRoot: root}
	installed, err := b.loadConfig()
	if err != nil || !installed || b.Config != a.Config || b.Data != a.Data {
		t.Fatalf("config roundtrip: installed=%v err=%v config=%+v", installed, err, b.Config)
	}
	after, _ := os.ReadFile(b.Config.SecretKey)
	if string(after) != string(key) {
		t.Fatal("key changed")
	}
	reopened, err := security.NewVault(b.Config.SecretKey)
	if err != nil {
		t.Fatal(err)
	}
	plaintext, err := reopened.Open(ciphertext)
	if err != nil || plaintext != "isolated-service-password" {
		t.Fatal("credential did not survive INI reload")
	}
	if _, err := os.Stat(filepath.Join(root, "config.json")); !os.IsNotExist(err) {
		t.Fatal("legacy config was created")
	}
	if err := os.Remove(b.Config.SecretKey); err != nil {
		t.Fatal(err)
	}
	if _, err := New(root, "", nil, http.NotFoundHandler()); err == nil {
		t.Fatal("missing installed key accepted")
	}
	if _, err := os.Stat(b.Config.SecretKey); !os.IsNotExist(err) {
		t.Fatal("missing key regenerated")
	}
}

func TestMirrorConversionTimeoutConfiguration(t *testing.T) {
	root := t.TempDir()
	if err := PrepareWorkPath(root); err != nil {
		t.Fatal(err)
	}
	a := &App{WorkRoot: root, Data: filepath.Join(root, "data"), Config: Config{Listen: "127.0.0.1:3000", Database: filepath.Join(root, "db"), Repositories: filepath.Join(root, "repositories"), SecretKey: filepath.Join(root, "key")}}
	for _, value := range []time.Duration{0, 48 * time.Hour, -time.Second} {
		a.Config.MirrorConversionTimeout = value
		if err := a.saveConfig(); err != nil {
			t.Fatal(err)
		}
		b := &App{WorkRoot: root}
		_, err := b.loadConfig()
		if value < 0 {
			if err == nil {
				t.Fatal("negative timeout accepted")
			}
			continue
		}
		if err != nil || b.Config.MirrorConversionTimeout != conversionTimeout(value) {
			t.Fatalf("timeout roundtrip %v %v", b.Config.MirrorConversionTimeout, err)
		}
	}
}
