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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// CanRead checks if the user can read time entries for a task (requires read access to the task)
func (te *TaskTimeEntry) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	t := &Task{ID: te.TaskID}
	return t.CanRead(s, a)
}

// CanCreate checks if the user can create time entries for a task (requires write access to the task)
func (te *TaskTimeEntry) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	t := &Task{ID: te.TaskID}
	return t.CanUpdate(s, a)
}

// CanUpdate checks if the user can update this time entry (requires write access to the task)
func (te *TaskTimeEntry) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	t := &Task{ID: te.TaskID}
	return t.CanUpdate(s, a)
}

// CanDelete checks if the user can delete this time entry (requires write access to the task)
func (te *TaskTimeEntry) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	t := &Task{ID: te.TaskID}
	return t.CanUpdate(s, a)
}
