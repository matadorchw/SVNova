// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package models

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

type User struct {
	Created      int64 `xorm:"created"`
	LastLogin    int64
	ID           int64  `xorm:"pk autoincr"`
	Name         string `xorm:"unique notnull"`
	DisplayName  string
	Email        string
	Password     string `json:"-"`
	Salt         string `json:"-"`
	Admin        bool
	Active       bool
	LDAPAccount  bool   `xorm:"'ldap_account'"`
	LDAPKey      string `xorm:"'ldap_key'"`
	LDAPDisabled bool   `xorm:"'ldap_disabled'"`
}
type Session struct {
	Token   string `xorm:"pk" json:"-"`
	UserID  int64  `xorm:"index"`
	CSRF    string
	Expires int64
}
type SiteSetting struct {
	ID       int64 `xorm:"pk"`
	SiteName string
}

type Organization struct {
	FullName, Email, Description, Website, Location string
	Created                                         int64
	ID                                              int64  `xorm:"pk autoincr"`
	Name                                            string `xorm:"unique notnull"`
}
type Membership struct {
	ID     int64 `xorm:"pk autoincr"`
	OrgID  int64 `xorm:"unique(member)"`
	UserID int64 `xorm:"unique(member)"`
	Owner  bool  `json:"-"`
}
type Team struct {
	Description string
	IsOwnerTeam bool
	ID          int64 `xorm:"pk autoincr"`
	OrgID       int64
	Name        string
}
type TeamMember struct {
	ID     int64 `xorm:"pk autoincr"`
	TeamID int64 `xorm:"unique(member)"`
	UserID int64 `xorm:"unique(member)"`
}
type TeamRepository struct {
	ID           int64 `xorm:"pk autoincr"`
	TeamID       int64 `xorm:"unique(grant)"`
	RepositoryID int64 `xorm:"unique(grant)"`
}
type Repository struct {
	Size        int64
	SizeUpdated int64
	SizeChecked int64
	ID          int64  `xorm:"pk autoincr"`
	OrgID       int64  `xorm:"unique(repo)"`
	Name        string `xorm:"unique(repo)"`
	Kind        string
	Status      string
	Path        string `json:"-"`
	UUID        string
	Remote      string
	Username    string
	Password    string `json:"-"`
	SkipTLS     bool
	Interval    int64
	LastSuccess int64
	NextSync    int64
	Revision    int64
	Halted      bool
}
type MirrorConversion struct {
	RepositoryID int64  `xorm:"pk"`
	Snapshot     string `xorm:"TEXT"`
}

type RepositoryPathMove struct {
	RepositoryID  int64 `xorm:"pk"`
	OldPath       string
	NewPath       string
	Token         string
	UUID          string
	Revision      int64
	HadRepository bool
	HadStaging    bool
}
type Task struct {
	ActorID              int64
	ActorName            string
	Verification         string
	VerificationPosition int64
	ProgressStarted      int64
	ProgressUpdated      int64
	Trigger              string
	ID                   int64 `xorm:"pk autoincr"`
	RepositoryID         int64 `xorm:"index"`
	Kind                 string
	State                string
	Phase                string
	Source               string `json:"-"`
	Error                string
	Started              int64
	Finished             int64
	CurrentRevision      int64
	TotalRevision        int64
}
type LDAPConfig struct {
	SkipVerify         bool
	UsePagedSearch     bool
	SearchPageSize     uint32
	SyncEnabled        bool
	AllowDeactivateAll bool
	Created            int64
	Updated            int64
	ID                 int64 `xorm:"pk" json:"-"`
	Enabled            bool
	URL                string
	BindDN             string
	Password           string `json:"-"`
	BaseDN             string
	Filter             string
	UsernameAttribute  string
	KeyAttribute       string
	NameAttribute      string
	EmailAttribute     string
	LastSync           int64
	LastError          string
	LastSyncReport     string `xorm:"TEXT"`
}
type Migration struct {
	ID      int64 `xorm:"pk"`
	Applied int64
}

type LDAPSyncRun struct {
	ID                int64  `xorm:"pk autoincr" json:"id"`
	Trigger           string `json:"trigger"`
	ActorID           int64  `json:"actorID"`
	ActorName         string `json:"actorName"`
	Started           int64  `json:"started"`
	Finished          int64  `json:"finished"`
	DurationMS        int64  `json:"durationMS"`
	Status            string `xorm:"index" json:"status"`
	MatchedUsers      int    `json:"matchedUsers"`
	CreatedUsers      int    `json:"createdUsers"`
	UpdatedUsers      int    `json:"updatedUsers"`
	DisabledUsers     int    `json:"disabledUsers"`
	SkippedLocalCount int    `json:"skippedLocalCount"`
	Report            string `xorm:"TEXT" json:"-"`
	FailureReason     string `json:"failureReason"`
}

