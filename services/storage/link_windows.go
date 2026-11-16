// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package storage

import (
	"golang.org/x/sys/windows"
	"os"
)

func linked(path string, info os.FileInfo) bool {
	if info.Mode()&os.ModeSymlink != 0 {
		return true
	}
	pointer, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return true
	}
	attrs, err := windows.GetFileAttributes(pointer)
	return err != nil || attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
