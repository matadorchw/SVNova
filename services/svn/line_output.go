// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"context"
	"errors"
	"strings"
)

// lineOutput consumes progress without accumulating output across revisions.
type lineOutput struct {
	line    func(string) error
	cancel  context.CancelFunc
	partial []byte
	err     error
}

func (o *lineOutput) Write(p []byte) (int, error) {
	if o.err != nil {
		return 0, o.err
	}
	for _, b := range p {
		if b == '\r' || b == '\n' {
			o.finish()
		} else {
			o.partial = append(o.partial, b)
			if len(o.partial) > 4096 {
				o.err = errors.New("command output line exceeds limit")
			}
		}
		if o.err != nil {
			o.cancel()
			return len(p), o.err
		}
	}
	return len(p), nil
}
func (o *lineOutput) finish() {
	if o.err != nil || len(o.partial) == 0 {
		return
	}
	o.err = o.line(strings.TrimSpace(string(o.partial)))
	o.partial = o.partial[:0]
}
