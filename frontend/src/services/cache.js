/**
 * Simple in-memory cache with TTL support for API responses.
 * Used to reduce redundant API calls for frequently accessed data.
 */

// Cache storage
const cache = new Map();

/**
 * Cache configuration by endpoint
 */
const CACHE_CONFIG = {
    // GET /api/directory - Department list (5 minutes)
    '/directory': { ttl: 5 * 60 * 1000 },

    // GET /api/admin/departments - Admin department list (5 minutes)
    '/admin/departments': { ttl: 5 * 60 * 1000 },

    // GET /api/admin/stats - Admin statistics (1 minute)
    '/admin/stats': { ttl: 60 * 1000 },

    // GET /api/analyst/stats/departments - Analyst department stats (1 minute)
    '/analyst/stats/departments': { ttl: 60 * 1000 },

    // GET /api/analyst/stats/doctors - Analyst doctor stats (1 minute)
    '/analyst/stats/doctors': { ttl: 60 * 1000 },
};

/**
 * Generate cache key from request config
 * @param {string} method - HTTP method
 * @param {string} url - Request URL
 * @param {object} params - Request params
 * @returns {string} Cache key
 */
const generateCacheKey = (method, url, params = {}) => {
    const baseKey = `${method}:${url}`;
    if (Object.keys(params).length > 0) {
        return `${baseKey}:${JSON.stringify(params)}`;
    }
    return baseKey;
};

/**
 * Get cached data if valid
 * @param {string} key - Cache key
 * @returns {object|null} Cached data or null if expired/not found
 */
export const getCached = (key) => {
    const cached = cache.get(key);
    if (!cached) return null;

    const now = Date.now();
    if (now > cached.expiry) {
        cache.delete(key);
        return null;
    }

    return cached.data;
};

/**
 * Set cached data with TTL
 * @param {string} key - Cache key
 * @param {any} data - Data to cache
 * @param {number} ttl - Time to live in milliseconds
 */
export const setCached = (key, data, ttl) => {
    cache.set(key, {
        data,
        expiry: Date.now() + ttl,
    });
};

/**
 * Get cache TTL for a specific endpoint
 * @param {string} url - API URL
 * @returns {number|null} TTL in ms or null if not cacheable
 */
export const getCacheTTL = (url) => {
    // Check exact match first
    if (CACHE_CONFIG[url]) {
        return CACHE_CONFIG[url].ttl;
    }

    // Check pattern match
    for (const [pattern, config] of Object.entries(CACHE_CONFIG)) {
        if (url.includes(pattern)) {
            return config.ttl;
        }
    }

    return null;
};

/**
 * Check if a request should be cached
 * @param {string} method - HTTP method
 * @param {string} url - Request URL
 * @returns {boolean} Whether to cache this response
 */
export const shouldCache = (method, url) => {
    if (method !== 'GET') return false;
    return getCacheTTL(url) !== null;
};

/**
 * Invalidate cache for specific endpoints
 * Call this after mutations (create/update/delete)
 * @param {string} pattern - URL pattern to invalidate
 */
export const invalidateCache = (pattern = null) => {
    if (!pattern) {
        // Clear all cache
        cache.clear();
        return;
    }

    // Clear matching patterns
    for (const key of cache.keys()) {
        if (key.includes(pattern)) {
            cache.delete(key);
        }
    }
};

// Export common invalidation helpers
export const invalidateDirectoryCache = () => invalidateCache('/directory');
export const invalidateAdminCache = () => invalidateCache('/admin');
export const invalidateAnalystCache = () => invalidateCache('/analyst');

export default {
    getCached,
    setCached,
    getCacheTTL,
    shouldCache,
    invalidateCache,
    invalidateDirectoryCache,
    invalidateAdminCache,
    invalidateAnalystCache,
};
