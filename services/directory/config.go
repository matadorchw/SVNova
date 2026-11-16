// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"errors"
	"net/url"
	"strings"
	"svnova/models"
)

var ErrBindPassword = errors.New("LDAP bind password required")
var ErrSyncDisabled = errors.New("LDAP user synchronization is disabled")
var ErrEmptyResult = errors.New("empty LDAP result; users retained for manual review")

type LocalAccountConflict struct{ Name string }

func (e *LocalAccountConflict) Error() string { return "LDAP username conflicts with a local account" }

type SyncReport struct {
	CreatedUsers         int      `json:"createdUsers"`
	UpdatedUsers         int      `json:"updatedUsers"`
	MatchedUsers         int      `json:"matchedUsers"`
	SyncedUsers          int      `json:"syncedUsers"`
	SkippedLocalCount    int      `json:"skippedLocalCount"`
	SkippedLocalUsers    []string `json:"skippedLocalUsers"`
	SkippedDisabledCount int      `json:"skippedDisabledCount"`
	DeactivatedUsers     int      `json:"deactivatedUsers"`
}

func sameEndpoint(left, right string) bool {
	a, err := url.Parse(left)
	if err != nil {
		return false
	}
	b, err := url.Parse(right)
	if err != nil {
		return false
	}
	port := func(u *url.URL) string {
		if u.Port() != "" {
			return u.Port()
		}
		if u.Scheme == "ldaps" {
			return "636"
		}
		return "389"
	}
	return a.Scheme == b.Scheme && strings.EqualFold(a.Hostname(), b.Hostname()) && port(a) == port(b)
}

func (d *Directory) Save(candidate *models.LDAPConfig, replacePassword bool, password string) error {
	return d.SaveConfirmed(candidate, replacePassword, password, "", 0)
}
