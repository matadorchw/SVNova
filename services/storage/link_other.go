//go:build !windows

// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package storage

import "os"

func linked(_ string, info os.FileInfo) bool { return info.Mode()&os.ModeSymlink != 0 }
