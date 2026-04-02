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

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type taskTimeEntries20260402143643 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk"`
	TaskID      int64     `xorm:"bigint not null index"`
	UserID      int64     `xorm:"bigint not null"`
	Start       time.Time `xorm:"datetime not null"`
	End         time.Time `xorm:"datetime null"`
	Duration    int64     `xorm:"bigint null"`
	Billable    bool      `xorm:"bool default true"`
	Description string    `xorm:"text null"`
	Created     time.Time `xorm:"created not null"`
	Updated     time.Time `xorm:"updated not null"`
}

func (taskTimeEntries20260402143643) TableName() string {
	return "task_time_entries"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260402143643",
		Description: "add task_time_entries table for time tracking",
		Migrate: func(tx *xorm.Engine) error {
			_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS task_time_entries (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				task_id INTEGER NOT NULL,
				user_id INTEGER NOT NULL,
				start DATETIME NOT NULL,
				"end" DATETIME NULL,
				duration INTEGER NULL,
				billable INTEGER DEFAULT 1,
				description TEXT NULL,
				created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS IDX_task_time_entries_task_id ON task_time_entries (task_id)`)
			return err
		},
		Rollback: func(tx *xorm.Engine) error {
			_, err := tx.Exec(`DROP TABLE IF EXISTS task_time_entries`)
			return err
		},
	})
}
