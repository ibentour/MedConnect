# Repositories — `internal/repository`

> **Sources:**
> - [`backend/internal/repository/database.go`](../internal/repository/database.go)
> - [`backend/internal/repository/referral.go`](../internal/repository/referral.go)
> - [`backend/internal/repository/user.go`](../internal/repository/user.go)
> - [`backend/internal/repository/patient.go`](../internal/repository/patient.go)
> - [`backend/internal/repository/department.go`](../internal/repository/department.go)
> - [`backend/internal/repository/auditlog.go`](../internal/repository/auditlog.go)
> - [`backend/internal/repository/seeder.go`](../internal/repository/seeder.go)

---

## Overview

The repository layer implements the **Repository Pattern** with Go interfaces and GORM-backed implementations. Each entity has an interface defining available operations and a concrete implementation using GORM.

> **Current state:** Repository interfaces are defined and implemented, but handlers currently access `h.DB` directly. Wiring the repositories into `HandlerContext` is planned for future refactoring.

---

## Database Connection

> **Source:** [`database.go`](../internal/repository/database.go)

### `ConnectDB(dsn string) (*gorm.DB, error)`

Connects to PostgreSQL using GORM's Postgres driver and runs auto-migrations for all models.

```go
db, err := repository.ConnectDB(os.Getenv("DB_DSN"))
// → Runs AutoMigrate for: User, Department, Patient, Referral, Attachment, AuditLog, Notification
```

**Auto-migrated tables:**
1. `users`
2. `departments`
3. `patients`
4. `referrals`
5. `attachments`
6. `audit_logs`
7. `notifications`

---

## Interface Contracts

### ReferralRepository

```go
type ReferralRepository interface {
    Create(referral *models.Referral) error
    GetByID(id uuid.UUID) (*models.Referral, error)
    Update(referral *models.Referral) error
    Delete(id uuid.UUID) error
    FindByStatus(status models.ReferralStatus) ([]models.Referral, error)
    FindByCreatorID(creatorID uuid.UUID) ([]models.Referral, error)
    FindByDepartmentID(deptID uuid.UUID) ([]models.Referral, error)
    FindByPatientID(patientID uuid.UUID) ([]models.Referral, error)
    FindQueueByDepartmentID(deptID uuid.UUID) ([]models.Referral, error)
    CountByStatus(status models.ReferralStatus) (int64, error)
    CountByDepartmentID(deptID uuid.UUID) (int64, error)
    FindAll(limit, offset int) ([]models.Referral, error)
    FindWithPreload(conditions map[string]interface{}, preloads ...string) ([]models.Referral, error)
}
```

**Key method:** `FindQueueByDepartmentID` returns pending referrals ordered by urgency (CRITICAL → HIGH → MEDIUM → LOW).

### UserRepository

```go
type UserRepository interface {
    Create(user *models.User) error
    GetByID(id uuid.UUID) (*models.User, error)
    Update(user *models.User) error
    Delete(id uuid.UUID) error
    FindByUsername(username string) (*models.User, error)
    FindByRole(role models.Role) ([]models.User, error)
    FindByDepartmentID(deptID uuid.UUID) ([]models.User, error)
    FindActiveUsers() ([]models.User, error)
    Count() (int64, error)
    CountByRole(role models.Role) (int64, error)
    CountByDepartmentID(deptID uuid.UUID) (int64, error)
    FindAll(limit, offset int) ([]models.User, error)
    FindWithPreload(conditions map[string]interface{}, preloads ...string) ([]models.User, error)
}
```

### PatientRepository

```go
type PatientRepository interface {
    Create(patient *models.Patient) error
    GetByID(id uuid.UUID) (*models.Patient, error)
    Update(patient *models.Patient) error
    Delete(id uuid.UUID) error
    FindByCIN(cin string) (*models.Patient, error)
    FindByPhone(phone string) (*models.Patient, error)
    Count() (int64, error)
    FindAll(limit, offset int) ([]models.Patient, error)
    FindWithPreload(conditions map[string]interface{}, preloads ...string) ([]models.Patient, error)
}
```

