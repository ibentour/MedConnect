package repository

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"medconnect-oriental/backend/internal/models"
)

// ──────────────────────────────────────────────────────────────────────
// Test Account Seeder
// ──────────────────────────────────────────────────────────────────────
//
// Seeds the database with development/test accounts.
// ⚠  FOR DEVELOPMENT ONLY — replace with Super Admin dashboard in production.
//
// Accounts are only created if the username does not already exist,
// making this function idempotent (safe to run on every startup).
// ──────────────────────────────────────────────────────────────────────

type seedAccount struct {
	Username     string
	Password     string
	Role         models.Role
	FacilityName string
	DeptName     string // empty string if no department attachment
}

// SeedTestAccounts creates the 6 development test accounts and their
// associated departments if they don't already exist.
func SeedTestAccounts(db *gorm.DB) error {
	// ── 1. Seed CHU Departments ──────────────────────────────
	departments := []models.Department{
		{Name: "Cardiology", PhoneExtension: "3201", WorkHours: "08:00-16:00", WorkDays: "Lundi-Vendredi", IsAccepting: true},
		{Name: "Neurology", PhoneExtension: "3202", WorkHours: "08:00-16:00", WorkDays: "Lundi-Jeudi", IsAccepting: true},
		{Name: "Pediatric Surgery", PhoneExtension: "3203", WorkHours: "08:00-14:00", WorkDays: "Lundi-Vendredi", IsAccepting: true},
		{Name: "General Traumatology", PhoneExtension: "3204", WorkHours: "08:00-18:00", WorkDays: "24/7", IsAccepting: true},
		{Name: "Oncology", PhoneExtension: "3205", WorkHours: "08:00-16:00", WorkDays: "Lundi-Mercredi", IsAccepting: true},
		{Name: "Nephrology", PhoneExtension: "3206", WorkHours: "08:00-15:00", WorkDays: "Mardi-Vendredi", IsAccepting: true},
	}

	for i := range departments {
		var existing models.Department
		result := db.Where("name = ?", departments[i].Name).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&departments[i]).Error; err != nil {
				return err
			}
			log.Printf("  🏥 Created department: %s", departments[i].Name)
		} else {
			departments[i] = existing
		}
	}

	// Build a name→ID lookup for department associations
	deptMap := make(map[string]models.Department)
	for _, d := range departments {
		deptMap[d.Name] = d
	}

	// ── 2. Seed Test User Accounts ───────────────────────────
	accounts := []seedAccount{
		{
			Username:     "admin_oujda",
			Password:     "OujdaSuper2026!",
			Role:         models.RoleSuperAdmin,
			FacilityName: "Regional Health HQ",
			DeptName:     "",
		},
		{
			Username:     "analyst_oriental",
			Password:     "DataStats2026#",
			Role:         models.RoleAnalyst,
			FacilityName: "Regional Observatory",
			DeptName:     "",
		},
		{
			Username:     "dr_cardiologue",
			Password:     "HeartSaver123",
			Role:         models.RoleCHUDoc,
			FacilityName: "CHU Mohammed VI - Oujda",
			DeptName:     "Cardiology",
		},
		{
			Username:     "dr_neurologue",
			Password:     "BrainCheck456",
			Role:         models.RoleCHUDoc,
			FacilityName: "CHU Mohammed VI - Oujda",
			DeptName:     "Neurology",
		},
		{
			Username:     "doc_berkane",
			Password:     "Provincial_Berkane1",
			Role:         models.RoleLevel2Doc,
			FacilityName: "Hôpital de Berkane",
			DeptName:     "",
		},
		{
			Username:     "doc_ahfir",
			Password:     "Provincial_Ahfir2",
			Role:         models.RoleLevel2Doc,
			FacilityName: "Hôpital de Ahfir",
			DeptName:     "",
		},
	}

	for _, acct := range accounts {
		var existing models.User
		result := db.Where("username = ?", acct.Username).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			// Hash password with bcrypt (cost 12)
			hash, err := bcrypt.GenerateFromPassword([]byte(acct.Password), 12)
			if err != nil {
				return err
			}

			user := models.User{
				Username:     acct.Username,
				PasswordHash: string(hash),
				Role:         acct.Role,
				FacilityName: acct.FacilityName,
				IsActive:     true,
			}

			// Attach department if specified
			if acct.DeptName != "" {
				if dept, ok := deptMap[acct.DeptName]; ok {
					user.DeptID = &dept.ID
				}
			}

			if err := db.Create(&user).Error; err != nil {
				return err
			}
			log.Printf("  👤 Created user: %s (%s)", acct.Username, acct.Role)
		} else {
			log.Printf("  ⏭️  User already exists: %s", acct.Username)
		}
	}

	log.Println("✅ Database seeding completed")
	return nil
}
