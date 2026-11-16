// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"svnova/models"
	"svnova/services/security"
	"svnova/services/svn"
	"svnova/services/tasks"
)

func conversionFixture(t *testing.T) (*App, *models.Repository, *models.User, string) {
	t.Helper()
	bin := os.Getenv("SVN_BIN")
	if bin == "" {
		t.Skip("set SVN_BIN for real SVN conversion")
	}
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	vault, err := security.NewVault(filepath.Join(root, "key"))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{DB: db, Vault: vault, SVN: svn.Runner{Bin: bin}, Config: Config{Repositories: filepath.Join(root, "managed")}}
	actor := &models.User{Name: "owner", Active: true}
	org := &models.Organization{Name: "org"}
	if _, err = db.Insert(actor, org); err != nil {
		t.Fatal(err)
	}
	team := &models.Team{OrgID: org.ID, Name: "Owners", IsOwnerTeam: true}
	if _, err = db.Insert(team); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Insert(&models.TeamMember{TeamID: team.ID, UserID: actor.ID}, &models.Membership{OrgID: org.ID, UserID: actor.ID}); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "source")
	run := func(tool string, args ...string) {
		t.Helper()
		if _, err := a.SVN.Run(context.Background(), tool, nil, args...); err != nil {
			t.Fatal(err)
		}
	}
	run("svnadmin", "create", source)
	fixture := filepath.Join(root, "fixture")
	if err = os.Mkdir(fixture, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(fixture, "data.txt"), []byte("source history"), 0600); err != nil {
		t.Fatal(err)
	}
	run("svn", "import", fixture, svn.LocalURL(source), "-m", "fixture", "--non-interactive")
	if err = os.MkdirAll(a.Config.Repositories, 0700); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(a.Config.Repositories, "repo")
	run("svnadmin", "hotcopy", source, dest)
	uuid, revision, err := a.SVN.LocalInfo(context.Background(), dest)
	if err != nil {
		t.Fatal(err)
	}
	repo := &models.Repository{OrgID: org.ID, Name: "repo", Path: dest, Kind: "local", Status: "ready", UUID: uuid, Revision: revision}
	if _, err = db.Insert(repo); err != nil {
		t.Fatal(err)
	}
	a.Worker = &tasks.Worker{DB: db, Root: a.Config.Repositories, SVN: a.SVN, Vault: vault, Convert: a.executeConversion}
	return a, repo, actor, source
}

func queueFixtureConversion(t *testing.T, a *App, r *models.Repository, u *models.User, remote string) *models.Task {
	t.Helper()
	// File sources are confined to this internal integration fixture; the public API rejects them.
	data, err := json.Marshal(conversionInput{Remote: remote, Password: a.Vault.Seal("fixture-password"), FullAccess: true})
	if err != nil {
		t.Fatal(err)
	}
	task := &models.Task{RepositoryID: r.ID, Kind: "conversion", State: "queued", ActorID: u.ID, ActorName: u.Name, Source: string(data), Started: 1}
	if _, err = a.DB.Insert(task); err != nil {
		t.Fatal(err)
	}
	return task
}

func TestConversionBackgroundQueueAndEncryptedCredentials(t *testing.T) {
	a, r, u, _ := conversionFixture(t)
	request := httptest.NewRequest("POST", "/", strings.NewReader(`{"Remote":"https://example.com/repo","Password":"secret-value","FullAccess":true}`))
	ctx, cancel := context.WithCancel(request.Context())
	request = request.WithContext(ctx)
	response := httptest.NewRecorder()
	if err := a.convertMirror(response, request, u, r); err != nil {
		t.Fatal(err)
	}
	cancel()
	var task models.Task
	if ok, err := a.DB.Where("kind=?", "conversion").Get(&task); err != nil || !ok {
		t.Fatal(err)
	}
	if task.State != "queued" || task.ActorID != u.ID || strings.Contains(task.Source, "secret-value") || !strings.Contains(response.Body.String(), "taskURL") {
		t.Fatalf("unsafe queue: %+v", task)
	}
	var input conversionInput
	if err := json.Unmarshal([]byte(task.Source), &input); err != nil {
		t.Fatal(err)
	}
	if clear, err := a.Vault.Open(input.Password); err != nil || clear != "secret-value" {
		t.Fatal("credential encryption failed", err)
	}
	if err := a.convertMirror(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(`{"Remote":"https://example.com/repo","FullAccess":true}`)), u, r); err == nil {
		t.Fatal("duplicate accepted")
	}
	if _, err := a.DB.ID(u.ID).Cols("active").Update(&models.User{Active: false}); err != nil {
		t.Fatal(err)
	}
	a.Worker.Tick(context.Background())
	taskID := task.ID
	task = models.Task{}
	if _, err := a.DB.ID(taskID).Get(&task); err != nil {
		t.Fatal(err)
	}
	if task.State != "failed" {
		t.Fatal("revoked actor executed conversion")
	}
}

