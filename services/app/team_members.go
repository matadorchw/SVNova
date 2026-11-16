// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package app

import (
	"errors"
	"net/http"
	"svnova/models"
)

// Membership follows Gitea: joining a team joins the organization; leaving the last team leaves it.
func (a *App) changeTeamMember(w http.ResponseWriter, u *models.User, team *models.Team, targetID int64, remove, self bool) error {
	if u == nil || !u.Active {
		return forbidden()
	}
	tx := a.DB.NewSession()
	defer tx.Close()
	if err := tx.Begin(); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE organization SET id=id WHERE id=?", team.OrgID); err != nil {
		return err
	}
	if self {
		if !remove || targetID != u.ID {
			return forbidden()
		}
		exists, err := tx.Where("team_id=? AND user_id=?", team.ID, u.ID).Exist(new(models.TeamMember))
		if err != nil {
			return err
		}
		if !exists {
			return forbidden()
		}
	} else if err := checkOrganizationOwner(tx, u, team.OrgID); err != nil {
		return err
	}
	if remove {
		exists, err := tx.Where("team_id=? AND user_id=?", team.ID, targetID).Exist(new(models.TeamMember))
		if err != nil {
			return err
		}
		if exists && team.IsOwnerTeam {
			count, err := tx.Where("team_id=?", team.ID).Count(new(models.TeamMember))
			if err != nil {
				return err
			}
			if count <= 1 {
				return errors.New("organization requires an owner")
			}
		}
		if _, err := tx.Where("team_id=? AND user_id=?", team.ID, targetID).Delete(new(models.TeamMember)); err != nil {
			return err
		}
		var remaining int64
		if _, err := tx.SQL("SELECT COUNT(*) FROM team_member tm JOIN team t ON t.id=tm.team_id WHERE t.org_id=? AND tm.user_id=?", team.OrgID, targetID).Get(&remaining); err != nil {
			return err
		}
		if remaining == 0 {
			if _, err := tx.Where("org_id=? AND user_id=?", team.OrgID, targetID).Delete(new(models.Membership)); err != nil {
				return err
			}
		}
	} else {
		target := new(models.User)
		exists, err := tx.ID(targetID).Get(target)
		if err != nil {
			return err
		}
		if !exists || !target.Active {
			return errors.New("active user required")
		}
		member, err := tx.Where("org_id=? AND user_id=?", team.OrgID, targetID).Exist(new(models.Membership))
		if err != nil {
			return err
		}
		if !member {
			if _, err := tx.Insert(&models.Membership{OrgID: team.OrgID, UserID: targetID}); err != nil {
				return err
			}
		}
		exists, err = tx.Where("team_id=? AND user_id=?", team.ID, targetID).Exist(new(models.TeamMember))
		if err != nil {
			return err
		}
		if !exists {
			if _, err := tx.Insert(&models.TeamMember{TeamID: team.ID, UserID: targetID}); err != nil {
				return err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return send(w, map[string]bool{"ok": true})
}