func Open(path string) (*xorm.Engine, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	db, err := xorm.NewEngine("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMapper(names.GonicMapper{})
	for _, q := range []string{"PRAGMA foreign_keys=ON", "PRAGMA journal_mode=WAL", "PRAGMA busy_timeout=5000"} {
		if _, err = db.Exec(q); err != nil {
			db.Close()
			return nil, err
		}
	}
	if err = checkLDAPUpgrade(db); err != nil {
		db.Close()
		return nil, err
	}
	if err = db.Sync2(new(Migration)); err != nil {
		db.Close()
		return nil, err
	}
	var latest Migration
	if _, err = db.Desc("id").Get(&latest); err != nil {
		db.Close()
		return nil, err
	}
	if latest.ID > 15 {
		db.Close()
		return nil, errors.New("database schema newer than application")
	}
	err = db.Sync2(new(LDAPSchedule), new(NotificationConfig), new(NotificationDelivery), new(NotificationAttempt), new(TaskExecution), new(LDAPSyncRun), new(MirrorConversion), new(RepositoryPathMove), new(RevisionEvent), new(ActivityIndex), new(SiteSetting), new(User), new(Session), new(Organization), new(Membership), new(Team), new(TeamMember), new(TeamRepository), new(Repository), new(Task), new(LDAPConfig))
	if err != nil {
		db.Close()
		return nil, err
	}
	if err = removeNotificationExclusions(db); err != nil {
		db.Close()
		return nil, err
	}
	// SQLite adds nullable columns to existing tables; normalize unknown size metadata.
	if _, err = db.Exec("UPDATE repository SET size=COALESCE(size,0),size_updated=COALESCE(size_updated,0),size_checked=COALESCE(size_checked,0) WHERE size IS NULL OR size_updated IS NULL OR size_checked IS NULL"); err != nil {
		db.Close()
		return nil, err
	}
	if latest.ID == 0 {
		if _, err = db.Insert(&Migration{ID: 1, Applied: time.Now().Unix()}); err != nil {
			db.Close()
			return nil, err
		}
	}
	if latest.ID < 2 {
		if err = migrateOwnerTeams(db); err != nil {
			db.Close()
			return nil, err
		}
	}
	if err = finishLDAPMigration(db); err != nil {
		db.Close()
		return nil, err
	}
	if _, err = db.Exec("INSERT OR IGNORE INTO migration(id,applied) VALUES(10,?),(11,?),(12,?),(13,?),(14,?)", time.Now().Unix(), time.Now().Unix(), time.Now().Unix(), time.Now().Unix(), time.Now().Unix()); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
func Owner(db *xorm.Engine, u *User, org int64) bool {
	if u == nil || !u.Active {
		return false
	}
	if u.Admin {
		return true
	}
	var count int64
	_, err := db.SQL("SELECT COUNT(*) AS n FROM team t JOIN team_member tm ON tm.team_id=t.id JOIN membership m ON m.org_id=t.org_id AND m.user_id=tm.user_id WHERE t.org_id=? AND t.is_owner_team=? AND tm.user_id=?", org, true, u.ID).Get(&count)
	ok := count > 0
	return err == nil && ok
}
func CanRead(db *xorm.Engine, u *User, r *Repository) bool {
	if u == nil || !u.Active {
		return false
	}
	if Owner(db, u, r.OrgID) {
		return true
	}
	member, err := db.Where("org_id=? AND user_id=?", r.OrgID, u.ID).Exist(new(Membership))
	if err != nil || !member {
		return false
	}
	var rows []struct{ N int64 }
	err = db.SQL("SELECT COUNT(*) AS n FROM team_repository tr JOIN team t ON t.id=tr.team_id JOIN team_member tm ON tm.team_id=t.id WHERE tr.repository_id=? AND t.org_id=? AND tm.user_id=?", r.ID, r.OrgID, u.ID).Find(&rows)
	return err == nil && len(rows) == 1 && rows[0].N > 0
}

var ErrActiveTask = errors.New("repository already has an active task")

func Enqueue(db *xorm.Engine, repoID int64, kind, source string) error {
	return EnqueueTriggered(db, repoID, kind, source, "manual")
}
func EnqueueTriggered(db *xorm.Engine, repoID int64, kind, source, trigger string) error {
	s := db.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		return err
	}
	// Acquire SQLite's writer lock before checking for another active task.
	if _, err := s.Exec("UPDATE repository SET id=id WHERE id=?", repoID); err != nil {
		return err
	}
	exists, err := s.Where("repository_id=? AND state IN (?,?)", repoID, "queued", "running").Exist(new(Task))
	if err != nil {
		return err
	}
	if exists {
		return ErrActiveTask
	}
	r := new(Repository)
	ok, err := s.ID(repoID).Get(r)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("repository not found")
	}
	if kind == "sync" && (r.Kind != "mirror" || r.Status != "ready" || r.Halted) {
		return errors.New("repository cannot synchronize")
	}
	_, err = s.Insert(&Task{RepositoryID: repoID, Kind: kind, Source: source, Trigger: trigger, State: "queued", Started: time.Now().Unix()})
	if err != nil {
		return err
	}
	return s.Commit()
}

func (o Organization) DisplayName() string {
	if o.FullName != "" {
		return o.FullName
	}
	return o.Name
}
