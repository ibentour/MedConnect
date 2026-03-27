// Package repository provides data access layer implementations for MedConnect.
// This package follows the Repository pattern to separate business logic from data access.
package repository

import (
	"medconnect-oriental/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReferralRepository defines the interface for referral data operations.
type ReferralRepository interface {
	// Basic CRUD operations
	Create(referral *models.Referral) error
	GetByID(id uuid.UUID) (*models.Referral, error)
	Update(referral *models.Referral) error
	Delete(id uuid.UUID) error

	// Query methods
	FindByStatus(status models.ReferralStatus) ([]models.Referral, error)
	FindByCreatorID(creatorID uuid.UUID) ([]models.Referral, error)
	FindByDepartmentID(deptID uuid.UUID) ([]models.Referral, error)
	FindByPatientID(patientID uuid.UUID) ([]models.Referral, error)

	// Queue operations (pending + redirected for a department)
	FindQueueByDepartmentID(deptID uuid.UUID) ([]models.Referral, error)

	// Count methods
	CountByStatus(status models.ReferralStatus) (int64, error)
	CountByDepartmentID(deptID uuid.UUID) (int64, error)

	// Pagination and eager loading
	FindAll(limit, offset int) ([]models.Referral, error)
	FindWithPreload(conditions map[string]interface{}, preload []string) (*models.Referral, error)
}

// GormReferralRepository implements ReferralRepository using GORM.
type GormReferralRepository struct {
	db *gorm.DB
}

// NewReferralRepository creates a new ReferralRepository instance.
func NewReferralRepository(db *gorm.DB) *GormReferralRepository {
	return &GormReferralRepository{db: db}
}

// Create inserts a new referral record.
func (r *GormReferralRepository) Create(referral *models.Referral) error {
	return r.db.Create(referral).Error
}

// GetByID retrieves a referral by its UUID.
func (r *GormReferralRepository) GetByID(id uuid.UUID) (*models.Referral, error) {
	var referral models.Referral
	err := r.db.Preload("Patient").Preload("Creator").Preload("Department").Preload("Attachments").
		First(&referral, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &referral, nil
}

// Update updates an existing referral record.
func (r *GormReferralRepository) Update(referral *models.Referral) error {
	return r.db.Save(referral).Error
}

// Delete removes a referral by its UUID.
func (r *GormReferralRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Referral{}, "id = ?", id).Error
}

// FindByStatus retrieves all referrals with a specific status.
func (r *GormReferralRepository) FindByStatus(status models.ReferralStatus) ([]models.Referral, error) {
	var referrals []models.Referral
	err := r.db.Where("status = ?", status).
		Preload("Patient").Preload("Creator").Preload("Department").
		Order("created_at DESC").
		Find(&referrals).Error
	return referrals, err
}

// FindByCreatorID retrieves all referrals created by a specific user.
func (r *GormReferralRepository) FindByCreatorID(creatorID uuid.UUID) ([]models.Referral, error) {
	var referrals []models.Referral
	err := r.db.Where("creator_id = ?", creatorID).
		Preload("Patient").Preload("Department").
		Order("created_at DESC").
		Find(&referrals).Error
	return referrals, err
}

// FindByDepartmentID retrieves all referrals for a specific department.
func (r *GormReferralRepository) FindByDepartmentID(deptID uuid.UUID) ([]models.Referral, error) {
	var referrals []models.Referral
	err := r.db.Where("current_dept_id = ?", deptID).
		Preload("Patient").Preload("Creator").Preload("Attachments").
		Order("created_at DESC").
		Find(&referrals).Error
	return referrals, err
}

// FindByPatientID retrieves all referrals for a specific patient.
func (r *GormReferralRepository) FindByPatientID(patientID uuid.UUID) ([]models.Referral, error) {
	var referrals []models.Referral
	err := r.db.Where("patient_id = ?", patientID).
		Preload("Creator").Preload("Department").
		Order("created_at DESC").
		Find(&referrals).Error
	return referrals, err
}

// FindQueueByDepartmentID retrieves pending referrals for a department,
// ordered by urgency and creation time. Does NOT include SCHEDULED, REDIRECTED, DENIED or CANCELED referrals.
func (r *GormReferralRepository) FindQueueByDepartmentID(deptID uuid.UUID) ([]models.Referral, error) {
	var referrals []models.Referral
	err := r.db.Where("current_dept_id = ? AND status = ?", deptID, string(models.StatusPending)).
		Preload("Patient").
		Preload("Creator").
		Preload("Attachments").
		Order("CASE urgency WHEN 'CRITICAL' THEN 1 WHEN 'HIGH' THEN 2 WHEN 'MEDIUM' THEN 3 WHEN 'LOW' THEN 4 END, created_at ASC").
		Find(&referrals).Error
	return referrals, err
}

// CountByStatus returns the count of referrals with a specific status.
func (r *GormReferralRepository) CountByStatus(status models.ReferralStatus) (int64, error) {
	var count int64
	err := r.db.Model(&models.Referral{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountByDepartmentID returns the count of referrals for a specific department.
func (r *GormReferralRepository) CountByDepartmentID(deptID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Referral{}).Where("current_dept_id = ?", deptID).Count(&count).Error
	return count, err
}

// FindAll retrieves all referrals with pagination.
func (r *GormReferralRepository) FindAll(limit, offset int) ([]models.Referral, error) {
	var referrals []models.Referral
	err := r.db.Preload("Patient").Preload("Creator").Preload("Department").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&referrals).Error
	return referrals, err
}

// FindWithPreload retrieves a referral with specific conditions and eager loads associations.
func (r *GormReferralRepository) FindWithPreload(conditions map[string]interface{}, preload []string) (*models.Referral, error) {
	var referral models.Referral
	query := r.db

	// Apply preloads
	for _, p := range preload {
		query = query.Preload(p)
	}

	// Apply conditions
	query = query.Where(conditions)

	err := query.First(&referral).Error
	if err != nil {
		return nil, err
	}
	return &referral, nil
}
