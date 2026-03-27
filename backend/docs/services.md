# Services — `internal/service`

> **Sources:**
> - [`backend/internal/service/whatsapp_service.go`](../internal/service/whatsapp_service.go)
> - [`backend/internal/service/templates.go`](../internal/service/templates.go)
> - [`backend/internal/service/notification_service.go`](../internal/service/notification_service.go)
> - [`backend/internal/service/cache.go`](../internal/service/cache.go)

---

## Overview

The service layer contains business logic that doesn't belong in handlers or repositories. Four services exist:

| Service                | Purpose                                         | Async? |
| ---------------------- | ----------------------------------------------- | ------ |
| `WhatsAppService`      | Send WhatsApp messages via Evolution API        | Yes (goroutines) |
| `NotificationService`  | In-app notification CRUD                        | No     |
| `DecryptionCache`      | Thread-safe in-memory cache for decrypted PII   | N/A    |
| `Templates`            | WhatsApp message templates (French + Darija)    | N/A    |

---

## WhatsAppService

> **Source:** [`whatsapp_service.go`](../internal/service/whatsapp_service.go)

Integrates with **Evolution API** (open-source WhatsApp gateway) to send clinical notifications.

### Struct

```go
type WhatsAppService struct {
    url      string  // Evolution API base URL
    token    string  // API auth token
    instance string  // Instance name (e.g., "medconnect")
}
```

### Functions

| Function                                          | Description                                              |
| ------------------------------------------------- | -------------------------------------------------------- |
| `NewWhatsAppService(url, token, instance)`        | Constructor                                              |
| `SendTextMessage(phone, message)`                 | Send a raw text message                                  |
| `SendAppointmentNotification(phone, params)`      | Send formatted appointment confirmation                  |
| `SendReferralDeniedNotification(phone, params)`   | Send denial notification                                 |
| `SendReferralRedirectedNotification(phone, params)`| Send redirect notification                              |

### Phone Normalization

Moroccan phone numbers are normalized from local format to E.164:

```
06XXXXXXXX  →  2126XXXXXXXX
07XXXXXXXX  →  2127XXXXXXXX
```

### Integration

WhatsApp notifications are **always async** (fire-and-forget goroutines):

```go
go func() {
    if err := h.WhatsApp.SendAppointmentNotification(phone, params); err != nil {
        log.Printf("WhatsApp send failed: %v", err)
    }
}()
```

### Configuration

| Variable     | Default                  | Description              |
| ------------ | ------------------------ | ------------------------ |
| `WA_URL`     | `http://localhost:8080`  | Evolution API base URL   |
| `WA_TOKEN`   | *(empty — disabled)*     | API authentication token |
| `WA_INSTANCE`| `medconnect`             | Instance name            |

> **Note:** If `WA_TOKEN` is empty, WhatsApp is effectively disabled (messages are silently skipped).

---

## Templates

> **Source:** [`templates.go`](../internal/service/templates.go)

WhatsApp message templates in **French** and **Moroccan Darija** (Arabic script).

### Template Functions

| Function                                             | Description                                        |
| ---------------------------------------------------- | -------------------------------------------------- |
| `AppointmentScheduledTemplate(params)`               | Appointment confirmation with document checklist    |
| `ReferralDeniedTemplate(params)`                     | Denial notification                                 |
| `ReferralRedirectedTemplate(params)`                 | Redirect notification                               |
| `GetDefaultInstructions(departmentName)`             | Department-specific prep instructions               |
| `DefaultCHUAddress()` / `DefaultCHUContact()`        | Static CHU contact information                      |

### Department-Specific Instructions

Pre-defined document checklists for major departments:

| Department          | Instructions Include                                |
| ------------------- | --------------------------------------------------- |
| Cardiology          | ECG récent, bilan lipidique, échocardiographie      |
| Neurology           | IRM cérébrale, EEG, bilan neurologique              |
| Chirurgie           | Bilan pré-opératoire, groupage sanguin              |
| Traumatologie       | Radiographies, scanner, bilan de coagulation        |
| Oncologie           | Compte-rendu anatomo-pathologique, bilan d'extension|
| Néphrologie         | Créatininémie, DFG, échographie rénale              |
| Pédiatrie           | Carnet de vaccination, bilan de croissance          |

