// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

type NotificationConfig struct {
	ID             int64 `xorm:"pk"`
	Enabled        bool
	SVNEnabled     bool
	LDAPEnabled    bool
	SVNNoChange    bool
	SVNFailureOnly bool
	LDAPNoChange   bool
	Webhook        string `xorm:"TEXT" json:"-"`
	Secret         string `xorm:"TEXT" json:"-"`
	RetentionDays  int
	Generation     int64
	Updated        int64
}

type NotificationDelivery struct {
	ID           int64  `xorm:"pk autoincr"`
	EventKey     string `xorm:"unique"`
	Kind         string
	RepositoryID int64
	Object       string
	Payload      string `xorm:"TEXT" json:"-"`
	Generation   int64
	Status       string `xorm:"index"`
	Attempts     int
	NextAttempt  int64
	Created      int64
	Finished     int64
	LastError    string
	ParentID     int64
}

type NotificationAttempt struct {
	ID         int64 `xorm:"pk autoincr"`
	DeliveryID int64 `xorm:"index"`
	Number     int
	Started    int64
	Finished   int64
	Status     string
	Reason     string
}

type TaskExecution struct {
	Trigger        string
	Error          string
	ID             int64 `xorm:"pk autoincr"`
	TaskID         int64 `xorm:"index"`
	Started        int64
	Finished       int64
	State          string
	BeforeRevision int64
	AfterRevision  int64
	RevisionKnown  bool
}
