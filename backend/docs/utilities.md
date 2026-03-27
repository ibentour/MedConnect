# Utilities — Middleware & Crypto

> **Sources:**
> - [`backend/internal/middleware/auth.go`](../internal/middleware/auth.go)
> - [`backend/internal/middleware/validation.go`](../internal/middleware/validation.go)
> - [`backend/internal/middleware/rate_limit.go`](../internal/middleware/rate_limit.go)
> - [`backend/internal/middleware/security_headers.go`](../internal/middleware/security_headers.go)
> - [`backend/internal/middleware/audit.go`](../internal/middleware/audit.go)
> - [`backend/internal/crypto/aes_gcm.go`](../internal/crypto/aes_gcm.go)
> - [`backend/internal/websocket/hub.go`](../internal/websocket/hub.go)

---

## Overview

Utility packages provide cross-cutting concerns: authentication, input validation, rate limiting, security headers, audit logging, encryption, and real-time WebSocket communication.

---

## Middleware

### JWT Authentication (`auth.go`)

Provides JWT-based authentication and role-based access control.

#### Claims

```go
type Claims struct {
    UserID       string `json:"user_id"`
    Username     string `json:"username"`
    Role         string `json:"role"`
    DeptID       string `json:"dept_id"`
    FacilityName string `json:"facility_name"`
    jwt.RegisteredClaims
}
```

#### Key Functions

| Function                                                     | Description                                      |
| ------------------------------------------------------------ | ------------------------------------------------ |
| `GenerateToken(user *models.User, secret string) (string, error)` | Generate 24-hour HS256 JWT                  |
| `JWTAuthMiddleware(secret string) gin.HandlerFunc`           | Validate Bearer token or `?token=` query param    |
| `RBACMiddleware(allowedRoles ...string) gin.HandlerFunc`     | Gate endpoint by role                             |
| `GetUserIDFromContext(c *gin.Context) string`                | Extract user ID from JWT claims                   |
| `GetUsernameFromContext(c *gin.Context) string`              | Extract username from JWT claims                  |
| `GetUserRoleFromContext(c *gin.Context) string`              | Extract role from JWT claims                      |
| `GetDeptIDFromContext(c *gin.Context) string`                | Extract department ID from JWT claims             |
| `ValidateTokenString(token, secret string) (*Claims, error)` | Standalone token validation                       |

#### Token Flow

```
Client                 Server
  │                       │
  ├── POST /api/login ───▶│ Validate credentials
  │◀── { token } ─────────┤ Return JWT (24h expiry)
  │                       │
  ├── GET /api/queue ────▶│ JWTAuthMiddleware
  │   Authorization:      │   ├─ Parse & validate token
  │   Bearer <token>      │   ├─ Inject claims into context
  │                       │   └─ RBACMiddleware(["CHU_DOC"])
  │◀── 200 OK ────────────┘
```

### Input Validation (`validation.go`)

Input sanitization and structured validation error handling.

#### Key Functions

| Function                                            | Description                                        |
| --------------------------------------------------- | -------------------------------------------------- |
| `SanitizeInput(input string) string`                | Remove null bytes, control chars, trim whitespace   |
| `SanitizeStruct(s interface{}) interface{}`         | Sanitize all string fields in a struct              |
| `BindAndValidate(c, dest) error`                    | Bind JSON → Sanitize → Validate pipeline            |
| `ParseValidationErrors(err) []ValidationError`      | Convert validator errors to structured format       |
| `ValidatePhoneNumber(phone string) (string, error)` | International phone validation                      |
| `ValidateMoroccanPhoneNumber(phone string)`         | Moroccan-specific (06XX/07XX → +2126XX/+2127XX)    |

#### ValidationError

