// Package repository provides data access layer implementations for MedConnect.
// This package follows the Repository pattern to separate business logic from data access.
package repository

import (
	"medconnect-oriental/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	// Basic CRUD operations
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error

	// Query methods
	FindByUsername(username string) (*models.User, error)
	FindByRole(role models.Role) ([]models.User, error)
	FindByDepartmentID(deptID uuid.UUID) ([]models.User, error)
	FindActiveUsers() ([]models.User, error)

	// Count methods
	Count() (int64, error)
	CountByRole(role models.Role) (int64, error)
	CountByDepartmentID(deptID uuid.UUID) (int64, error)

	// Pagination and eager loading
	FindAll(limit, offset int) ([]models.User, error)
	FindWithPreload(conditions map[string]interface{}, preload []string) (*models.User, error)
}

// GormUserRepository implements UserRepository using GORM.
type GormUserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// Create inserts a new user record.
func (r *GormUserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByID retrieves a user by their UUID.
func (r *GormUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Department").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates an existing user record.
func (r *GormUserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete removes a user by their UUID.
func (r *GormUserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// FindByUsername retrieves a user by their username.
func (r *GormUserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Department").First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByRole retrieves all users with a specific role.
func (r *GormUserRepository) FindByRole(role models.Role) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("role = ?", role).Preload("Department").Find(&users).Error
	return users, err
}

// FindByDepartmentID retrieves all users belonging to a specific department.
func (r *GormUserRepository) FindByDepartmentID(deptID uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("dept_id = ?", deptID).Find(&users).Error
	return users, err
}

// FindActiveUsers retrieves all active users.
func (r *GormUserRepository) FindActiveUsers() ([]models.User, error) {
	var users []models.User
	err := r.db.Where("is_active = ?", true).Preload("Department").Find(&users).Error
	return users, err
}

// Count returns the total number of users.
func (r *GormUserRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

// CountByRole returns the number of users with a specific role.
func (r *GormUserRepository) CountByRole(role models.Role) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("role = ?", role).Count(&count).Error
	return count, err
}

// CountByDepartmentID returns the number of users in a specific department.
func (r *GormUserRepository) CountByDepartmentID(deptID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("dept_id = ?", deptID).Count(&count).Error
	return count, err
}

// FindAll retrieves all users with pagination.
func (r *GormUserRepository) FindAll(limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.Preload("Department").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&users).Error
	return users, err
}

// FindWithPreload retrieves a user with specific conditions and eager loads associations.
func (r *GormUserRepository) FindWithPreload(conditions map[string]interface{}, preload []string) (*models.User, error) {
	var user models.User
	query := r.db

	// Apply preloads
	for _, p := range preload {
		query = query.Preload(p)
	}

	// Apply conditions
	query = query.Where(conditions)

	err := query.First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
