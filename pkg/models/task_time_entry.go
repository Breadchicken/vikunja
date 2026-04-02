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

	"xorm.io/xorm"
)

// TaskTimeEntry represents a time tracking entry for a task
type TaskTimeEntry struct {
	// The unique, numeric id of this time entry.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"timeentry"`
	// The task this time entry belongs to.
	TaskID int64 `xorm:"bigint not null index" json:"task_id" param:"task"`
	// The ID of the user who created this time entry.
	UserID int64 `xorm:"bigint not null" json:"-"`
	// The user who created this time entry.
	User *user.User `xorm:"-" json:"user"`
	// The start time of this time entry.
	Start time.Time `xorm:"datetime not null" json:"start"`
	// The end time of this time entry. If null, the timer is still running.
	End time.Time `xorm:"datetime null" json:"end"`
	// The duration of this time entry in seconds. Computed when the timer is stopped.
	Duration int64 `xorm:"bigint null" json:"duration"`
	// Whether this time entry is billable.
	Billable bool `xorm:"bool default true" json:"billable"`
	// An optional description for this time entry.
	Description string `xorm:"text null" json:"description"`

	// A timestamp when this entry was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this entry was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName holds the table name for task time entries
func (*TaskTimeEntry) TableName() string {
	return "task_time_entries"
}

// Create creates a new time entry
// @Summary Create a new time entry
// @Description Create a new time entry for a task. The user needs at least write access to the task.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Param entry body models.TaskTimeEntry true "The time entry object"
// @Success 201 {object} models.TaskTimeEntry "The created time entry."
// @Failure 400 {object} web.HTTPError "Invalid time entry object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/time-entries [put]
func (te *TaskTimeEntry) Create(s *xorm.Session, a web.Auth) (err error) {
	// Check if the task exists
	_, err = GetTaskSimple(s, &Task{ID: te.TaskID})
	if err != nil {
		return err
	}

	te.UserID = a.GetID()

	if te.Start.IsZero() {
		te.Start = time.Now()
	}

	// If end is set, compute duration
	if !te.End.IsZero() {
		te.Duration = int64(te.End.Sub(te.Start).Seconds())
	}

	_, err = s.Insert(te)
	if err != nil {
		return err
	}

	te.User, err = user.GetUserByID(s, te.UserID)
	return
}

// ReadOne returns a single time entry
// @Summary Get a time entry
// @Description Get a single time entry. The user needs at least read access to the task.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Param timeEntryID path int true "Time Entry ID"
// @Success 200 {object} models.TaskTimeEntry "The time entry."
// @Failure 404 {object} web.HTTPError "The time entry was not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/time-entries/{timeEntryID} [get]
func (te *TaskTimeEntry) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	entry := &TaskTimeEntry{}
	exists, err := s.Where("id = ? AND task_id = ?", te.ID, te.TaskID).Get(entry)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTaskTimeEntryDoesNotExist{ID: te.ID, TaskID: te.TaskID}
	}

	*te = *entry

	te.User, err = user.GetUserByID(s, te.UserID)
	return
}

// ReadAll returns all time entries for a task
// @Summary Get all time entries for a task
// @Description Get all time entries for a task. The user needs at least read access to the task.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Success 200 {array} models.TaskTimeEntry "All time entries for the task."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/time-entries [get]
func (te *TaskTimeEntry) ReadAll(s *xorm.Session, auth web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	// Check if the user has access to the task
	canRead, _, err := te.CanRead(s, auth)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	entries := []*TaskTimeEntry{}
	query := s.Where("task_id = ?", te.TaskID).OrderBy("start DESC")
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&entries)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get all user IDs
	var userIDs []int64
	for _, entry := range entries {
		userIDs = append(userIDs, entry.UserID)
	}

	users, err := getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return nil, 0, 0, err
	}

	for _, entry := range entries {
		entry.User = users[entry.UserID]
	}

	numberOfTotalItems, err = s.Where("task_id = ?", te.TaskID).Count(&TaskTimeEntry{})
	return entries, len(entries), numberOfTotalItems, err
}

