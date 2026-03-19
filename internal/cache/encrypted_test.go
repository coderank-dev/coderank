package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptedCacheRoundTrip(t *testing.T) {
	cache, err := NewEncryptedCache(t.TempDir(), "test-license-key-abc123")
	require.NoError(t, err)

	key := EncryptedCacheKey("react", "19.1.0", "query", []string{"useState"})
	content := "# useState\n\nReturns a stateful value and a setter function."

	require.NoError(t, cache.Put(key, content))

	got, cachedAt, err := cache.Get(key, 1*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, content, got)
	assert.WithinDuration(t, time.Now(), cachedAt, 5*time.Second)
}

func TestEncryptedCacheMiss(t *testing.T) {
	cache, err := NewEncryptedCache(t.TempDir(), "test-license-key-abc123")
	require.NoError(t, err)

	key := EncryptedCacheKey("react", "19.1.0", "query", []string{"nonexistent"})
	got, _, err := cache.Get(key, 1*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, "", got, "cache miss should return empty string")
}

func TestEncryptedCacheTTLExpiry(t *testing.T) {
	cache, err := NewEncryptedCache(t.TempDir(), "test-license-key-abc123")
	require.NoError(t, err)

	key := EncryptedCacheKey("react", "19.1.0", "query", []string{"useState"})
	require.NoError(t, cache.Put(key, "cached content"))

	got, _, err := cache.Get(key, 1*time.Nanosecond)
	require.NoError(t, err)
	assert.Equal(t, "", got, "expired entry should return empty string (stale)")
}

func TestEncryptedCacheWrongKey(t *testing.T) {
	dir := t.TempDir()

	cache1, err := NewEncryptedCache(dir, "license-key-user-1")
	require.NoError(t, err)
	key := EncryptedCacheKey("react", "19.1.0", "query", []string{"useState"})
	require.NoError(t, cache1.Put(key, "secret docs"))

	// Different license key — should fail to decrypt, treat as miss
	cache2, err := NewEncryptedCache(dir, "license-key-user-2")
	require.NoError(t, err)
	got, _, err := cache2.Get(key, 1*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, "", got, "wrong key should treat as cache miss and auto-evict")
}

func TestEncryptedCacheEvictLibrary(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewEncryptedCache(dir, "test-key")
	require.NoError(t, err)

	key := EncryptedCacheKey("react", "19.1.0", "query", []string{"useState"})
	require.NoError(t, cache.Put(key, "content"))

	// Verify file exists before eviction
	_, statErr := os.Stat(filepath.Join(dir, key))
	require.NoError(t, statErr, "cache file should exist before eviction")

	require.NoError(t, cache.EvictLibrary("react"))

	got, _, err := cache.Get(key, 1*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, "", got, "evicted library should produce cache miss")
}

func TestEncryptedCacheKeyDeterministic(t *testing.T) {
	key1 := EncryptedCacheKey("react", "19.1.0", "query", []string{"useState"})
	key2 := EncryptedCacheKey("react", "19.1.0", "query", []string{"useState"})
	assert.Equal(t, key1, key2, "same inputs should produce same cache key")

	key3 := EncryptedCacheKey("react", "19.1.0", "query", []string{"useEffect"})
	assert.NotEqual(t, key1, key3, "different args should produce different cache key")
}

func TestNewEncryptedCacheRejectsEmptyKey(t *testing.T) {
	_, err := NewEncryptedCache(t.TempDir(), "")
	assert.ErrorContains(t, err, "license key required")
}
