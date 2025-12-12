package main

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

// Cache prefixes for different types of cached items
const (
	PrefixJWT    = "jwt:"
	PrefixSecret = "secret:"
	PrefixToken  = "token:"
)

// Default expiration times
const (
	DefaultJWTExpiration    = 5 * time.Minute
	DefaultSecretExpiration = 30 * time.Minute
	DefaultTokenExpiration  = 10 * time.Minute
	CleanupInterval         = 1 * time.Minute
)

// AppCache wraps go-cache with convenience methods for tokens and secrets
type AppCache struct {
	c *cache.Cache
}

// CachedToken represents a cached Kerberos or JWT token
type CachedToken struct {
	Value     string
	ExpiresAt time.Time
	SPN       string
}

// CachedSecret represents a cached secret
type CachedSecret struct {
	Value     string
	ExpiresAt time.Time
	Metadata  map[string]string
}

var appCache *AppCache

// InitCache initializes the application cache
func InitCache() {
	appCache = &AppCache{
		c: cache.New(DefaultTokenExpiration, CleanupInterval),
	}
}

// GetCache returns the application cache instance
func GetCache() *AppCache {
	if appCache == nil {
		InitCache()
	}
	return appCache
}

// SetJWT stores a JWT token with the given key and expiration
func (ac *AppCache) SetJWT(key string, token string, expiration time.Duration) {
	ct := &CachedToken{
		Value:     token,
		ExpiresAt: time.Now().Add(expiration),
	}
	ac.c.Set(PrefixJWT+key, ct, expiration)
}

// GetJWT retrieves a JWT token by key
func (ac *AppCache) GetJWT(key string) (string, bool) {
	if val, found := ac.c.Get(PrefixJWT + key); found {
		if ct, ok := val.(*CachedToken); ok {
			return ct.Value, true
		}
	}
	return "", false
}

// SetSecret stores a secret with the given key and expiration
func (ac *AppCache) SetSecret(key string, secret string, expiration time.Duration) {
	cs := &CachedSecret{
		Value:     secret,
		ExpiresAt: time.Now().Add(expiration),
		Metadata:  make(map[string]string),
	}
	ac.c.Set(PrefixSecret+key, cs, expiration)
}

// SetSecretWithMetadata stores a secret with metadata
func (ac *AppCache) SetSecretWithMetadata(key string, secret string, metadata map[string]string, expiration time.Duration) {
	cs := &CachedSecret{
		Value:     secret,
		ExpiresAt: time.Now().Add(expiration),
		Metadata:  metadata,
	}
	ac.c.Set(PrefixSecret+key, cs, expiration)
}

// GetSecret retrieves a secret by key
func (ac *AppCache) GetSecret(key string) (string, bool) {
	if val, found := ac.c.Get(PrefixSecret + key); found {
		if cs, ok := val.(*CachedSecret); ok {
			return cs.Value, true
		}
	}
	return "", false
}

// GetSecretWithMetadata retrieves a secret with its metadata
func (ac *AppCache) GetSecretWithMetadata(key string) (*CachedSecret, bool) {
	if val, found := ac.c.Get(PrefixSecret + key); found {
		if cs, ok := val.(*CachedSecret); ok {
			return cs, true
		}
	}
	return nil, false
}

// SetToken stores a Kerberos token for an SPN
func (ac *AppCache) SetToken(spn string, token string, expiration time.Duration) {
	ct := &CachedToken{
		Value:     token,
		ExpiresAt: time.Now().Add(expiration),
		SPN:       spn,
	}
	ac.c.Set(PrefixToken+spn, ct, expiration)
}

// GetToken retrieves a cached Kerberos token for an SPN
func (ac *AppCache) GetToken(spn string) (string, bool) {
	if val, found := ac.c.Get(PrefixToken + spn); found {
		if ct, ok := val.(*CachedToken); ok {
			return ct.Value, true
		}
	}
	return "", false
}

