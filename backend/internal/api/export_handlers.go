// Package api provides HTTP handlers for MedConnect API endpoints.
package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"medconnect-oriental/backend/internal/models"
)

// AuditLogFilter holds filter parameters for audit log queries
type AuditLogFilter struct {
	UserID    *uuid.UUID `form:"user_id"`
	Action    string     `form:"action"`
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
	Limit     int        `form:"limit"`
	Offset    int        `form:"offset"`
}

// GetAuditLogs returns audit logs with filtering and pagination
func (h *HandlerContext) GetAuditLogs(c *gin.Context) {
	var filter AuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	// Build query with filters
	query := h.DB.Model(&models.AuditLog{}).Order("timestamp DESC")

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Action != "" {
		query = query.Where("action LIKE ?", "%"+filter.Action+"%")
	}
	if filter.StartDate != nil {
		query = query.Where("timestamp >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("timestamp <= ?", *filter.EndDate)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Fetch paginated logs
	var logs []models.AuditLog
	if err := query.Limit(filter.Limit).Offset(filter.Offset).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load audit logs"})
		return
	}

	// Build pagination metadata
	pagination := PaginationMeta{
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		Total:   total,
		HasNext: int64(filter.Offset+filter.Limit) < total,
		HasPrev: filter.Offset > 0,
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":       logs,
		"pagination": pagination,
	})
}

// GetAuditLogExport handles CSV and PDF export of audit logs
func (h *HandlerContext) GetAuditLogExport(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")

	// Parse date filters
	var startDate, endDate *time.Time
	if sd := c.Query("start_date"); sd != "" {
		if t, err := time.Parse(time.RFC3339, sd); err == nil {
			startDate = &t
		}
	}
	if ed := c.Query("end_date"); ed != "" {
		if t, err := time.Parse(time.RFC3339, ed); err == nil {
			endDate = &t
		}
	}

	// Build query with filters
	query := h.DB.Model(&models.AuditLog{}).Order("timestamp DESC")

	if startDate != nil {
		query = query.Where("timestamp >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("timestamp <= ?", *endDate)
	}

	// Fetch all logs for export (no limit)
	var logs []models.AuditLog
	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load audit logs for export"})
		return
	}

	switch format {
	case "csv":
		h.exportAuditLogsCSV(c, logs)
	case "pdf":
		h.exportAuditLogsPDF(c, logs)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format. Use 'csv' or 'pdf'"})
	}
}

// exportAuditLogsCSV writes audit logs as CSV
func (h *HandlerContext) exportAuditLogsCSV(c *gin.Context, logs []models.AuditLog) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=audit_logs.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"ID", "User ID", "Username", "Action", "Target ID", "IP Address", "Status", "Timestamp"})

	// Write data
	for _, log := range logs {
		writer.Write([]string{
			log.ID.String(),
			stringOrEmpty(log.UserID),
			log.Username,
			log.Action,
			log.TargetID,
			log.IPAddress,
			fmt.Sprintf("%d", log.Status),
			log.Timestamp.Format(time.RFC3339),
		})
	}
}

