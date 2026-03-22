// Package api provides HTTP handlers for the MedConnect Oriental API.
// All handlers receive dependencies via the HandlerContext struct.
package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"medconnect-oriental/backend/internal/ai"
	"medconnect-oriental/backend/internal/crypto"
	"medconnect-oriental/backend/internal/middleware"
	"medconnect-oriental/backend/internal/models"
	"medconnect-oriental/backend/internal/service"
)

// HandlerContext holds shared dependencies for all API handlers.
type HandlerContext struct {
	DB           *gorm.DB
	Crypto       *crypto.AESCrypto
	AI           *ai.Service
	WhatsApp     *service.WhatsAppService
	Notification *service.NotificationService
}

// ══════════════════════════════════════════════════════════════════════
// GET /api/directory — List CHU Departments
// Role: ALL authenticated users
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) GetDirectory(c *gin.Context) {
	var departments []models.Department
	if err := h.DB.Order("name ASC").Find(&departments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load departments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"departments": departments,
		"count":       len(departments),
	})
}

// ══════════════════════════════════════════════════════════════════════
// POST /api/referrals — Create a New Referral
// Role: LEVEL_2_DOC
// Encrypts: patient CIN, name, symptoms
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) CreateReferral(c *gin.Context) {
	var req CreateReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		details := middleware.ParseValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": details,
		})
		return
	}

	// Sanitize input strings
	req.PatientCIN = middleware.SanitizeInput(req.PatientCIN)
	req.PatientName = middleware.SanitizeInput(req.PatientName)
	req.Symptoms = middleware.SanitizeInput(req.Symptoms)
	req.PatientPhone = middleware.SanitizeInput(req.PatientPhone)

	userID, _ := middleware.GetUserIDFromContext(c)

	// Parse department UUID
	deptID, err := uuid.Parse(req.DepartmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
		return
	}

	// Verify department exists and is accepting
	var dept models.Department
	if err := h.DB.First(&dept, "id = ?", deptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
		return
	}
	if !dept.IsAccepting {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Department '%s' is not currently accepting referrals", dept.Name)})
		return
	}

	// ── Encrypt sensitive fields ─────────────────────────────
	encryptedCIN, err := h.Crypto.Encrypt(req.PatientCIN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption failed"})
		return
	}

	encryptedName, err := h.Crypto.Encrypt(req.PatientName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption failed"})
		return
	}

	encryptedSymptoms, err := h.Crypto.Encrypt(req.Symptoms)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption failed"})
		return
	}

	dob, err := time.Parse("2006-01-02", req.PatientDOB)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date of birth format. Use YYYY-MM-DD"})
		return
	}

	// ── Create patient record ────────────────────────────────
	patient := models.Patient{
		CIN:         encryptedCIN,
		FullName:    encryptedName,
		DateOfBirth: dob,
		PhoneNumber: req.PatientPhone,
	}

	if err := h.DB.Create(&patient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create patient record"})
		return
	}

	// ── Create referral record ───────────────────────────────
	referral := models.Referral{
		PatientID:       patient.ID,
		CreatorID:       userID,
		CurrentDeptID:   deptID,
		Status:          models.StatusPending,
		Urgency:         req.Urgency,
		Symptoms:        encryptedSymptoms,
		AISuggestedDept: req.AISuggestedDept,
		AISummary:       nil, // Will be updated asynchronously
	}

	if err := h.DB.Create(&referral).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create referral"})
		return
	}

	// ── Generate AI summary (async, non-blocking) ────────────
	if h.AI != nil {
		go func(refID uuid.UUID, symptoms string) {
			summary, err := h.AI.SummarizeSymptoms(symptoms)
			if err != nil {
				log.Printf("[AI ERROR] Summarization failed for referral %s: %v", refID, err)
				return
			}
			// Update the referral with the summary
			if err := h.DB.Model(&models.Referral{}).Where("id = ?", refID).Update("ai_summary", summary).Error; err != nil {
				log.Printf("[DATABASE ERROR] Failed to update AI summary for referral %s: %v", refID, err)
			}
		}(referral.ID, req.Symptoms)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Referral created successfully",
		"referral_id": referral.ID,
		"patient_id":  patient.ID,
		"department":  dept.Name,
		"status":      referral.Status,
	})
}

