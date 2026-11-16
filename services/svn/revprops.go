// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"context"
	"strconv"
	"strings"
)

func (r Runner) CopyRevprops(ctx context.Context, dest string, c Credentials, progress ...func(int64) error) error {
	var onLine func(string) error
	if len(progress) > 0 && progress[0] != nil {
		onLine = func(line string) error {
			n := parseRevpropRevision(line)
			if n < 0 {
				return nil
			}
			return progress[0](n)
		}
	}
	_, err := r.run(ctx, "svnsync", nil, onLine, append([]string{"copy-revprops", LocalURL(dest), "--skip-unchanged"}, c.syncArgs(onLine == nil)...)...)
	return err
}

func parseRevpropRevision(line string) int64 {
	const prefix = "Copied properties for revision "
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, ".") {
		return -1
	}
	n, err := strconv.ParseInt(strings.TrimSuffix(strings.TrimPrefix(line, prefix), "."), 10, 64)
	if err != nil || n < 0 {
		return -1
	}
	return n
}