// Update updates a time entry
// @Summary Update a time entry
// @Description Update a time entry. The user needs at least write access to the task.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Param timeEntryID path int true "Time Entry ID"
// @Param entry body models.TaskTimeEntry true "The time entry object"
// @Success 200 {object} models.TaskTimeEntry "The updated time entry."
// @Failure 400 {object} web.HTTPError "Invalid time entry object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task."
// @Failure 404 {object} web.HTTPError "The time entry was not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/time-entries/{timeEntryID} [post]
func (te *TaskTimeEntry) Update(s *xorm.Session, _ web.Auth) (err error) {
	existing := &TaskTimeEntry{}
	exists, err := s.Where("id = ? AND task_id = ?", te.ID, te.TaskID).Get(existing)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTaskTimeEntryDoesNotExist{ID: te.ID, TaskID: te.TaskID}
	}

	// Recompute duration if end is set
	if !te.End.IsZero() && !te.Start.IsZero() {
		te.Duration = int64(te.End.Sub(te.Start).Seconds())
	}

	_, err = s.Where("id = ?", te.ID).
		Cols("start", "end", "duration", "billable", "description").
		Update(te)
	if err != nil {
		return err
	}

	te.User, err = user.GetUserByID(s, existing.UserID)
	return
}

// Delete deletes a time entry
// @Summary Delete a time entry
// @Description Delete a time entry. The user needs at least write access to the task.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Param timeEntryID path int true "Time Entry ID"
// @Success 200 {object} models.Message "The time entry was deleted successfully."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task."
// @Failure 404 {object} web.HTTPError "The time entry was not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/time-entries/{timeEntryID} [delete]
func (te *TaskTimeEntry) Delete(s *xorm.Session, _ web.Auth) (err error) {
	existing := &TaskTimeEntry{}
	exists, err := s.Where("id = ? AND task_id = ?", te.ID, te.TaskID).Get(existing)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTaskTimeEntryDoesNotExist{ID: te.ID, TaskID: te.TaskID}
	}

	_, err = s.Where("id = ?", te.ID).Delete(&TaskTimeEntry{})
	return
}

// StartTimer starts a new timer for the current user on the given task
func StartTimer(s *xorm.Session, taskID int64, a web.Auth) (entry *TaskTimeEntry, err error) {
	// Check if the task exists
	_, err = GetTaskSimple(s, &Task{ID: taskID})
	if err != nil {
		return nil, err
	}

	// Check if the user already has a running timer (on any task)
	running := &TaskTimeEntry{}
	exists, err := s.Where("user_id = ? AND `end` IS NULL", a.GetID()).Get(running)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTimerAlreadyRunning{TaskID: running.TaskID}
	}

	entry = &TaskTimeEntry{
		TaskID:   taskID,
		UserID:   a.GetID(),
		Start:    time.Now(),
		Billable: true,
	}

	_, err = s.Insert(entry)
	if err != nil {
		return nil, err
	}

	entry.User, err = user.GetUserByID(s, entry.UserID)
	return
}

// StopTimer stops the running timer for the current user
func StopTimer(s *xorm.Session, taskID int64, a web.Auth) (entry *TaskTimeEntry, err error) {
	entry = &TaskTimeEntry{}
	exists, err := s.Where("user_id = ? AND task_id = ? AND `end` IS NULL", a.GetID(), taskID).Get(entry)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNoActiveTimer{TaskID: taskID}
	}

	now := time.Now()
	entry.End = now
	entry.Duration = int64(now.Sub(entry.Start).Seconds())

	_, err = s.Where("id = ?", entry.ID).
		Cols("end", "duration").
		Update(entry)
	if err != nil {
		return nil, err
	}

	entry.User, err = user.GetUserByID(s, entry.UserID)
	return
}

// GetActiveTimer returns the currently running timer for the authenticated user
func GetActiveTimer(s *xorm.Session, a web.Auth) (entry *TaskTimeEntry, err error) {
	entry = &TaskTimeEntry{}
	exists, err := s.Where("user_id = ? AND `end` IS NULL", a.GetID()).Get(entry)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	entry.User, err = user.GetUserByID(s, entry.UserID)
	return
}