// GetTokenWithExpiry retrieves a cached token with its expiry time
func (ac *AppCache) GetTokenWithExpiry(spn string) (string, time.Time, bool) {
	if val, found := ac.c.Get(PrefixToken + spn); found {
		if ct, ok := val.(*CachedToken); ok {
			return ct.Value, ct.ExpiresAt, true
		}
	}
	return "", time.Time{}, false
}

// Delete removes an item from the cache
func (ac *AppCache) Delete(key string) {
	ac.c.Delete(key)
}

// DeleteJWT removes a JWT from the cache
func (ac *AppCache) DeleteJWT(key string) {
	ac.c.Delete(PrefixJWT + key)
}

// DeleteSecret removes a secret from the cache
func (ac *AppCache) DeleteSecret(key string) {
	ac.c.Delete(PrefixSecret + key)
}

// DeleteToken removes a token from the cache
func (ac *AppCache) DeleteToken(spn string) {
	ac.c.Delete(PrefixToken + spn)
}

// Clear removes all items from the cache
func (ac *AppCache) Clear() {
	ac.c.Flush()
}

// ItemCount returns the number of items in the cache
func (ac *AppCache) ItemCount() int {
	return ac.c.ItemCount()
}

// Stats returns cache statistics as a formatted string
func (ac *AppCache) Stats() string {
	return fmt.Sprintf("Cache items: %d", ac.c.ItemCount())
}

// CacheEntry represents a cache entry for display
type CacheEntry struct {
	Key       string
	Value     string
	ExpiresAt time.Time
	Type      string // "jwt", "secret", "token", or "custom"
}

// ListKeys returns all cache keys
func (ac *AppCache) ListKeys() []string {
	items := ac.c.Items()
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	return keys
}

// ListEntries returns all cache entries with metadata
func (ac *AppCache) ListEntries() []CacheEntry {
	items := ac.c.Items()
	entries := make([]CacheEntry, 0, len(items))

	for k, item := range items {
		entry := CacheEntry{
			Key: k,
		}

		// Determine type and extract value
		switch v := item.Object.(type) {
		case *CachedToken:
			entry.Value = v.Value
			entry.ExpiresAt = v.ExpiresAt
			if len(k) > len(PrefixToken) && k[:len(PrefixToken)] == PrefixToken {
				entry.Type = "token"
			} else if len(k) > len(PrefixJWT) && k[:len(PrefixJWT)] == PrefixJWT {
				entry.Type = "jwt"
			}
		case *CachedSecret:
			entry.Value = v.Value
			entry.ExpiresAt = v.ExpiresAt
			entry.Type = "secret"
		case string:
			entry.Value = v
			entry.Type = "custom"
			// Calculate expiry from item.Expiration (Unix nano timestamp)
			if item.Expiration > 0 {
				entry.ExpiresAt = time.Unix(0, item.Expiration)
			}
		default:
			entry.Value = fmt.Sprintf("%v", item.Object)
			entry.Type = "custom"
			if item.Expiration > 0 {
				entry.ExpiresAt = time.Unix(0, item.Expiration)
			}
		}

		entries = append(entries, entry)
	}

	return entries
}

// GetValue retrieves any cached value as a string by its full key
func (ac *AppCache) GetValue(key string) (string, bool) {
	item, found := ac.c.Get(key)
	if !found {
		return "", false
	}

	switch v := item.(type) {
	case *CachedToken:
		return v.Value, true
	case *CachedSecret:
		return v.Value, true
	case string:
		return v, true
	default:
		return fmt.Sprintf("%v", item), true
	}
}

// Set stores a custom string value with the given key and expiration
func (ac *AppCache) Set(key string, value string, expiration time.Duration) {
	ac.c.Set(key, value, expiration)
}

// Get retrieves a custom string value by key
func (ac *AppCache) Get(key string) (string, bool) {
	if val, found := ac.c.Get(key); found {
		if s, ok := val.(string); ok {
			return s, true
		}
	}
	return "", false
}