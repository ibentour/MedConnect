# Services — API, WebSocket & Cache

> **Sources:**
> - [`frontend/src/services/api.js`](../src/services/api.js) (327 lines)
> - [`frontend/src/services/websocket.js`](../src/services/websocket.js) (147 lines)
> - [`frontend/src/services/cache.js`](../src/services/cache.js) (141 lines)

---

## Overview

Three service modules handle all backend communication:

| Service        | Purpose                                    | Pattern              |
| -------------- | ------------------------------------------ | -------------------- |
| `api.js`       | HTTP API client with caching & auth        | Axios + interceptors |
| `websocket.js` | Real-time event subscriptions              | WebSocket + pub/sub  |
| `cache.js`     | In-memory TTL cache for GET responses      | Map-based cache      |

---

## API Service (`api.js`)

### Axios Instance

```js
const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL || 'http://localhost:3000/api',
});
```

### Request Interceptor

Attaches JWT Bearer token from `localStorage` and serves cached GET responses:

```js
api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) config.headers.Authorization = `Bearer ${token}`;
    
    if (config.method === 'get') {
        const cached = getCached(generateCacheKey(config));
        if (cached) {
            config.adapter = () => Promise.resolve({ data: cached, status: 200 });
        }
    }
    return config;
});
```

### Response Interceptor

Caches GET responses and handles 401 errors:

```js
api.interceptors.response.use((response) => {
    if (response.config.method === 'get') {
        setCached(key, response.data, getCacheTTL(url));
    }
    return response;
}, (error) => {
    if (error.response?.status === 401) {
        localStorage.clear();
        window.location.href = '/login';
    }
    return Promise.reject(error);
});
```

### Cache TTL Configuration

| Endpoint Pattern           | TTL      |
| -------------------------- | -------- |
| `/directory`               | 5 min    |
| `/admin/departments`       | 5 min    |
| `/admin/stats`             | 1 min    |
| `/analyst/stats/*`         | 1 min    |
| All other GET endpoints    | 30 sec   |

### Exported Functions

#### Authentication

| Function                      | Method | Endpoint    | Description           |
| ----------------------------- | ------ | ----------- | --------------------- |
| `login(username, password)`   | POST   | `/login`    | Authenticate user     |
| `logout()`                    | —      | —           | Clear localStorage + cache |

#### Directory

| Function             | Method | Endpoint       | Description          |
| -------------------- | ------ | -------------- | -------------------- |
| `getDirectory()`     | GET    | `/directory`   | Department listing   |

#### Referrals

| Function                          | Method | Endpoint                            | Description              |
| --------------------------------- | ------ | ----------------------------------- | ------------------------ |
| `suggestDepartment(payload)`      | POST   | `/referrals/suggest`                | AI triage suggestion     |
| `createReferral(payload)`         | POST   | `/referrals`                        | Create referral          |
| `uploadAttachments(refID, files)` | POST   | `/referrals/:id/attachments`        | Upload files (multipart) |
| `getQueue()`                      | GET    | `/queue`                            | CHU pending queue        |
| `getReferralDetails(id)`          | GET    | `/referrals/:id`                    | Single referral          |
| `scheduleReferral(id, date)`      | PATCH  | `/referrals/:id/schedule`           | Schedule appointment     |
| `redirectReferral(id, deptId, reason)` | PATCH | `/referrals/:id/redirect`       | Redirect referral        |
| `denyReferral(id, reason)`        | PATCH  | `/referrals/:id/deny`               | Deny referral            |
| `getHistory()`                    | GET    | `/history`                          | Referral history         |
| `rescheduleReferral(id, date)`    | PATCH  | `/referrals/:id/reschedule`         | Reschedule appointment   |
| `cancelReferral(id)`              | PATCH  | `/referrals/:id/cancel`             | Cancel referral          |

#### Attachments

| Function                      | Description                                      |
| ----------------------------- | ------------------------------------------------ |
| `getAttachmentUrl(id)`        | Returns `{url, authHeader}` for inline viewing   |
| `getAttachmentUrlLegacy(id)`  | Returns plain URL (no auth)                      |
| `downloadAttachment(id)`      | Download as blob (uses `fetch`, not axios)       |

#### Notifications

