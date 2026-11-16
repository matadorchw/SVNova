// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package util

func Iif[T any](condition bool, yes, no T) T {
	if condition {
		return yes
	}
	return no
}