func TestQuickConversionExecutionAndRollback(t *testing.T) {
	for _, scenario := range []string{"success", "divergent", "permission", "halted", "cancel", "timeout", "initialize", "recovery"} {
		t.Run(scenario, func(t *testing.T) {
			a, r, u, source := conversionFixture(t)
			if scenario == "divergent" {
				logFile := filepath.Join(filepath.Dir(source), "log.txt")
				if err := os.WriteFile(logFile, []byte("independently changed revision property"), 0600); err != nil {
					t.Fatal(err)
				}
				if _, err := a.SVN.Run(context.Background(), "svnadmin", nil, "setrevprop", r.Path, "-r", "1", "svn:log", logFile); err != nil {
					t.Fatal(err)
				}
			}
			task := queueFixtureConversion(t, a, r, u, svn.LocalURL(source))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			dumps := 0
			initialized := false
			a.SVN.OnCommand = func(tool string) {
				if tool == "svnrdump" {
					dumps++
				}
				if tool != "svnsync" || initialized {
					return
				}
				initialized = true
				switch scenario {
				case "permission":
					if _, err := a.DB.ID(u.ID).Cols("active").Update(&models.User{Active: false}); err != nil {
						t.Fatal(err)
					}
				case "halted":
					if _, err := a.DB.ID(r.ID).Cols("halted").Update(&models.Repository{Halted: true}); err != nil {
						t.Fatal(err)
					}
				case "cancel":
					cancel()
				case "timeout":
					time.Sleep(2100 * time.Millisecond)
				case "recovery":
					hook := "pre-revprop-change"
					if runtime.GOOS == "windows" {
						hook += ".bat"
					}
					if err := os.WriteFile(filepath.Join(r.Path, "hooks", hook), []byte("changed hook requiring manual recovery"), 0700); err != nil {
						t.Fatal(err)
					}
				case "initialize":
					if err := os.Rename(source, source+"-offline"); err != nil {
						t.Fatal(err)
					}
				}
			}
			if scenario == "success" {
				a.SVN.Timeout = time.Nanosecond
			}
			if scenario == "timeout" {
				a.Config.MirrorConversionTimeout = 2 * time.Second
			}
			a.Worker.Tick(ctx)
			taskID, repoID := task.ID, r.ID
			task = &models.Task{}
			r = &models.Repository{}
			if _, err := a.DB.ID(taskID).Get(task); err != nil {
				t.Fatal(err)
			}
			if _, err := a.DB.ID(repoID).Get(r); err != nil {
				t.Fatal(err)
			}
			if dumps != 0 {
				t.Fatal("quick conversion executed revision dumps")
			}
			if task.Started <= 1 || task.Finished == 0 {
				t.Fatal("attempt timestamps not persisted")
			}
			if scenario == "success" || scenario == "divergent" {
				if task.State != "succeeded" || r.Kind != "mirror" || r.LastSuccess != 0 || task.Source != "" {
					t.Fatalf("conversion failed: %+v %+v", task, r)
				}
				var next models.Task
				if ok, err := a.DB.Where("kind=? AND repository_id=?", "sync", r.ID).Get(&next); err != nil || !ok || next.State != "queued" {
					t.Fatal("initial sync missing", err)
				}
				if !strings.Contains(taskResult(*task), "未逐修订") {
					t.Fatal("missing risk result")
				}
			} else {
				if task.State != "failed" || r.Kind != "local" {
					t.Fatalf("failed conversion committed: %+v %+v", task, r)
				}
				if scenario == "halted" && !r.Halted {
					t.Fatal("worker overwrote concurrent state change")
				}
				if scenario != "recovery" {
					if _, err := a.SVN.SnapshotConversion(context.Background(), r.Path); err != nil {
						t.Fatal("rollback failed", err)
					}
				}
			}
			if pending, err := a.DB.Exist(new(models.MirrorConversion)); err != nil || pending != (scenario == "recovery") {
				t.Fatal("unexpected recovery journal state", err)
			}
			if scenario == "recovery" && a.Worker.TryLock() {
				a.Worker.Unlock()
				t.Fatal("unrecovered conversion allowed mutation")
			}
			if _, err := os.Stat(filepath.Join(r.Path, "db")); err != nil {
				t.Fatal("local repository lost", err)
			}
		})
	}
}