// exportAuditLogsPDF writes audit logs as HTML (printable to PDF)
func (h *HandlerContext) exportAuditLogsPDF(c *gin.Context, logs []models.AuditLog) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>MedConnect Audit Logs Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #1e40af; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; font-size: 12px; }
        th { background-color: #1e40af; color: white; }
        tr:nth-child(even) { background-color: #f2f2f2; }
        .meta { color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <h1>MedConnect Audit Logs Report</h1>
    <p class="meta">Generated: %s</p>
    <p class="meta">Total Records: %d</p>
    <table>
        <thead>
            <tr>
                <th>Timestamp</th>
                <th>Username</th>
                <th>Action</th>
                <th>Target ID</th>
                <th>IP Address</th>
                <th>Status</th>
            </tr>
        </thead>
        <tbody>
%s        </tbody>
    </table>
</body>
</html>`

	rows := ""
	for _, log := range logs {
		rows += fmt.Sprintf(`            <tr>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%d</td>
            </tr>
`, log.Timestamp.Format("2006-01-02 15:04:05"), escapeHTML(log.Username), escapeHTML(log.Action), escapeHTML(log.TargetID), escapeHTML(log.IPAddress), log.Status)
	}

	finalHTML := fmt.Sprintf(html, time.Now().Format("2006-01-02 15:04:05"), len(logs), rows)

	c.Header("Content-Type", "text/html")
	c.Header("Content-Disposition", "attachment; filename=audit_logs.html")
	c.Writer.Write([]byte(finalHTML))
}

// ReferralExportFilter holds filter parameters for referral export
type ReferralExportFilter struct {
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
	DepartmentID *uuid.UUID `form:"department_id"`
	Status       string     `form:"status"`
}

// GetReferralsExport handles CSV and PDF export of referrals
func (h *HandlerContext) GetReferralsExport(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")

	var filter ReferralExportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build query with filters
	query := h.DB.Model(&models.Referral{}).Preload("Patient").Preload("Department").Preload("Creator").Order("created_at DESC")

	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}
	if filter.DepartmentID != nil {
		query = query.Where("current_dept_id = ?", *filter.DepartmentID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Fetch referrals for export
	var referrals []models.Referral
	if err := query.Find(&referrals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load referrals for export"})
		return
	}

	switch format {
	case "csv":
		h.exportReferralsCSV(c, referrals)
	case "pdf":
		h.exportReferralsPDF(c, referrals)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format. Use 'csv' or 'pdf'"})
	}
}

// exportReferralsCSV writes referrals as CSV
func (h *HandlerContext) exportReferralsCSV(c *gin.Context, referrals []models.Referral) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=referrals.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"ID", "Patient ID", "Creator", "Department", "Status", "Urgency", "Created At", "Appointment Date"})

	// Write data
	for _, ref := range referrals {
		creatorName := ""
		if ref.Creator.Username != "" {
			creatorName = ref.Creator.Username
		}
		deptName := ""
		if ref.Department.Name != "" {
			deptName = ref.Department.Name
		}
		appointmentDate := ""
		if ref.AppointmentDate != nil {
			appointmentDate = ref.AppointmentDate.Format(time.RFC3339)
		}

		writer.Write([]string{
			ref.ID.String(),
			ref.PatientID.String(),
			creatorName,
			deptName,
			string(ref.Status),
			string(ref.Urgency),
			ref.CreatedAt.Format(time.RFC3339),
			appointmentDate,
		})
	}
}

// exportReferralsPDF writes referrals as HTML (printable to PDF)
func (h *HandlerContext) exportReferralsPDF(c *gin.Context, referrals []models.Referral) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>MedConnect Referrals Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #1e40af; }
        table { width: 100%%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; font-size: 12px; }
        th { background-color: #1e40af; color: white; }
        tr:nth-child(even) { background-color: #f2f2f2; }
        .meta { color: #666; font-size: 12px; }
        .status { padding: 2px 6px; border-radius: 3px; }
        .status-PENDING { background: #fef3c7; color: #92400e; }
        .status-SCHEDULED { background: #d1fae5; color: #065f46; }
        .status-REDIRECTED { background: #dbeafe; color: #1e40af; }
        .status-DENIED { background: #fee2e2; color: #991b1b; }
        .status-CANCELED { background: #f3f4f6; color: #374151; }
    </style>
</head>
<body>
    <h1>MedConnect Referrals Report</h1>
    <p class="meta">Generated: %s</p>
    <p class="meta">Total Referrals: %d</p>
    <table>
        <thead>
            <tr>
                <th>Date</th>
                <th>Creator</th>
                <th>Department</th>
                <th>Status</th>
                <th>Urgency</th>
            </tr>
        </thead>
        <tbody>
%s        </tbody>
    </table>
</body>
</html>`

	rows := ""
	for _, ref := range referrals {
		rows += fmt.Sprintf(`            <tr>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td><span class="status status-%s">%s</span></td>
                <td>%s</td>
            </tr>
`, ref.CreatedAt.Format("2006-01-02 15:04"), escapeHTML(ref.Creator.Username), escapeHTML(ref.Department.Name), ref.Status, ref.Status, ref.Urgency)
	}

	finalHTML := fmt.Sprintf(html, time.Now().Format("2006-01-02 15:04:05"), len(referrals), rows)

	c.Header("Content-Type", "text/html")
	c.Header("Content-Disposition", "attachment; filename=referrals.html")
	c.Writer.Write([]byte(finalHTML))
}

// GetUsersForFilter returns list of users for filter dropdown
func (h *HandlerContext) GetUsersForFilter(c *gin.Context) {
	var users []models.User
	if err := h.DB.Select("id, username").Order("username ASC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetActionsForFilter returns unique actions for filter dropdown
func (h *HandlerContext) GetActionsForFilter(c *gin.Context) {
	var actions []string
	if err := h.DB.Model(&models.AuditLog{}).Distinct("action").Order("action ASC").Pluck("action", &actions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load actions"})
		return
	}
	c.JSON(http.StatusOK, actions)
}

// Helper functions
func stringOrEmpty(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