// ══════════════════════════════════════════════════════════════════════
// POST /api/referrals/suggest — AI Department Suggestion
// Role: LEVEL_2_DOC
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) SuggestDepartment(c *gin.Context) {
	var req SuggestDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.AI == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service is not available"})
		return
	}

	// Fetch all accepting departments
	var departments []models.Department
	h.DB.Where("is_accepting = ?", true).Find(&departments)

	deptNames := make([]string, len(departments))
	for i, d := range departments {
		deptNames[i] = d.Name
	}

	dob, _ := time.Parse("2006-01-02", req.PatientDOB)
	age := time.Now().Year() - dob.Year()

	suggestion, err := h.AI.SuggestDepartment(req.Symptoms, age, deptNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI triage failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuggestDepartmentResponse{
		SuggestedDepartment: suggestion.Department,
		Urgency:             suggestion.Urgency,
		Confidence:          suggestion.Confidence,
		Reasoning:           suggestion.Reasoning,
	})
}

// ══════════════════════════════════════════════════════════════════════
// GET /api/queue — CHU Doctor Triage Queue
// Role: CHU_DOC
// Returns only referrals for the doctor's assigned department
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) GetQueue(c *gin.Context) {
	deptID := middleware.GetDeptIDFromContext(c)
	if deptID == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No department assigned to your account"})
		return
	}

	var referrals []models.Referral
	err := h.DB.Where("current_dept_id = ? AND status IN ?", *deptID, []string{
		string(models.StatusPending),
		string(models.StatusRedirected),
	}).
		Preload("Patient").
		Preload("Creator").
		Preload("Attachments").
		Order("CASE urgency WHEN 'CRITICAL' THEN 1 WHEN 'HIGH' THEN 2 WHEN 'MEDIUM' THEN 3 WHEN 'LOW' THEN 4 END, created_at ASC").
		Find(&referrals).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load queue"})
		return
	}

	// Build queue items with decrypted names
	queue := make([]QueueItem, 0, len(referrals))
	for _, r := range referrals {
		patientName, err := h.Crypto.Decrypt(r.Patient.FullName)
		if err != nil {
			patientName = "[Decryption Error]"
		}

		queue = append(queue, QueueItem{
			ID:              r.ID,
			PatientName:     patientName,
			PatientDOB:      r.Patient.DateOfBirth.Format("2006-01-02"),
			Urgency:         r.Urgency,
			AISummary:       r.AISummary,
			Status:          r.Status,
			CreatorFacility: r.Creator.FacilityName,
			CreatedAt:       r.CreatedAt,
			HasAttachments:  len(r.Attachments) > 0,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"queue": queue,
		"count": len(queue),
	})
}

// ══════════════════════════════════════════════════════════════════════
// GET /api/referrals/:id — Get Single Referral (Decrypted)
// Role: CHU_DOC (own dept), LEVEL_2_DOC (own referrals)
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) GetReferral(c *gin.Context) {
	referralID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid referral ID"})
		return
	}

	userID, _ := middleware.GetUserIDFromContext(c)
	userRole := middleware.GetUserRoleFromContext(c)
	deptID := middleware.GetDeptIDFromContext(c)

	var referral models.Referral
	err = h.DB.Preload("Patient").Preload("Creator").Preload("Department").Preload("Attachments").
		First(&referral, "id = ?", referralID).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Referral not found"})
		return
	}

	// ── Access control ─────────────────────────────────
	switch userRole {
	case models.RoleLevel2Doc:
		if referral.CreatorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only view your own referrals"})
			return
		}
	case models.RoleCHUDoc:
		if deptID == nil || referral.CurrentDeptID != *deptID {
			c.JSON(http.StatusForbidden, gin.H{"error": "This referral is not in your department"})
			return
		}
	case models.RoleSuperAdmin, models.RoleAnalyst:
		// Full access
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// ── Decrypt sensitive fields ─────────────────────────
	patientCIN, _ := h.Crypto.Decrypt(referral.Patient.CIN)
	patientName, _ := h.Crypto.Decrypt(referral.Patient.FullName)
	symptoms, _ := h.Crypto.Decrypt(referral.Symptoms)

	response := ReferralResponse{
		ID:              referral.ID,
		PatientCIN:      patientCIN,
		PatientName:     patientName,
		PatientDOB:      referral.Patient.DateOfBirth.Format("2006-01-02"),
		PatientPhone:    referral.Patient.PhoneNumber,
		CreatorUsername: referral.Creator.Username,
		CreatorFacility: referral.Creator.FacilityName,
		Department:      referral.Department.Name,
		DepartmentID:    referral.CurrentDeptID,
		Status:          referral.Status,
		Urgency:         referral.Urgency,
		Symptoms:        symptoms,
		AISuggestedDept: referral.AISuggestedDept,
		AISummary:       referral.AISummary,
		AppointmentDate: referral.AppointmentDate,
		RejectionReason: referral.RejectionReason,
		CreatedAt:       referral.CreatedAt,
		UpdatedAt:       referral.UpdatedAt,
	}

	for _, att := range referral.Attachments {
		response.Attachments = append(response.Attachments, AttachmentResponse{
			ID:       att.ID,
			FileName: att.FileName,
			FileType: att.FileType,
			FileSize: att.FileSize,
		})
	}

	c.JSON(http.StatusOK, response)
}

