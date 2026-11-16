// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ConversionError struct {
	Code     string
	Revision int64
}

func (e *ConversionError) Error() string {
	return fmt.Sprintf("mirror conversion %s at revision %d", e.Code, e.Revision)
}

type ConversionSnapshot struct{ Properties map[string][]byte }

func ValidateConversionIdentity(localUUID string, localRevision int64, remoteUUID string, remoteRevision int64) error {
	if localUUID == "" || remoteUUID != localUUID {
		return &ConversionError{Code: "uuid"}
	}
	if remoteRevision < localRevision {
		return &ConversionError{Code: "revision"}
	}
	return nil
}

func (r Runner) RevisionProperties(ctx context.Context, dest string) (map[string][]byte, error) {
	return r.RemoteRevisionProperties(ctx, LocalURL(dest), Credentials{})
}
func (r Runner) RemoteRevisionProperties(ctx context.Context, target string, c Credentials) (map[string][]byte, error) {
	args := append([]string{"proplist", "--revprop", "-r", "0", "--xml", "--verbose", target}, c.Args()...)
	b, err := r.Run(ctx, "svn", c.Input(), args...)
	if err != nil {
		return nil, err
	}
	var result struct {
		Properties []struct {
			Name     string `xml:"name,attr"`
			Encoding string `xml:"encoding,attr"`
			Value    string `xml:",chardata"`
		} `xml:"revprops>property"`
	}
	if err = xml.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	props := map[string][]byte{}
	for _, p := range result.Properties {
		value := []byte(p.Value)
		if p.Encoding == "base64" {
			value, err = base64.StdEncoding.DecodeString(p.Value)
			if err != nil {
				return nil, err
			}
		}
		props[p.Name] = value
	}
	return props, nil
}

func (r Runner) SnapshotConversion(ctx context.Context, dest string) (ConversionSnapshot, error) {
	props, err := r.RevisionProperties(ctx, dest)
	if err != nil {
		return ConversionSnapshot{}, err
	}
	for key := range props {
		if strings.HasPrefix(key, "svn:sync-") {
			return ConversionSnapshot{}, &ConversionError{Code: "metadata"}
		}
	}
	for _, name := range []string{"pre-revprop-change", "post-revprop-change", "start-commit", "pre-commit", "post-commit"} {
		for _, suffix := range []string{"", ".bat", ".cmd", ".exe"} {
			if _, err = os.Lstat(filepath.Join(dest, "hooks", name+suffix)); err == nil {
				return ConversionSnapshot{}, &ConversionError{Code: "hook"}
			} else if !os.IsNotExist(err) {
				return ConversionSnapshot{}, err
			}
		}
	}
	return ConversionSnapshot{Properties: props}, nil
}

func (r Runner) InitializeExistingMirror(ctx context.Context, dest, remote string, c Credentials, uuid string, revision int64) error {
	hook := filepath.Join(dest, "hooks", hookFilename())
	f, err := os.OpenFile(hook, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(hookContent()))
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	args := append([]string{"initialize", LocalURL(dest), remote, "--allow-non-empty"}, c.SyncArgs()...)
	if _, err = r.Run(ctx, "svnsync", nil, args...); err != nil {
		return err
	}
	return r.CheckExistingMirror(ctx, dest, remote, uuid, revision)
}

func (r Runner) CheckExistingMirror(ctx context.Context, dest, remote, uuid string, revision int64) error {
	props, err := r.RevisionProperties(ctx, dest)
	if err != nil {
		return err
	}
	if !SameRepositoryURL(string(props["svn:sync-from-url"]), remote) || string(props["svn:sync-from-uuid"]) != uuid || string(props["svn:sync-last-merged-rev"]) != strconv.FormatInt(revision, 10) || len(props["svn:sync-lock"]) > 0 {
		return &ConversionError{Code: "initialization"}
	}
	return nil
}

func SameRepositoryURL(left, right string) bool {
	a, err := url.Parse(left)
	if err != nil {
		return false
	}
	b, err := url.Parse(right)
	if err != nil {
		return false
	}
	return a.Scheme == b.Scheme && strings.EqualFold(a.Host, b.Host) && strings.TrimRight(a.Path, "/") == strings.TrimRight(b.Path, "/") && a.RawQuery == b.RawQuery && a.Fragment == b.Fragment
}

func (r Runner) RestoreConversion(ctx context.Context, dest string, snapshot ConversionSnapshot) error {
	if snapshot.Properties == nil {
		return errors.New("missing conversion snapshot")
	}
	current, err := r.RevisionProperties(ctx, dest)
	if err != nil {
		return err
	}
	for key := range current {
		if _, ok := snapshot.Properties[key]; !ok {
			if _, err = r.Run(ctx, "svnadmin", nil, "delrevprop", dest, "-r", "0", key); err != nil {
				return err
			}
		}
	}
	for key, value := range snapshot.Properties {
		if existing, ok := current[key]; ok && bytes.Equal(existing, value) {
			continue
		}
		f, err := os.CreateTemp(filepath.Join(dest, "db"), ".svnova-revprop-*")
		if err != nil {
			return err
		}
		name := f.Name()
		_, err = f.Write(value)
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
		if err == nil {
			_, err = r.Run(ctx, "svnadmin", nil, "setrevprop", dest, "-r", "0", key, name)
		}
		os.Remove(name)
		if err != nil {
			return err
		}
	}
	hook := filepath.Join(dest, "hooks", hookFilename())
	if b, err := os.ReadFile(hook); err == nil {
		if !strings.HasPrefix(hookContent(), string(b)) {
			return errors.New("conversion hook changed; manual recovery required")
		}
		return os.Remove(hook)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}
