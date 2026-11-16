// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"svnova/models"
	"svnova/services/security"
	"time"
)

func (s *Service) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}
func (s *Service) Recover() error {
	tx, err := begin(s.DB)
	if err != nil {
		return err
	}
	defer tx.Close()
	now := s.clock().UnixMilli()
	if _, err = tx.Exec("UPDATE notification_attempt SET status='interrupted',finished=?,reason=? WHERE status='sending'", now, "服务中断，送达状态未知"); err != nil {
		return err
	}
	if _, err = tx.Exec("UPDATE notification_delivery SET status=CASE WHEN attempts>=4 THEN 'failed' ELSE 'retry' END,next_attempt=?+CASE attempts WHEN 1 THEN 60000 WHEN 2 THEN 300000 WHEN 3 THEN 900000 ELSE 0 END,finished=CASE WHEN attempts>=4 THEN ? ELSE 0 END,last_error=? WHERE status='sending'", now, now, "服务中断，可能已送达"); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Service) Run(ctx context.Context) {
	// A single sender spaces attempts by three seconds, including tests and retries.
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	nextPrune := time.Time{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Tick(ctx); err != nil && ctx.Err() == nil {
				log.Print("Notification delivery tick failed")
			}
			if s.clock().After(nextPrune) {
				if err := s.Prune(); err != nil {
					log.Print("Notification retention cleanup failed")
				}
				nextPrune = s.clock().Add(time.Hour)
			}
		}
	}
}
func (s *Service) Tick(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	tx, err := begin(s.DB)
	if err != nil {
		return err
	}
	defer tx.Close()
	c, err := config(tx)
	if err != nil {
		return err
	}
	now := s.clock().UnixMilli()
	var latest models.NotificationAttempt
	if _, err = tx.Desc("id").Get(&latest); err != nil {
		return err
	}
	if latest.Started > 0 && now-latest.Started < 3000 {
		return nil
	}
	var d models.NotificationDelivery
	ok, err := tx.In("status", []string{"pending", "retry"}).Where("next_attempt<=?", now).Asc("id").Get(&d)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	var payload Event
	if d.Generation != c.Generation || json.Unmarshal([]byte(d.Payload), &payload) != nil || !eligible(c, payload) {
		_, err = tx.ID(d.ID).Cols("status", "finished", "last_error").Update(&models.NotificationDelivery{Status: "cancelled", Finished: now, LastError: "配置已更改，投递已取消"})
		if err != nil {
			return err
		}
		return tx.Commit()
	}
	d.Status = "sending"
	d.Attempts++
	if _, err = tx.ID(d.ID).Cols("status", "attempts").Update(&d); err != nil {
		return err
	}
	attempt := models.NotificationAttempt{DeliveryID: d.ID, Number: d.Attempts, Started: now, Status: "sending"}
	if _, err = tx.Insert(&attempt); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	tx.Close()
	// Recheck cancellation immediately before starting I/O; an in-flight request cannot be recalled.
	var current models.NotificationDelivery
	if _, err = s.DB.ID(d.ID).Get(&current); err != nil {
		return err
	}
	if current.Status != "sending" {
		return s.finish(d, attempt, SendResult{Reason: "配置已更改，投递已取消"}, true)
	}
	raw, e1 := s.Vault.Open(c.Webhook)
	secret, e2 := s.Vault.Open(c.Secret)
	result := SendResult{Reason: "无法读取机器人配置"}
	var event Event
	if e1 == nil && e2 == nil && json.Unmarshal([]byte(d.Payload), &event) == nil {
		title, message := Format(event, d.ID, s.Origin)
		sender := s.sender
		if sender == nil {
			sender = DingTalk{}
		}
		result = sender.Send(ctx, raw, secret, title, message)
	}
	return s.finish(d, attempt, result, false)
}
func (s *Service) finish(d models.NotificationDelivery, a models.NotificationAttempt, result SendResult, cancelled bool) error {
	tx, err := begin(s.DB)
	if err != nil {
		return err
	}
	defer tx.Close()
	now := s.clock().UnixMilli()
	a.Finished = now
	a.Status = "success"
	a.Reason = result.Reason
	if result.Reason != "" {
		a.Status = "failed"
	}
	if cancelled {
		a.Status = "cancelled"
	}
	if _, err = tx.ID(a.ID).AllCols().Update(&a); err != nil {
		return err
	}
	d.LastError = result.Reason
	d.Status = "success"
	d.Finished = now
	if result.Reason != "" {
		d.Status = "failed"
		if result.Retry && d.Attempts < 4 {
			d.Status = "retry"
			d.Finished = 0
			d.NextAttempt = now + []int64{60000, 300000, 900000}[d.Attempts-1]
		}
	}
	if _, err = tx.ID(d.ID).Where("status=?", "sending").Cols("status", "last_error", "finished", "next_attempt").Update(&d); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Service) Prune() error {
	tx, err := begin(s.DB)
	if err != nil {
		return err
	}
	defer tx.Close()
	c, err := config(tx)
	if err != nil {
		return err
	}
	if c.RetentionDays == 0 {
		return nil
	}
	cutoff := s.clock().Add(-time.Duration(c.RetentionDays) * 24 * time.Hour).UnixMilli()
	if _, err = tx.Exec("DELETE FROM notification_attempt WHERE delivery_id IN (SELECT id FROM notification_delivery WHERE status IN ('success','failed','cancelled') AND finished>0 AND finished<?)", cutoff); err != nil {
		return err
	}
	if _, err = tx.In("status", []string{"success", "failed", "cancelled"}).Where("finished>0 AND finished<?", cutoff).Delete(new(models.NotificationDelivery)); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Service) Test() error {
	tx, err := begin(s.DB)
	if err != nil {
		return err
	}
	defer tx.Close()
	c, err := config(tx)
	if err != nil {
		return err
	}
	raw, e := s.Vault.Open(c.Webhook)
	secret, e2 := s.Vault.Open(c.Secret)
	if e != nil || e2 != nil || ValidateWebhook(raw) != nil || secret == "" {
		return errors.New("请先保存有效的机器人配置")
	}
	now := s.clock().UnixMilli()
	if err = Enqueue(tx, Event{Key: "test/" + security.Token(), Kind: "test", Object: "钉钉机器人", Trigger: "test", Success: true, Started: now, Finished: now}); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Service) Redeliver(id int64) error {
	tx, err := begin(s.DB)
	if err != nil {
		return err
	}
	defer tx.Close()
	var old models.NotificationDelivery
	ok, err := tx.ID(id).Get(&old)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("投递记录不存在")
	}
	if old.Status != "failed" {
		return errors.New("仅失败的投递记录可以重发")
	}
	c, err := config(tx)
	if err != nil {
		return err
	}
	var event Event
	if json.Unmarshal([]byte(old.Payload), &event) != nil || !eligible(c, event) {
		return errors.New("当前通知开关或排除规则不允许重发")
	}
	if c.Webhook == "" || c.Secret == "" {
		return errors.New("请先保存机器人配置")
	}
	old.ParentID = old.ID
	old.ID = 0
	old.EventKey = fmt.Sprintf("redelivery/%s", security.Token())
	old.Generation = c.Generation
	old.Status = "pending"
	old.Attempts = 0
	old.Finished = 0
	old.LastError = ""
	old.Created = s.clock().UnixMilli()
	old.NextAttempt = old.Created
	if _, err = tx.Insert(&old); err != nil {
		return err
	}
	return tx.Commit()
}
