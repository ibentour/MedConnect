// Package repository provides data access layer implementations for MedConnect.
// This package follows the Repository pattern to separate business logic from data access.
package repository

import (
	"medconnect-oriental/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DepartmentRepository defines the interface for department data operations.
type DepartmentRepository interface {
	// Basic CRUD operations
	Create(department *models.Department) error
	GetByID(id uuid.UUID) (*models.Department, error)
	Update(department *models.Department) error
	Delete(id uuid.UUID) error

	// Query methods
	FindByName(name string) (*models.Department, error)
	FindAll() ([]models.Department, error)
	FindAccepting() ([]models.Department, error)
	FindByPhoneExtension(extension string) (*models.Department, error)

	// Count methods
	Count() (int64, error)
	CountAccepting() (int64, error)

	// Pagination and eager loading
	FindWithPagination(limit, offset int) ([]models.Department, error)
	FindWithPreload(conditions map[string]interface{}, preload []string) (*models.Department, error)
}

// GormDepartmentRepository implements DepartmentRepository using GORM.
type GormDepartmentRepository struct {
	db *gorm.DB
}

// NewDepartmentRepository creates a new DepartmentRepository instance.
func NewDepartmentRepository(db *gorm.DB) *GormDepartmentRepository {
	return &GormDepartmentRepository{db: db}
}

// Create inserts a new department record.
func (r *GormDepartmentRepository) Create(department *models.Department) error {
	return r.db.Create(department).Error
}

// GetByID retrieves a department by its UUID.
func (r *GormDepartmentRepository) GetByID(id uuid.UUID) (*models.Department, error) {
	var department models.Department
	err := r.db.First(&department, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &department, nil
}

// Update updates an existing department record.
func (r *GormDepartmentRepository) Update(department *models.Department) error {
	return r.db.Save(department).Error
}

// Delete removes a department by its UUID.
func (r *GormDepartmentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Department{}, "id = ?", id).Error
}

// FindByName retrieves a department by its name.
func (r *GormDepartmentRepository) FindByName(name string) (*models.Department, error) {
	var department models.Department
	err := r.db.First(&department, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &department, nil
}

// FindAll retrieves all departments ordered by name.
func (r *GormDepartmentRepository) FindAll() ([]models.Department, error) {
	var departments []models.Department
	err := r.db.Order("name ASC").Find(&departments).Error
	return departments, err
}

// FindAccepting retrieves all departments that are currently accepting referrals.
func (r *GormDepartmentRepository) FindAccepting() ([]models.Department, error) {
	var departments []models.Department
	err := r.db.Where("is_accepting = ?", true).Order("name ASC").Find(&departments).Error
	return departments, err
}

// FindByPhoneExtension retrieves a department by its phone extension.
func (r *GormDepartmentRepository) FindByPhoneExtension(extension string) (*models.Department, error) {
	var department models.Department
	err := r.db.First(&department, "phone_extension = ?", extension).Error
	if err != nil {
		return nil, err
	}
	return &department, nil
}

// Count returns the total number of departments.
func (r *GormDepartmentRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Department{}).Count(&count).Error
	return count, err
}

// CountAccepting returns the number of departments that are accepting referrals.
func (r *GormDepartmentRepository) CountAccepting() (int64, error) {
	var count int64
	err := r.db.Model(&models.Department{}).Where("is_accepting = ?", true).Count(&count).Error
	return count, err
}

// FindWithPagination retrieves departments with pagination.
func (r *GormDepartmentRepository) FindWithPagination(limit, offset int) ([]models.Department, error) {
	var departments []models.Department
	err := r.db.Order("name ASC").Limit(limit).Offset(offset).Find(&departments).Error
	return departments, err
}

// FindWithPreload retrieves a department with specific conditions and eager loads associations.
func (r *GormDepartmentRepository) FindWithPreload(conditions map[string]interface{}, preload []string) (*models.Department, error) {
	var department models.Department
	query := r.db

	// Apply preloads
	for _, p := range preload {
		query = query.Preload(p)
	}

	// Apply conditions
	query = query.Where(conditions)

	err := query.First(&department).Error
	if err != nil {
		return nil, err
	}
	return &department, nil
}
