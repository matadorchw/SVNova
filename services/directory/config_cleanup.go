// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"svnova/models"
	"time"
	"xorm.io/xorm"
)

var ErrCleanupConfirmation = errors.New("LDAP configuration cleanup confirmation required or expired")
var ErrCleanupOwner = errors.New("cleanup would remove an organization's last owner")
var ErrCleanupAdmin = errors.New("cleanup would remove the last active administrator")
var ErrConfigAdmin = errors.New("administrator access is required")

type CleanupOwnerError struct{ Organization string }

func (e *CleanupOwnerError) Error() string {
	return "cleanup would remove an organization's last owner"
}
func (e *CleanupOwnerError) Unwrap() error { return ErrCleanupOwner }

type CleanupPreview struct {
	RequiresConfirmation bool   `json:"requiresConfirmation"`
	DeleteUsers          int    `json:"deleteUsers"`
	Sessions             int64  `json:"sessions"`
	Memberships          int64  `json:"memberships"`
	TeamMemberships      int64  `json:"teamMemberships"`
	Token                string `json:"token"`
}
type cleanupPlan struct {
	CleanupPreview
	users       []int64
	fingerprint string
	changed     bool
	exists      bool
	previous    models.LDAPConfig
}
type cleanupToken struct {
	Fingerprint string
	Expires     int64
}

func ldapConfigSnapshot(s models.LDAPConfig) models.LDAPConfig {
	s.Created = 0
	s.Updated = 0
	s.LastSync = 0
	s.LastError = ""
	s.LastSyncReport = ""
	return s
}

func (d *Directory) cleanupPlan(tx *xorm.Session, c *models.LDAPConfig, replace bool, password string) (cleanupPlan, error) {
	p := cleanupPlan{users: []int64{}}
	if err := Validate(c); err != nil {
		return p, err
	}
	c.ID = 1
	ok, err := tx.ID(1).Get(&p.previous)
	if err != nil {
		return p, err
	}
	if ok {
		p.exists = true
		c.Password = p.previous.Password
		c.Created = p.previous.Created
		c.LastSync = p.previous.LastSync
		c.LastError = p.previous.LastError
		c.LastSyncReport = p.previous.LastSyncReport
		if sameEndpoint(c.URL, p.previous.URL) {
			c.URL = p.previous.URL
		}
	}
	if c.Enabled && c.BindDN != "" {
		value := ""
		if replace {
			value = password
		} else {
			value, err = d.Vault.Open(c.Password)
		}
		if err != nil || value == "" {
			return p, ErrBindPassword
		}
	}
	p.changed = !p.exists || replace || !reflect.DeepEqual(ldapConfigSnapshot(*c), ldapConfigSnapshot(p.previous))
	if p.exists && p.changed {
		var users []models.User
		if err = tx.Cols("id").Where("ldap_account=1").Asc("id").Find(&users); err != nil {
			return p, err
		}
		for _, u := range users {
			p.users = append(p.users, u.ID)
		}
	}
	p.DeleteUsers = len(p.users)
	p.RequiresConfirmation = p.exists && p.changed
	var memberships []models.Membership
	var teamMembers []models.TeamMember
	if len(p.users) > 0 {
		p.Sessions, err = tx.In("user_id", p.users).Count(new(models.Session))
		if err != nil {
			return p, err
		}
		if err = tx.In("user_id", p.users).Asc("id").Find(&memberships); err != nil {
			return p, err
		}
		p.Memberships = int64(len(memberships))
		if err = tx.In("user_id", p.users).Asc("id").Find(&teamMembers); err != nil {
			return p, err
		}
		p.TeamMemberships = int64(len(teamMembers))
		admins, err := tx.Where("admin=1 AND active=1").NotIn("id", p.users).Count(new(models.User))
		if err != nil {
			return p, err
		}
		if admins == 0 {
			return p, ErrCleanupAdmin
		}
		var owners []struct {
			OrgID  int64
			UserID int64
			Active bool
		}
		if err = tx.SQL("SELECT t.org_id AS org_id,u.id AS user_id,u.active AS active FROM team t JOIN team_member tm ON tm.team_id=t.id JOIN user u ON u.id=tm.user_id JOIN membership m ON m.org_id=t.org_id AND m.user_id=u.id WHERE t.is_owner_team=1").Find(&owners); err != nil {
			return p, err
		}
		deleted := map[int64]bool{}
		for _, id := range p.users {
			deleted[id] = true
		}
		affected := map[int64]bool{}
		remaining := map[int64]bool{}
		for _, o := range owners {
			if deleted[o.UserID] {
				affected[o.OrgID] = true
			} else if o.Active {
				remaining[o.OrgID] = true
			}
		}
		for org := range affected {
			if !remaining[org] {
				var organization models.Organization
				if _, err = tx.ID(org).Get(&organization); err != nil {
					return p, err
				}
				return p, &CleanupOwnerError{Organization: organization.Name}
			}
		}
	}
	hashPassword := sha256.Sum256([]byte(password))
	snapshot := struct {
		Candidate    models.LDAPConfig
		Previous     models.LDAPConfig
		Users        []int64
		Memberships  []models.Membership
		Teams        []models.TeamMember
		Replace      bool
		PasswordHash string
	}{ldapConfigSnapshot(*c), ldapConfigSnapshot(p.previous), p.users, memberships, teamMembers, replace, hex.EncodeToString(hashPassword[:])}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return p, err
	}
	sum := sha256.Sum256(encoded)
	p.fingerprint = hex.EncodeToString(sum[:])
	return p, nil
}

