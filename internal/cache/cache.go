package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Cache represents an in-memory cache with TTL (Time-To-Live) support.
// This implementation:
// - Thread-safe (uses sync.RWMutex)
// - Automatic expiration
// - Memory-bounded (via maxEntries)
// - LRU-like eviction when full
//
// Use cases:
// - Cache API responses to reduce external calls
// - Store computed results
// - Rate limiting data
type Cache struct {
	mu         sync.RWMutex
	items      map[string]*cacheItem
	maxEntries int           // Maximum number of entries
	defaultTTL time.Duration // Default time-to-live for items
}

// cacheItem represents a single cache entry
type cacheItem struct {
	value      interface{}
	expiration time.Time
	createdAt  time.Time
}

// NewCache creates a new cache with specified max entries and default TTL
// Parameters:
//   - maxEntries: Maximum number of items to store (0 = unlimited, not recommended)
//   - defaultTTL: Default time-to-live for cache entries
//
// Example:
//
//	cache := NewCache(1000, 5*time.Minute)
//	// Cache with max 1000 entries, 5-minute TTL
func NewCache(maxEntries int, defaultTTL time.Duration) *Cache {
	c := &Cache{
		items:      make(map[string]*cacheItem),
		maxEntries: maxEntries,
		defaultTTL: defaultTTL,
	}

	// Start cleanup goroutine to remove expired items
	go c.startCleanup()

	return c
}

// Set stores a value in the cache with default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If at capacity, remove oldest entry
	if c.maxEntries > 0 && len(c.items) >= c.maxEntries {
		c.evictOldest()
	}

	c.items[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
		createdAt:  time.Now(),
	}
}

// Get retrieves a value from the cache
// Returns (value, true) if found and not expired
// Returns (nil, false) if not found or expired
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.expiration) {
		return nil, false
	}

	return item.value, true
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
}

// Len returns the current number of items in the cache
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// evictOldest removes the oldest entry from the cache
// This implements a simple FIFO eviction policy
// For production, consider implementing LRU (Least Recently Used)
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	// Find the oldest entry
	for key, item := range c.items {
		if oldestKey == "" || item.createdAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.createdAt
		}
	}

	// Remove it
	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// startCleanup runs a background goroutine to remove expired entries
// Runs every minute to prevent memory leaks from expired items
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.removeExpired()
	}
}

// removeExpired removes all expired entries
func (c *Cache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiration) {
			delete(c.items, key)
		}
	}
}

// GenerateKey creates a deterministic cache key from multiple parameters
// Uses SHA-256 to create a fixed-length key from variable inputs
//
// Example:
//
//	key := GenerateKey("search", "golang tutorial", 10)
//	// Returns: "abc123..." (SHA-256 hash)
func GenerateKey(parts ...interface{}) string {
	h := sha256.New()
	for _, part := range parts {
		// Convert to string and hash
		h.Write([]byte(toString(part)))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// toString converts various types to string for hashing
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return string(rune(val))
	case int64:
		return string(rune(val))
	default:
		return ""
	}
}

// Stats returns cache statistics
type Stats struct {
	TotalItems    int
	MaxEntries    int
	DefaultTTL    time.Duration
	OldestItemAge time.Duration
	NewestItemAge time.Duration
}

// GetStats returns cache statistics for monitoring
func (c *Cache) GetStats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := Stats{
		TotalItems: len(c.items),
		MaxEntries: c.maxEntries,
		DefaultTTL: c.defaultTTL,
	}

	if len(c.items) > 0 {
		now := time.Now()
		var oldest, newest time.Time

		for _, item := range c.items {
			if oldest.IsZero() || item.createdAt.Before(oldest) {
				oldest = item.createdAt
			}
			if newest.IsZero() || item.createdAt.After(newest) {
				newest = item.createdAt
			}
		}

		stats.OldestItemAge = now.Sub(oldest)
		stats.NewestItemAge = now.Sub(newest)
	}

	return stats
}
