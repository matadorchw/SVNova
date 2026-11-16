// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package notifications

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"svnova/models"
	"svnova/services/security"
	"time"
	"xorm.io/xorm"
)

type Service struct {
	DB     *xorm.Engine
	Vault  *security.Vault
	Origin string
	sender Sender
	now    func() time.Time
}

func defaults() models.NotificationConfig {
	return models.NotificationConfig{ID: 1, SVNEnabled: true, LDAPEnabled: true, RetentionDays: 90, Generation: 1}
}
func config(tx *xorm.Session) (models.NotificationConfig, error) {
	c := defaults()
	var stored models.NotificationConfig
	ok, err := tx.ID(1).Get(&stored)
	if ok {
		c = stored
	}
	return c, err
}
func (s *Service) Config() (models.NotificationConfig, error) {
	tx := s.DB.NewSession()
	defer tx.Close()
	return config(tx)
}
func ValidateWebhook(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host != "oapi.dingtalk.com" || u.Path != "/robot/send" || u.RawPath != "" || u.User != nil || u.Fragment != "" || len(u.Query()["access_token"]) != 1 || strings.TrimSpace(u.Query().Get("access_token")) == "" {
		return errors.New("请输入官方 HTTPS 钉钉机器人 Webhook 地址")
	}
	for key := range u.Query() {
		if key != "access_token" {
			return errors.New("Webhook 地址仅保留 access_token 参数，签名由系统生成")
		}
	}
	if len(raw) > 4096 {
		return errors.New("Webhook 地址过长")
	}
	return nil
}

type Settings struct {
	Enabled        bool
	SVNEnabled     bool
	LDAPEnabled    bool
	SVNNoChange    bool
	SVNFailureOnly bool
	LDAPNoChange   bool
	RetentionDays  int
	Webhook        string
	Secret         string
	ReplaceWebhook bool
	ReplaceSecret  bool
}

func begin(db *xorm.Engine) (*xorm.Session, error) {
	tx := db.NewSession()
	if err := tx.Begin(); err != nil {
		tx.Close()
		return nil, err
	}
	if _, err := tx.Exec("UPDATE migration SET id=id WHERE id=10"); err != nil {
		tx.Close()
		return nil, err
	}
	return tx, nil
}
func (s *Service) Save(in Settings) error {
	if in.RetentionDays != 0 && in.RetentionDays != 30 && in.RetentionDays != 90 && in.RetentionDays != 180 {
		return errors.New("记录保留期限无效")
	}
	tx, err := begin(s.DB)
	if err != nil {
		return err
	}
	defer tx.Close()
	old, err := config(tx)
	if err != nil {
		return err
	}
	c := old
	c.Enabled = in.Enabled
	c.SVNEnabled = in.SVNEnabled
	c.LDAPEnabled = in.LDAPEnabled
	c.SVNNoChange = in.SVNNoChange
	c.SVNFailureOnly = in.SVNFailureOnly
	c.LDAPNoChange = in.LDAPNoChange
	c.RetentionDays = in.RetentionDays
	if in.ReplaceWebhook {
		if err = ValidateWebhook(in.Webhook); err != nil {
			return err
		}
		c.Webhook = s.Vault.Seal(in.Webhook)
	}
	if in.ReplaceSecret {
		if strings.TrimSpace(in.Secret) == "" || len(in.Secret) > 1024 {
			return errors.New("签名密钥不能为空且最多1024字节")
		}
		c.Secret = s.Vault.Seal(in.Secret)
	}
	if c.Enabled {
		raw, e := s.Vault.Open(c.Webhook)
		secret, e2 := s.Vault.Open(c.Secret)
		if e != nil || e2 != nil || secret == "" {
			return errors.New("请先配置 Webhook 和签名密钥")
		}
		if err = ValidateWebhook(raw); err != nil {
			return err
		}
	}
	if in.ReplaceWebhook || in.ReplaceSecret {
		c.Generation++
	}
	c.Updated = time.Now().UnixMilli()
	exists, err := tx.ID(1).Exist(new(models.NotificationConfig))
	if err != nil {
		return err
	}
	if exists {
		_, err = tx.ID(1).AllCols().Update(&c)
	} else {
		_, err = tx.Insert(&c)
	}
	if err != nil {
		return err
	}
	var pending []models.NotificationDelivery
	if err = tx.In("status", []string{"pending", "retry", "sending"}).Find(&pending); err != nil {
		return err
	}
	for _, d := range pending {
		var event Event
		if (old.Enabled && !c.Enabled) || d.Generation != c.Generation || json.Unmarshal([]byte(d.Payload), &event) != nil || !eligible(c, event) {
			if _, err = tx.ID(d.ID).Cols("status", "finished", "last_error").Update(&models.NotificationDelivery{Status: "cancelled", Finished: c.Updated, LastError: "配置已更改，投递已取消"}); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}
func allowed(c models.NotificationConfig, kind string) bool {
	if kind == "test" {
		return true
	}
	if !c.Enabled {
		return false
	}
	if kind == "ldap" {
		return c.LDAPEnabled
	}
	if kind != "svn" || !c.SVNEnabled {
		return false
	}
	return true
}