// ══════════════════════════════════════════════════════════════════════
// POST /api/referrals/:id/attachments — Upload Attachments
// Role: LEVEL_2_DOC (only own)
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) UploadAttachments(c *gin.Context) {
	// File upload security constants
	const (
		maxFileSize      = int64(10 * 1024 * 1024) // 10MB per file
		maxTotalSize     = int64(50 * 1024 * 1024) // 50MB total per request
		maxFilesPerBatch = 10
	)

	// Allowed MIME types for medical document uploads
	allowedMimeTypes := map[string]bool{
		"application/pdf":    true, // PDF documents
		"image/jpeg":         true, // JPEG images
		"image/png":          true, // PNG images
		"image/gif":          true, // GIF images
		"application/msword": true, // DOC files
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // DOCX files
	}

	referralID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid referral ID"})
		return
	}

	userID, _ := middleware.GetUserIDFromContext(c)

	var referral models.Referral
	if err := h.DB.First(&referral, "id = ?", referralID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Referral not found"})
		return
	}

	if referral.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only attach files to your own referrals"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
		return
	}

	files := form.File["attachments"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	// Validate number of files
	if len(files) > maxFilesPerBatch {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Too many files. Maximum %d files per batch", maxFilesPerBatch),
		})
		return
	}

	// Validate total size and individual file sizes
	var totalSize int64
	for _, file := range files {
		// Check individual file size
		if file.Size > maxFileSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("File %q exceeds maximum size of 10MB", file.Filename),
			})
			return
		}

		// Check MIME type
		contentType := file.Header.Get("Content-Type")
		if !allowedMimeTypes[contentType] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("File type %q is not allowed. Allowed types: PDF, JPEG, PNG, GIF, DOC, DOCX", contentType),
			})
			return
		}

		totalSize += file.Size
	}

	// Check total size
	if totalSize > maxTotalSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Total upload size exceeds maximum of 50MB",
		})
		return
	}

	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	uploadedFiles := []string{}
	for _, file := range files {
		fileID := uuid.New()
		// Save with UUID prefix to prevent collisions and path traversal
		safeFileName := filepath.Base(file.Filename) // Sanitize filename
		fileName := fmt.Sprintf("%s_%s", fileID.String(), safeFileName)
		filePath := filepath.Join(uploadDir, fileName)

		if err := c.SaveUploadedFile(file, filePath); err != nil {
			// Clean up any already uploaded files on failure
			for _, uploaded := range uploadedFiles {
				os.Remove(uploaded)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		uploadedFiles = append(uploadedFiles, filePath)

		attachment := models.Attachment{
			ID:         fileID,
			ReferralID: referralID,
			FilePath:   filePath,
			FileName:   safeFileName,
			FileType:   file.Header.Get("Content-Type"),
			FileSize:   file.Size,
		}

		if err := h.DB.Create(&attachment).Error; err != nil {
			// Clean up uploaded files on database failure
			for _, uploaded := range uploadedFiles {
				os.Remove(uploaded)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record attachment in database"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Attachments uploaded successfully"})
}

// ══════════════════════════════════════════════════════════════════════
// GET /api/attachments/:id — Download/View Attachment
// Role: ALL authenticated (with referral access)
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) GetAttachment(c *gin.Context) {
	attachmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attachment ID"})
		return
	}

	userID, _ := middleware.GetUserIDFromContext(c)
	userRole := middleware.GetUserRoleFromContext(c)
	userDeptID := middleware.GetDeptIDFromContext(c)

	var att models.Attachment
	if err := h.DB.First(&att, "id = ?", attachmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attachment not found"})
		return
	}

	// Authorization check: User must have access to the referral
	var referral models.Referral
	if err := h.DB.First(&referral, "id = ?", att.ReferralID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Associated referral not found"})
		return
	}

	// Check access based on role:
	// - SUPER_ADMIN, ANALYST: Full access to all attachments
	// - CHU_DOC: Access to attachments in their department's referrals
	// - LEVEL_2_DOC: Access only to attachments of referrals they created
	hasAccess := false
	switch userRole {
	case models.RoleSuperAdmin, models.RoleAnalyst:
		hasAccess = true
	case models.RoleCHUDoc:
		// CHU docs can access attachments for referrals in their department
		if userDeptID != nil {
			hasAccess = referral.CurrentDeptID == *userDeptID
		}
	case models.RoleLevel2Doc:
		// Level 2 docs can only access attachments of referrals they created
		hasAccess = referral.CreatorID == userID
	}

	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this attachment"})
		return
	}

	c.File(att.FilePath)
}

