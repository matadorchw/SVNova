// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRevpropOutputChunksAndLimits(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var revisions []int64
	observer := &lineOutput{cancel: cancel, line: func(line string) error {
		if n := parseRevpropRevision(line); n >= 0 {
			revisions = append(revisions, n)
		}
		return nil
	}}
	for _, part := range []string{"Copied proper", "ties for revision 0.\r", "\nignored\nCopied properties for revision 2."} {
		if _, err := observer.Write([]byte(part)); err != nil {
			t.Fatal(err)
		}
	}
	observer.finish()
	if len(revisions) != 2 || revisions[0] != 0 || revisions[1] != 2 {
		t.Fatal(revisions)
	}
	for _, line := range []string{"Copied properties for revision -1.", "Copied properties for revision x.", "Copied properties for revision 1", "Copied properties for revision 9223372036854775808."} {
		if parseRevpropRevision(line) != -1 {
			t.Fatal(line)
		}
	}
	failed := errors.New("persist failed")
	observer = &lineOutput{cancel: cancel, line: func(string) error { return failed }}
	if _, err := observer.Write([]byte("line\n")); !errors.Is(err, failed) || ctx.Err() == nil {
		t.Fatal("callback failure did not cancel", err)
	}
	observer = &lineOutput{cancel: cancel, line: func(string) error { return nil }}
	if _, err := observer.Write([]byte(strings.Repeat("x", 4097))); err == nil {
		t.Fatal("unbounded partial output")
	}
}

func TestRevpropProgressExceedsCaptureLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	count := 0
	observer := &lineOutput{cancel: cancel, line: func(line string) error {
		if parseRevpropRevision(line) != 1234 {
			t.Fatal("corrupt progress line")
		}
		count++
		return nil
	}}
	line := []byte("Copied properties for revision 1234.\r\n")
	lines := MaxOutput/len(line) + 1024
	for i := 0; i < lines; i++ {
		if n, err := observer.Write(line); err != nil || n != len(line) {
			t.Fatalf("progress stopped after %d bytes: %v", i*len(line), err)
		}
	}
	observer.finish()
	if count != lines || len(observer.partial) != 0 || cap(observer.partial) > 4096 || ctx.Err() != nil {
		t.Fatalf("progress not bounded or incomplete: %d/%d", count, lines)
	}
}
