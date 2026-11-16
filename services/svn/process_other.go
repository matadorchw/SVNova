//go:build !windows

// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import "os/exec"

func hideWindow(cmd *exec.Cmd)           {}
func platformExecutable(s string) string { return s }
func hookFilename() string               { return "pre-revprop-change" }
func hookContent() string                { return "#!/bin/sh\nexit 0\n" }
