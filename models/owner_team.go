// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

import (
	"fmt"
	"strings"
	"time"
	"xorm.io/xorm"
)

const OwnerTeamName = "Owners"

// MigrateOwnerTeams preserves legacy authority without elevating existing ordinary teams.
func migrateOwnerTeams(db *xorm.Engine) error {
	tx := db.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return err
	}
	var orgs []Organization
	if err := tx.Find(&orgs); err != nil {
		return err
	}
	for _, org := range orgs {
		var members []Membership
		if err := tx.Where("org_id=? AND owner=?", org.ID, true).Find(&members); err != nil {
			return err
		}
		if len(members) == 0 {
			return fmt.Errorf("organization %d has no legacy owner", org.ID)
		}
		var teams []Team
		if err := tx.Where("org_id=?", org.ID).Find(&teams); err != nil {
			return err
		}
		for _, team := range teams {
			if strings.EqualFold(team.Name, OwnerTeamName) {
				if _, err := tx.ID(team.ID).Cols("name").Update(&Team{Name: fmt.Sprintf("Owners-legacy-%d", team.ID)}); err != nil {
					return err
				}
			}
		}
		team := &Team{OrgID: org.ID, Name: OwnerTeamName, IsOwnerTeam: true}
		if _, err := tx.Insert(team); err != nil {
			return err
		}
		for _, member := range members {
			if _, err := tx.Insert(&TeamMember{TeamID: team.ID, UserID: member.UserID}); err != nil {
				return err
			}
		}
	}
	if _, err := tx.Exec("UPDATE membership SET owner=0"); err != nil {
		return err
	}
	if _, err := tx.Insert(&Migration{ID: 2, Applied: time.Now().Unix()}); err != nil {
		return err
	}
	return tx.Commit()
}