```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code"`
}
```

#### Supported Phone Formats

| Country | Format Example  | Prefix |
| ------- | --------------- | ------ |
| Morocco | 0612345678      | +212   |
| France  | 0612345678      | +33    |
| Spain   | 612345678       | +34    |
| Germany | 015112345678    | +49    |
| Italy   | 3123456789      | +39    |
| UK      | 07123456789     | +44    |
| US/CA   | 2125551234      | +1     |
| Algeria | 0512345678      | +213   |
| Tunisia | 20123456        | +216   |

### Rate Limiting (`rate_limit.go`)

In-memory IP-based rate limiting with periodic cleanup.

```go
type RateLimiter struct {
    mu       sync.Mutex
    requests map[string][]time.Time
    limit    int
    window   time.Duration
}
```

| Function                          | Description                              |
| --------------------------------- | ---------------------------------------- |
| `NewRateLimiter(limit int, window)` | Create limiter (e.g., 10 requests/minute) |
| `Middleware() gin.HandlerFunc`     | Gin middleware that returns 429 on excess |

**Applied to:** `POST /api/login` — 10 requests per minute per IP.

### Security Headers (`security_headers.go`)

Sets security-focused HTTP headers on all responses.

| Header                        | Value                                             |
| ----------------------------- | ------------------------------------------------- |
| `X-Content-Type-Options`      | `nosniff`                                         |
| `X-Frame-Options`             | `DENY`                                            |
| `X-XSS-Protection`            | `1; mode=block`                                   |
| `Referrer-Policy`             | `strict-origin-when-cross-origin`                 |
| `X-DNS-Prefetch-Control`      | `off`                                             |
| `Permissions-Policy`          | `geolocation=(), microphone=(), camera=()`        |
| `Content-Security-Policy`     | `default-src 'self'; ...` (full policy)           |

### Audit Logging (`audit.go`)

Asynchronous request audit logging for Moroccan Law 09-08 compliance.

**Captures per request:**
- User ID, Username
- HTTP method + path (as action)
- `:id` route parameter (as target)
- IP address, User agent
- HTTP response status
- Timestamp

> All writes are non-blocking (goroutines) to avoid impacting request latency.

---

## Crypto — AES-256-GCM (`aes_gcm.go`)

Authenticated encryption for patient PII at rest.

### Struct

```go
type AESCrypto struct {
    key []byte  // 256-bit key (32 bytes)
}
```

### Functions

| Function                          | Description                                     |
| --------------------------------- | ----------------------------------------------- |
| `NewAESCrypto(hexKey string)`    | Constructor — validates 64-char hex key          |
| `MustNewAESCrypto(hexKey string)`| Panics on invalid key                           |
| `Encrypt(plaintext string)`      | Encrypt → `hex(nonce):hex(ciphertext+tag)`      |
| `Decrypt(encoded string)`        | Decrypt → plaintext                             |

### Encrypted Format

```
<12-byte-nonce-hex>:<ciphertext+GCM-tag-hex>
```

Example:
```
4e6f6e636531323334353637:656e637279707465644461746148455245...
```

### Error Types

| Error                       | Condition                    |
| --------------------------- | ---------------------------- |
| `ErrInvalidKeyLength`       | Key is not 64 hex chars      |
| `ErrInvalidKeyHex`          | Non-hex characters in key    |
| `ErrInvalidCiphertextFmt`   | Missing `:` separator        |
| `ErrCiphertextTooShort`     | Ciphertext < 12 bytes (nonce)|
| `ErrDecryptionFailed`       | Authentication tag mismatch  |

### Usage

```go
crypto := crypto.MustNewAESCrypto(os.Getenv("AES_KEY"))

encrypted, _ := crypto.Encrypt("Patient CIN: AB123456")
// → "4e6f6e636531323334353637:656e63727970746564..."

decrypted, _ := crypto.Decrypt(encrypted)
// → "Patient CIN: AB123456"
```

---

## WebSocket Hub (`websocket/hub.go`)

Real-time notification delivery using the Hub pattern.

### Architecture

```
┌────────────────────────────────────────────┐
│                   Hub                      │
│                                            │
│  clients: map[string]*Client  (by userID)  │
│  register:    chan *Client                 │
│  unregister:  chan *Client                 │
│  broadcast:   chan *Message                │
│                                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ Client 1 │  │ Client 2 │  │ Client 3 │  │
│  │ (user_A) │  │ (user_B) │  │ (user_A) │  │
│  └──────────┘  └──────────┘  └──────────┘  │
└────────────────────────────────────────────┘
```

### Key Functions

| Function                                   | Description                              |
| ------------------------------------------ | ---------------------------------------- |
| `NewHub() *Hub`                            | Create and start hub goroutine           |
| `GinWebSocket(hub, jwtSecret) gin.HandlerFunc` | Gin handler for WS upgrade           |
| `BroadcastToDepartment(deptID, event, payload)` | Send to all dept clients         |
| `BroadcastToUser(userID, event, payload)`  | Send to specific user's clients          |

### Client Lifecycle

1. **Connect:** Client connects via `GET /ws?token=<jwt>`, JWT is validated
2. **Register:** Client added to hub's client map
3. **Read pump:** Reads incoming messages, handles ping/pong (60s deadline)
4. **Write pump:** Writes queued messages, sends 30s keepalive pings
5. **Disconnect:** Client removed from hub, connection closed

### Event Types

| Event                 | Payload         | When Sent                        |
| --------------------- | --------------- | -------------------------------- |
| `new_referral`        | Referral data   | New referral created in dept     |
| `referral_updated`    | Status update   | Referral status changed          |
| `referral_redirected` | Redirect details| Referral redirected to dept      |
