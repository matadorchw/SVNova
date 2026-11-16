// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"os/exec"
	"syscall"
)

func hideWindow(cmd *exec.Cmd)           { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
func platformExecutable(s string) string { return s + ".exe" }
func hookFilename() string               { return "pre-revprop-change.bat" }
func hookContent() string                { return "@echo off\r\nexit /b 0\r\n" }
