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