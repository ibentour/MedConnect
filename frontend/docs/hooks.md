# Hooks — Custom React Hooks

> **Source:** [`frontend/src/context/AuthContext.jsx`](../src/context/AuthContext.jsx)

---

## Overview

The codebase has **one custom hook**: `useAuth()`. All other state management uses built-in React hooks (`useState`, `useEffect`, `useMemo`, `useCallback`) directly in components.

---

## useAuth()

> **Source:** `context/AuthContext.jsx`

The sole custom hook provides authentication state and actions.

### Signature

```jsx
function useAuth(): {
    user: User | null,
    isAuthenticated: boolean,
    login: (username: string, password: string) => Promise<void>,
    logout: () => void
}
```

### Usage

```jsx
import { useAuth } from '../context/AuthContext';

function MyComponent() {
    const { user, isAuthenticated, login, logout } = useAuth();
    
    if (!isAuthenticated) return <Navigate to="/login" />;
    
    return (
        <div>
            <p>Welcome, {user.username}</p>
            <p>Role: {user.role}</p>
            <button onClick={logout}>Logout</button>
        </div>
    );
}
```

### Error Handling

Throws if used outside `AuthProvider`:

```jsx
export function useAuth() {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}
```

---

## Built-in Hook Usage Patterns

### useState + useEffect (Data Fetching)

Used extensively across all page components:

```jsx
const [data, setData] = useState([]);
const [loading, setLoading] = useState(true);
const [error, setError] = useState(null);

useEffect(() => {
    async function fetchData() {
        try {
            const result = await getQueue();
            setData(result);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }
    fetchData();
}, []);
```

### useMemo (Derived State)

Used for filtering/search:

```jsx
const filteredReferrals = useMemo(() => {
    return referrals.filter(r =>
        r.patient_name.toLowerCase().includes(search.toLowerCase())
    );
}, [referrals, search]);
```

### useNavigate (Programmatic Navigation)

```jsx
const navigate = useNavigate();

// After form submission
navigate('/dashboard');

// After logout
navigate('/login');
```

### useLocation (Active Route Detection)

Used in `Navbar.jsx`:

```jsx
const location = useLocation();
const isActive = (path) => location.pathname === path;
```

### useRef (DOM Access)

Used for file input refs and scroll containers:

```jsx
const fileInputRef = useRef(null);
```

---

## Recommended Custom Hooks

The following hooks could be extracted to reduce code duplication:

### useFetch

```jsx
function useFetch(fetcher, deps = []) {
    const [data, setData] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    
    useEffect(() => {
        let cancelled = false;
        setLoading(true);
        fetcher()
            .then(result => !cancelled && setData(result))
            .catch(err => !cancelled && setError(err))
            .finally(() => !cancelled && setLoading(false));
        return () => { cancelled = true; };
    }, deps);
    
    return { data, loading, error, refetch: () => { /* ... */ } };
}
```

### useWebSocket

```jsx
function useWebSocket(eventType) {
    const [lastMessage, setLastMessage] = useState(null);
    
    useEffect(() => {
        const unsubscribe = subscribe((msg) => {
            if (msg.event === eventType) setLastMessage(msg.payload);
        });
        return unsubscribe;
    }, [eventType]);
    
    return lastMessage;
}
```

### usePagination

```jsx
function usePagination(initialLimit = 20) {
    const [offset, setOffset] = useState(0);
    const [limit] = useState(initialLimit);
    const next = () => setOffset(prev => prev + limit);
    const prev = () => setOffset(prev => Math.max(0, prev - limit));
    return { offset, limit, next, prev, reset: () => setOffset(0) };
}
```

### calculateAge (Utility)

Currently duplicated in `chu/Dashboard.jsx` and `shared/History.jsx`:

```jsx
function useAge(dob) {
    return useMemo(() => {
        if (!dob) return null;
        const today = new Date();
        const birth = new Date(dob);
        let age = today.getFullYear() - birth.getFullYear();
        const m = today.getMonth() - birth.getMonth();
        if (m < 0 || (m === 0 && today.getDate() < birth.getDate())) age--;
        return age;
    }, [dob]);
}
```
