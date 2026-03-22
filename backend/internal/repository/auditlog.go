// Package repository provides data access layer implementations for MedConnect.
// This package follows the Repository pattern to separate business logic from data access.
package repository

import (
	"medconnect-oriental/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLogRepository defines the interface for audit log data operations.
type AuditLogRepository interface {
	// Basic CRUD operations
	Create(log *models.AuditLog) error
	GetByID(id uuid.UUID) (*models.AuditLog, error)

	// Query methods
	FindByUserID(userID uuid.UUID) ([]models.AuditLog, error)
	FindByAction(action string) ([]models.AuditLog, error)
	FindByTargetID(targetID string) ([]models.AuditLog, error)
	FindByIPAddress(ipAddress string) ([]models.AuditLog, error)
	FindByStatus(status int) ([]models.AuditLog, error)
	FindByDateRange(start, end interface{}) ([]models.AuditLog, error)

	// Count methods
	Count() (int64, error)
	CountByUserID(userID uuid.UUID) (int64, error)
	CountByAction(action string) (int64, error)

	// Pagination and eager loading
	FindAll(limit, offset int) ([]models.AuditLog, error)
	FindWithConditions(conditions map[string]interface{}, limit, offset int) ([]models.AuditLog, error)
}

// GormAuditLogRepository implements AuditLogRepository using GORM.
type GormAuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new AuditLogRepository instance.
func NewAuditLogRepository(db *gorm.DB) *GormAuditLogRepository {
	return &GormAuditLogRepository{db: db}
}

// Create inserts a new audit log record.
func (r *GormAuditLogRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

// GetByID retrieves an audit log by its UUID.
func (r *GormAuditLogRepository) GetByID(id uuid.UUID) (*models.AuditLog, error) {
	var log models.AuditLog
	err := r.db.First(&log, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// FindByUserID retrieves all audit logs for a specific user.
func (r *GormAuditLogRepository) FindByUserID(userID uuid.UUID) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where("user_id = ?", userID).Order("timestamp DESC").Find(&logs).Error
	return logs, err
}

// FindByAction retrieves all audit logs for a specific action.
func (r *GormAuditLogRepository) FindByAction(action string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where("action = ?", action).Order("timestamp DESC").Find(&logs).Error
	return logs, err
}

// FindByTargetID retrieves all audit logs for a specific target resource.
func (r *GormAuditLogRepository) FindByTargetID(targetID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where("target_id = ?", targetID).Order("timestamp DESC").Find(&logs).Error
	return logs, err
}

// FindByIPAddress retrieves all audit logs from a specific IP address.
func (r *GormAuditLogRepository) FindByIPAddress(ipAddress string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where("ip_address = ?", ipAddress).Order("timestamp DESC").Find(&logs).Error
	return logs, err
}

// FindByStatus retrieves all audit logs with a specific HTTP status code.
func (r *GormAuditLogRepository) FindByStatus(status int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where("status = ?", status).Order("timestamp DESC").Find(&logs).Error
	return logs, err
}

// FindByDateRange retrieves all audit logs within a specific date range.
func (r *GormAuditLogRepository) FindByDateRange(start, end interface{}) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where("timestamp BETWEEN ? AND ?", start, end).Order("timestamp DESC").Find(&logs).Error
	return logs, err
}

// Count returns the total number of audit logs.
func (r *GormAuditLogRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.AuditLog{}).Count(&count).Error
	return count, err
}

// CountByUserID returns the number of audit logs for a specific user.
func (r *GormAuditLogRepository) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.AuditLog{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountByAction returns the number of audit logs for a specific action.
func (r *GormAuditLogRepository) CountByAction(action string) (int64, error) {
	var count int64
	err := r.db.Model(&models.AuditLog{}).Where("action = ?", action).Count(&count).Error
	return count, err
}

// FindAll retrieves all audit logs with pagination.
func (r *GormAuditLogRepository) FindAll(limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Order("timestamp DESC").
		Limit(limit).Offset(offset).
		Find(&logs).Error
	return logs, err
}

// FindWithConditions retrieves audit logs with specific conditions and pagination.
func (r *GormAuditLogRepository) FindWithConditions(conditions map[string]interface{}, limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where(conditions).
		Order("timestamp DESC").
		Limit(limit).Offset(offset).
		Find(&logs).Error
	return logs, err
}