---

## NotificationService

> **Source:** [`notification_service.go`](../internal/service/notification_service.go)

In-app notification management for real-time user alerts.

### Struct

```go
type NotificationService struct {
    db *gorm.DB
}
```

### Functions

| Function                           | Description                                         |
| ---------------------------------- | --------------------------------------------------- |
| `NewNotificationService(db)`      | Constructor                                          |
| `Create(userID, referralID, msg)` | Create a notification record                         |
| `GetForUser(userID)`             | Retrieve last 50 notifications for a user            |
| `MarkAsRead(id, userID)`         | Mark a specific notification as read                 |
| `GetUnreadCount(userID)`         | Count unread notifications                           |

### Usage in Handlers

```go
// After scheduling a referral
h.Notification.Create(
    referral.CreatorID,
    referral.ID,
    "✅ Votre demande pour Cardiologie a été programmée le 15/03 à 09:00",
)
```

---

## DecryptionCache

> **Source:** [`cache.go`](../internal/service/cache.go)

Thread-safe in-memory cache that reduces redundant AES-256-GCM decryption operations on patient PII.

### Design

```
Key format: decrypt:{entity}:{id}:{field}
Example:    decrypt:patient:550e8400-e29b-41d4-a716-446655440000:cin
```

### Struct

```go
type DecryptionCache struct {
    mu       sync.RWMutex
    entries  map[string]*CacheEntry
    ttl      time.Duration  // Default: 5 minutes
    maxSize  int            // Default: 10,000 entries
}
```

### Functions

| Function                        | Description                                         |
| ------------------------------- | --------------------------------------------------- |
| `NewDecryptionCache(ttl, max)` | Constructor with configurable TTL and max size      |
| `Get(key)`                     | Retrieve cached value (returns `""` if miss/expired)|
| `Set(key, value)`              | Store with default TTL                               |
| `SetWithTTL(key, value, ttl)`  | Store with custom TTL                                |
| `Invalidate(key)`              | Remove single entry                                  |
| `InvalidateEntity(prefix)`     | Remove all entries matching entity prefix            |
| `InvalidatePatient(patientID)` | Remove all entries for a patient                     |
| `InvalidateReferral(refID)`    | Remove all entries for a referral                    |
| `Clear()`                      | Flush entire cache                                   |
| `Stats()`                      | Return entry count and memory usage estimate         |

### Eviction Policy

When the cache reaches `maxSize`, approximately **10% of the oldest entries** are evicted (sorted by creation time). This is a simple LRU-approximation strategy.

### Usage

```go
// In handler decryption helper
cacheKey := fmt.Sprintf("decrypt:patient:%s:cin", patient.ID.String())
if cached := h.DecryptionCache.Get(cacheKey); cached != "" {
    return cached
}
decrypted, err := h.Crypto.Decrypt(patient.CIN)
if err == nil {
    h.DecryptionCache.Set(cacheKey, decrypted)
}
```

---

## Architecture Diagram

```
┌────────────────────────────────────────────────────────────────┐
│                      Handler Layer                             │
│                                                                │
│  CreateReferral ──▶ NotificationService.Create()               │
│       │                                                        │
│       ├─ (async) ──▶ AIService.SummarizeSymptoms()             │
│                                                                │
│  ScheduleReferral ──▶ NotificationService.Create()             │
│       │                                                        │
│       ├─ (async) ──▶ WhatsAppService.SendAppointment()         │
│       │                  └──▶ Templates.AppointmentTemplate()  │
│       │                                                        │
│  GetReferral ──▶ DecryptionCache.Get() ──▶ AES.Decrypt()       │
│                                                                │
│  WebSocket Hub ──▶ BroadcastToDepartment() / BroadcastToUser() │
└────────────────────────────────────────────────────────────────┘
```
