// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package log

import "log"

func Error(format string, args ...any) { log.Printf(format, args...) }
