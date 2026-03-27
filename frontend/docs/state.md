# State Management — AuthContext

> **Source:** [`frontend/src/context/AuthContext.jsx`](../src/context/AuthContext.jsx)

---

## Overview

The application uses **React Context** for global state management. There is no Redux, Zustand, or other state library. The only global state is **authentication** — all other state is component-local via `useState`.

---

## AuthContext

### Context Shape

```jsx
{
  user: {
    id: string,
    username: string,
    role: "LEVEL_2_DOC" | "CHU_DOC" | "SUPER_ADMIN" | "ANALYST",
    facility_name: string,
    department_id: string | null
  } | null,
  isAuthenticated: boolean,
  login: (username, password) => Promise<void>,
  logout: () => void
}
```

### AuthProvider

Wraps the entire application in `App.jsx`:

```jsx
<AuthProvider>
  <BrowserRouter>
    <AppRoutes />
  </BrowserRouter>
</AuthProvider>
```

### Initialization

On mount, `AuthProvider` checks `localStorage` for existing session:

```jsx
useEffect(() => {
    const storedUser = localStorage.getItem('user');
    const storedToken = localStorage.getItem('token');
    if (storedUser && storedToken) {
        setUser(JSON.parse(storedUser));
    }
    setLoading(false);
}, []);
```

### login(username, password)

1. Calls `loginApi(username, password)` from `services/api.js`
2. Stores user and token in `localStorage`
3. Sets `user` state → `isAuthenticated` becomes `true`
4. Connects WebSocket: `wsConnect()`
5. Redirects based on role:
   - `LEVEL_2_DOC`, `CHU_DOC` → `/dashboard`
   - Others → `/directory`

### logout()

1. Calls `logoutApi()` — clears `localStorage` and cache
2. Sets `user` to `null` → `isAuthenticated` becomes `false`
3. Disconnects WebSocket: `wsDisconnect()`
4. Redirects to `/login`

### useAuth() Hook

```jsx
export function useAuth() {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within AuthProvider');
    }
    return context;
}
```

### Loading State

While checking `localStorage`, the provider shows a centered loading spinner:

```jsx
if (loading) {
    return (
        <div className="min-h-screen flex items-center justify-center">
            <div className="animate-spin h-8 w-8 border-4 border-teal-500 rounded-full border-t-transparent" />
        </div>
    );
}
```

---

## Component-Local State

Beyond AuthContext, all state is managed locally with `useState`. No global store exists for:

| Data                    | State Location              | Pattern                    |
| ----------------------- | --------------------------- | -------------------------- |
| Referral queue          | `chu/Dashboard.jsx`         | `useState` + `useEffect`   |
| Admin users/departments | `admin/Dashboard.jsx`       | `useState` + `useEffect`   |
| Notifications           | `level2/Notifications.jsx`  | `useState` + `useEffect`   |
| Form state              | `CreateReferral.jsx`        | Multiple `useState` calls  |
| Search/filter           | `shared/History.jsx`        | `useState` + `useMemo`     |
| Modal visibility        | Various                     | `useState(boolean)`        |

---

## Derived State

No `useReducer` or memoized selectors are used. Derived state is computed inline:

```jsx
// Filtering referrals by search term
const filteredReferrals = referrals.filter(r =>
    r.patient_name.toLowerCase().includes(search.toLowerCase())
);

// Counting unread notifications
const unreadCount = notifications.filter(n => !n.is_read).length;
```

---

## Data Flow Diagram

```
localStorage (persist)
      │
      ▼
AuthProvider (Context)
      │
      ├── login() ──▶ api.login() ──▶ Backend
      │                 │
      │                 ├── stores token + user in localStorage
      │                 ├── sets user state
      │                 └── wsConnect()
      │
      ├── logout() ──▶ api.logout()
      │                 │
      │                 ├── clears localStorage + cache
      │                 ├── sets user = null
      │                 └── wsDisconnect()
      │
      └── useAuth() ──▶ Components read: user, isAuthenticated
                              │
                              ├── ProtectedRoute (auth check)
                              ├── Navbar (user display)
                              ├── RoleBasedDashboard (role routing)
                              └── Pages (user info display)
```

---

## WebSocket State

WebSocket connection state is managed in `services/websocket.js`, not in React state. Components subscribe via `useEffect`:

```jsx
useEffect(() => {
    const unsubscribe = subscribe((message) => {
        if (message.event === 'new_referral') {
            setQueue(prev => [...prev, message.payload]);
            showToast('Nouvelle référence reçue');
        }
    });
    return unsubscribe;
}, []);
```

---

## Recommendations

1. **Extract shared state** — Consider React Query or SWR for server state (replaces manual `useEffect` + `useState` for API data)
2. **Form library** — React Hook Form or Formik would reduce boilerplate in `CreateReferral.jsx`
3. **Shared utility state** — `calculateAge()` is duplicated; extract to a shared utils file
4. **Global toast state** — Toast notifications could use a context instead of local state per component
