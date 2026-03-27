# Architecture — Frontend

> **Source:** [`frontend/src/App.jsx`](../src/App.jsx), [`frontend/src/main.jsx`](../src/main.jsx)

---

## Overview

MedConnect frontend is a **React 18 SPA** built with Vite, styled with Tailwind CSS, and routed via React Router v6. The UI is primarily in **French** (Moroccan context) and follows a role-based dashboard pattern where each user role sees a different primary view.

---

## Tech Stack

| Technology           | Version | Purpose                           |
| -------------------- | ------- | --------------------------------- |
| React                | 18.3.1  | UI framework                      |
| Vite                 | 5.4.0   | Build tool & dev server           |
| React Router DOM     | 6.24.1  | Client-side routing               |
| Tailwind CSS         | 3.4.4   | Utility-first styling             |
| Axios                | 1.7.2   | HTTP client                       |
| Lucide React         | 0.395.0 | Icon library                      |
| clsx + tailwind-merge| 2.x     | Class merging (installed, unused) |

---

## Directory Structure

```
frontend/
├── index.html                    # Vite HTML entry point
├── package.json                  # Dependencies & scripts
├── vite.config.js                # Vite configuration
├── tailwind.config.js            # Tailwind custom theme
├── postcss.config.js             # PostCSS config
├── eslint.config.js              # ESLint flat config
├── public/                       # Static assets
│   ├── vite.svg
│   └── vite.png
└── src/
    ├── main.jsx                  # React entry point
    ├── App.jsx                   # Root component + routing
    ├── App.css                   # Legacy boilerplate (unused)
    ├── index.css                 # Tailwind directives + global styles
    ├── assets/                   # Static imports
    │   └── react.svg
    ├── context/                  # React Context providers
    │   └── AuthContext.jsx
    ├── services/                 # API & real-time services
    │   ├── api.js
    │   ├── websocket.js
    │   └── cache.js
    ├── components/               # Shared UI components
    │   ├── Navbar.jsx
    │   ├── StatusBadge.jsx
    │   └── Security.jsx
    └── pages/                    # Route-level page components
        ├── Login.jsx
        ├── Directory.jsx
        ├── admin/
        │   ├── Dashboard.jsx
        │   └── AuditLogs.jsx
        ├── level2/
        │   ├── Dashboard.jsx
        │   ├── CreateReferral.jsx
        │   └── Notifications.jsx
        ├── chu/
        │   └── Dashboard.jsx
        ├── analyst/
        │   └── Stats.jsx
        └── shared/
            └── History.jsx
```

---

## Component Hierarchy

```
<StrictMode>
  └── <AuthProvider>              # Auth state (Context)
        └── <BrowserRouter>       # Routing
              └── <Routes>
                    ├── /login → <Login>
                    ├── / → <ProtectedRoute> → <MainLayout> → <RoleBasedDashboard>
                    │         ├── LEVEL_2_DOC  → <Level2Dashboard>
                    │         ├── CHU_DOC      → <ChuDashboard>
                    │         ├── SUPER_ADMIN  → <AdminDashboard>
                    │         └── ANALYST      → <AnalystDashboard>
                    ├── /dashboard → (same as /)
                    ├── /directory → <ProtectedRoute> → <MainLayout> → <Directory>
                    ├── /referrals/new → <ProtectedRoute[LEVEL_2_DOC]> → <CreateReferral>
                    ├── /notifications → <ProtectedRoute[LEVEL_2_DOC]> → <Notifications>
                    ├── /history → <ProtectedRoute[LEVEL_2_DOC, CHU_DOC]> → <HistoryTab>
                    ├── /analyst → <ProtectedRoute[ANALYST, SUPER_ADMIN]> → <AnalystDashboard>
                    ├── /admin → <ProtectedRoute[SUPER_ADMIN]> → <AdminDashboard>
                    ├── /admin/audit-logs → <ProtectedRoute[SUPER_ADMIN]> → <AuditLogs>
                    └── * → Navigate to /dashboard
```

### MainLayout

Wraps authenticated pages:
```
<MainLayout>
  ├── <Watermark />           # Anti-screenshot overlay
  ├── <Navbar />              # Responsive navigation
  └── <main>                  # max-w-7xl container
        └── {children}         # Page content
</MainLayout>
```

### ProtectedRoute

Guards routes requiring authentication:

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

---

## Data Flow

```
┌───────────────────────────────────────────────────────────────┐
│                      AuthContext                              │
│  ┌──────────┐   ┌──────────────────┐   ┌───────────────────┐  │
│  │  user    │   │ isAuthenticated  │   │ login() / logout()│  │
│  │  (state) │   │  (derived)       │   │  (actions)        │  │
│  └────┬─────┘   └────────┬─────────┘   └───────────────────┘  │
│       │                  │                                    │
└───────┼──────────────────┼────────────────────────────────────┘
        │                  │
        ▼                  ▼
  ┌──────────┐     ┌────────────────────────────────┐
  │ Navbar   │     │ ProtectedRoute / RoleDashboard │
  │ (reads)  │     │ (reads role for routing)       │
  └──────────┘     └────────────────────────────────┘

┌───────────────────────────────────────────────────────────────┐
│                      Service Layer                            │
│                                                               │
│  api.js ──── Axios (with interceptors) ──── Backend API       │
│     │                                                         │
│     ├── cache.js (in-memory TTL cache)                        │
│     │                                                         │
│  websocket.js ──── WebSocket ──── Real-time events            │
│     │                                                         │
│     └── subscribe() → Components (toast, refresh)             │
└───────────────────────────────────────────────────────────────┘
```

---

## Security Architecture

```
┌──────────────────────────────────────┐
│         Security Layer               │
│                                      │
│  ┌────────────────────────────────┐  │
│  │ Watermark (anti-screenshot)    │  │
│  │ SVG overlay with username+time │  │
│  └────────────────────────────────┘  │
│                                      │
│  ┌────────────────────────────────┐  │
│  │ AntiLeak (anti-copy)           │  │
│  │ - Disable text selection       │  │
│  │ - Block right-click            │  │
│  │ - Block copy (confidential)    │  │
│  │ - Block drag operations        │  │
│  └────────────────────────────────┘  │
│                                      │
│  ┌────────────────────────────────┐  │
│  │ CSS Print Block                │  │
│  │ @media print: display none     │  │
│  └────────────────────────────────┘  │
│                                      │
│  ┌────────────────────────────────┐  │
│  │ JWT Token (localStorage)       │  │
│  │ - Bearer header injection      │  │
│  │ - Auto-redirect on 401         │  │
│  └────────────────────────────────┘  │
└──────────────────────────────────────┘
```

---

## Build & Dev

```bash
# Development
npm run dev         # Start Vite dev server (port 5173)

# Production
npm run build       # Build to dist/
npm run preview     # Preview production build

# Linting
npm run lint        # ESLint check
```

### Vite Config

Minimal configuration — just enables React plugin:

```js
export default defineConfig({
  plugins: [react()],
})
```

### Environment Variables

| Variable          | Default                      | Description          |
| ----------------- | ---------------------------- | -------------------- |
| `VITE_API_URL`    | `http://localhost:3000/api`  | Backend API base URL |
| `VITE_WS_URL`     | `ws://localhost:3000/ws`     | WebSocket URL        |
