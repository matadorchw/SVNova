// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LocalMigrationTimeout allows full-history verification and cross-volume copies of large repositories.
const LocalMigrationTimeout = 24 * time.Hour

type verifyOutput struct {
	sync.Mutex
	partial  string
	tail     string
	progress func(int64) error
	metadata func(int64) error
	err      error
}

func (v *verifyOutput) Write(b []byte) (int, error) {
	v.Lock()
	defer v.Unlock()
	if v.err != nil {
		return 0, v.err
	}
	v.tail += string(b)
	if len(v.tail) > 65536 {
		v.tail = v.tail[len(v.tail)-65536:]
	}
	for _, c := range b {
		if c == '\n' || c == '\r' {
			line := strings.TrimSpace(v.partial)
			v.partial = ""
			if v.metadata != nil {
				position := int64(-2)
				if line == "* Verifying repository metadata ..." {
					position = -1
				}
				if strings.HasPrefix(line, "* Verifying metadata at revision ") {
					if n, e := strconv.ParseInt(strings.TrimSuffix(strings.TrimPrefix(line, "* Verifying metadata at revision "), " ..."), 10, 64); e == nil {
						position = n
					}
				}
				if position >= -1 {
					if e := v.metadata(position); e != nil {
						v.err = e
						return 0, e
					}
				}
			}
			if strings.HasPrefix(line, "* Verified revision ") {
				number := strings.TrimSuffix(strings.TrimPrefix(line, "* Verified revision "), ".")
				if n, e := strconv.ParseInt(number, 10, 64); e == nil {
					if e = v.progress(n); e != nil {
						v.err = e
						return 0, e
					}
				}
			}
		} else {
			v.partial += string(c)
			if len(v.partial) > 4096 {
				v.err = errors.New("verification output line exceeds limit")
				return 0, v.err
			}
		}
	}
	return len(b), nil
}
func (r Runner) Verify(ctx context.Context, target string, progress func(int64) error, metadata ...func(int64) error) error {
	timeout := r.Timeout
	if timeout == 0 {
		timeout = LocalMigrationTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(r.Bin, platformExecutable("svnadmin")), "verify", target)
	hideWindow(cmd)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	output := &verifyOutput{progress: progress}
	if len(metadata) > 0 {
		output.metadata = metadata[0]
	}
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Run()
	if output.err != nil {
		return output.err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return &CommandError{Tool: "svnadmin", Reason: safeError(output.tail)}
	}
	return nil
}