func TestConversionProgressAndDefaultTimeout(t *testing.T) {
	if conversionTimeout(0) != 24*time.Hour {
		t.Fatal("wrong timeout default")
	}
	task := models.Task{Kind: "conversion", State: "running", Phase: "conversion-initialize", Started: time.Now().Unix() - 5, ProgressStarted: time.Now().Unix() - 2}
	if phase := taskPhase(task); !strings.Contains(phase, "初始化镜像") || !strings.Contains(phase, "已用时") {
		t.Fatal(phase)
	}
}

func TestQuickConversionRejectsUnsafeSources(t *testing.T) {
	for _, scenario := range []string{"uuid", "revision", "root", "r0", "hook", "metadata", "notready"} {
		t.Run(scenario, func(t *testing.T) {
			a, r, u, source := conversionFixture(t)
			remote := svn.LocalURL(source)
			run := func(tool string, args ...string) {
				t.Helper()
				if _, err := a.SVN.Run(context.Background(), tool, nil, args...); err != nil {
					t.Fatal(err)
				}
			}
			property := filepath.Join(filepath.Dir(source), "property.txt")
			if err := os.WriteFile(property, []byte("fixture"), 0600); err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "uuid":
				run("svnadmin", "setuuid", source)
			case "revision":
				older := source + "-older"
				run("svnadmin", "create", older)
				run("svnadmin", "setuuid", older, r.UUID)
				remote = svn.LocalURL(older)
			case "root":
				remote += "/data.txt"
			case "r0":
				run("svnadmin", "setrevprop", source, "-r", "0", "fixture:r0", property)
			case "metadata":
				run("svnadmin", "setrevprop", r.Path, "-r", "0", "svn:sync-from-uuid", property)
			case "hook":
				if err := os.WriteFile(filepath.Join(r.Path, "hooks", "pre-commit.bat"), []byte("exit 1"), 0700); err != nil {
					t.Fatal(err)
				}
			case "notready":
				if _, err := a.DB.ID(r.ID).Cols("status").Update(&models.Repository{Status: "failed"}); err != nil {
					t.Fatal(err)
				}
			}
			task := queueFixtureConversion(t, a, r, u, remote)
			id := task.ID
			initialized := false
			a.SVN.OnCommand = func(tool string) {
				if tool == "svnsync" {
					initialized = true
				}
			}
			a.Worker.Tick(context.Background())
			task = &models.Task{}
			if _, err := a.DB.ID(id).Get(task); err != nil {
				t.Fatal(err)
			}
			if task.State != "failed" || initialized {
				t.Fatalf("unsafe source accepted: %s %s", scenario, task.Error)
			}
			if pending, err := a.DB.Exist(new(models.MirrorConversion)); err != nil || pending {
				t.Fatal("precheck left recovery journal", err)
			}
		})
	}
}

func TestConversionOwnerRetry(t *testing.T) {
	a, r, u, source := conversionFixture(t)
	task := queueFixtureConversion(t, a, r, u, svn.LocalURL(source))
	if _, err := a.DB.ID(task.ID).Cols("state", "phase", "error", "finished").Update(&models.Task{State: "failed", Phase: "conversion-recovery", Error: "old failure", Finished: 1}); err != nil {
		t.Fatal(err)
	}
	if err := a.repositories(httptest.NewRecorder(), httptest.NewRequest("POST", "/", nil), u, []string{"1", "retry"}); err != nil {
		t.Fatal("Owners cannot retry conversion", err)
	}
	a.Worker.Tick(context.Background())
	id := task.ID
	task = &models.Task{}
	if _, err := a.DB.ID(id).Get(task); err != nil {
		t.Fatal(err)
	}
	if task.State != "succeeded" || task.Started <= 1 || task.Error != "" {
		t.Fatalf("retry failed: %+v", task)
	}
}
