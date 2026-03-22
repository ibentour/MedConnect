// Package service provides caching utilities for MedConnect backend.
// The decryption cache reduces redundant decryption operations for frequently
// accessed patient data, improving performance while maintaining security.
package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CacheConfig holds configuration for the decryption cache.
type CacheConfig struct {
	// DefaultTTL is the time-to-live for cache entries (default: 5 minutes)
	DefaultTTL time.Duration
	// MaxEntries limits the maximum number of cached items
	MaxEntries int
}

// DefaultCacheConfig returns the default cache configuration.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		DefaultTTL: 5 * time.Minute,
		MaxEntries: 10000,
	}
}

// cacheEntry holds a cached value with its expiration time.
type cacheEntry struct {
	value      string
	expiration time.Time
}

// DecryptionCache provides thread-safe in-memory caching for decrypted data.
// This cache is used to avoid redundant decryption operations for the same
// encrypted patient data (CIN, Name, Symptoms).
//
// Cache Key Format: "decrypt:{entity}:{id}:{field}"
// Example: "decrypt:patient:uuid-123:fullname"
type DecryptionCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	config  CacheConfig
}

// NewDecryptionCache creates a new decryption cache with the given configuration.
func NewDecryptionCache(config CacheConfig) *DecryptionCache {
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 5 * time.Minute
	}
	if config.MaxEntries == 0 {
		config.MaxEntries = 10000
	}
	return &DecryptionCache{
		entries: make(map[string]cacheEntry),
		config:  config,
	}
}

// cacheKey generates a cache key for decrypted data.
// Format: "decrypt:{entity}:{id}:{field}"
func cacheKey(entity string, id uuid.UUID, field string) string {
	return fmt.Sprintf("decrypt:%s:%s:%s", entity, id.String(), field)
}

// Get retrieves a cached decrypted value if it exists and hasn't expired.
func (c *DecryptionCache) Get(entity string, id uuid.UUID, field string) (string, bool) {
	key := cacheKey(entity, id, field)

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return "", false
	}

	// Check if entry has expired
	if time.Now().After(entry.expiration) {
		// Entry expired - will be cleaned up on next write or lazily
		return "", false
	}

	return entry.value, true
}

// Set stores a decrypted value in the cache with the default TTL.
func (c *DecryptionCache) Set(entity string, id uuid.UUID, field string, value string) {
	c.SetWithTTL(entity, id, field, value, c.config.DefaultTTL)
}

// SetWithTTL stores a decrypted value in the cache with a custom TTL.
func (c *DecryptionCache) SetWithTTL(entity string, id uuid.UUID, field string, value string, ttl time.Duration) {
	key := cacheKey(entity, id, field)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if cache is full
	if len(c.entries) >= c.config.MaxEntries {
		c.evictOldest()
	}

	c.entries[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Invalidate removes a specific cached entry.
func (c *DecryptionCache) Invalidate(entity string, id uuid.UUID, field string) {
	key := cacheKey(entity, id, field)

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// InvalidateEntity removes all cached entries for a specific entity ID.
// This should be called when patient data is updated.
func (c *DecryptionCache) InvalidateEntity(entity string, id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	idStr := id.String()
	prefix := fmt.Sprintf("decrypt:%s:%s:", entity, idStr)

	for key := range c.entries {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			delete(c.entries, key)
		}
	}
}

// InvalidatePatient removes all cached entries for a patient.
// Convenience method that calls InvalidateEntity with "patient".
func (c *DecryptionCache) InvalidatePatient(patientID uuid.UUID) {
	c.InvalidateEntity("patient", patientID)
}

// InvalidateReferral removes all cached entries related to a referral.
// This includes the referral's patient data.
func (c *DecryptionCache) InvalidateReferral(referralID uuid.UUID, patientID uuid.UUID) {
	c.InvalidateEntity("referral", referralID)
	c.InvalidatePatient(patientID)
}

// Clear removes all entries from the cache.
func (c *DecryptionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
}

// evictOldest removes approximately 10% of the oldest entries.
// This is called when the cache reaches its maximum size.
func (c *DecryptionCache) evictOldest() {
	if len(c.entries) == 0 {
		return
	}

	// Calculate how many entries to remove (10%)
	count := len(c.entries) / 10
	if count < 1 {
		count = 1
	}

	// Find and remove oldest entries
	now := time.Now()
	for key, entry := range c.entries {
		if entry.expiration.Before(now) {
			delete(c.entries, key)
			count--
			if count <= 0 {
				return
			}
		}
	}

	// If we still need to remove more, just remove arbitrary entries
	for key := range c.entries {
		delete(c.entries, key)
		count--
		if count <= 0 {
			return
		}
	}
}

// Stats returns current cache statistics.
func (c *DecryptionCache) Stats() (total, expired int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	total = len(c.entries)
	for _, entry := range c.entries {
		if now.After(entry.expiration) {
			expired++
		}
	}
	return total, expired
}
