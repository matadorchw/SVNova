// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"context"
	"errors"
	"github.com/go-ldap/ldap/v3"
	"path/filepath"
	"svnova/models"
	"testing"
	"time"
)

func scheduleFixture(t *testing.T) *Directory {
	t.Helper()
	db, err := models.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err = db.Insert(&models.LDAPConfig{ID: 1, Enabled: true, SyncEnabled: true, AllowDeactivateAll: true}); err != nil {
		t.Fatal(err)
	}
	return &Directory{DB: db, readUsers: func(*models.LDAPConfig) ([]*ldap.Entry, error) { return nil, nil }}
}
func TestLDAPScheduledSyncWaitsForLockAndCancels(t *testing.T) {
	d := scheduleFixture(t)
	d.mu.Lock()
	done := make(chan error, 1)
	go func() { done <- d.ScheduledSync(context.Background(), "0 0 * * *") }()
	select {
	case err := <-done:
		d.mu.Unlock()
		t.Fatal("busy scheduled sync did not wait", err)
	case <-time.After(25 * time.Millisecond):
	}
	d.mu.Unlock()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("scheduled sync did not resume")
	}
	if n, _ := d.DB.Where("trigger=?", "scheduled").Count(new(models.LDAPSyncRun)); n != 1 {
		t.Fatal("scheduled execution missing")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	go func() { done <- d.ScheduledSync(ctx, "0 0 * * *") }()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal("wrong cancellation", err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown blocked on lock")
	}
}
func TestLDAPScheduledSyncRevalidatesConfiguration(t *testing.T) {
	d := scheduleFixture(t)
	d.mu.Lock()
	done := make(chan error, 1)
	go func() { done <- d.ScheduledSync(context.Background(), "0 0 * * *") }()
	if err := d.SaveSchedule("30 2 * * *"); err != nil {
		d.mu.Unlock()
		t.Fatal(err)
	}
	d.mu.Unlock()
	if err := <-done; !errors.Is(err, ErrScheduleChanged) {
		t.Fatal("stale schedule executed", err)
	}
	d.mu.Lock()
	go func() { done <- d.ScheduledSync(context.Background(), "30 2 * * *") }()
	if _, err := d.DB.ID(1).Cols("enabled").Update(&models.LDAPConfig{Enabled: false}); err != nil {
		d.mu.Unlock()
		t.Fatal(err)
	}
	d.mu.Unlock()
	if err := <-done; err == nil {
		t.Fatal("disabled source executed")
	}
	if n, _ := d.DB.Count(new(models.LDAPSyncRun)); n != 0 {
		t.Fatal("cancelled scheduled execution created history")
	}
}
func TestLDAPScheduleValidationCalendarAndNoStartupCatchup(t *testing.T) {
	d := scheduleFixture(t)
	for _, value := range []string{"02:30", "0 0 0 * * *", "@daily", "24:00", "2:30", "02:60", "02:30:00", "", " 02:30"} {
		if err := d.SaveSchedule(value); err == nil {
			t.Fatal("invalid schedule accepted", value)
		}
	}
	for _, value := range []string{"*/15 * * * *", "0 2,8 * * 1-5", "0 0 1 * *"} {
		if !validSchedule(value) {
			t.Fatal("valid cron rejected", value)
		}
	}
	utc := time.Date(2026, 11, 13, 18, 0, 0, 0, time.UTC)
	if next := nextSync(utc, "30 2 * * *"); next.Hour() != 2 || next.Minute() != 30 || next.Day() != 14 || next.Sub(utc) != 30*time.Minute {
		t.Fatal("schedule not in Beijing time", next)
	}
	now := time.Date(2026, 11, 13, 23, 59, 58, 0, beijing)
	ticks := make(chan time.Time)
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	called := make(chan struct{}, 4)
	d.readUsers = func(*models.LDAPConfig) ([]*ldap.Entry, error) { called <- struct{}{}; return nil, nil }
	go func() { d.runScheduler(ctx, ticks); close(stopped) }()
	ticks <- now
	d.mu.Lock()
	ticks <- now.Add(2 * time.Second)
	ticks <- now.Add(3 * time.Second)
	select {
	case <-called:
		d.mu.Unlock()
		t.Fatal("scheduled execution bypassed lock")
	default:
	}
	d.mu.Unlock()
	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("due schedule skipped")
	}
	cancel()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop")
	}
	ctx, cancel = context.WithCancel(context.Background())
	stopped = make(chan struct{})
	go func() { d.runScheduler(ctx, ticks); close(stopped) }()
	ticks <- now.Add(10 * time.Second)
	select {
	case <-called:
		t.Fatal("restart backfilled missed schedule")
	case <-time.After(25 * time.Millisecond):
	}
	cancel()
	<-stopped
}

func TestLDAPScheduleWithoutNextOccurrenceDoesNotRun(t *testing.T) {
	d := scheduleFixture(t)
	if err := d.SaveSchedule("0 0 31 2 *"); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 14, 0, 0, 0, 0, beijing)
	if next := nextSync(now, "0 0 31 2 *"); !next.IsZero() {
		t.Fatal("impossible date has next occurrence", next)
	}
	called := make(chan struct{}, 4)
	d.readUsers = func(*models.LDAPConfig) ([]*ldap.Entry, error) { called <- struct{}{}; return nil, nil }
	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time)
	stopped := make(chan struct{})
	go func() { d.runScheduler(ctx, ticks); close(stopped) }()
	defer func() { cancel(); <-stopped }()
	for i := 0; i < 5; i++ {
		ticks <- now.Add(time.Duration(i) * time.Second)
	}
	select {
	case <-called:
		t.Fatal("impossible cron triggered synchronization")
	case <-time.After(25 * time.Millisecond):
	}
	// A subsequent valid configuration must reactivate the scheduler.
	if err := d.SaveSchedule("* * * * *"); err != nil {
		t.Fatal(err)
	}
	ticks <- now.Add(10 * time.Second)
	select {
	case <-called:
		t.Fatal("impossible cron triggered synchronization")
	default:
	}
	ticks <- now.Add(time.Minute)
	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("valid cron did not resume")
	}
}
