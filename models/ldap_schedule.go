// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

type LDAPSchedule struct {
	ID      int64 `xorm:"pk"`
	Time    string
	Updated int64
}
