package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"medconnect-oriental/backend/internal/middleware"
	"medconnect-oriental/backend/internal/models"
)

// ══════════════════════════════════════════════════════════════════════
// GET /api/history — View Referral History
// Role: CHU_DOC (own dept history) or LEVEL_2_DOC (own created history)
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) GetReferralHistory(c *gin.Context) {
	userID, _ := middleware.GetUserIDFromContext(c)
	userRole := middleware.GetUserRoleFromContext(c)
	deptID := middleware.GetDeptIDFromContext(c)

	// Parse pagination parameters with defaults
	limit, offset := parsePaginationParams(c)

	var referrals []models.Referral
	query := h.DB.Preload("Patient").Preload("Creator").Preload("Department").Preload("Attachments").
		Order("updated_at DESC")

	if userRole == models.RoleLevel2Doc {
		// Level 2 sees all their sent referrals regardless of status
		query = query.Where("creator_id = ?", userID)
	} else if userRole == models.RoleCHUDoc {
		// CHU doc sees all resolved referrals for their department
		if deptID == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "No department assigned"})
			return
		}
		query = query.Where("current_dept_id = ? AND status IN ?", *deptID, []string{
			string(models.StatusScheduled),
			string(models.StatusDenied),
			string(models.StatusCanceled),
			string(models.StatusRedirected),
		})
	}

	// Get total count for pagination metadata
	var total int64
	query.Model(&models.Referral{}).Count(&total)

	// Apply pagination
	if err := query.Limit(limit).Offset(offset).Find(&referrals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load history"})
		return
	}

	// Format response with decrypted fields using caching and logging
	var history []ReferralResponse
	for _, r := range referrals {
		// Use caching decrypt helper with logging
		patientCIN := h.decryptPatientField(r.Patient.ID, "cin", r.Patient.CIN, "[Decryption Error]")
		patientName := h.decryptPatientField(r.Patient.ID, "fullname", r.Patient.FullName, "[Decryption Error]")
		symptoms := h.decryptReferralField(r.ID, r.Patient.ID, "symptoms", r.Symptoms, "[Decryption Error]")

		res := ReferralResponse{
			ID:              r.ID,
			PatientCIN:      patientCIN,
			PatientName:     patientName,
			PatientDOB:      r.Patient.DateOfBirth.Format("2006-01-02"),
			PatientPhone:    r.Patient.PhoneNumber,
			CreatorUsername: r.Creator.Username,
			CreatorFacility: r.Creator.FacilityName,
			Department:      r.Department.Name,
			DepartmentID:    r.CurrentDeptID,
			Status:          r.Status,
			Urgency:         r.Urgency,
			Symptoms:        symptoms,
			AISuggestedDept: r.AISuggestedDept,
			AISummary:       r.AISummary,
			AppointmentDate: r.AppointmentDate,
			RejectionReason: r.RejectionReason,
			CreatedAt:       r.CreatedAt,
			UpdatedAt:       r.UpdatedAt,
		}

		for _, att := range r.Attachments {
			res.Attachments = append(res.Attachments, AttachmentResponse{
				ID:       att.ID,
				FileName: att.FileName,
				FileType: att.FileType,
				FileSize: att.FileSize,
			})
		}

		history = append(history, res)
	}

	// Build pagination metadata
	pagination := PaginationMeta{
		Limit:   limit,
		Offset:  offset,
		Total:   total,
		HasNext: int64(offset+limit) < total,
		HasPrev: offset > 0,
	}

	c.JSON(http.StatusOK, gin.H{
		"history":    history,
		"count":      len(history),
		"pagination": pagination,
	})
}

// ══════════════════════════════════════════════════════════════════════
// PATCH /api/referrals/:id/reschedule — Reschedule Appointment
// Role: CHU_DOC
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) RescheduleReferral(c *gin.Context) {
	referralID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid referral ID"})
		return
	}

	var req ScheduleReferralRequest // Reuse the scheduling DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appointmentDate, err := time.Parse(time.RFC3339, req.AppointmentDate)
	if err != nil || appointmentDate.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid future appointment date format"})
		return
	}

	deptID := middleware.GetDeptIDFromContext(c)

	var referral models.Referral
	if err := h.DB.Preload("Patient").Preload("Department").First(&referral, "id = ? AND current_dept_id = ?", referralID, *deptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Referral not found in your department"})
		return
	}

	// Update appointment date
	h.DB.Model(&referral).Updates(map[string]interface{}{
		"appointment_date": appointmentDate,
		"status":           models.StatusScheduled, // ensure it's scheduled
	})

	// Notify Level 2
	username := middleware.GetUsernameFromContext(c)
	h.Notification.Create(
		referral.CreatorID,
		referral.ID,
		fmt.Sprintf("📅 Appointment Rescheduled for %s at %s by Dr. %s.",
			appointmentDate.Format("02/01/2006 15:04"), referral.Department.Name, username),
	)

	// Async WhatsApp Update with decryption logging
	go func() {
		if h.WhatsApp != nil && h.AI != nil {
			// Use caching decrypt helper with logging
			patientName := h.decryptPatientField(referral.Patient.ID, "fullname", referral.Patient.FullName, "")
			symptoms := h.decryptReferralField(referral.ID, referral.Patient.ID, "symptoms", referral.Symptoms, "")
			if patientName == "" || symptoms == "" {
				log.Printf("[WHATSAPP ERROR] Failed to decrypt patient data for referral %s", referral.ID)
				return
			}
			msg, err := h.AI.GenerateWhatsAppMessage(patientName, symptoms, referral.Department.Name, appointmentDate)
			if err == nil {
				h.WhatsApp.SendTextMessage(referral.Patient.PhoneNumber, "[UPDATE/MODIFICATION] "+msg)
			} else {
				log.Println("[WHATSAPP] Failed to generate reschedule AI message:", err)
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Appointment rescheduled", "appointment_date": appointmentDate})
}

// ══════════════════════════════════════════════════════════════════════
// PATCH /api/referrals/:id/cancel — Cancel an Appointment
// Role: CHU_DOC
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) CancelReferral(c *gin.Context) {
	referralID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid referral ID"})
		return
	}

	deptID := middleware.GetDeptIDFromContext(c)
	username := middleware.GetUsernameFromContext(c)

	var referral models.Referral
	if err := h.DB.First(&referral, "id = ? AND current_dept_id = ?", referralID, *deptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Referral not found in your department"})
		return
	}

	h.DB.Model(&referral).Update("status", models.StatusCanceled)

	// Notify Level 2
	h.Notification.Create(
		referral.CreatorID,
		referral.ID,
		fmt.Sprintf("🚫 Appointment Canceled by CHU Dr. %s.", username),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Appointment canceled successfully", "status": models.StatusCanceled})
}
