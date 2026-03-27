# Styling — Tailwind CSS Conventions

> **Sources:**
> - [`frontend/tailwind.config.js`](../tailwind.config.js)
> - [`frontend/src/index.css`](../src/index.css)
> - [`frontend/postcss.config.js`](../postcss.config.js)

---

## Overview

The frontend uses **Tailwind CSS 3.4** as the primary styling system. There is no CSS-in-JS, no SCSS, and no component library (e.g., Material UI, Chakra). All styling is done through Tailwind utility classes applied directly in JSX.

---

## Tailwind Configuration

> **Source:** [`tailwind.config.js`](../tailwind.config.js)

### Custom Theme

```js
theme: {
    extend: {
        fontFamily: {
            sans: ['Inter', 'system-ui', 'sans-serif'],
        },
        colors: {
            brand: {
                50: '#f0fdfa', 100: '#ccfbf1', 200: '#99f6e4', 300: '#5eead4',
                400: '#2dd4bf', 500: '#14b8a6', 600: '#0d9488', 700: '#0f766e',
                800: '#115e59', 900: '#134e4a',
            },
            chu: {
                50: '#eef2ff', 100: '#e0e7ff', 200: '#c7d2fe', 300: '#a5b4fc',
                400: '#818cf8', 500: '#6366f1', 600: '#4f46e5', 700: '#4338ca',
                800: '#3730a3', 900: '#312e81',
            },
        },
    },
},
```

### Color Palette

| Token       | Hex       | Usage                          |
| ----------- | --------- | ------------------------------ |
| `brand-500` | `#14b8a6` | Primary buttons, active links  |
| `brand-600` | `#0d9488` | Primary hover state            |
| `brand-50`  | `#f0fdfa` | Light backgrounds              |
| `chu-500`   | `#6366f1` | CHU-specific elements          |
| `chu-600`   | `#4f46e5` | CHU hover state                |
| `chu-50`    | `#eef2ff` | CHU light backgrounds          |

---

## Global CSS

> **Source:** [`index.css`](../src/index.css)

### Base Layer

```css
@layer base {
    body {
        @apply bg-gray-50 font-sans antialiased;
    }
}
```

### Utility Classes

```css
/* Disable text selection (security) */
.secure-text {
    @apply select-none;
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
}

/* Watermark overlay */
.watermark-overlay {
    @apply fixed inset-0 pointer-events-none z-50;
    background-image: var(--watermark-svg);
    background-repeat: repeat;
    opacity: 0.03;
}

/* Print blocking (security) */
@media print {
    * { display: none !important; }
}

/* Custom scrollbar */
::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-thumb { @apply bg-gray-300 rounded-full; }
::-webkit-scrollbar-track { @apply bg-transparent; }
```

---

## Design Patterns

### Cards

Consistent card styling across the application:

```jsx
<div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
```

| Property           | Value                           | Purpose             |
| ------------------ | ------------------------------- | ------------------- |
| Background         | `bg-white`                      | Clean surface       |
| Border radius      | `rounded-2xl` or `rounded-3xl`  | Soft, modern feel   |
| Shadow             | `shadow-sm`                     | Subtle elevation    |
| Border             | `border border-gray-100`        | Subtle definition   |
| Padding            | `p-6`                           | Comfortable spacing |

### Buttons

#### Primary Action
```jsx
<button className="bg-teal-600 hover:bg-teal-700 text-white px-4 py-2 rounded-xl font-medium transition-colors">
```

#### Secondary Action
```jsx
<button className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-4 py-2 rounded-xl font-medium transition-colors">
```

#### Danger Action
```jsx
<button className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-xl font-medium transition-colors">
```

### Input Fields

```jsx
<input className="w-full border border-gray-200 rounded-xl px-3 py-2 focus:ring-2 focus:ring-teal-500 focus:border-transparent outline-none transition-all" />
```

### Badges

```jsx
<span className="px-2 py-1 rounded-full text-xs font-medium bg-teal-100 text-teal-700">
```

### Section Headings

```jsx
<h2 className="text-lg font-semibold text-gray-800">
```

---

## Status Colors

| Status       | Background        | Text              | Usage                  |
| ------------ | ----------------- | ----------------- | ---------------------- |
| Pending      | `bg-yellow-100`   | `text-yellow-700` | Awaiting review        |
| Scheduled    | `bg-green-100`    | `text-green-700`  | Appointment set        |
| Redirected   | `bg-blue-100`     | `text-blue-700`   | Sent to other dept     |
| Denied       | `bg-red-100`      | `text-red-700`    | Refused                |
| Canceled     | `bg-gray-100`     | `text-gray-700`   | Canceled appointment   |

## Urgency Colors

| Urgency  | Dot Color   | Badge Background   |
| -------- | ----------- | ------------------ |
| LOW      | `bg-green-500`  | `bg-green-100 text-green-700`   |
| MEDIUM   | `bg-yellow-500` | `bg-yellow-100 text-yellow-700` |
| HIGH     | `bg-orange-500` | `bg-orange-100 text-orange-700` |
| CRITICAL | `bg-red-500`    | `bg-red-100 text-red-700`       |

---

## Responsive Design

### Breakpoints (Tailwind Defaults)

| Prefix | Min Width | Target             |
| ------ | --------- | ------------------ |
| `sm:`  | 640px     | Small tablets      |
| `md:`  | 768px     | Tablets            |
| `lg:`  | 1024px    | Small desktops     |
| `xl:`  | 1280px    | Desktops           |
| `2xl:` | 1536px    | Large desktops     |

### Common Responsive Patterns

```jsx
{/* Grid: 1 col → 2 cols → 3 cols */}
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">

{/* Hide on mobile, show on desktop */}
<div className="hidden md:block">

{/* Mobile bottom nav, desktop top nav */}
<nav className="md:top-0 fixed bottom-0 md:bottom-auto">

{/* Max-width container */}
<div className="max-w-7xl mx-auto px-4 sm:px-6">
```

---

## Mobile Navigation

The mobile layout uses a fixed bottom navigation bar:

```
┌────────────────────────────────┐
│         Page Content           │
│                                │
├────────┬────────┬────┬─────────┤
│ Dash   │ Dir    │ +  │ History │
│ (icon) │ (icon) │ FAB│ (icon)  │
└────────┴────────┴────┴─────────┘
```

The `+` FAB (Floating Action Button) for Level 2 doctors:

```jsx
<button className="absolute -top-6 left-1/2 -translate-x-1/2 w-14 h-14 bg-teal-600 rounded-full shadow-lg flex items-center justify-center text-white">
    <Plus />
</button>
```

---

## App.css (Legacy)

> **Source:** [`App.css`](../src/App.css)

This file contains **unused Vite boilerplate** CSS (logo animations, root max-width). It is imported in `App.jsx` but effectively does nothing relevant to the application. Can be safely removed.

---

## Recommendations

1. **Remove `App.css`** — Unused boilerplate
2. **Create `cn()` utility** — Combine `clsx` + `tailwind-merge` for cleaner conditional classes
3. **Extract design tokens** — Define spacing, border-radius, and shadow scales in `tailwind.config.js`
4. **Add dark mode** — Tailwind supports `darkMode: 'class'` out of the box
5. **Component variants** — Consider `cva` (class-variance-authority) for button/badge variants
