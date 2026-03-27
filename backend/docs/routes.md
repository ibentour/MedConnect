# Routes — `cmd/server/main.go`

> **Source:** [`backend/cmd/server/main.go`](../cmd/server/main.go)

---

## Overview

Routes are registered in `main.go` using Gin's router group pattern. The middleware stack applies security headers globally, JWT auth on protected routes, and RBAC on role-specific endpoints.

---

## Middleware Stack

```
Request
  │
  ▼
SecurityHeaders (all routes)
  │
  ▼
CORS (all routes)
  │
  ▼
AuditLogger (all routes, async)
  │
  ├──▶ Public routes (login, health)
  │
  ▼
JWTAuth (protected routes)
  │
  ├──▶ All authenticated routes
  │
  ▼
RBAC (role-specific routes)
  │
  ▼
Handler
```

---

## Public Routes

| Method | Path          | Handler           | Rate Limit       | Description          |
| ------ | ------------- | ----------------- | ---------------- | -------------------- |
| POST   | `/api/login`  | `loginHandler`    | 10/min per IP    | Authenticate, return JWT |
| GET    | `/api/health` | Inline handler    | None             | Health check         |
| GET    | `/ws`         | `GinWebSocket`    | None             | WebSocket upgrade    |

---

## Authenticated Routes (All Roles)

All routes require a valid JWT in the `Authorization: Bearer <token>` header or `?token=<token>` query parameter.

| Method | Path                        | Handler                | Description                    |
| ------ | --------------------------- | ---------------------- | ------------------------------ |
| GET    | `/api/directory`            | `GetDirectory`         | List CHU departments           |
| GET    | `/api/referrals/:id`        | `GetReferral`          | Single referral detail         |
| PATCH  | `/api/notifications/:id/read`| `MarkNotificationRead` | Mark notification as read     |
| GET    | `/api/attachments/:id`      | `GetAttachment`        | Download/view attachment        |
| GET    | `/api/history`              | `GetReferralHistory`   | Referral history (role-filtered)|

---

## Level 2 Doctor Routes

Requires `LEVEL_2_DOC` role.

| Method | Path                          | Handler              | Description                    |
| ------ | ----------------------------- | -------------------- | ------------------------------ |
| POST   | `/api/referrals`              | `CreateReferral`     | Create new referral             |
| POST   | `/api/referrals/suggest`      | `SuggestDepartment`  | AI department triage            |
| GET    | `/api/notifications`          | `GetNotifications`   | List notifications              |
| POST   | `/api/referrals/:id/attachments`| `UploadAttachments` | Upload referral files           |

---

## CHU Doctor Routes

Requires `CHU_DOC` role.

| Method | Path                            | Handler                | Description                    |
| ------ | ------------------------------- | ---------------------- | ------------------------------ |
| GET    | `/api/queue`                    | `GetQueue`             | Triage queue (ordered by urgency) |
| PATCH  | `/api/referrals/:id/schedule`   | `ScheduleReferral`     | Schedule appointment            |
| PATCH  | `/api/referrals/:id/redirect`   | `RedirectReferral`     | Redirect to another department  |
| PATCH  | `/api/referrals/:id/deny`       | `DenyReferral`         | Deny referral                   |
| PATCH  | `/api/referrals/:id/reschedule` | `RescheduleReferral`   | Reschedule appointment          |
| PATCH  | `/api/referrals/:id/cancel`     | `CancelReferral`       | Cancel appointment              |

---

## Analyst Routes

Requires `ANALYST` or `SUPER_ADMIN` role.

| Method | Path                          | Handler                | Description                    |
| ------ | ----------------------------- | ---------------------- | ------------------------------ |
| GET    | `/api/analyst/stats/departments` | `GetAdminDepartments` | Department statistics           |
| GET    | `/api/analyst/stats/doctors`     | `GetAnalystDoctorStats`| Doctor referral statistics      |

---

## Admin Routes

Requires `SUPER_ADMIN` role.

