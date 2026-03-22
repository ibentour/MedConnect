package api

import (

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"medconnect-oriental/backend/internal/models"
)

// ══════════════════════════════════════════════════════════════════════
// ADMIN: Users Management
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) GetUsers(c *gin.Context) {
	var users []models.User
	if err := h.DB.Preload("Department").Order("created_at DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *HandlerContext) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username already exists
	var count int64
	h.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         req.Role,
		FacilityName: req.FacilityName,
		IsActive:     true,
	}

	if req.DepartmentID != nil && *req.DepartmentID != "" {
		deptID, err := uuid.Parse(*req.DepartmentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
			return
		}
		user.DeptID = &deptID
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created", "user": user})
}

func (h *HandlerContext) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.DB.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// ══════════════════════════════════════════════════════════════════════
// ADMIN: Departments Management
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) GetAdminDepartments(c *gin.Context) {
	var depts []models.Department
	if err := h.DB.Order("name ASC").Find(&depts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load departments"})
		return
	}

	// Calculate stats for each department manually
	type DeptStats struct {
		models.Department
		TotalReferrals     int64         `json:"total_referrals"`
		PendingReferrals   int64         `json:"pending_referrals"`
		ScheduledReferrals int64         `json:"scheduled_referrals"`
		LowUrgency         int64         `json:"low_urgency"`
		MediumUrgency      int64         `json:"medium_urgency"`
		HighUrgency        int64         `json:"high_urgency"`
		CriticalUrgency    int64         `json:"critical_urgency"`
		Doctors            []models.User `json:"doctors"`
	}

	var stats []DeptStats
	for _, d := range depts {
		var total, pending, scheduled, low, med, high, crit int64
		h.DB.Model(&models.Referral{}).Where("current_dept_id = ?", d.ID).Count(&total)
		h.DB.Model(&models.Referral{}).Where("current_dept_id = ? AND status IN ?", d.ID, []string{"PENDING", "REDIRECTED"}).Count(&pending)
		h.DB.Model(&models.Referral{}).Where("current_dept_id = ? AND status = ?", d.ID, "SCHEDULED").Count(&scheduled)
		
		h.DB.Model(&models.Referral{}).Where("current_dept_id = ? AND urgency = ?", d.ID, "LOW").Count(&low)
		h.DB.Model(&models.Referral{}).Where("current_dept_id = ? AND urgency = ?", d.ID, "MEDIUM").Count(&med)
		h.DB.Model(&models.Referral{}).Where("current_dept_id = ? AND urgency = ?", d.ID, "HIGH").Count(&high)
		h.DB.Model(&models.Referral{}).Where("current_dept_id = ? AND urgency = ?", d.ID, "CRITICAL").Count(&crit)

		var docs []models.User
		h.DB.Where("dept_id = ? AND role = ?", d.ID, models.RoleCHUDoc).Find(&docs)
		
		stats = append(stats, DeptStats{
			Department:         d,
			TotalReferrals:     total,
			PendingReferrals:   pending,
			ScheduledReferrals: scheduled,
			LowUrgency:         low,
			MediumUrgency:      med,
			HighUrgency:        high,
			CriticalUrgency:    crit,
			Doctors:            docs,
		})
	}

	c.JSON(http.StatusOK, stats)
}

func (h *HandlerContext) CreateDepartment(c *gin.Context) {
	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dept := models.Department{
		Name:           req.Name,
		PhoneExtension: req.PhoneExtension,
		WorkHours:      req.WorkHours,
		WorkDays:       req.WorkDays,
		IsAccepting:    true,
	}

	if err := h.DB.Create(&dept).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create department"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Department created", "department": dept})
}

func (h *HandlerContext) UpdateDepartment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
		return
	}

	var req UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var dept models.Department
	if err := h.DB.First(&dept, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.PhoneExtension != nil {
		updates["phone_extension"] = *req.PhoneExtension
	}
	if req.WorkHours != nil {
		updates["work_hours"] = *req.WorkHours
	}
	if req.WorkDays != nil {
		updates["work_days"] = *req.WorkDays
	}
	if req.IsAccepting != nil {
		updates["is_accepting"] = *req.IsAccepting
	}

	if err := h.DB.Model(&dept).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update department"})
		return
	}

	// Refetch to return populated object
	h.DB.First(&dept, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"message": "Department updated", "department": dept})
}

func (h *HandlerContext) DeleteDepartment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.DB.Delete(&models.Department{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete department"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Department deleted"})
}

// ══════════════════════════════════════════════════════════════════════
// ADMIN: Super Admin Dashboard Stats
// ══════════════════════════════════════════════════════════════════════

func (h *HandlerContext) GetAdminStats(c *gin.Context) {
	var totalUsers, totalDepts, totalReferrals, pendingReferrals int64

	h.DB.Model(&models.User{}).Count(&totalUsers)
	h.DB.Model(&models.Department{}).Count(&totalDepts)
	h.DB.Model(&models.Referral{}).Count(&totalReferrals)
	h.DB.Model(&models.Referral{}).Where("status IN ('PENDING', 'REDIRECTED')").Count(&pendingReferrals)

	c.JSON(http.StatusOK, gin.H{
		"total_users":       totalUsers,
		"total_departments": totalDepts,
		"total_referrals":   totalReferrals,
		"pending_referrals": pendingReferrals,
	})
}

// GetAnalystDoctorStats returns a list of all Level 2 Doctors and their referral counts/destinations.
func (h *HandlerContext) GetAnalystDoctorStats(c *gin.Context) {
	var doctors []models.User
	if err := h.DB.Where("role = ?", models.RoleLevel2Doc).Find(&doctors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load doctors"})
		return
	}

	type DestStat struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}

	type DocStats struct {
		ID             uuid.UUID  `json:"id"`
		Username       string     `json:"username"`
		FacilityName   string     `json:"facility_name"`
		TotalReferrals int64      `json:"total_referrals"`
		ByDepartment   []DestStat `json:"by_department"`
	}

	var stats []DocStats
	for _, d := range doctors {
		var total int64
		h.DB.Model(&models.Referral{}).Where("creator_id = ?", d.ID).Count(&total)

		var depts []DestStat
		h.DB.Table("referrals").
			Select("departments.name as name, count(*) as count").
			Joins("join departments on departments.id = referrals.current_dept_id").
			Where("referrals.creator_id = ?", d.ID).
			Group("departments.name").
			Scan(&depts)

		stats = append(stats, DocStats{
			ID:             d.ID,
			Username:       d.Username,
			FacilityName:   d.FacilityName,
			TotalReferrals: total,
			ByDepartment:   depts,
		})
	}
	c.JSON(http.StatusOK, stats)
}
