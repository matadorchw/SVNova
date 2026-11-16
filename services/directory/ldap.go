// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"svnova/models"
	"svnova/services/notifications"
	"svnova/services/security"
	"xorm.io/xorm"
)

type Directory struct {
	DB        *xorm.Engine
	Vault     *security.Vault
	mu        directoryMutex
	readUsers func(*models.LDAPConfig) ([]*ldap.Entry, error)
}

func (d *Directory) Lock()   { d.mu.Lock() }
func (d *Directory) Unlock() { d.mu.Unlock() }

func Validate(s *models.LDAPConfig) error {
	u, e := url.Parse(s.URL)
	if e != nil || u.Hostname() == "" || (u.Scheme != "ldap" && u.Scheme != "ldaps") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return errors.New("invalid LDAP URL")
	}
	if u.Port() != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil || port < 1 || port > 65535 {
			return errors.New("invalid LDAP port")
		}
	}
	if s.BaseDN == "" || s.Filter == "" || s.UsernameAttribute == "" || s.KeyAttribute == "" {
		return errors.New("missing LDAP settings")
	}
	if s.UsePagedSearch && (s.SearchPageSize == 0 || s.SearchPageSize > 1000) {
		return errors.New("LDAP page size must be between 1 and 1000")
	}
	attribute := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9;-]*$`)
	for _, name := range []string{s.UsernameAttribute, s.KeyAttribute, s.NameAttribute, s.EmailAttribute} {
		if name != "" && !attribute.MatchString(name) {
			return errors.New("invalid LDAP attribute mapping")
		}
	}
	_, e = ldap.CompileFilter(filter(s))
	return e
}
func (d *Directory) connect(s *models.LDAPConfig) (*ldap.Conn, error) {
	if err := Validate(s); err != nil {
		return nil, err
	}
	u, _ := url.Parse(s.URL)
	tlsConfig := &tls.Config{ServerName: u.Hostname(), MinVersion: tls.VersionTLS12, InsecureSkipVerify: s.SkipVerify} // Explicit administrator option.
	conn, err := ldap.DialURL(s.URL, ldap.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}), ldap.DialWithTLSConfig(tlsConfig))
	if err != nil {
		return nil, errors.New("LDAP connection failed")
	}
	conn.SetTimeout(15 * time.Second)
	if u.Scheme == "ldap" {
		if err = conn.StartTLS(tlsConfig); err != nil {
			conn.Close()
			return nil, errors.New("LDAP StartTLS failed")
		}
	}
	p, err := d.Vault.Open(s.Password)
	if err == nil && s.BindDN != "" {
		if p == "" {
			err = errors.New("LDAP bind password required")
		} else {
			err = conn.Bind(s.BindDN, p)
		}
	}
	if err != nil {
		conn.Close()
		return nil, errors.New("LDAP service bind failed")
	}
	return conn, nil
}
func (d *Directory) Source() (*models.LDAPConfig, error) {
	s := new(models.LDAPConfig)
	ok, e := d.DB.ID(1).Where("enabled=?", true).Get(s)
	if e != nil {
		return nil, e
	}
	if !ok {
		return nil, errors.New("LDAP is disabled")
	}
	return s, nil
}
func attributes(s *models.LDAPConfig) []string {
	result := []string{s.UsernameAttribute, s.KeyAttribute, s.NameAttribute, "userAccountControl"}
	if s.EmailAttribute != "" {
		result = append(result, s.EmailAttribute)
	}
	return result
}
func expandFilter(pattern, value string) string {
	return strings.NewReplacer("%[1]s", value, "%s", value).Replace(pattern)
}
func filter(s *models.LDAPConfig) string { return expandFilter(s.Filter, "*") }
func loginFilter(s *models.LDAPConfig, name string) string {
	escaped := ldap.EscapeFilter(name)
	return fmt.Sprintf("(&%s(%s=%s))", expandFilter(s.Filter, escaped), s.UsernameAttribute, escaped)
}
func disabled(e *ldap.Entry) bool {
	flags, _ := strconv.Atoi(e.GetAttributeValue("userAccountControl"))
	return flags&2 != 0
}
func (d *Directory) Login(name, password string) (*models.User, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if password == "" {
		return nil, errors.New("invalid credentials")
	}
	s, err := d.Source()
	if err != nil {
		return nil, err
	}
	conn, err := d.connect(s)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	f := loginFilter(s, name)
	res, err := conn.Search(ldap.NewSearchRequest(s.BaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 2, 15, false, f, attributes(s), nil))
	if err != nil || len(res.Entries) != 1 {
		return nil, errors.New("invalid credentials")
	}
	entry := res.Entries[0]
	if disabled(entry) {
		return nil, errors.New("account disabled")
	}
	if err = conn.Bind(entry.DN, password); err != nil {
		return nil, errors.New("invalid credentials")
	}
	u, err := d.upsert(s, entry)
	if err != nil {
		return nil, err
	}
	if !u.Active {
		return nil, errors.New("account disabled by administrator")
	}
	return u, nil
}
func (d *Directory) upsert(s *models.LDAPConfig, e *ldap.Entry) (*models.User, error) {
	tx := d.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return nil, err
	}
	if _, err := tx.Exec("UPDATE user SET id=id WHERE id=0"); err != nil {
		return nil, err
	}
	u, _, err := d.upsertTx(tx, s, e)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return u, nil
}
func (d *Directory) upsertTx(tx *xorm.Session, s *models.LDAPConfig, e *ldap.Entry) (*models.User, bool, error) {
	key := e.GetRawAttributeValue(s.KeyAttribute)
	name := e.GetAttributeValue(s.UsernameAttribute)
	if name == "" {
		return nil, false, errors.New("LDAP stable identity missing")
	}
	stable := fmt.Sprintf("%x", key)
	u := new(models.User)
	ok := false
	var err error
	if len(key) > 0 {
		ok, err = tx.Where("ldap_account=1 AND ldap_key=?", stable).Get(u)
		if err != nil {
			return nil, false, err
		}
	}
	if !ok {
		var existing models.User
		exists, err := tx.Where("name=?", name).Get(&existing)
		if err != nil {
			return nil, false, err
		}
		if exists {
			if !existing.LDAPAccount {
				return nil, false, &LocalAccountConflict{Name: existing.Name}
			}
			return nil, false, errors.New("LDAP username belongs to another authentication identity")
		}
	}
	if len(key) == 0 {
		return nil, false, errors.New("LDAP stable identity missing")
	}
	email := ""
	if s.EmailAttribute != "" {
		email = strings.TrimSpace(e.GetAttributeValue(s.EmailAttribute))
		if email != "" {
			address, err := mail.ParseAddress(email)
			if err != nil || address.Address != email || len(email) > 254 {
				return nil, false, errors.New("LDAP email attribute is invalid")
			}
		}
	}
	if !ok {
		u = &models.User{Name: name, Email: email, DisplayName: e.GetAttributeValue(s.NameAttribute), LDAPAccount: true, LDAPKey: stable, Active: false}
		if _, err = tx.Insert(u); err != nil {
			return nil, false, err
		}
	} else {
		if !u.Active && !u.LDAPDisabled {
			return nil, false, errors.New("account disabled by administrator")
		}
		u.DisplayName = e.GetAttributeValue(s.NameAttribute)
		if s.EmailAttribute != "" {
			u.Email = email
		}
		u.Active = true
		u.LDAPDisabled = false
		if _, err = tx.ID(u.ID).Cols("display_name", "email", "active", "ldap_disabled").Update(u); err != nil {
			return nil, false, err
		}
	}
	return u, !ok, nil
}
func (d *Directory) Sync() error { return d.synchronize(true, 0, "") }
func (d *Directory) SyncAs(actorID int64, actorName string) error {
	return d.synchronize(false, actorID, actorName)
}
func (d *Directory) synchronize(automatic bool, actorID int64, actorName string) error {
	if !d.mu.TryLock() {
		return ErrSyncBusy
	}
	defer d.mu.Unlock()
	return d.syncLocked(automatic, actorID, actorName)
}
func (d *Directory) syncLocked(automatic bool, actorID int64, actorName string) error {
	s, err := d.Source()
	if err != nil {
		return err
	}
	if !s.SyncEnabled {
		if automatic {
			return nil
		}
		return ErrSyncDisabled
	}
	trigger := "manual"
	if automatic {
		trigger = "scheduled"
	}
	run := &models.LDAPSyncRun{Trigger: trigger, ActorID: actorID, ActorName: actorName, Started: time.Now().UnixMilli(), Status: "running"}
	if _, err = d.DB.Insert(run); err != nil {
		return err
	}
	s.LastSyncReport = ""
	err = d.sync(s, run)
	if err == nil {
		return nil
	}
	var report SyncReport
	_ = json.Unmarshal([]byte(s.LastSyncReport), &report)
	completeRun(run, report, err)
	tx := d.DB.NewSession()
	defer tx.Close()
	if storeErr := tx.Begin(); storeErr != nil {
		return storeErr
	}
	s.LastSync = time.Now().Unix()
	s.LastError = err.Error()
	if _, storeErr := tx.ID(s.ID).Cols("last_sync", "last_error", "last_sync_report").Update(s); storeErr != nil {
		return storeErr
	}
	if _, storeErr := tx.ID(run.ID).AllCols().Update(run); storeErr != nil {
		return storeErr
	}
	if storeErr := notifications.Enqueue(tx, notifications.LDAPEvent(*run)); storeErr != nil {
		return storeErr
	}
	if storeErr := tx.Commit(); storeErr != nil {
		return storeErr
	}
	return err
}
func (d *Directory) sync(s *models.LDAPConfig, run *models.LDAPSyncRun) error {
	if d.readUsers != nil {
		entries, err := d.readUsers(s)
		if err != nil {
			return err
		}
		return d.syncEntriesRun(s, entries, run)
	}
	conn, err := d.connect(s)
	if err != nil {
		return err
	}
	defer conn.Close()
	res, err := searchUsers(conn, s)
	if err != nil {
		return errors.New("LDAP synchronization query failed")
	}
	return d.syncEntriesRun(s, res.Entries, run)
}

type searchClient interface {
	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
	SearchWithPaging(*ldap.SearchRequest, uint32) (*ldap.SearchResult, error)
}

func searchUsers(conn searchClient, s *models.LDAPConfig) (*ldap.SearchResult, error) {
	request := ldap.NewSearchRequest(s.BaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 15, false, filter(s), attributes(s), nil)
	if s.UsePagedSearch {
		return conn.SearchWithPaging(request, s.SearchPageSize)
	}
	return conn.Search(request)
}
func (d *Directory) syncEntries(s *models.LDAPConfig, entries []*ldap.Entry) error {
	return d.syncEntriesRun(s, entries, nil)
}
func (d *Directory) syncEntriesRun(s *models.LDAPConfig, entries []*ldap.Entry, run *models.LDAPSyncRun) (result error) {
	report := SyncReport{MatchedUsers: len(entries), SkippedLocalUsers: []string{}}
	committed := false
	defer func() {
		if !committed {
			report.SyncedUsers = 0
			report.CreatedUsers = 0
			report.UpdatedUsers = 0
			report.DeactivatedUsers = 0
		}
		encoded, _ := json.Marshal(report)
		s.LastSyncReport = string(encoded)
	}()
	if len(entries) == 0 && !s.AllowDeactivateAll {
		return ErrEmptyResult
	}
	tx := d.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE user SET id=id WHERE id=0"); err != nil {
		return err
	}
	seen := map[string]bool{}
	identities := map[string]bool{}
	for _, entry := range entries {
		if disabled(entry) {
			report.SkippedDisabledCount++
			continue
		}
		key := entry.GetRawAttributeValue(s.KeyAttribute)
		identity := fmt.Sprintf("%x", key)
		if len(key) > 0 && identities[identity] {
			return errors.New("duplicate LDAP stable identity")
		}
		if len(key) > 0 {
			identities[identity] = true
		}
		var before models.User
		if _, err := tx.Where("ldap_account=1 AND ldap_key=?", identity).Get(&before); err != nil {
			return err
		}
		u, created, err := d.upsertTx(tx, s, entry)
		if err != nil {
			var conflict *LocalAccountConflict
			if errors.As(err, &conflict) {
				report.SkippedLocalCount++
				if len(report.SkippedLocalUsers) < 20 {
					report.SkippedLocalUsers = append(report.SkippedLocalUsers, conflict.Name)
				}
				log.Printf("LDAP synchronization skipped local account %q", conflict.Name)
				continue
			}
			if err.Error() == "account disabled by administrator" {
				report.SkippedDisabledCount++
				continue
			}
			return err
		}
		seen[u.LDAPKey] = true
		report.SyncedUsers++
		if created {
			report.CreatedUsers++
		} else if before.DisplayName != u.DisplayName || before.Email != u.Email || before.Active != u.Active || before.LDAPDisabled != u.LDAPDisabled {
			report.UpdatedUsers++
		}
	}
	var users []models.User
	if err := tx.Where("ldap_account=1").Find(&users); err != nil {
		return err
	}
	for _, u := range users {
		if seen[u.LDAPKey] || !u.Active {
			continue
		}
		u.Active = false
		u.LDAPDisabled = true
		if _, err := tx.ID(u.ID).Cols("active", "ldap_disabled").Update(&u); err != nil {
			return err
		}
		if _, err := tx.Where("user_id=?", u.ID).Delete(new(models.Session)); err != nil {
			return err
		}
		report.DeactivatedUsers++
	}
	encoded, _ := json.Marshal(report)
	s.LastSyncReport = string(encoded)
	s.LastSync = time.Now().Unix()
	s.LastError = ""
	if _, err := tx.ID(s.ID).Cols("last_sync", "last_error", "last_sync_report").Update(s); err != nil {
		return err
	}
	if run != nil {
		completeRun(run, report, nil)
		if _, err := tx.ID(run.ID).AllCols().Update(run); err != nil {
			return err
		}
		if err := notifications.Enqueue(tx, notifications.LDAPEvent(*run)); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}
