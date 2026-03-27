# Routing — React Router v6

> **Source:** [`frontend/src/App.jsx`](../src/App.jsx)

---

## Overview

Routing is handled by **React Router DOM v6** with nested routes, layout routes, and role-based protected routes.

---

## Route Configuration

```jsx
<Routes>
  {/* Public */}
  <Route path="/login" element={<Login />} />

  {/* Authenticated — All Roles */}
  <Route path="/" element={<ProtectedRoute><MainLayout><RoleBasedDashboard /></MainLayout></ProtectedRoute>} />
  <Route path="/dashboard" element={<ProtectedRoute><MainLayout><RoleBasedDashboard /></MainLayout></ProtectedRoute>} />
  <Route path="/directory" element={<ProtectedRoute><MainLayout><Directory /></MainLayout></ProtectedRoute>} />

  {/* Level 2 Doctor */}
  <Route path="/referrals/new" element={<ProtectedRoute allowedRoles={['LEVEL_2_DOC']}><MainLayout><CreateReferral /></MainLayout></ProtectedRoute>} />
  <Route path="/notifications" element={<ProtectedRoute allowedRoles={['LEVEL_2_DOC']}><MainLayout><Notifications /></MainLayout></ProtectedRoute>} />

  {/* Level 2 + CHU Doctor */}
  <Route path="/history" element={<ProtectedRoute allowedRoles={['LEVEL_2_DOC', 'CHU_DOC']}><MainLayout><HistoryTab /></MainLayout></ProtectedRoute>} />

  {/* Analyst + Super Admin */}
  <Route path="/analyst" element={<ProtectedRoute allowedRoles={['ANALYST', 'SUPER_ADMIN']}><MainLayout><AnalystDashboard /></MainLayout></ProtectedRoute>} />

  {/* Super Admin */}
  <Route path="/admin" element={<ProtectedRoute allowedRoles={['SUPER_ADMIN']}><MainLayout><AdminDashboard /></MainLayout></ProtectedRoute>} />
  <Route path="/admin/audit-logs" element={<ProtectedRoute allowedRoles={['SUPER_ADMIN']}><MainLayout><AuditLogs /></MainLayout></ProtectedRoute>} />

  {/* Catch-all */}
  <Route path="*" element={<Navigate to="/dashboard" />} />
</Routes>
```

---

## Route Map

| Path                | Page Component      | Roles Allowed               | Layout       |
| ------------------- | ------------------- | --------------------------- | ------------ |
| `/login`            | `Login`             | Public                      | None         |
| `/`                 | `RoleBasedDashboard`| All authenticated           | MainLayout   |
| `/dashboard`        | `RoleBasedDashboard`| All authenticated           | MainLayout   |
| `/directory`        | `Directory`         | All authenticated           | MainLayout   |
| `/referrals/new`    | `CreateReferral`    | `LEVEL_2_DOC`               | MainLayout   |
| `/notifications`    | `Notifications`     | `LEVEL_2_DOC`               | MainLayout   |
| `/history`          | `HistoryTab`        | `LEVEL_2_DOC`, `CHU_DOC`    | MainLayout   |
| `/analyst`          | `AnalystDashboard`  | `ANALYST`, `SUPER_ADMIN`    | MainLayout   |
| `/admin`            | `AdminDashboard`    | `SUPER_ADMIN`               | MainLayout   |
| `/admin/audit-logs` | `AuditLogs`         | `SUPER_ADMIN`               | MainLayout   |
| `*`                 | Redirect → `/dashboard` | Any                     | None         |

---

## ProtectedRoute

```jsx
function ProtectedRoute({ children, allowedRoles }) {
    const { isAuthenticated, user } = useAuth();
    
    if (!isAuthenticated) return <Navigate to="/login" />;
    
    if (allowedRoles && !allowedRoles.includes(user.role)) {
        return <Navigate to="/directory" />;
    }
    
    return children;
}
```

### Redirect Behavior

| Condition                      | Redirect Target |
| ------------------------------ | --------------- |
| Not authenticated              | `/login`        |
| Authenticated but wrong role   | `/directory`    |

---

## RoleBasedDashboard

Dynamically renders the correct dashboard based on `user.role`:

```jsx
function RoleBasedDashboard() {
    const { user } = useAuth();
    switch (user.role) {
        case 'LEVEL_2_DOC':  return <Level2Dashboard />;
        case 'CHU_DOC':      return <ChuDashboard />;
        case 'SUPER_ADMIN':  return <AdminDashboard />;
        case 'ANALYST':      return <AnalystDashboard />;
        default:             return <Navigate to="/directory" />;
    }
}
```

---

## Navigation Flow

```
Login Success
     │
     ├─ LEVEL_2_DOC / CHU_DOC → /dashboard (role-specific dashboard)
     └─ SUPER_ADMIN / ANALYST → /directory

Navbar Links (by role)
     │
     ├─ LEVEL_2_DOC:  /dashboard, /directory, /referrals/new, /history
     ├─ CHU_DOC:      /dashboard, /directory, /history
     ├─ SUPER_ADMIN:  /directory, /analyst, /admin, /admin/audit-logs
     └─ ANALYST:      /analyst, /directory

Programmatic Navigation
     │
     ├─ CreateReferral success → /dashboard
     ├─ Logout → /login
     ├─ 401 response → /login
     └─ Unknown route → /dashboard
```

---

## useNavigate Usage

Programmatic navigation is used for:

```jsx
// After successful referral creation
const navigate = useNavigate();
navigate('/dashboard');

// After login
navigate(user.role === 'LEVEL_2_DOC' ? '/dashboard' : '/directory');

// After logout
navigate('/login');
```

---

## useLocation Usage

The `Navbar` uses `useLocation()` to highlight the active link:

```jsx
const location = useLocation();
const isActive = (path) => location.pathname === path;
```
