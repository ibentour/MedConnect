package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"medconnect-oriental/backend/internal/models"
)

// ──────────────────────────────────────────────────────────────────────
// Audit Middleware
// ──────────────────────────────────────────────────────────────────────
//
// Logs every HTTP request to the audit_logs table for compliance with
// Moroccan Law 09-08. The audit entry is written asynchronously after
// the request completes, so it captures the final HTTP status code.
//
// Captured fields:
//   - UserID:    from JWT claims (nil if unauthenticated)
//   - Username:  from JWT claims (empty if unauthenticated)
//   - Action:    HTTP method + path (e.g. "POST /api/referrals")
//   - TargetID:  resource ID from URL param :id (if present)
//   - IPAddress: client IP via Gin's ClientIP (proxy-aware)
//   - UserAgent: client User-Agent header
//   - Status:    HTTP response status code
//   - Timestamp: server time at request completion
// ──────────────────────────────────────────────────────────────────────

// AuditMiddleware returns a Gin middleware that writes an audit log entry
// for every HTTP request. Requires a GORM database connection.
func AuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Let the request proceed first
		c.Next()

		// ── Collect audit data after handler execution ────────

		// User identity (may be nil for unauthenticated endpoints like /login)
		var userID *uuid.UUID
		var username string

		if uid, exists := c.Get("user_id"); exists {
			if id, ok := uid.(uuid.UUID); ok {
				userID = &id
			}
		}
		if uname, exists := c.Get("username"); exists {
			if name, ok := uname.(string); ok {
				username = name
			}
		}

		// Action: method + path (e.g. "PATCH /api/referrals/abc-123/schedule")
		action := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		if action == fmt.Sprintf("%s ", c.Request.Method) {
			// FullPath() returns empty for unregistered routes; fall back to raw URL
			action = fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		}

		// Target ID: extract from URL param ":id" if present
		targetID := c.Param("id")

		// Client IP (proxy-aware)
		ipAddress := c.ClientIP()

		// User-Agent
		userAgent := c.Request.UserAgent()

		// HTTP status code
		statusCode := c.Writer.Status()

		// Timestamp
		timestamp := time.Now()

		// ── Write audit log asynchronously ────────────────────
		entry := models.AuditLog{
			UserID:    userID,
			Username:  username,
			Action:    action,
			TargetID:  targetID,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Status:    statusCode,
			Timestamp: timestamp,
		}

		go func(log models.AuditLog) {
			if err := db.Create(&log).Error; err != nil {
				// Log to stderr — in production, this should go to a structured logger
				fmt.Printf("[AUDIT ERROR] Failed to write audit log: %v | Entry: %+v\n", err, log)
			}
		}(entry)
	}
}
