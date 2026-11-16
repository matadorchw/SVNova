// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"net/url"
	"svnova/services/directory"
)

func (p *webPage) LDAPSyncReport() directory.SyncReport {
	var report directory.SyncReport
	_ = json.Unmarshal([]byte(p.LDAP.LastSyncReport), &report)
	return report
}

func (p *webPage) LDAPHost() string {
	u, err := url.Parse(p.LDAP.URL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}
func (p *webPage) LDAPProtocol() string {
	u, err := url.Parse(p.LDAP.URL)
	if err == nil && u.Scheme == "ldap" {
		return "starttls"
	}
	return "ldaps"
}
func (p *webPage) LDAPPort() string {
	u, err := url.Parse(p.LDAP.URL)
	if err == nil && u.Port() != "" {
		return u.Port()
	}
	if p.LDAPProtocol() == "starttls" {
		return "389"
	}
	return "636"
}
