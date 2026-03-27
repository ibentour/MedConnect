# Models — `internal/models`

> **Sources:**
> - [`backend/internal/models/models.go`](../internal/models/models.go) — Core GORM models
> - [`backend/internal/models/notification.go`](../internal/models/notification.go) — Notification model

---

## Overview

MedConnect uses **GORM** (Go ORM) with **PostgreSQL** as the database. All models use UUID primary keys and auto-migrate on startup. Patient PII (CIN, FullName, Symptoms) is stored **encrypted** using AES-256-GCM.

---

## Entity Relationship Diagram

```
┌──────────────┐       ┌──────────────┐       ┌───────────────┐
│    Users     │       │ Departments  │       │   Patients    │
├──────────────┤       ├──────────────┤       ├───────────────┤
│ id (PK)      │──┐    │ id (PK)      │───┐   │ id (PK)       │
│ username     │  │    │ name (uniq)  │   │   │ cin (enc)     │
│ password_hash│  │    │ phone_ext    │   │   │ full_name(enc)│
│ role         │  │    │ work_hours   │   │   │ date_of_birth │
│ dept_id (FK) │──┘    │ work_days    │   │   │ phone_number  │
│ facility_name│       │ is_accepting │   │   └──────┬────────┘
│ is_active    │       └──────────────┘   │          │
└──────┬───────┘                          │          │
       │                                  │          │
       │ creates                          │          │
       ▼                                  ▼          │
┌───────────────┐       ┌───────────────┐ │          │
│  Referrals    │       │ Attachments   │ │          │
├───────────────┤       ├───────────────┤ │          │
│ id (PK)       │◀──────│ referral_id   │ │          │
│ patient_id(FK)│───┘   │ file_path     │ │          │
│ creator_id    │   |   │ file_name     │ │          │
│ dept_id (FK)  │───┘   │ file_type     │ │          │
│ status        │       │ file_size     │ │          │
│ urgency       │       └───────────────┘ │          │
│ symptoms(enc) │                         │          │
│ ai_dept       │      ┌──────────────┐   │          │
│ ai_summary    │      │ AuditLogs    │   │          │
│ apt_date      │      ├──────────────┤   │          │
│ reject_reason │      │ id (PK)      │   │          │
└──────┬────────┘      │ user_id (FK) │───┘          │
       │               │ username     │              │
       │               │ action       │              │
       │               │ target_id    │              │
       ▼               │ ip_address   │              │
┌──────────────┐       │ user_agent   │              │
│Notifications │       │ status       │              │
├──────────────┤       │ timestamp    │              │
│ id (PK)      │       └──────────────┘              │
│ user_id (FK) │                                     │
│ ref_id (FK)  │◀────────────────────────────────────┘
│ message      │
│ is_read      │
│ created_at   │
└──────────────┘
```

---

## Enums

### `Role`

```go
type Role string

const (
    RoleSuperAdmin Role = "SUPER_ADMIN"
    RoleAnalyst    Role = "ANALYST"
    RoleCHUDoc     Role = "CHU_DOC"
    RoleLevel2Doc  Role = "LEVEL_2_DOC"
)
```

| Role           | Description                                        |
| -------------- | -------------------------------------------------- |
| `SUPER_ADMIN`  | Full system access — user/dept management, audit   |
| `ANALYST`      | Read-only analytics dashboard                      |
| `CHU_DOC`      | CHU specialist — triage queue, schedule/deny/redirect |
| `LEVEL_2_DOC`  | Provincial doctor — create referrals, view history |

### `ReferralStatus`

```go
type ReferralStatus string

const (
    StatusPending    ReferralStatus = "PENDING"
    StatusScheduled  ReferralStatus = "SCHEDULED"
    StatusRedirected ReferralStatus = "REDIRECTED"
    StatusDenied     ReferralStatus = "DENIED"
    StatusCanceled   ReferralStatus = "CANCELED"
)
```

```
PENDING ──▶ SCHEDULED
    │
    ├──▶ REDIRECTED ──▶ SCHEDULED
    │                ──▶ DENIED
    │
    └──▶ DENIED
    │
    └──▶ CANCELED
```

### `UrgencyLevel`

