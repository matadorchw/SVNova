// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"svnova/models"
	"testing"
	"time"
)

func TestDeleteLocalUserCleanupAndGuards(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	admin := &models.User{ID: 1, Name: "admin", Admin: true, Active: true}
	local := &models.User{ID: 2, Name: "local", Active: true}
	ldap := &models.User{ID: 3, Name: "ldap", LDAPAccount: true, Active: true}
	_, err = db.Insert(admin, local, ldap, &models.Session{Token: "local", UserID: 2}, &models.Organization{ID: 1, Name: "org"}, &models.Team{ID: 1, OrgID: 1, Name: "Owners", IsOwnerTeam: true}, &models.Membership{OrgID: 1, UserID: 2}, &models.TeamMember{TeamID: 1, UserID: 2}, &models.Repository{ID: 1, OrgID: 1, Name: "repo"})
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	call := func(actor *models.User, id int64, name string) error {
		return a.deleteLocalUser(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(`{"Name":"`+name+`"}`)), actor, id)
	}
	for _, c := range []struct {
		actor *models.User
		id    int64
		name  string
	}{{local, 2, "local"}, {admin, 1, "admin"}, {admin, 3, "ldap"}, {admin, 2, "wrong"}, {admin, 2, "local"}, {&models.User{ID: 99, Admin: true, Active: true}, 2, "local"}} {
		if call(c.actor, c.id, c.name) == nil {
			t.Fatalf("guard bypassed: %+v", c)
		}
	}
	if count, _ := db.Count(new(models.Session)); count != 1 {
		t.Fatal("failed delete removed session")
	}
	if _, err = db.Insert(&models.Membership{OrgID: 1, UserID: 1}, &models.TeamMember{TeamID: 1, UserID: 1}); err != nil {
		t.Fatal(err)
	}
	// A mid-cleanup failure must roll back earlier deletions.
	if _, err = db.Exec("CREATE TRIGGER reject_user_delete BEFORE DELETE ON user BEGIN SELECT RAISE(ABORT,'fixture'); END"); err != nil {
		t.Fatal(err)
	}
	if call(admin, 2, "local") == nil {
		t.Fatal("injected failure ignored")
	}
	if count, _ := db.Where("user_id=2").Count(new(models.Session)); count != 1 {
		t.Fatal("partial cleanup committed")
	}
	for _, model := range []any{new(models.Membership), new(models.TeamMember)} {
		if n, err := db.Where("user_id=2").Count(model); err != nil || n != 1 {
			t.Fatal("rollback lost member relationship", err)
		}
	}
	if ok, err := db.ID(2).Exist(new(models.User)); err != nil || !ok {
		t.Fatal("rollback lost user", err)
	}
	if _, err = db.Exec("DROP TRIGGER reject_user_delete"); err != nil {
		t.Fatal(err)
	}
	if err = call(admin, 2, "local"); err != nil {
		t.Fatal(err)
	}
	for _, v := range []any{new(models.Session), new(models.TeamMember), new(models.Membership)} {
		if count, err := db.Where("user_id=2").Count(v); err != nil || count != 0 {
			t.Fatal("associated records remain", err)
		}
	}
	if ok, _ := db.ID(2).Exist(new(models.User)); ok {
		t.Fatal("user remains")
	}
	if count, _ := db.Count(new(models.Repository)); count != 1 {
		t.Fatal("repository removed")
	}
	if call(admin, 2, "local") == nil {
		t.Fatal("missing user accepted")
	}
}

func TestDeletedAdministratorCannotReuseOrganizationAuthority(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	actor := &models.User{ID: 1, Name: "stale", Active: true, Admin: true}
	if _, err = db.Insert(&models.Organization{ID: 1, Name: "org"}); err != nil {
		t.Fatal(err)
	}
	tx := db.NewSession()
	defer tx.Close()
	if err = tx.Begin(); err != nil {
		t.Fatal(err)
	}
	if err = checkOrganizationOwner(tx, actor, 1); err == nil {
		t.Fatal("deleted administrator retained authority")
	}
}

