package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Cache provides a simple file-based JSON cache with TTL support.
type Cache struct {
	dir string
}

// New creates a new Cache instance, ensuring the ~/.termf1/cache directory exists.
func New() (*Cache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".termf1", "cache")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &Cache{dir: dir}, nil
}

type cacheEntry struct {
	ExpiresAt time.Time       `json:"expires_at"`
	Data      json.RawMessage `json:"data"`
}

// Get attempts to read the cached data into target. If missing or expired,
// it calls fetchFn() to get fresh data, saves it to cache with the given ttl,
// and then unmarshals into target.
// A ttl of 0 means do not cache.
func (c *Cache) Get(key string, ttl time.Duration, target interface{}, fetchFn func() error) error {
	if ttl <= 0 {
		return fetchFn()
	}

	hash := sha256.Sum256([]byte(key))
	safeKey := hex.EncodeToString(hash[:])
	path := filepath.Join(c.dir, safeKey+".json")

	// Try reading from cache
	if b, err := os.ReadFile(path); err == nil {
		var entry cacheEntry
		if err := json.Unmarshal(b, &entry); err == nil {
			if time.Now().Before(entry.ExpiresAt) {
				// Cache hit, still valid
				if err := json.Unmarshal(entry.Data, target); err == nil {
					return nil
				}
			}
		}
	}

	// Fetch fresh data
	if err := fetchFn(); err != nil {
		return err
	}

	// Save fresh data to cache
	raw, err := json.Marshal(target)
	if err == nil {
		entry := cacheEntry{
			ExpiresAt: time.Now().Add(ttl),
			Data:      raw,
		}
		if eb, err := json.Marshal(entry); err == nil {
			_ = os.WriteFile(path, eb, 0644)
		}
	}

	return nil
}

// Clear removes all cached files.
func (c *Cache) Clear() error {
	return os.RemoveAll(c.dir)
}