| Function                        | Method | Endpoint                      | Description         |
| ------------------------------- | ------ | ----------------------------- | ------------------- |
| `checkNotifications()`          | GET    | `/notifications`              | Get notifications   |
| `markNotificationRead(id)`      | PATCH  | `/notifications/:id/read`     | Mark as read        |

#### Admin

| Function                         | Method | Endpoint                    | Description            |
| -------------------------------- | ------ | --------------------------- | ---------------------- |
| `getAdminStats()`                | GET    | `/admin/stats`              | Dashboard statistics   |
| `getUsers()`                     | GET    | `/admin/users`              | User list              |
| `createUser(payload)`            | POST   | `/admin/users`              | Create user            |
| `deleteUser(id)`                 | DELETE | `/admin/users/:id`          | Delete user            |
| `getAdminDepartments()`          | GET    | `/admin/departments`        | Department list        |
| `createDepartment(payload)`      | POST   | `/admin/departments`        | Create department      |
| `updateDepartment(id, payload)`  | PATCH  | `/admin/departments/:id`    | Update department      |
| `deleteDepartment(id)`           | DELETE | `/admin/departments/:id`    | Delete department      |

#### Analyst

| Function                    | Method | Endpoint                        | Description             |
| --------------------------- | ------ | ------------------------------- | ----------------------- |
| `getAnalystStats()`         | GET    | `/analyst/stats/departments`    | Department analytics    |
| `getAnalystDoctorStats()`   | GET    | `/analyst/stats/doctors`        | Doctor analytics        |

### Cache Invalidation

Mutations automatically invalidate related caches:

```js
export async function createReferral(payload) {
    const { data } = await api.post('/referrals', payload);
    invalidateCache('/admin');        // Admin stats may change
    invalidateCache('/analyst');      // Analyst stats may change
    invalidateCache('/directory');    // Directory may update
    return data;
}
```

---

## WebSocket Service (`websocket.js`)

### Connection

```js
const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:3000/ws';
```

WebSocket connects with JWT authentication via query parameter:

```js
ws = new WebSocket(`${WS_URL}?token=${token}`);
```

### Event Types

```js
export const WS_EVENTS = {
    NEW_REFERRAL: 'new_referral',
    REFERRAL_UPDATE: 'referral_updated',
    REFERRAL_REDIRECT: 'referral_redirected',
};
```

### Auto-Reconnect

- Up to **5 retry attempts**
- **3-second delay** between retries
- Resets retry counter on successful connection

### API

| Function                           | Description                                     |
| ---------------------------------- | ----------------------------------------------- |
| `connect()`                        | Establish WebSocket connection                  |
| `disconnect()`                     | Close connection and clear state                |
| `subscribe(handler)`               | Register event handler, returns unsubscribe fn  |
| `onConnectionChange(handler)`      | Monitor connection status changes               |
| `isConnected()`                    | Returns current connection status               |
| `send(event, data)`                | Send message to server                          |

### Usage in Components

```jsx
useEffect(() => {
    const unsubscribe = subscribe((message) => {
        if (message.event === WS_EVENTS.NEW_REFERRAL) {
            setQueue(prev => [message.payload, ...prev]);
            showToast('Nouvelle référence reçue');
        }
    });
    return unsubscribe;
}, []);
```

---

## Cache Service (`cache.js`)

In-memory TTL cache using a JavaScript `Map`.

### API

| Function                          | Description                                      |
| --------------------------------- | ------------------------------------------------ |
| `getCached(key)`                  | Returns data if not expired, `null` otherwise    |
| `setCached(key, data, ttlMs)`     | Store with TTL                                   |
| `invalidateCache(pattern)`        | Clear matching entries or all if no pattern      |
| `invalidateDirectoryCache()`      | Clear `/directory` entries                       |
| `invalidateAdminCache()`          | Clear `/admin` entries                           |
| `invalidateAnalystCache()`        | Clear `/analyst` entries                         |

### Cache Key Generation

```js
generateCacheKey(method, url, params)
// → "GET:/api/admin/users:limit=20&offset=0"
```

### TTL Lookup

```js
getCacheTTL(url)
// Matches URL patterns to return appropriate TTL in milliseconds
```

### Cleanup

Expired entries are cleaned up on every `getCached` call (lazy eviction).
