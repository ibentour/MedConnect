# API Layer — `internal/api`

> **Sources:**
> - [`backend/internal/api/handlers.go`](../internal/api/handlers.go) — Core handlers (1145 lines)
> - [`backend/internal/api/admin_handlers.go`](../internal/api/admin_handlers.go) — Admin CRUD (447 lines)
> - [`backend/internal/api/history_handlers.go`](../internal/api/history_handlers.go) — History & reschedule (217 lines)
> - [`backend/internal/api/export_handlers.go`](../internal/api/export_handlers.go) — CSV/HTML export (407 lines)
> - [`backend/internal/api/dto.go`](../internal/api/dto.go) — Request/Response DTOs (153 lines)

---

## Overview

The API layer is built on **Gin** and follows a handler-centric architecture. Each handler is a method on `HandlerContext`, a dependency container holding database, services, and middleware references.

---

## HandlerContext

The central dependency injection struct:

```go
type HandlerContext struct {
    DB                *gorm.DB
    Crypto            *crypto.AESCrypto
    AI                *ai.AIService
    WhatsApp          *service.WhatsAppService
    Notification      *service.NotificationService
    DecryptionCache   *service.DecryptionCache
    WSHub             *websocket.Hub

    // Repository interfaces (defined but not yet wired)
    ReferralRepo      repository.ReferralRepository
    UserRepo          repository.UserRepository
    PatientRepo       repository.PatientRepository
    DeptRepo          repository.DepartmentRepository
    AuditLogRepo      repository.AuditLogRepository
}
```

> **Note:** Handlers currently use `h.DB` directly for queries. Repository interfaces exist but are not yet connected in `main.go`.

---

## DTOs (Data Transfer Objects)

### Request DTOs

| DTO                          | Fields                                                         | Used By              |
| ---------------------------- | -------------------------------------------------------------- | -------------------- |
| `CreateReferralRequest`      | patient_cin, patient_full_name, patient_dob, patient_phone, department_id, symptoms, urgency, ai_suggested_dept | `CreateReferral` |
| `SuggestDepartmentRequest`   | symptoms, patient_dob                                          | `SuggestDepartment`  |
| `ScheduleReferralRequest`    | appointment_date (ISO 8601)                                    | `ScheduleReferral`   |
| `RedirectReferralRequest`    | department_id, reason                                          | `RedirectReferral`   |
| `DenyReferralRequest`        | reason (min 10 chars)                                          | `DenyReferral`       |
| `CreateUserRequest`          | username, password, role, facility_name, department_id         | `CreateUser`         |
| `CreateDepartmentRequest`    | name, phone_extension, work_hours, work_days                   | `CreateDepartment`   |
| `UpdateDepartmentRequest`    | All optional fields for partial update                         | `UpdateDepartment`   |

### Response DTOs

| DTO                    | Description                                    |
| ---------------------- | ---------------------------------------------- |
| `ReferralResponse`     | Full decrypted referral view for API consumers |
| `AttachmentResponse`   | ID, file_name, file_type, file_size            |
| `QueueItem`            | Lighter referral view for triage queue         |
| `NotificationResponse` | Notification item for frontend                 |
| `PaginationMeta`       | limit, offset, total, has_next, has_prev       |

---

## Core Handlers (`handlers.go`)

### Referral Workflow

| Handler               | Method | Route                           | Role        | Description                              |
| --------------------- | ------ | --------------------------------- | ----------- | ---------------------------------------- |
| `CreateReferral`      | POST   | `/api/referrals`                 | LEVEL_2_DOC | Create referral with encrypted patient data + async AI summary |
| `SuggestDepartment`   | POST   | `/api/referrals/suggest`         | LEVEL_2_DOC | AI-powered department triage suggestion  |
| `GetQueue`            | GET    | `/api/queue`                     | CHU_DOC     | Triage queue ordered by urgency          |
| `GetReferral`         | GET    | `/api/referrals/:id`             | ALL         | Single referral detail (decrypted)       |
| `ScheduleReferral`    | PATCH  | `/api/referrals/:id/schedule`    | CHU_DOC     | Schedule appointment + WhatsApp notify   |
| `RedirectReferral`    | PATCH  | `/api/referrals/:id/redirect`    | CHU_DOC     | Redirect to another department           |
| `DenyReferral`        | PATCH  | `/api/referrals/:id/deny`        | CHU_DOC     | Deny with mandatory reason               |

### Attachments

| Handler              | Method | Route                        | Role        | Description                              |
| -------------------- | ------ | ------------------------------ | ----------- | ---------------------------------------- |
| `UploadAttachments`  | POST   | `/api/referrals/:id/attachments` | LEVEL_2_DOC | Upload files (max 10MB/file, 50MB total) |
| `GetAttachment`      | GET    | `/api/attachments/:id`       | ALL         | Download/view attachment                 |

### Notifications

| Handler                  | Method | Route                          | Role        | Description              |
| ------------------------ | ------ | ------------------------------ | ----------- | ------------------------ |
| `GetNotifications`       | GET    | `/api/notifications`          | LEVEL_2_DOC | List notifications       |
| `MarkNotificationRead`   | PATCH  | `/api/notifications/:id/read` | ALL         | Mark notification as read |

### Directory

