// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"svnova/models"
	"svnova/services/security"
	"svnova/services/svn"
	"svnova/services/tasks"
)

func TestMirrorPasswordCanBeSavedWithoutRemoteConnection(t *testing.T) {
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	vault, err := security.NewVault(filepath.Join(root, "key"))
	if err != nil {
		t.Fatal(err)
	}
	repo := &models.Repository{OrgID: 1, Name: "mirror", Kind: "mirror", Status: "failed", Remote: "https://offline.invalid/repo", Password: vault.Seal("old-password"), Interval: 28800}
	if _, err = db.Insert(repo); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Vault: vault, Worker: &tasks.Worker{DB: db, Vault: vault}, SVN: svn.Runner{Bin: filepath.Join(root, "nonexistent-tools")}}
	user := &models.User{ID: 1, Admin: true, Active: true}
	request := httptest.NewRequest("POST", "/", strings.NewReader(`{"Remote":"https://offline.invalid/repo","Username":"remote-user","ReplacePassword":true,"Password":"replacement-password","SkipTLS":false,"Interval":28800}`))
	response := httptest.NewRecorder()
	if err = a.settings(response, request, user, repo); err != nil {
		t.Fatalf("offline credential save failed: %v", err)
	}
	var saved models.Repository
	if _, err = db.ID(repo.ID).Get(&saved); err != nil {
		t.Fatal(err)
	}
	password, err := vault.Open(saved.Password)
	if err != nil || password != "replacement-password" {
		t.Fatal("replacement password not persisted")
	}
	if strings.Contains(response.Body.String(), "replacement-password") || strings.Contains(response.Body.String(), saved.Password) {
		t.Fatal("password leaked in response")
	}
	request = httptest.NewRequest("POST", "/", strings.NewReader(`{"Remote":"https://different.invalid/repo","Username":"remote-user","ReplacePassword":true,"Password":"must-not-be-saved","Interval":28800}`))
	if err = a.settings(httptest.NewRecorder(), request, user, &saved); err == nil {
		t.Fatal("changed address bypassed online identity verification")
	}
	var retained models.Repository
	_, _ = db.ID(repo.ID).Get(&retained)
	if retained.Remote != repo.Remote || retained.Password != saved.Password {
		t.Fatal("failed address verification changed saved configuration")
	}
}
