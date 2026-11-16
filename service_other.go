//go:build !windows

// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package main

import "errors"

func runService(data string) error { return errors.New("this distribution supports Windows x64 only") }
func serviceCommand(data string, args []string) error {
	return errors.New("Windows service commands require Windows")
}
func lockInstance(data string) (func(), error) { return func() {}, nil }
