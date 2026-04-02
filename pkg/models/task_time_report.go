// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// TimeReportEntry represents a joined row for time report queries
type TimeReportEntry struct {
	TimeEntryID int64     `xorm:"'id'" json:"time_entry_id"`
	TaskID      int64     `xorm:"'task_id'" json:"task_id"`
	TaskTitle   string    `xorm:"'task_title'" json:"task_title"`
	ProjectID   int64     `xorm:"'project_id'" json:"project_id"`
	ProjectName string    `xorm:"'project_name'" json:"project_name"`
	UserID      int64     `xorm:"'user_id'" json:"user_id"`
	Username    string    `xorm:"'username'" json:"username"`
	Start       time.Time `xorm:"'start'" json:"start"`
	End         time.Time `xorm:"'end'" json:"end"`
	Duration    int64     `xorm:"'duration'" json:"duration"`
	Billable    bool      `xorm:"'billable'" json:"billable"`
	Description string    `xorm:"'description'" json:"description"`
}

// GetTimeReportEntries returns time entries for reporting with task and project info joined.
// It filters by the user's accessible projects, and optionally by project, user, date range, and billable status.
func GetTimeReportEntries(s *xorm.Session, a web.Auth, projectID, filterUserID int64, fromDate, toDate time.Time, billableStr string) ([]*TimeReportEntry, error) {

	// Build conditions
	conds := []builder.Cond{}

	// Only completed entries (timer not running)
	conds = append(conds, builder.NotNull{"tte.`end`"})

	if projectID > 0 {
		// Verify the user has access to this project
		p := &Project{ID: projectID}
		canRead, _, err := p.CanRead(s, a)
		if err != nil {
			return nil, err
		}
		if !canRead {
			return nil, ErrGenericForbidden{}
		}
		conds = append(conds, builder.Eq{"t.project_id": projectID})
	} else {
		// Get all projects the user can access
		u, err := user.GetUserByID(s, a.GetID())
		if err != nil {
			return nil, err
		}
		allProjects, _, err := getAllProjectsForUser(s, a.GetID(), &projectOptions{
			user:    u,
			page:    0,
			perPage: 0,
		})
		if err != nil {
			return nil, err
		}

		var projectIDs []int64
		for _, p := range allProjects {
			projectIDs = append(projectIDs, p.ID)
		}
		if len(projectIDs) == 0 {
			return []*TimeReportEntry{}, nil
		}
		conds = append(conds, builder.In("t.project_id", projectIDs))
	}

	if filterUserID > 0 {
		conds = append(conds, builder.Eq{"tte.user_id": filterUserID})
	}

	if !fromDate.IsZero() {
		conds = append(conds, builder.Gte{"tte.start": fromDate})
	}

	if !toDate.IsZero() {
		conds = append(conds, builder.Lte{"tte.start": toDate})
	}

	if billableStr == "true" {
		conds = append(conds, builder.Eq{"tte.billable": true})
	} else if billableStr == "false" {
		conds = append(conds, builder.Eq{"tte.billable": false})
	}

	var entries []*TimeReportEntry

	err := s.
		Select("tte.id, tte.task_id, t.title as task_title, t.project_id, p.title as project_name, tte.user_id, u.username, tte.start, tte.`end`, tte.duration, tte.billable, tte.description").
		Table("task_time_entries").Alias("tte").
		Join("INNER", []string{"tasks", "t"}, "t.id = tte.task_id").
		Join("INNER", []string{"projects", "p"}, "p.id = t.project_id").
		Join("INNER", []string{"users", "u"}, "u.id = tte.user_id").
		Where(builder.And(conds...)).
		OrderBy("tte.start DESC").
		Find(&entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}