func TestDeleteUserHTTPAuthorizationAndCSRF(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Insert(&models.User{ID: 1, Name: "admin", Active: true, Admin: true}, &models.User{ID: 2, Name: "local", Active: true})
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte("token"))
	if _, err = db.Insert(&models.Session{Token: hex.EncodeToString(sum[:]), UserID: 1, CSRF: "csrf", Expires: time.Now().Add(time.Hour).Unix()}); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Config: Config{Origin: "http://example.com"}}
	for _, c := range []struct {
		csrf, origin string
		code         int
	}{{"", "http://example.com", 403}, {"csrf", "http://evil.example", 403}, {"csrf", "http://example.com", 200}} {
		r := httptest.NewRequest("POST", "http://example.com/api/v1/admin/users/2/delete", strings.NewReader(`{"Name":"local"}`))
		r.AddCookie(&http.Cookie{Name: "svnova_session", Value: "token"})
		r.Header.Set("X-CSRF-Token", c.csrf)
		r.Header.Set("Origin", c.origin)
		w := httptest.NewRecorder()
		a.Handler().ServeHTTP(w, r)
		if w.Code != c.code {
			t.Fatalf("unexpected HTTP response %d: %s", w.Code, w.Body.String())
		}
	}
}

func TestMembershipRechecksTargetAfterConcurrentDeletion(t *testing.T) {
	db, err := models.Open(filepath.Join(t.TempDir(), "race.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	actor := &models.User{ID: 1, Name: "admin", Admin: true, Active: true}
	if _, err = db.Insert(actor, &models.User{ID: 2, Name: "target", Active: true}, &models.Organization{ID: 1, Name: "org"}); err != nil {
		t.Fatal(err)
	}
	// Inject deletion at the writer boundary, after the request's initial target lookup.
	if _, err = db.Exec("CREATE TRIGGER delete_target BEFORE UPDATE ON organization BEGIN DELETE FROM user WHERE id=2; END"); err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db}
	err = a.organizations(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(`{"UserID":2}`)), actor, []string{"1", "members"})
	if err == nil {
		t.Fatal("stale membership request accepted deleted user")
	}
	if n, err := db.Where("user_id=2").Count(new(models.Membership)); err != nil || n != 0 {
		t.Fatal("orphan membership created", err)
	}
}

func TestStaleAdministratorRequestAfterDeletionOrRevocation(t *testing.T) {
	for _, change := range []string{"DELETE FROM user WHERE id=2", "UPDATE user SET admin=0 WHERE id=2", "UPDATE user SET active=0 WHERE id=2"} {
		t.Run(change, func(t *testing.T) {
			db, err := models.Open(filepath.Join(t.TempDir(), "race.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			actor := &models.User{ID: 2, Name: "stale", Admin: true, Active: true}
			if _, err = db.Insert(actor, &models.Organization{ID: 1, Name: "org"}, &models.User{ID: 3, Name: "target", Active: true}); err != nil {
				t.Fatal(err)
			}
			ready, resume := make(chan struct{}), make(chan struct{})
			done := make(chan error, 1)
			go func() {
				snapshot := *actor
				close(ready)
				<-resume
				a := &App{DB: db}
				done <- a.changeTeamMember(httptest.NewRecorder(), &snapshot, &models.Team{ID: 1, OrgID: 1}, 3, false, false)
			}()
			<-ready
			if _, err = db.Exec(change); err != nil {
				close(resume)
				<-done
				t.Fatal(err)
			}
			close(resume)
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("stale administrator request accepted")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("request did not finish")
			}
			if n, err := db.Count(new(models.Membership)); err != nil || n != 0 {
				t.Fatal("unauthorized membership created", err)
			}
		})
	}
}
