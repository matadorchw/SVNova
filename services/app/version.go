// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import "runtime/debug"

// BuildCommit is supplied by release builds, including builds from source archives.
var BuildCommit string

func buildVersion() string {
	revision := BuildCommit
	if revision == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					revision = setting.Value
					break
				}
			}
		}
	}
	if revision == "" {
		return "开发版"
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	return revision
}