> **Note:** `FindByCIN` performs encrypted CIN lookup — production deployments should consider a hash-based search index for encrypted field lookups.

### DepartmentRepository

```go
type DepartmentRepository interface {
    Create(dept *models.Department) error
    GetByID(id uuid.UUID) (*models.Department, error)
    Update(dept *models.Department) error
    Delete(id uuid.UUID) error
    FindByName(name string) (*models.Department, error)
    FindAll() ([]models.Department, error)
    FindAccepting() ([]models.Department, error)
    FindByPhoneExtension(ext string) (*models.Department, error)
    Count() (int64, error)
    CountAccepting() (int64, error)
    FindWithPagination(limit, offset int) ([]models.Department, error)
    FindWithPreload(conditions map[string]interface{}, preloads ...string) ([]models.Department, error)
}
```

### AuditLogRepository

```go
type AuditLogRepository interface {
    Create(log *models.AuditLog) error
    GetByID(id uuid.UUID) (*models.AuditLog, error)
    FindByUserID(userID uuid.UUID, limit, offset int) ([]models.AuditLog, error)
    FindByAction(action string, limit, offset int) ([]models.AuditLog, error)
    FindByTargetID(targetID string) ([]models.AuditLog, error)
    FindByIPAddress(ip string, limit, offset int) ([]models.AuditLog, error)
    FindByStatus(status int, limit, offset int) ([]models.AuditLog, error)
    FindByDateRange(start, end time.Time, limit, offset int) ([]models.AuditLog, error)
    Count() (int64, error)
    CountByUserID(userID uuid.UUID) (int64, error)
    CountByAction(action string) (int64, error)
    FindAll(limit, offset int) ([]models.AuditLog, error)
    FindWithConditions(conditions map[string]interface{}, limit, offset int, orderBy string) ([]models.AuditLog, error)
}
```

---

## GORM Implementation Pattern

All repositories follow the same implementation pattern:

```go
type GORMReferralRepository struct {
    db *gorm.DB
}

func NewReferralRepository(db *gorm.DB) ReferralRepository {
    return &GORMReferralRepository{db: db}
}

func (r *GORMReferralRepository) Create(referral *models.Referral) error {
    return r.db.Create(referral).Error
}

func (r *GORMReferralRepository) GetByID(id uuid.UUID) (*models.Referral, error) {
    var referral models.Referral
    err := r.db.Preload(clause.Associations).First(&referral, "id = ?", id).Error
    return &referral, err
}
// ... remaining methods
```

---

## Seeder

> **Source:** [`seeder.go`](../internal/repository/seeder.go)

`SeedTestAccounts(db *gorm.DB)` creates test data on startup (idempotent — skips if accounts exist).

### Seeded Accounts

| Username             | Password                 | Role          | Facility                | Department  |
| -------------------- | ------------------------ | ------------- | ----------------------- | ----------- |
| `admin_oujda`        | `OujdaSuper2026!`       | SUPER_ADMIN   | Regional Health HQ      | —           |
| `analyst_oriental`   | `DataStats2026#`        | ANALYST       | Regional Observatory    | —           |
| `dr_cardiologue`     | `HeartSaver123`         | CHU_DOC       | CHU Mohammed VI         | Cardiology  |
| `dr_neurologue`      | `BrainCheck456`         | CHU_DOC       | CHU Mohammed VI         | Neurology   |
| `doc_berkane`        | `Provincial_Berkane1`   | LEVEL_2_DOC   | Hôpital de Berkane      | —           |
| `doc_ahfir`          | `Provincial_Ahfir2`     | LEVEL_2_DOC   | Hôpital de Ahfir        | —           |

### Seeded Departments

1. Cardiology
2. Neurology
3. Chirurgie Pédiatrique
4. Traumatologie Générale
5. Oncologie
6. Néphrologie

---

## Performance Notes

- **N+1 avoidance:** Admin department stats and analyst doctor stats use raw SQL aggregation queries instead of iterating through records.
- **Composite indexes** on referrals table optimize queue and history queries.
- **Pagination** is supported on all `FindAll` and `FindWith*` methods.