```go
type UrgencyLevel string

const (
    UrgencyLow      UrgencyLevel = "LOW"
    UrgencyMedium   UrgencyLevel = "MEDIUM"
    UrgencyHigh     UrgencyLevel = "HIGH"
    UrgencyCritical UrgencyLevel = "CRITICAL"
)
```

---

## Core Models

### `User`

```go
type User struct {
    ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Username      string     `gorm:"uniqueIndex;not null"`
    PasswordHash  string     `gorm:"not null" json:"-"`
    Role          Role       `gorm:"not null"`
    DepartmentID  *uuid.UUID `gorm:"type:uuid"`
    FacilityName  string
    IsActive      bool       `gorm:"default:true"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

**Relationships:** Belongs to `Department` (optional for Level 2/Analyst/Admin).

### `Department`

```go
type Department struct {
    ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Name           string    `gorm:"uniqueIndex;not null"`
    PhoneExtension string
    WorkHours      string    // e.g., "08:00-16:00"
    WorkDays       string    // e.g., "Mon,Tue,Wed,Thu,Fri"
    IsAccepting    bool      `gorm:"default:true"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### `Patient`

```go
type Patient struct {
    ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    CIN          string     // AES-256-GCM encrypted
    FullName     string     // AES-256-GCM encrypted
    DateOfBirth  time.Time
    PhoneNumber  string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

> **Security:** `CIN` and `FullName` are encrypted at the application layer before storage. Decryption uses the `AESCrypto` service with a cached decryption layer.

### `Referral`

```go
type Referral struct {
    ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    PatientID       uuid.UUID      `gorm:"type:uuid;not null"`
    CreatorID       uuid.UUID      `gorm:"type:uuid;not null"`
    CurrentDeptID   uuid.UUID      `gorm:"type:uuid;not null"`
    Status          ReferralStatus `gorm:"default:'PENDING'"`
    Urgency         UrgencyLevel   `gorm:"default:'MEDIUM'"`
    Symptoms        string                                          // AES-256-GCM encrypted
    AISuggestedDept string
    AISummary       string
    AppointmentDate *time.Time
    RejectionReason string
    CreatedAt       time.Time
    UpdatedAt       time.Time

    // Associations
    Patient      Patient       `gorm:"foreignKey:PatientID"`
    Creator      User          `gorm:"foreignKey:CreatorID"`
    Department   Department    `gorm:"foreignKey:CurrentDeptID"`
    Attachments  []Attachment  `gorm:"foreignKey:ReferralID"`
}
```

**Indexes:**
- Composite on `(current_dept_id, status)` — Queue queries
- Composite on `(creator_id, status)` — History queries
- Composite on `(current_dept_id, urgency)` — Urgency sorting

### `Attachment`

```go
type Attachment struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    ReferralID uuid.UUID `gorm:"type:uuid;not null"`
    FilePath   string    `gorm:"not null" json:"-"`  // Hidden from API responses
    FileName   string
    FileType   string
    FileSize   int64
    CreatedAt  time.Time
}
```

> **Note:** `FilePath` is excluded from JSON serialization to prevent path leakage.

### `AuditLog`

```go
type AuditLog struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id"`
    UserID    *uuid.UUID
    Username  string    // Denormalized for audit integrity
    Action    string
    TargetID  string
    IPAddress string
    UserAgent string
    Status    int       // HTTP status code
    Timestamp time.Time `gorm:"autoCreateTime"`

    User *User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

func (AuditLog) TableName() string {
    return "audit_logs"
}
```

> **Compliance:** Satisfies Moroccan Law 09-08 data protection requirements. Audit records are append-only with denormalized username for historical accuracy.

### `Notification`

```go
type Notification struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID     uuid.UUID `gorm:"type:uuid;not null"`
    ReferralID uuid.UUID `gorm:"type:uuid"`
    Message    string
    IsRead     bool      `gorm:"default:false"`
    CreatedAt  time.Time
}
```

**Index:** Composite on `(user_id, is_read)` for unread count queries.

---

## Data Flow

1. **Patient data** is encrypted via `AESCrypto.Encrypt()` before `db.Create()`
2. **Referral creation** stores encrypted symptoms, then triggers async AI summary
3. **Read operations** decrypt fields through `DecryptionCache` for performance
4. **Audit logs** are written asynchronously for every HTTP request (see `middleware/audit.go`)
