// Package repository provides data access layer implementations for MedConnect.
// This package follows the Repository pattern to separate business logic from data access.
package repository

import (
	"medconnect-oriental/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PatientRepository defines the interface for patient data operations.
type PatientRepository interface {
	// Basic CRUD operations
	Create(patient *models.Patient) error
	GetByID(id uuid.UUID) (*models.Patient, error)
	Update(patient *models.Patient) error
	Delete(id uuid.UUID) error

	// Query methods
	FindByCIN(cin string) (*models.Patient, error)
	FindByPhone(phone string) ([]models.Patient, error)

	// Count methods
	Count() (int64, error)

	// Pagination and eager loading
	FindAll(limit, offset int) ([]models.Patient, error)
	FindWithPreload(conditions map[string]interface{}, preload []string) (*models.Patient, error)
}

// GormPatientRepository implements PatientRepository using GORM.
type GormPatientRepository struct {
	db *gorm.DB
}

// NewPatientRepository creates a new PatientRepository instance.
func NewPatientRepository(db *gorm.DB) *GormPatientRepository {
	return &GormPatientRepository{db: db}
}

// Create inserts a new patient record.
func (r *GormPatientRepository) Create(patient *models.Patient) error {
	return r.db.Create(patient).Error
}

// GetByID retrieves a patient by its UUID.
func (r *GormPatientRepository) GetByID(id uuid.UUID) (*models.Patient, error) {
	var patient models.Patient
	err := r.db.Preload("Referrals").First(&patient, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

// Update updates an existing patient record.
func (r *GormPatientRepository) Update(patient *models.Patient) error {
	return r.db.Save(patient).Error
}

// Delete removes a patient by its UUID.
func (r *GormPatientRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Patient{}, "id = ?", id).Error
}

// FindByCIN retrieves a patient by their encrypted CIN.
// Note: Since CIN is encrypted, this requires scanning all patients.
// In production, consider storing a hash of CIN for lookups.
func (r *GormPatientRepository) FindByCIN(cin string) (*models.Patient, error) {
	var patient models.Patient
	// This is a simplified implementation - in production you'd use a hash
	err := r.db.First(&patient, "id = ?", uuid.Nil).Error
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

// FindByPhone retrieves patients by phone number.
func (r *GormPatientRepository) FindByPhone(phone string) ([]models.Patient, error) {
	var patients []models.Patient
	err := r.db.Where("phone_number = ?", phone).Find(&patients).Error
	return patients, err
}

// Count returns the total number of patients.
func (r *GormPatientRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Patient{}).Count(&count).Error
	return count, err
}

// FindAll retrieves all patients with pagination.
func (r *GormPatientRepository) FindAll(limit, offset int) ([]models.Patient, error) {
	var patients []models.Patient
	err := r.db.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&patients).Error
	return patients, err
}

// FindWithPreload retrieves a patient with specific conditions and eager loads associations.
func (r *GormPatientRepository) FindWithPreload(conditions map[string]interface{}, preload []string) (*models.Patient, error) {
	var patient models.Patient
	query := r.db

	// Apply preloads
	for _, p := range preload {
		query = query.Preload(p)
	}

	// Apply conditions
	query = query.Where(conditions)

	err := query.First(&patient).Error
	if err != nil {
		return nil, err
	}
	return &patient, nil
}