// ══════════════════════════════════════════════════════════════════════
// PATCH /api/referrals/:id/schedule — Schedule Appointment
// Role: CHU_DOC
// Triggers: AI WhatsApp message → Evolution API
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) ScheduleReferral(c *gin.Context) {
	referralID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid referral ID"})
		return
	}

	var req ScheduleReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appointmentDate, err := time.Parse(time.RFC3339, req.AppointmentDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use ISO 8601 (e.g., 2026-04-15T10:00:00Z)"})
		return
	}

	if appointmentDate.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Appointment date must be in the future"})
		return
	}

	deptID := middleware.GetDeptIDFromContext(c)

	var referral models.Referral
	err = h.DB.Preload("Patient").Preload("Department").
		First(&referral, "id = ? AND current_dept_id = ?", referralID, *deptID).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Referral not found in your department"})
		return
	}

	if referral.Status != models.StatusPending && referral.Status != models.StatusRedirected {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Cannot schedule: referral status is '%s'", referral.Status)})
		return
	}

	// ── Update referral ────────────────────────────────────
	h.DB.Model(&referral).Updates(map[string]interface{}{
		"status":           models.StatusScheduled,
		"appointment_date": appointmentDate,
	})

	// ── Trigger WhatsApp notification (Goroutine) ──────────
	go func(ref models.Referral, apptDate time.Time) {
		// Decrypt patient data for the AI message
		patientName, err := h.Crypto.Decrypt(ref.Patient.FullName)
		if err != nil {
			log.Printf("[WHATSAPP ERROR] Failed to decrypt patient name: %v", err)
			return
		}

		symptoms, err := h.Crypto.Decrypt(ref.Symptoms)
		if err != nil {
			log.Printf("[WHATSAPP ERROR] Failed to decrypt symptoms: %v", err)
			return
		}

		// Generate AI message
		if h.AI != nil {
			message, err := h.AI.GenerateWhatsAppMessage(
				patientName,
				symptoms,
				ref.Department.Name,
				apptDate,
			)
			if err != nil {
				log.Printf("[WHATSAPP ERROR] AI message generation failed: %v", err)
				return
			}

			// Send via Evolution API
			if h.WhatsApp != nil {
				msgID, err := h.WhatsApp.SendTextMessage(ref.Patient.PhoneNumber, message)
				if err != nil {
					log.Printf("[WHATSAPP ERROR] Failed to send message: %v", err)
				} else {
					log.Printf("[WHATSAPP] Message sent to %s (ID: %s)", ref.Patient.PhoneNumber, msgID)
				}
			}
		}
	}(referral, appointmentDate)

	// ── Notify Level 2 doctor ──────────────────────────────
	h.Notification.Create(
		referral.CreatorID,
		referral.ID,
		fmt.Sprintf("✅ Your referral has been scheduled for %s at %s",
			appointmentDate.Format("02/01/2006 15:04"), referral.Department.Name),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":          "Referral scheduled successfully",
		"referral_id":      referral.ID,
		"appointment_date": appointmentDate,
		"whatsapp_status":  "sending",
	})
}

