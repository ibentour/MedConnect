# Utilities — Helper Functions

---

## Overview

The frontend has **no dedicated utility files**. Helper functions are defined inline within components or service files. This section documents all utility functions and their locations.

---

## Utility Functions

### API & Cache Utilities (`services/api.js`, `services/cache.js`)

| Function                        | Location         | Description                                   |
| ------------------------------- | ---------------- | --------------------------------------------- |
| `generateCacheKey(method, url, params)` | `api.js` / `cache.js` | Creates unique cache key from request params |
| `getCacheTTL(url)`             | `cache.js`       | Maps URL patterns to TTL values               |
| `shouldCache(method, url)`     | `cache.js`       | Determines if response should be cached       |
| `invalidateDirectoryCache()`   | `cache.js`       | Clears directory cache entries                |
| `invalidateAdminCache()`       | `cache.js`       | Clears admin cache entries                    |
| `invalidateAnalystCache()`     | `cache.js`       | Clears analyst cache entries                  |

### Phone Validation (`pages/level2/CreateReferral.jsx`)

```jsx
function validatePhoneNumber(phone, countryCode) {
    // Validates phone number by country
    // Morocco: 06XX or 07XX (10 digits)
    // France: 10 digits starting with 0
    // Spain: 9 digits starting with 6 or 7
    // International: minimum 8 digits
    // Returns: { valid: boolean, error?: string }
}
```

| Country | Format       | Validation Rule              |
| ------- | ------------ | ---------------------------- |
| Morocco | `06XXXXXXXX` | 10 digits, starts with 06/07 |
| France  | `06XXXXXXXX` | 10 digits, starts with 0     |
| Spain   | `6XXXXXXXX`  | 9 digits                     |
| Other   | varies       | Minimum 8 digits             |

### Age Calculation (`pages/chu/Dashboard.jsx`, `pages/shared/History.jsx`)

> **Note:** This function is duplicated in two files.

```jsx
function calculateAge(dob) {
    if (!dob) return null;
    const today = new Date();
    const birth = new Date(dob);
    let age = today.getFullYear() - birth.getFullYear();
    const m = today.getMonth() - birth.getMonth();
    if (m < 0 || (m === 0 && today.getDate() < birth.getDate())) age--;
    return age;
}
```

### Urgency Color Mapping (`pages/chu/Dashboard.jsx`)

```jsx
function getUrgencyColor(urgency) {
    const colors = {
        LOW: 'bg-green-100 text-green-700',
        MEDIUM: 'bg-yellow-100 text-yellow-700',
        HIGH: 'bg-orange-100 text-orange-700',
        CRITICAL: 'bg-red-100 text-red-700',
    };
    return colors[urgency] || 'bg-gray-100 text-gray-700';
}
```

### Day Toggle (`pages/admin/Dashboard.jsx`)

```jsx
function toggleDay(currentStr, day) {
    // Toggles a day in comma-separated string
    // "Mon,Tue,Wed" + "Thu" → "Mon,Tue,Wed,Thu"
    // "Mon,Tue,Wed" + "Tue" → "Mon,Wed"
    const days = currentStr ? currentStr.split(',') : [];
    if (days.includes(day)) {
        return days.filter(d => d !== day).join(',');
    }
    return [...days, day].join(',');
}
```

### Notification Icon Mapper (`pages/shared/History.jsx`)

```jsx
function getNotificationIcon(message) {
    if (message.includes('✅')) return 'green-check';
    if (message.includes('❌')) return 'red-x';
    if (message.includes('🔄')) return 'blue-refresh';
    return 'default-bell';
}
```

### TimePicker Sub-component (`pages/admin/Dashboard.jsx`)

Inline-defined component for selecting work hours:

```jsx
function TimePicker({ value, onChange }) {
    // Renders HH:MM hour and minute selectors
    // Range: 00:00 to 23:59
    // Returns: "HH:MM" string format
}
```

---

## Class Merging (Installed but Unused)

`clsx` and `tailwind-merge` are installed as dependencies but **not used** in any source file:

```jsx
// These packages are available but unused:
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

// Recommended pattern for conditional Tailwind classes:
function cn(...inputs) {
    return twMerge(clsx(inputs));
}
```

---

## Recommendations

1. **Create `utils/` directory** — Extract shared helpers:
   - `utils/phone.js` — Phone validation
   - `utils/date.js` — Age calculation, date formatting
   - `utils/classNames.js` — `cn()` function using clsx + tailwind-merge
   - `utils/colors.js` — Urgency/status color mappings

2. **Eliminate duplication** — `calculateAge()` exists in two files; extract once.

3. **Use `cn()`** — The `clsx` + `tailwind-merge` packages are installed; create the `cn()` utility and use it for conditional classes.

4. **Type safety** — If migrating to TypeScript, utility functions should be the first to receive types.
