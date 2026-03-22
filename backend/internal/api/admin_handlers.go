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

// DeptStats holds department statistics for API response
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

// ReferralStatsRow represents a row from the aggregated referral stats query
type ReferralStatsRow struct {
	DeptID          uuid.UUID `gorm:"column:dept_id"`
	Total           int64     `gorm:"column:total"`
	Pending         int64     `gorm:"column:pending"`
	Scheduled       int64     `gorm:"column:scheduled"`
	LowUrgency      int64     `gorm:"column:low_urgency"`
	MediumUrgency   int64     `gorm:"column:medium_urgency"`
	HighUrgency     int64     `gorm:"column:high_urgency"`
	CriticalUrgency int64     `gorm:"column:critical_urgency"`
}

func (h *HandlerContext) GetAdminDepartments(c *gin.Context) {
	var depts []models.Department
	if err := h.DB.Order("name ASC").Find(&depts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load departments"})
		return
	}

	// Single aggregated query for all referral statistics per department
	// This replaces N+1 queries (8 queries per department) with just 1 query
	var referralStats []ReferralStatsRow
	err := h.DB.Raw(`
		SELECT 
			current_dept_id as dept_id,
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status IN ('PENDING', 'REDIRECTED')) as pending,
			COUNT(*) FILTER (WHERE status = 'SCHEDULED') as scheduled,
			COUNT(*) FILTER (WHERE urgency = 'LOW') as low_urgency,
			COUNT(*) FILTER (WHERE urgency = 'MEDIUM') as medium_urgency,
			COUNT(*) FILTER (WHERE urgency = 'HIGH') as high_urgency,
			COUNT(*) FILTER (WHERE urgency = 'CRITICAL') as critical_urgency
		FROM referrals
		WHERE current_dept_id IS NOT NULL
		GROUP BY current_dept_id
	`).Scan(&referralStats).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load referral statistics"})
		return
	}

	// Create a map for quick lookup of stats by department ID
	statsMap := make(map[uuid.UUID]ReferralStatsRow)
	for _, stat := range referralStats {
		statsMap[stat.DeptID] = stat
	}

	// Single query to get all doctors grouped by department
	type DoctorRow struct {
		ID       uuid.UUID `gorm:"column:id"`
		Username string    `gorm:"column:username"`
		DeptID   uuid.UUID `gorm:"column:dept_id"`
		Role     string    `gorm:"column:role"`
	}
	var doctors []DoctorRow
	err = h.DB.Table("users").
		Select("id, username, dept_id, role").
		Where("dept_id IS NOT NULL AND role = ?", models.RoleCHUDoc).
		Find(&doctors).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load doctors"})
		return
	}

	// Group doctors by department ID
	doctorsMap := make(map[uuid.UUID][]models.User)
	for _, doc := range doctors {
		doctorsMap[doc.DeptID] = append(doctorsMap[doc.DeptID], models.User{
			ID:       doc.ID,
			Username: doc.Username,
			Role:     models.Role(doc.Role),
			DeptID:   &doc.DeptID,
		})
	}

	// Combine all data into response
	var stats []DeptStats
	for _, d := range depts {
		stat := DeptStats{
			Department: d,
		}

		// Add referral stats if available
		if refStat, ok := statsMap[d.ID]; ok {
			stat.TotalReferrals = refStat.Total
			stat.PendingReferrals = refStat.Pending
			stat.ScheduledReferrals = refStat.Scheduled
			stat.LowUrgency = refStat.LowUrgency
			stat.MediumUrgency = refStat.MediumUrgency
			stat.HighUrgency = refStat.HighUrgency
			stat.CriticalUrgency = refStat.CriticalUrgency
		}

		// Add doctors if available
		if docs, ok := doctorsMap[d.ID]; ok {
			stat.Doctors = docs
		} else {
			stat.Doctors = []models.User{}
		}

		stats = append(stats, stat)
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