func (d *Directory) PreviewCleanup(candidate *models.LDAPConfig, replace bool, password string) (CleanupPreview, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	tx := d.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return CleanupPreview{}, err
	}
	p, err := d.cleanupPlan(tx, candidate, replace, password)
	if err != nil {
		return CleanupPreview{}, err
	}
	if p.RequiresConfirmation {
		b, _ := json.Marshal(cleanupToken{p.fingerprint, time.Now().Add(10 * time.Minute).Unix()})
		p.Token = d.Vault.Seal(string(b))
	}
	return p.CleanupPreview, nil
}

func (d *Directory) SaveConfirmed(c *models.LDAPConfig, replace bool, password, token string, actorID int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	tx := d.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE migration SET id=id WHERE id=15"); err != nil {
		return err
	}
	if actorID != 0 {
		var actor models.User
		ok, err := tx.ID(actorID).Get(&actor)
		if err != nil {
			return err
		}
		if !ok || !actor.Admin || !actor.Active {
			return ErrConfigAdmin
		}
	}
	p, err := d.cleanupPlan(tx, c, replace, password)
	if err != nil {
		return err
	}
	if p.RequiresConfirmation {
		raw, err := d.Vault.Open(token)
		if err != nil {
			return ErrCleanupConfirmation
		}
		var saved cleanupToken
		if json.Unmarshal([]byte(raw), &saved) != nil || saved.Expires < time.Now().Unix() || saved.Fingerprint != p.fingerprint {
			return ErrCleanupConfirmation
		}
	}
	if !p.changed {
		*c = p.previous
		return nil
	}
	if len(p.users) > 0 {
		if _, err = tx.In("user_id", p.users).Delete(new(models.Session)); err != nil {
			return err
		}
		if _, err = tx.In("user_id", p.users).Delete(new(models.TeamMember)); err != nil {
			return err
		}
		if _, err = tx.In("user_id", p.users).Delete(new(models.Membership)); err != nil {
			return err
		}
		if _, err = tx.In("id", p.users).And("ldap_account>0").Delete(new(models.User)); err != nil {
			return err
		}
	}
	if replace {
		c.Password = d.Vault.Seal(password)
	}
	c.Updated = time.Now().Unix()
	if c.Created == 0 {
		c.Created = c.Updated
	}
	if p.exists {
		_, err = tx.ID(c.ID).AllCols().Update(c)
	} else {
		_, err = tx.Insert(c)
	}
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	log.Printf("LDAP configuration changed; removed %d directory users", p.DeleteUsers)
	return nil
}
