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
		})
	}

	if err := query.Find(&referrals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load history"})
		return
	}

	// Format response
	var history []ReferralResponse
	for _, r := range referrals {
		patientCIN, _ := h.Crypto.Decrypt(r.Patient.CIN)
		patientName, _ := h.Crypto.Decrypt(r.Patient.FullName)
		symptoms, _ := h.Crypto.Decrypt(r.Symptoms)

		res := ReferralResponse{
			ID:              r.ID,
			PatientCIN:      patientCIN,
			PatientName:     patientName,
			PatientDOB:      r.Patient.DateOfBirth.Format("2006-01-02"),
			PatientPhone:    r.Patient.PhoneNumber,
			CreatorUsername:  r.Creator.Username,
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

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history),
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

	// Async WhatsApp Update
	go func() {
		if h.WhatsApp != nil && h.AI != nil {
			patientName, _ := h.Crypto.Decrypt(referral.Patient.FullName)
			symptoms, _ := h.Crypto.Decrypt(referral.Symptoms)
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
