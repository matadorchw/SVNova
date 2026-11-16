// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package notifications

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"svnova/models"
	"svnova/services/security"
	"testing"
	"time"
)

func fixture(t *testing.T) (*Service, *time.Time) {
	t.Helper()
	root := t.TempDir()
	db, err := models.Open(filepath.Join(root, "db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	v, err := security.NewVault(filepath.Join(root, "key"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().Add(time.Minute)
	return &Service{DB: db, Vault: v, Origin: "https://svnova.example", now: func() time.Time { return now }}, &now
}
func configured(t *testing.T, s *Service) Settings {
	t.Helper()
	in := Settings{Enabled: true, SVNEnabled: true, LDAPEnabled: true, RetentionDays: 90, ReplaceWebhook: true, ReplaceSecret: true, Webhook: "https://oapi.dingtalk.com/robot/send?access_token=private-token", Secret: "private-secret"}
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	in.ReplaceWebhook = false
	in.ReplaceSecret = false
	in.Webhook = ""
	in.Secret = ""
	return in
}
func queue(t *testing.T, s *Service, e Event) {
	t.Helper()
	tx := s.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		t.Fatal(err)
	}
	if err := Enqueue(tx, e); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}
func delivery(t *testing.T, s *Service, key string) models.NotificationDelivery {
	t.Helper()
	var d models.NotificationDelivery
	ok, err := s.DB.Where("event_key=?", key).Get(&d)
	if err != nil || !ok {
		t.Fatal("delivery missing", err)
	}
	return d
}

type fakeSender struct {
	calls  int
	result SendResult
	during func()
}

func (f *fakeSender) Send(context.Context, string, string, string, string) SendResult {
	f.calls++
	if f.during != nil {
		f.during()
	}
	return f.result
}
func TestFilteringEncryptionAndAtomicOutbox(t *testing.T) {
	s, now := fixture(t)
	c, err := s.Config()
	if err != nil || c.Enabled || !c.SVNEnabled || !c.LDAPEnabled || c.SVNNoChange || c.RetentionDays != 90 {
		t.Fatal("incorrect defaults", err)
	}
	e := Event{Key: "svn/1", Kind: "svn", Success: true, Changed: true, Finished: now.UnixMilli()}
	queue(t, s, e)
	if n, _ := s.DB.Count(new(models.NotificationDelivery)); n != 0 {
		t.Fatal("disabled notification enqueued")
	}
	in := configured(t, s)
	c, _ = s.Config()
	if strings.Contains(c.Webhook, "private-token") || strings.Contains(c.Secret, "private-secret") {
		t.Fatal("plaintext persisted")
	}
	b, _ := json.Marshal(c)
	if strings.Contains(string(b), c.Webhook) || strings.Contains(string(b), c.Secret) {
		t.Fatal("credentials in JSON")
	}
	e.Changed = false
	queue(t, s, e)
	if n, _ := s.DB.Count(new(models.NotificationDelivery)); n != 0 {
		t.Fatal("unchanged success enqueued")
	}
	e.Success = false
	queue(t, s, e)
	queue(t, s, e)
	if n, _ := s.DB.Count(new(models.NotificationDelivery)); n != 1 {
		t.Fatal("failure lost or duplicate enqueued")
	}
	e.Key = "ldap/rollback"
	e.Kind = "ldap"
	tx := s.DB.NewSession()
	if err = tx.Begin(); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Insert(&models.User{Name: "rolled-back"}); err != nil {
		t.Fatal(err)
	}
	if err = Enqueue(tx, e); err != nil {
		t.Fatal(err)
	}
	tx.Close()
	if n, _ := s.DB.Count(new(models.User)); n != 0 {
		t.Fatal("user rollback failed")
	}
	if n, _ := s.DB.Where("event_key=?", e.Key).Count(new(models.NotificationDelivery)); n != 0 {
		t.Fatal("outbox survived rollback")
	}
	if _, err = s.DB.Insert(&models.Repository{ID: 1, Kind: "mirror", Name: "included"}); err != nil {
		t.Fatal(err)
	}
	e.Kind = "svn"
	e.Key = "svn/all-repositories"
	e.RepositoryID = 1
	queue(t, s, e)
	if n, _ := s.DB.Where("event_key=?", e.Key).Count(new(models.NotificationDelivery)); n != 1 {
		t.Fatal("repository notification not enqueued")
	}
	in.LDAPEnabled = false
	if err = s.Save(in); err != nil {
		t.Fatal(err)
	}
	e.Kind = "ldap"
	e.Key = "disabled-ldap"
	queue(t, s, e)
	if n, _ := s.DB.Where("event_key=?", e.Key).Count(new(models.NotificationDelivery)); n != 0 {
		t.Fatal("disabled category enqueued")
	}
}
func TestRetryScheduleLimitAndRedelivery(t *testing.T) {
	s, now := fixture(t)
	configured(t, s)
	sender := &fakeSender{result: SendResult{Retry: true, Reason: "临时故障"}}
	s.sender = sender
	queue(t, s, Event{Key: "ldap/1", Kind: "ldap", Finished: now.UnixMilli()})
	for i, delay := range []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute, 0} {
		if err := s.Tick(context.Background()); err != nil {
			t.Fatal(err)
		}
		d := delivery(t, s, "ldap/1")
		if d.Attempts != i+1 {
			t.Fatal("wrong attempt number")
		}
		if delay > 0 {
			if d.Status != "retry" || d.NextAttempt != now.Add(delay).UnixMilli() {
				t.Fatalf("incorrect retry %+v", d)
			}
			*now = now.Add(delay - time.Millisecond)
			if err := s.Tick(context.Background()); err != nil {
				t.Fatal(err)
			}
			if sender.calls != i+1 {
				t.Fatal("sent before due")
			}
			*now = now.Add(time.Millisecond)
		} else if d.Status != "failed" || d.Finished == 0 {
			t.Fatal("retry limit ignored")
		}
	}
	if sender.calls != 4 {
		t.Fatal("wrong attempt limit")
	}
	old := delivery(t, s, "ldap/1")
	for _, status := range []string{"success", "cancelled", "pending", "retry", "sending"} {
		if _, err := s.DB.ID(old.ID).Cols("status").Update(&models.NotificationDelivery{Status: status}); err != nil {
			t.Fatal(err)
		}
		if err := s.Redeliver(old.ID); err == nil {
			t.Fatal("nonfailed redelivery accepted", status)
		}
	}
	if _, err := s.DB.ID(old.ID).Cols("status").Update(&models.NotificationDelivery{Status: "failed"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Redeliver(old.ID); err != nil {
		t.Fatal(err)
	}
	var repeat models.NotificationDelivery
	_, _ = s.DB.Desc("id").Get(&repeat)
	if repeat.ID == old.ID || repeat.ParentID != old.ID || repeat.Attempts != 0 || repeat.Generation != old.Generation {
		t.Fatal("invalid redelivery")
	}
	sender.result = SendResult{}
	*now = now.Add(3 * time.Second)
	if err := s.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	repeat = delivery(t, s, repeat.EventKey)
	if repeat.Status != "success" {
		t.Fatal("redelivery failed")
	}
}
func TestCancellationRecoveryRateAndRetention(t *testing.T) {
	s, now := fixture(t)
	in := configured(t, s)
	sender := &fakeSender{}
	s.sender = sender
	queue(t, s, Event{Key: "old", Kind: "ldap", Finished: now.UnixMilli()})
	in.Enabled = false
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	if d := delivery(t, s, "old"); d.Status != "cancelled" {
		t.Fatal("disable did not cancel")
	}
	in.Enabled = true
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	if err := s.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if sender.calls != 0 {
		t.Fatal("cancelled message replayed")
	}
	queue(t, s, Event{Key: "target", Kind: "ldap", Finished: now.UnixMilli()})
	in.ReplaceWebhook = true
	in.Webhook = "https://oapi.dingtalk.com/robot/send?access_token=new-target"
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	if d := delivery(t, s, "target"); d.Status != "cancelled" {
		t.Fatal("old target not cancelled")
	}
	queue(t, s, Event{Key: "one", Kind: "ldap", Finished: now.UnixMilli()})
	queue(t, s, Event{Key: "two", Kind: "ldap", Finished: now.UnixMilli()})
	if err := s.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if sender.calls != 1 {
		t.Fatal("rate limit bypassed")
	}
	d := delivery(t, s, "two")
	_, _ = s.DB.ID(d.ID).Cols("status", "attempts").Update(&models.NotificationDelivery{Status: "sending", Attempts: 1})
	_, _ = s.DB.Insert(&models.NotificationAttempt{DeliveryID: d.ID, Status: "sending", Number: 1, Started: now.UnixMilli()})
	if err := s.Recover(); err != nil {
		t.Fatal(err)
	}
	d = delivery(t, s, "two")
	if d.Status != "retry" || d.LastError == "" {
		t.Fatal("restart did not recover uncertain send")
	}
	if err := s.Recover(); err != nil {
		t.Fatal(err)
	}
	if n, _ := s.DB.Count(new(models.NotificationDelivery)); n != 4 {
		t.Fatal("recovery duplicated messages")
	}
	*now = now.Add(91 * 24 * time.Hour)
	if err := s.Prune(); err != nil {
		t.Fatal(err)
	}
	if n, _ := s.DB.Count(new(models.NotificationDelivery)); n != 1 {
		t.Fatal("retention did not preserve pending retry only")
	}
	if n, _ := s.DB.Where("delivery_id=?", d.ID).Count(new(models.NotificationAttempt)); n != 1 {
		t.Fatal("pending attempt history deleted")
	}
}
func TestInFlightCancellationDoesNotOverwriteState(t *testing.T) {
	s, now := fixture(t)
	in := configured(t, s)
	queue(t, s, Event{Key: "flight", Kind: "ldap", Finished: now.UnixMilli()})
	s.sender = &fakeSender{during: func() {
		in.Enabled = false
		if err := s.Save(in); err != nil {
			t.Fatal(err)
		}
	}}
	if err := s.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if d := delivery(t, s, "flight"); d.Status != "cancelled" {
		t.Fatal("completed HTTP revived cancelled delivery")
	}
}

func TestCategoryNoChangeAndPermanentFailure(t *testing.T) {
	s, now := fixture(t)
	in := configured(t, s)
	if _, err := s.DB.Insert(&models.Repository{ID: 1, Name: "mirror", Kind: "mirror"}); err != nil {
		t.Fatal(err)
	}
	in.SVNNoChange = true
	in.LDAPNoChange = true
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	queue(t, s, Event{Key: "svn/nochange", Kind: "svn", RepositoryID: 1, Success: true, Finished: now.UnixMilli()})
	queue(t, s, Event{Key: "ldap/nochange", Kind: "ldap", Success: true, Finished: now.UnixMilli()})
	in.LDAPEnabled = false
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	if delivery(t, s, "ldap/nochange").Status != "cancelled" || delivery(t, s, "svn/nochange").Status != "pending" {
		t.Fatal("category cancellation crossed boundary")
	}
	in.SVNEnabled = false
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	if delivery(t, s, "svn/nochange").Status != "cancelled" {
		t.Fatal("disabled SVN category did not cancel pending message")
	}
	if err := s.Test(); err != nil {
		t.Fatal(err)
	}
	sender := &fakeSender{result: SendResult{Reason: "签名错误"}}
	s.sender = sender
	if err := s.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	var latest models.NotificationDelivery
	_, _ = s.DB.Desc("id").Get(&latest)
	if latest.Status != "failed" || latest.Attempts != 1 {
		t.Fatal("permanent error retried")
	}
	in.RetentionDays = 0
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	*now = now.Add(365 * 24 * time.Hour)
	if err := s.Prune(); err != nil {
		t.Fatal(err)
	}
	if n, _ := s.DB.Count(new(models.NotificationDelivery)); n != 3 {
		t.Fatal("permanent retention deleted history")
	}
}

func TestSVNFailureOnlyOverridesSuccessAndCancelsQueue(t *testing.T) {
	s, now := fixture(t)
	in := configured(t, s)
	in.SVNNoChange = true
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	queue(t, s, Event{Key: "success-before", Kind: "svn", Success: true, Changed: true, Finished: now.UnixMilli()})
	in.SVNFailureOnly = true
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	d := delivery(t, s, "success-before")
	if d.Status != "cancelled" {
		t.Fatal("queued success not cancelled")
	}
	if err := s.Redeliver(d.ID); err == nil {
		t.Fatal("success redelivered in failure-only mode")
	}
	for _, event := range []Event{{Key: "changed", Kind: "svn", Success: true, Changed: true}, {Key: "unchanged", Kind: "svn", Success: true}, {Key: "failed", Kind: "svn"}, {Key: "ldap", Kind: "ldap", Success: true, Changed: true}} {
		event.Finished = now.UnixMilli()
		queue(t, s, event)
	}
	if n, _ := s.DB.Count(new(models.NotificationDelivery)); n != 3 {
		t.Fatal("failure-only filtering affected LDAP or allowed SVN success")
	}
	if delivery(t, s, "failed").Status != "pending" || delivery(t, s, "ldap").Status != "pending" {
		t.Fatal("required notifications missing")
	}
}

func TestDisablingCancelsExistingTestsButAllowsNewExplicitTest(t *testing.T) {
	s, _ := fixture(t)
	in := configured(t, s)
	if err := s.Test(); err != nil {
		t.Fatal(err)
	}
	in.Enabled = false
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	var old models.NotificationDelivery
	_, _ = s.DB.Desc("id").Get(&old)
	if old.Status != "cancelled" {
		t.Fatal("disabling did not cancel queued test")
	}
	if err := s.Test(); err != nil {
		t.Fatal(err)
	}
	var next models.NotificationDelivery
	_, _ = s.DB.Desc("id").Get(&next)
	if next.ID == old.ID || next.Status != "pending" {
		t.Fatal("explicit test unavailable while disabled")
	}
}

func TestRestartRecoveryKeepsBackoffAndAttemptLimit(t *testing.T) {
	s, now := fixture(t)
	configured(t, s)
	for attempt, delay := range map[int]time.Duration{1: time.Minute, 2: 5 * time.Minute, 3: 15 * time.Minute, 4: 0} {
		key := fmt.Sprintf("restart/%d", attempt)
		queue(t, s, Event{Key: key, Kind: "ldap", Finished: now.UnixMilli()})
		d := delivery(t, s, key)
		if _, err := s.DB.ID(d.ID).Cols("status", "attempts").Update(&models.NotificationDelivery{Status: "sending", Attempts: attempt}); err != nil {
			t.Fatal(err)
		}
		if err := s.Recover(); err != nil {
			t.Fatal(err)
		}
		d = delivery(t, s, key)
		if attempt < 4 {
			if d.Status != "retry" || d.NextAttempt != now.Add(delay).UnixMilli() {
				t.Fatal("restart lost retry schedule")
			}
		} else if d.Status != "failed" || d.Finished == 0 {
			t.Fatal("restart exceeded attempt limit")
		}
	}
}

func TestOutboxAndCredentialSurviveDatabaseReopen(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "db")
	db, err := models.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	vault, err := security.NewVault(filepath.Join(root, "key"))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	now := time.Now().Add(time.Minute)
	s := &Service{DB: db, Vault: vault, now: func() time.Time { return now }}
	configured(t, s)
	queue(t, s, Event{Key: "persistent", Kind: "ldap", Finished: now.UnixMilli()})
	d := delivery(t, s, "persistent")
	if _, err = db.ID(d.ID).Cols("status", "attempts").Update(&models.NotificationDelivery{Status: "sending", Attempts: 3}); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if _, err = db.Insert(&models.NotificationAttempt{DeliveryID: d.ID, Number: 3, Started: now.UnixMilli(), Status: "sending"}); err != nil {
		db.Close()
		t.Fatal(err)
	}
	db.Close()
	db, err = models.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	vault, err = security.NewVault(filepath.Join(root, "key"))
	if err != nil {
		t.Fatal(err)
	}
	s = &Service{DB: db, Vault: vault, now: func() time.Time { return now }}
	c, err := s.Config()
	if err != nil {
		t.Fatal(err)
	}
	secret, err := vault.Open(c.Secret)
	if err != nil || secret != "private-secret" {
		t.Fatal("credential did not survive reopen", err)
	}
	if err = s.Recover(); err != nil {
		t.Fatal(err)
	}
	d = delivery(t, s, "persistent")
	if d.Status != "retry" || d.NextAttempt != now.Add(15*time.Minute).UnixMilli() {
		t.Fatal("persisted uncertain delivery not recovered")
	}
	sender := &fakeSender{}
	s.sender = sender
	now = now.Add(15 * time.Minute)
	if err = s.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	d = delivery(t, s, "persistent")
	if sender.calls != 1 || d.Status != "success" || d.Attempts != 4 {
		t.Fatal("persisted retry not delivered")
	}
	var first models.NotificationAttempt
	_, _ = db.Asc("id").Get(&first)
	if first.Status != "interrupted" {
		t.Fatal("uncertain attempt history lost")
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestDingTalkSigningErrorsAndRedaction(t *testing.T) {
	for _, raw := range []string{"http://oapi.dingtalk.com/robot/send?access_token=a", "https://oapi.dingtalk.com.evil/robot/send?access_token=a", "https://oapi.dingtalk.com:443/robot/send?access_token=a", "https://user@oapi.dingtalk.com/robot/send?access_token=a", "https://oapi.dingtalk.com/robot/send?access_token=a&sign=old"} {
		if ValidateWebhook(raw) == nil {
			t.Fatal("unsafe URL accepted")
		}
	}
	for _, tc := range []struct {
		status        int
		body          string
		retry, failed bool
	}{{200, `{"errcode":0}`, false, false}, {200, `{"errcode":130101,"errmsg":"private-secret"}`, true, true}, {200, `{"errcode":310000,"errmsg":"private-secret"}`, false, true}, {200, `{"errcode":12345,"errmsg":"频率限制 private-secret"}`, true, true}, {503, "oops", true, true}, {302, "", false, true}, {200, "invalid", true, true}, {200, strings.Repeat("x", 65537), true, true}} {
		d := DingTalk{transport: roundTrip(func(r *http.Request) (*http.Response, error) {
			q := r.URL.Query()
			mac := hmac.New(sha256.New, []byte("private-secret"))
			mac.Write([]byte(q.Get("timestamp") + "\nprivate-secret"))
			if q.Get("sign") != base64.StdEncoding.EncodeToString(mac.Sum(nil)) {
				t.Fatal("incorrect signature")
			}
			return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: http.Header{}}, nil
		})}
		res := d.Send(context.Background(), "https://oapi.dingtalk.com/robot/send?access_token=private-token", "private-secret", "title", "body")
		if res.Retry != tc.retry || (res.Reason != "") != tc.failed || strings.Contains(res.Reason, "private") {
			t.Fatalf("unsafe or wrong response %+v", res)
		}
	}
	d := DingTalk{transport: roundTrip(func(*http.Request) (*http.Response, error) { return nil, errors.New("private-token") })}
	res := d.Send(context.Background(), "https://oapi.dingtalk.com/robot/send?access_token=private-token", "private-secret", "t", "m")
	if !res.Retry || strings.Contains(res.Reason, "private") {
		t.Fatal("network error leaked")
	}
}
func TestFormatBlocksEscapingAndUnknownRevision(t *testing.T) {
	title, text := Format(Event{Kind: "svn", Site: "站点", Object: "repo[bad](url)", Trigger: "manual", Started: 1, Finished: 1001, TaskID: 2, ExecutionID: 3, Phase: "synchronizing", Link: "/org/repo", HistoryLink: "/org/repo/tasks"}, 4, "https://nova.example")
	if title != "SVN 仓库同步结果" || !strings.Contains(text, "  \n\u200b  \n") || !strings.Contains(text, "最终修订：未知") || !strings.Contains(text, "同步修订") || !strings.Contains(text, "1970-01-01 08:00:01") || !strings.Contains(text, `repo\[bad\]`) {
		t.Fatal("incorrect format", text)
	}
	_, text = Format(Event{Kind: "ldap", Success: true, Created: 1, Updated: 2, Disabled: 3, Skipped: 4, HistoryLink: "//evil.example"}, 5, "https://nova.example")
	if strings.Contains(text, "evil") || !strings.Contains(text, "实际更新 2") {
		t.Fatal("unsafe link or missing stats")
	}
}