| Handler         | Method | Route             | Role | Description                 |
| --------------- | ------ | ----------------- | ---- | --------------------------- |
| `GetDirectory`  | GET    | `/api/directory`  | ALL  | List CHU departments (paginated) |

---

## Admin Handlers (`admin_handlers.go`)

### User Management

| Handler        | Method | Route                  | Role        | Description       |
| -------------- | ------ | ---------------------- | ----------- | ----------------- |
| `GetUsers`     | GET    | `/api/admin/users`    | SUPER_ADMIN | List users        |
| `CreateUser`   | POST   | `/api/admin/users`    | SUPER_ADMIN | Create user       |
| `DeleteUser`   | DELETE | `/api/admin/users/:id`| SUPER_ADMIN | Delete user       |

### Department Management

| Handler               | Method | Route                            | Role          | Description                   |
| --------------------- | ------ | -------------------------------- | ------------- | ----------------------------- |
| `GetAdminDepartments` | GET    | `/api/admin/departments`         | SUPER_ADMIN   | Departments with stats        |
|                       |        | `/api/analyst/stats/departments` | ANALYST       |                               |
| `CreateDepartment`    | POST   | `/api/admin/departments`         | SUPER_ADMIN   | Create department             |
| `UpdateDepartment`    | PATCH  | `/api/admin/departments/:id`     | SUPER_ADMIN   | Partial update department     |
| `DeleteDepartment`    | DELETE | `/api/admin/departments/:id`     | SUPER_ADMIN   | Delete department             |

### Statistics & Audit

| Handler                   | Method | Route                            | Role          | Description                        |
| ------------------------- | ------ | -------------------------------- | ------------- | ---------------------------------- |
| `GetAdminStats`           | GET    | `/api/admin/stats`               | SUPER_ADMIN   | Dashboard stats                    |
| `GetAnalystDoctorStats`   | GET    | `/api/analyst/stats/doctors`     | ANALYST, SUPER_ADMIN | Doctor referral stats    |
| `GetAuditLogs`            | GET    | `/api/admin/audit-logs`          | SUPER_ADMIN   | Audit logs with filters + pagination |
| `GetAuditLogExport`       | GET    | `/api/admin/audit-logs/export`   | SUPER_ADMIN   | Export audit logs (CSV/HTML)       |
| `GetUsersForFilter`       | GET    | `/api/admin/audit-logs/users`    | SUPER_ADMIN   | User list for filter dropdown      |
| `GetActionsForFilter`     | GET    | `/api/admin/audit-logs/actions`  | SUPER_ADMIN   | Unique actions for filter          |
| `GetReferralsExport`      | GET    | `/api/admin/referrals/export`    | SUPER_ADMIN   | Export referrals (CSV/HTML)        |

---

## History Handlers (`history_handlers.go`)

| Handler             | Method | Route                        | Role                    | Description                              |
| ------------------- | ------ | ------------------------------ | ----------------------- | ---------------------------------------- |
| `GetReferralHistory`| GET    | `/api/history`                | CHU_DOC, LEVEL_2_DOC   | Role-based referral history              |
| `RescheduleReferral`| PATCH  | `/api/referrals/:id/reschedule` | CHU_DOC              | Reschedule appointment + WhatsApp notify |
| `CancelReferral`    | PATCH  | `/api/referrals/:id/cancel`    | CHU_DOC               | Cancel appointment                       |

---

## Export Handlers (`export_handlers.go`)

| Handler              | Method | Route                          | Role        | Description                    |
| -------------------- | ------ | ------------------------------ | ----------- | ------------------------------ |
| `GetAuditLogExport`  | GET    | `/api/admin/audit-logs/export` | SUPER_ADMIN | Export audit logs as CSV or HTML/PDF |
| `GetReferralsExport` | GET    | `/api/admin/referrals/export`  | SUPER_ADMIN | Export referrals as CSV or HTML/PDF |

**Export formats:**
- `?format=csv` — Comma-separated values with BOM for Excel compatibility
- `?format=html` — Styled HTML table (printable as PDF)

---

## Helper Methods

| Method                       | Description                                                       |
| ---------------------------- | ----------------------------------------------------------------- |
| `decryptField()`             | Cached decryption with audit logging on failure                   |
| `decryptPatientField()`      | Patient-specific field decryption                                 |
| `decryptReferralField()`     | Referral-specific field decryption                                |
| `logDecryptionAudit()`       | Async audit log for decryption failures                           |
| `parsePaginationParams()`    | Extracts `limit`/`offset` from query params (max 100)             |

---

## Error Handling Pattern

All handlers follow a consistent error response format:

```json
{
  "error": "Human-readable error message"
}
```

HTTP status codes:
- `400` — Bad request (validation errors, malformed input)
- `401` — Unauthorized (missing/invalid token)
- `403` — Forbidden (insufficient role)
- `404` — Not found
- `409` — Conflict (e.g., department not accepting referrals)
- `500` — Internal server error

---

## Pagination

All list endpoints support pagination via query parameters:

```
GET /api/admin/users?limit=20&offset=0
```

Response includes `PaginationMeta`:
```json
{
  "data": [...],
  "meta": {
    "limit": 20,
    "offset": 0,
    "total": 45,
    "has_next": true,
    "has_prev": false
  }
}
```