// ══════════════════════════════════════════════════════════════════════
// PATCH /api/referrals/:id/redirect — Redirect to Another Department
// Role: CHU_DOC
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) RedirectReferral(c *gin.Context) {
	referralID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid referral ID"})
		return
	}

	var req RedirectReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newDeptID, err := uuid.Parse(req.NewDepartmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid new department ID"})
		return
	}

	deptID := middleware.GetDeptIDFromContext(c)
	username := middleware.GetUsernameFromContext(c)

	// Verify the referral exists and belongs to this CHU doc's dept
	var referral models.Referral
	err = h.DB.Preload("Department").
		First(&referral, "id = ? AND current_dept_id = ?", referralID, *deptID).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Referral not found in your department"})
		return
	}

	// Verify the new department exists
	var newDept models.Department
	if err := h.DB.First(&newDept, "id = ?", newDeptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target department not found"})
		return
	}

	// Cannot redirect to same department
	if newDeptID == *deptID {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot redirect to the same department"})
		return
	}

	oldDeptName := referral.Department.Name

	// ── Update referral ────────────────────────────────────
	h.DB.Model(&referral).Updates(map[string]interface{}{
		"current_dept_id": newDeptID,
		"status":          models.StatusRedirected,
	})

	// ── Notify Level 2 doctor ──────────────────────────────
	h.Notification.Create(
		referral.CreatorID,
		referral.ID,
		fmt.Sprintf("🔄 Your referral has been redirected from %s to %s by Dr. %s. Reason: %s",
			oldDeptName, newDept.Name, username, req.Reason),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Referral redirected successfully",
		"referral_id": referral.ID,
		"from_dept":   oldDeptName,
		"to_dept":     newDept.Name,
		"new_status":  models.StatusRedirected,
	})
}

// ══════════════════════════════════════════════════════════════════════
// PATCH /api/referrals/:id/deny — Deny a Referral
// Role: CHU_DOC (reason mandatory)
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) DenyReferral(c *gin.Context) {
	referralID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid referral ID"})
		return
	}

	var req DenyReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deptID := middleware.GetDeptIDFromContext(c)
	username := middleware.GetUsernameFromContext(c)

	var referral models.Referral
	err = h.DB.Preload("Department").
		First(&referral, "id = ? AND current_dept_id = ?", referralID, *deptID).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Referral not found in your department"})
		return
	}

	if referral.Status == models.StatusDenied {
		c.JSON(http.StatusConflict, gin.H{"error": "Referral is already denied"})
		return
	}

	// ── Update referral ────────────────────────────────────
	h.DB.Model(&referral).Updates(map[string]interface{}{
		"status":           models.StatusDenied,
		"rejection_reason": req.Reason,
	})

	// ── Notify Level 2 doctor ──────────────────────────────
	h.Notification.Create(
		referral.CreatorID,
		referral.ID,
		fmt.Sprintf("❌ Your referral to %s has been denied by Dr. %s. Reason: %s",
			referral.Department.Name, username, req.Reason),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Referral denied",
		"referral_id": referral.ID,
		"reason":      req.Reason,
	})
}

// ══════════════════════════════════════════════════════════════════════
// GET /api/notifications — Level 2 Doctor Notifications
// Role: LEVEL_2_DOC
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) GetNotifications(c *gin.Context) {
	userID, _ := middleware.GetUserIDFromContext(c)

	notifications, err := h.Notification.GetForUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load notifications"})
		return
	}

	unreadCount, _ := h.Notification.GetUnreadCount(userID)

	items := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		items[i] = NotificationResponse{
			ID:         n.ID,
			ReferralID: n.ReferralID,
			Message:    n.Message,
			IsRead:     n.IsRead,
			CreatedAt:  n.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": items,
		"unread_count":  unreadCount,
	})
}

// ══════════════════════════════════════════════════════════════════════
// PATCH /api/notifications/:id/read — Mark Notification as Read
// Role: ALL authenticated
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) MarkNotificationRead(c *gin.Context) {
	notifID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userID, _ := middleware.GetUserIDFromContext(c)

	if err := h.Notification.MarkAsRead(notifID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}
