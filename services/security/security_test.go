// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package security

import (
	"path/filepath"
	"testing"
)

func TestSecretsAndPasswords(t *testing.T) {
	h, s, err := Password("long-test-password")
	if err != nil || !Verify("long-test-password", h, s) || Verify("wrong", h, s) {
		t.Fatal("password verification failed")
	}
	p := filepath.Join(t.TempDir(), "key")
	v, err := NewVault(p)
	if err != nil {
		t.Fatal(err)
	}
	c := v.Seal("secret")
	again, err := NewVault(p)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := again.Open(c)
	if err != nil || plain != "secret" {
		t.Fatal("key persistence failed")
	}
	if _, err = v.Open(c[:len(c)-2] + "AA"); err == nil {
		t.Fatal("tampered ciphertext accepted")
	}
}
