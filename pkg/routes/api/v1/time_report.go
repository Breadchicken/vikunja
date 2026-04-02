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
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
)

// TimeReportSummary provides aggregated data for the time report
type TimeReportSummary struct {
	Entries          []*models.TimeReportEntry `json:"entries"`
	TotalDuration    int64                     `json:"total_duration"`
	BillableDuration int64                     `json:"billable_duration"`
	TotalEntries     int64                     `json:"total_entries"`
}

// GetTimeReport returns a time report for a specific project or globally
// @Summary Get time report
// @Description Get a time report, optionally filtered by project, user, date range, and billable status.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project_id query int false "Filter by project ID (0 or omit for all projects)"
// @Param user_id query int false "Filter by user ID (0 or omit for all users)"
// @Param from query string false "Filter entries from this date (RFC3339 format)"
// @Param to query string false "Filter entries until this date (RFC3339 format)"
// @Param billable query string false "Filter by billable status: 'true', 'false', or omit for all"
// @Success 200 {object} TimeReportSummary "The time report summary."
// @Failure 403 {object} web.HTTPError "The user does not have access."
// @Failure 500 {object} models.Message "Internal error"
// @Router /time-report [get]
func GetTimeReport(c *echo.Context) error {
	currentAuth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	// Parse filter parameters
	projectIDStr := c.QueryParam("project_id")
	userIDStr := c.QueryParam("user_id")
	fromStr := c.QueryParam("from")
	toStr := c.QueryParam("to")
	billableStr := c.QueryParam("billable")

	var projectID int64
	if projectIDStr != "" {
		projectID, err = strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid project_id parameter")
		}
	}

	var filterUserID int64
	if userIDStr != "" {
		filterUserID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user_id parameter")
		}
	}

	var fromDate, toDate time.Time
	if fromStr != "" {
		fromDate, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid from date. Use RFC3339 format.")
		}
	}
	if toStr != "" {
		toDate, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid to date. Use RFC3339 format.")
		}
	}

	entries, err := models.GetTimeReportEntries(s, currentAuth, projectID, filterUserID, fromDate, toDate, billableStr)
	if err != nil {
		return err
	}

	summary := &TimeReportSummary{
		Entries:      entries,
		TotalEntries: int64(len(entries)),
	}

	for _, e := range entries {
		summary.TotalDuration += e.Duration
		if e.Billable {
			summary.BillableDuration += e.Duration
		}
	}

	return c.JSON(http.StatusOK, summary)
}

// ExportTimeReportCSV exports the time report as CSV
// @Summary Export time report as CSV
// @Description Export a time report as CSV file, optionally filtered by project, user, date range, and billable status.
// @tags task
// @Accept json
// @Produce text/csv
// @Security JWTKeyAuth
// @Param project_id query int false "Filter by project ID"
// @Param user_id query int false "Filter by user ID"
// @Param from query string false "Filter entries from this date (RFC3339 format)"
// @Param to query string false "Filter entries until this date (RFC3339 format)"
// @Param billable query string false "Filter by billable status: 'true', 'false', or omit for all"
// @Success 200 {file} file "CSV file download"
// @Failure 403 {object} web.HTTPError "The user does not have access."
// @Failure 500 {object} models.Message "Internal error"
// @Router /time-report/csv [get]
func ExportTimeReportCSV(c *echo.Context) error {
	currentAuth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	// Parse filter parameters (same as JSON endpoint)
	projectIDStr := c.QueryParam("project_id")
	userIDStr := c.QueryParam("user_id")
	fromStr := c.QueryParam("from")
	toStr := c.QueryParam("to")
	billableStr := c.QueryParam("billable")

	var projectID int64
	if projectIDStr != "" {
		projectID, err = strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid project_id parameter")
		}
	}

	var filterUserID int64
	if userIDStr != "" {
		filterUserID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user_id parameter")
		}
	}

	var fromDate, toDate time.Time
	if fromStr != "" {
		fromDate, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid from date. Use RFC3339 format.")
		}
	}
	if toStr != "" {
		toDate, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid to date. Use RFC3339 format.")
		}
	}

	entries, err := models.GetTimeReportEntries(s, currentAuth, projectID, filterUserID, fromDate, toDate, billableStr)
	if err != nil {
		return err
	}

	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=time-report.csv")
	c.Response().WriteHeader(http.StatusOK)

	w := csv.NewWriter(c.Response())

	// Write header
	err = w.Write([]string{
		"Time Entry ID",
		"Task ID",
		"Task Title",
		"Project ID",
		"Project Name",
		"User ID",
		"Username",
		"Start",
		"End",
		"Duration (seconds)",
		"Duration (hours)",
		"Billable",
		"Description",
	})
	if err != nil {
		return fmt.Errorf("writing csv header: %w", err)
	}

	for _, e := range entries {
		billable := "No"
		if e.Billable {
			billable = "Yes"
		}

		hours := float64(e.Duration) / 3600.0

		err = w.Write([]string{
			strconv.FormatInt(e.TimeEntryID, 10),
			strconv.FormatInt(e.TaskID, 10),
			e.TaskTitle,
			strconv.FormatInt(e.ProjectID, 10),
			e.ProjectName,
			strconv.FormatInt(e.UserID, 10),
			e.Username,
			e.Start.Format(time.RFC3339),
			e.End.Format(time.RFC3339),
			strconv.FormatInt(e.Duration, 10),
			strconv.FormatFloat(hours, 'f', 2, 64),
			billable,
			e.Description,
		})
		if err != nil {
			return fmt.Errorf("writing csv row: %w", err)
		}
	}

	w.Flush()
	return w.Error()
}
