# Components — Shared UI

> **Sources:**
> - [`frontend/src/components/Navbar.jsx`](../src/components/Navbar.jsx)
> - [`frontend/src/components/StatusBadge.jsx`](../src/components/StatusBadge.jsx)
> - [`frontend/src/components/Security.jsx`](../src/components/Security.jsx)

---

## Overview

Three shared components provide navigation, status display, and security protection across the application.

---

## Navbar

> **Source:** [`Navbar.jsx`](../src/components/Navbar.jsx) (179 lines)

Responsive navigation with desktop top-bar and mobile bottom-tab bar.

### Props

None — uses `useAuth()` and `useLocation()` hooks.

### Role-Based Navigation

| Role           | Desktop Items                                                | Mobile Items                              |
| -------------- | ------------------------------------------------------------ | ----------------------------------------- |
| `LEVEL_2_DOC`  | Tableau de Bord, Annuaire, Nouveau Dossier, Historique       | Dashboard, Annuaire, **+** (FAB), Historique |
| `CHU_DOC`      | Tableau de Bord, Annuaire, Historique                        | Dashboard, Annuaire, Historique           |
| `SUPER_ADMIN`  | Annuaire, Analytique, Super Admin, Logs                      | Annuaire, Analytique, Admin, Logs         |
| `ANALYST`      | Analytique, Annuaire                                         | Analytique, Annuaire                      |

### Features

- **Active link highlighting** — Uses teal brand color (`text-teal-600`, `bg-teal-50`)
- **User info display** — Username + role badge in top-right (desktop)
- **Mobile FAB** — Floating action button for "Nouveau Dossier" (Level 2 only)
- **Logout button** — Calls `logout()` from AuthContext

### Desktop Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ [Logo]  Dashboard  Directory  Nouveau   Historique   [User] [↗] │
│                                        ▲ active                 │
└─────────────────────────────────────────────────────────────────┘
```

### Mobile Layout

```
┌─────────────────────────────────┐
│                                 │
│         Page Content            │
│                                 │
│                                 │
├────────┬────────┬────┬──────────┤
│ Dash   │ Dir    │ +  │ History  │
└────────┴────────┴────┴──────────┘
```

---

## StatusBadge

> **Source:** [`StatusBadge.jsx`](../src/components/StatusBadge.jsx) (38 lines)

Color-coded badge component for referral status display.

### Props

| Prop       | Type     | Description               |
| ---------- | -------- | ------------------------- |
| `status`   | `string` | Referral status enum      |
| `className`| `string` | Additional CSS classes    |

### Status Mappings

| Status       | Color   | Icon           | Label (French)          |
| ------------ | ------- | -------------- | ----------------------- |
| `PENDING`    | Yellow  | `Clock`        | En attente              |
| `SCHEDULED`  | Green   | `CheckCircle2` | Programmé               |
| `REDIRECTED` | Blue    | `ArrowRightLeft`| Redirigé               |
| `DENIED`     | Red     | `XCircle`      | Refusé                  |
| `CANCELED`   | Gray    | `AlertTriangle`| Annulé                  |

### Usage

```jsx
<StatusBadge status="PENDING" />
// → <span class="bg-yellow-100 text-yellow-700 ...">
//     <Clock /> En attente
//   </span>
```

---

## Security

> **Source:** [`Security.jsx`](../src/components/Security.jsx) (68 lines)

Two security-focused components for data protection.

### Watermark

Tiled SVG watermark overlay to deter screenshots.

**Features:**
- Displays `username — timestamp` text tiled across the screen
- Uses CSS `--watermark-svg` custom property for the background
- 3% opacity — visible but non-intrusive
- Fixed positioning, covers entire viewport

```jsx
<Watermark />
// → Creates a fixed overlay with tiled watermark text
```

### AntiLeak

Wrapper component that prevents data exfiltration:

| Protection         | Implementation                                        |
| ------------------ | ----------------------------------------------------- |
| Disable selection  | `className="select-none"` + cross-browser CSS         |
| Block right-click  | `onContextMenu={(e) => e.preventDefault()}`           |
| Block copy         | Replaces clipboard with "Confidential Medical Record" |
| Block drag         | `onDragStart={(e) => e.preventDefault()}`             |

```jsx
<AntiLeak>
  <SensitiveDataComponent />
</AntiLeak>
```

> **Applied to:** `Directory` page and CHU dashboard detail panel.

---

## Reusable Patterns

### Common Tailwind Classes

| Purpose          | Classes                                                                        |
| ---------------- | ------------------------------------------------------------------------------ |
| Card container   | `bg-white rounded-2xl shadow-sm border border-gray-100 p-6`                    |
| Primary button   | `bg-teal-600 hover:bg-teal-700 text-white px-4 py-2 rounded-xl`                |
| Danger button    | `bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-xl`                  |
| Input field      | `border border-gray-200 rounded-xl px-3 py-2 focus:ring-2 focus:ring-teal-500` |
| Badge            | `px-2 py-1 rounded-full text-xs font-medium`                                   |
| Section heading  | `text-lg font-semibold text-gray-800`                                          |
