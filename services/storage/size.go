// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package storage

import (
	"context"
	"errors"
	"io/fs"
	"math"
	"os"
	"path/filepath"
)

// DirectorySize sums file lengths, not allocated disk blocks or checked-out content.
func DirectorySize(ctx context.Context, root string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if err := CheckParents(root); err != nil {
		return 0, err
	}
	var total int64
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if cancel := ctx.Err(); cancel != nil {
			return cancel
		}
		if os.IsNotExist(err) && path != root {
			return nil
		}
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if os.IsNotExist(err) && path != root {
			return nil
		}
		if err != nil {
			return err
		}
		if linked(path, info) {
			return errors.New("linked repository paths are not supported")
		}
		if info.Mode().IsRegular() {
			if info.Size() > math.MaxInt64-total {
				return errors.New("repository size overflow")
			}
			total += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return total, nil
}
