// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"

	"svnova/modules/auth/password/hash"
)

func Token() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
func Password(password string) (string, string, error) {
	salt := Token()
	h, err := hash.Parse("pbkdf2$320000$50").Hash(password, salt)
	return h, salt, err
}
func Verify(password, stored, salt string) bool {
	return hash.Parse("pbkdf2$320000$50").VerifyPassword(password, stored, salt)
}

type Vault struct{ aead cipher.AEAD }

func NewVault(path string) (*Vault, error) {
	key, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		key = make([]byte, 32)
		if _, err = rand.Read(key); err != nil {
			return nil, err
		}
		f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return nil, e
		}
		_, err = f.Write(key)
		ce := f.Close()
		if err == nil {
			err = ce
		}
	}
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.New("invalid encryption key")
	}
	if err = protectKey(path); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	a, err := cipher.NewGCM(block)
	return &Vault{a}, err
}
func (v *Vault) Seal(s string) string {
	nonce := make([]byte, v.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(v.aead.Seal(nonce, nonce, []byte(s), nil))
}
func (v *Vault) Open(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil || len(b) < v.aead.NonceSize() {
		return "", errors.New("invalid encrypted credential")
	}
	p, err := v.aead.Open(nil, b[:v.aead.NonceSize()], b[v.aead.NonceSize():], nil)
	return string(p), err
}
