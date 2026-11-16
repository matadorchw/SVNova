//go:build !windows

// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package security

import "os"

func protectKey(path string) error { return os.Chmod(path, 0600) }
