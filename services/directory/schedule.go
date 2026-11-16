// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"context"
	"errors"

	"github.com/robfig/cron/v3"
	"log"
	"strings"
	"svnova/models"
	"time"
)

var ErrScheduleChanged = errors.New("LDAP scheduled synchronization configuration changed")
var beijing = time.FixedZone("Asia/Shanghai", 8*3600)

var scheduleParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

func validSchedule(value string) bool {
	if len(strings.Fields(value)) != 5 {
		return false
	}
	_, err := scheduleParser.Parse(value)
	return err == nil
}
func (d *Directory) Schedule() (models.LDAPSchedule, error) {
	value := models.LDAPSchedule{ID: 1, Time: "0 0 * * *"}
	var stored models.LDAPSchedule
	ok, err := d.DB.ID(1).Get(&stored)
	if ok {
		value = stored
	}

	return value, err
}
func (d *Directory) SaveSchedule(value string) error {
	if !validSchedule(value) {
		return errors.New("请输入五字段cron：分 时 日 月 周，例如 30 2 * * *（北京时间每天02:30）")
	}
	_, err := d.DB.Exec("INSERT INTO ldap_schedule(id,time,updated) VALUES(1,?,?) ON CONFLICT(id) DO UPDATE SET time=excluded.time,updated=excluded.updated", value, time.Now().Unix())
	return err
}
func nextSync(now time.Time, clock string) time.Time {
	schedule, err := scheduleParser.Parse(clock)
	if err != nil {
		return time.Time{}
	}
	return schedule.Next(now.In(beijing))
}
func (d *Directory) ScheduledSync(ctx context.Context, clock string) error {
	if err := d.mu.LockContext(ctx); err != nil {
		return err
	}
	defer d.mu.Unlock()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	schedule, err := d.Schedule()
	if err != nil {
		return err
	}
	if schedule.Time != clock {
		return ErrScheduleChanged
	}
	return d.syncLocked(true, 0, "")
}
func (d *Directory) RunScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	d.runScheduler(ctx, ticker.C)
}
func (d *Directory) runScheduler(ctx context.Context, ticks <-chan time.Time) {
	var enabled bool
	var clock string
	var next time.Time
	var cancel context.CancelFunc
	var pending bool
	done := make(chan error, 1)
	defer func() {
		if cancel != nil {
			cancel()
		}
		if pending {
			<-done
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-done:
			if cancel != nil {
				cancel()
			}
			pending = false
			cancel = nil
			if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, ErrScheduleChanged) {
				log.Print("Scheduled LDAP synchronization failed")
			}
		case now, ok := <-ticks:
			if !ok {
				return
			}
			s, err := d.Source()
			schedule, scheduleErr := d.Schedule()
			active := false
			value := ""
			if err == nil && scheduleErr == nil && s.SyncEnabled && validSchedule(schedule.Time) {
				active = true
				value = schedule.Time
			}
			if active != enabled || value != clock {
				if cancel != nil {
					cancel()
				}
				if pending {
					<-done
					pending = false
					cancel = nil
				}
				enabled = active
				clock = value
				next = time.Time{}
				if active {
					next = nextSync(now, value)
				}
			}
			if !enabled || pending || next.IsZero() || now.Before(next) {
				continue
			}
			jobCtx, stop := context.WithCancel(ctx)
			cancel = stop
			pending = true
			scheduledClock := clock
			go func() { done <- d.ScheduledSync(jobCtx, scheduledClock) }()
			next = nextSync(now, clock)
		}
	}
}
