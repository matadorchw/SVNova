// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package bundle

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed payload/*
var payload embed.FS

func Extract(data string) (string, error) {
	b, err := payload.ReadFile("payload/svn.zip")
	if err != nil {
		return "", errors.New("SVN payload missing; run scripts/build.ps1")
	}
	sum := sha256.Sum256(b)
	version := hex.EncodeToString(sum[:8])
	root := filepath.Join(data, "tools", version)
	bin := filepath.Join(root, "bin")
	if verify(root, b) == nil {
		return bin, nil
	}
	if _, err = os.Stat(root); !os.IsNotExist(err) {
		return "", errors.New("tool directory incomplete or modified; stop service and remove it before restart")
	}
	if err = os.MkdirAll(filepath.Dir(root), 0700); err != nil {
		return "", err
	}
	temp, err := os.MkdirTemp(filepath.Dir(root), "extract-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(temp)
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", err
	}
	for _, f := range z.File {
		p := filepath.Join(temp, filepath.FromSlash(f.Name))
		rel, e := filepath.Rel(temp, p)
		if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
			return "", errors.New("invalid payload path")
		}
		if f.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("linked payload")
		}
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(p, 0700); err != nil {
				return "", err
			}
			continue
		}
		if f.UncompressedSize64 > 100<<20 {
			return "", errors.New("oversized payload")
		}
		if err = os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			return "", err
		}
		in, e := f.Open()
		if e != nil {
			return "", e
		}
		out, e := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0700)
		if e != nil {
			in.Close()
			return "", e
		}
		_, err = io.Copy(out, io.LimitReader(in, 100<<20))
		ce := out.Close()
		in.Close()
		if err == nil {
			err = ce
		}
		if err != nil {
			return "", err
		}
	}
	if err = verify(temp, b); err != nil {
		return "", err
	}
	if err = os.Rename(temp, root); err != nil {
		return "", err
	}
	return bin, nil
}
func verify(root string, b []byte) error {
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}
	for _, f := range z.File {
		if f.FileInfo().IsDir() {
			continue
		}
		p := filepath.Join(root, filepath.FromSlash(f.Name))
		info, e := os.Lstat(p)
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("invalid tool file")
		}
		actual, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		in, e := f.Open()
		if e != nil {
			return e
		}
		h := sha256.New()
		_, e = io.Copy(h, in)
		in.Close()
		if e != nil {
			return e
		}
		sum := sha256.Sum256(actual)
		if !bytes.Equal(sum[:], h.Sum(nil)) {
			return errors.New("tool checksum mismatch")
		}
	}
	return nil
}
