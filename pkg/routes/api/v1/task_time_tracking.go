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

package v1

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
)

// StartTaskTimer starts a timer for the current user on a task
// @Summary Start a timer
// @Description Start a time tracking timer on a task. Only one timer can be running per user at a time.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Success 201 {object} models.TaskTimeEntry "The started timer entry."
// @Failure 400 {object} web.HTTPError "Invalid task ID."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task."
// @Failure 409 {object} web.HTTPError "A timer is already running."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/timers/start [post]
func StartTaskTimer(c *echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	currentAuth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.").Wrap(err)
	}

	// Check task write permission
	s := db.NewSession()
	defer s.Close()

	t := &models.Task{ID: taskID}
	canUpdate, err := t.CanUpdate(s, currentAuth)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if !canUpdate {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusForbidden, "You do not have write access to this task.")
	}

	if err = s.Begin(); err != nil {
		return err
	}

	entry, err := models.StartTimer(s, taskID, currentAuth)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err = s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, entry)
}

// StopTaskTimer stops the running timer for the current user on a task
// @Summary Stop a timer
// @Description Stop the running time tracking timer on a task.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Success 200 {object} models.TaskTimeEntry "The stopped timer entry with computed duration."
// @Failure 400 {object} web.HTTPError "Invalid task ID."
// @Failure 404 {object} web.HTTPError "No active timer found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/timers/stop [post]
func StopTaskTimer(c *echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID")
	}

	currentAuth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	if err = s.Begin(); err != nil {
		return err
	}

	entry, err := models.StopTimer(s, taskID, currentAuth)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err = s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, entry)
}

// GetActiveTimer returns the currently running timer for the authenticated user
// @Summary Get active timer
// @Description Get the currently running time tracking timer for the authenticated user.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} models.TaskTimeEntry "The active timer entry, or null if none is running."
// @Failure 500 {object} models.Message "Internal error"
// @Router /users/timers/active [get]
func GetActiveTimer(c *echo.Context) error {
	currentAuth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	entry, err := models.GetActiveTimer(s, currentAuth)
	if err != nil {
		return err
	}

	if entry == nil {
		return c.JSON(http.StatusOK, nil)
	}

	return c.JSON(http.StatusOK, entry)
}