### User Management

| Method | Path                    | Handler        | Description       |
| ------ | ----------------------- | -------------- | ----------------- |
| GET    | `/api/admin/users`      | `GetUsers`     | List users        |
| POST   | `/api/admin/users`      | `CreateUser`   | Create user       |
| DELETE | `/api/admin/users/:id`  | `DeleteUser`   | Delete user       |

### Department Management

| Method | Path                           | Handler              | Description           |
| ------ | ------------------------------ | -------------------- | --------------------- |
| GET    | `/api/admin/departments`       | `GetAdminDepartments`| List departments      |
| POST   | `/api/admin/departments`       | `CreateDepartment`   | Create department     |
| PATCH  | `/api/admin/departments/:id`   | `UpdateDepartment`   | Update department     |
| DELETE | `/api/admin/departments/:id`   | `DeleteDepartment`   | Delete department     |

### Statistics & Audit

| Method | Path                             | Handler              | Description                |
| ------ | -------------------------------- | -------------------- | -------------------------- |
| GET    | `/api/admin/stats`               | `GetAdminStats`      | Dashboard statistics       |
| GET    | `/api/admin/audit-logs`          | `GetAuditLogs`       | Audit log list + filters   |
| GET    | `/api/admin/audit-logs/export`   | `GetAuditLogExport`  | Export audit logs (CSV/HTML)|
| GET    | `/api/admin/audit-logs/users`    | `GetUsersForFilter`  | Users for filter dropdown  |
| GET    | `/api/admin/audit-logs/actions`  | `GetActionsForFilter`| Actions for filter dropdown|
| GET    | `/api/admin/referrals/export`    | `GetReferralsExport` | Export referrals (CSV/HTML)|

---

## WebSocket Events

Connected via `GET /ws` with JWT authentication (query param or header).

### Outbound Events (Server → Client)

| Event                  | Payload                              | Target               |
| ---------------------- | ------------------------------------ | -------------------- |
| `new_referral`         | Referral data                        | Department clients   |
| `referral_updated`     | Updated referral status              | Creator client       |
| `referral_redirected`  | Redirect details                     | Department clients   |

### Connection

```javascript
const ws = new WebSocket(`ws://localhost:3000/ws?token=${jwt}`);
```

---

## Route Diagram

```
/api
├── POST   /login              (public, rate-limited)
├── GET    /health             (public)
├── GET    /directory          (authenticated)
├── GET    /history            (authenticated, role-filtered)
├── /referrals
│   ├── POST   /               (LEVEL_2_DOC)
│   ├── POST   /suggest        (LEVEL_2_DOC)
│   └── /:id
│       ├── GET               (authenticated)
│       ├── PATCH /schedule    (CHU_DOC)
│       ├── PATCH /redirect    (CHU_DOC)
│       ├── PATCH /deny        (CHU_DOC)
│       ├── PATCH /reschedule  (CHU_DOC)
│       ├── PATCH /cancel      (CHU_DOC)
│       └── POST /attachments  (LEVEL_2_DOC)
├── /attachments/:id           (authenticated)
├── /notifications
│   ├── GET                    (LEVEL_2_DOC)
│   └── /:id/read              (authenticated)
├── /queue                     (CHU_DOC)
├── /analyst
│   ├── GET /stats/departments (ANALYST, SUPER_ADMIN)
│   └── GET /stats/doctors     (ANALYST, SUPER_ADMIN)
└── /admin                     (SUPER_ADMIN)
    ├── GET    /stats
    ├── /users
    │   ├── GET
    │   ├── POST
    │   └── DELETE /:id
    ├── /departments
    │   ├── GET
    │   ├── POST
    │   ├── PATCH  /:id
    │   └── DELETE /:id
    ├── /audit-logs
    │   ├── GET
    │   ├── GET /export
    │   ├── GET /users
    │   └── GET /actions
    └── GET /referrals/export

/ws                            (WebSocket, JWT required)
```
