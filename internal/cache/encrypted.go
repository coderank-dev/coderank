package cache

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/hkdf"
)

// EncryptedCache provides AES-256-GCM encrypted disk caching.
// Cache entries are keyed by library/version/command/args and encrypted
// with a key derived from the user's license key via HKDF-SHA256.
//
// On-disk format: [12-byte nonce][GCM ciphertext of (8-byte unix timestamp + content)]
// The raw markdown never touches disk in plaintext.
//
// This complements the SQLite Manager: SQLite tracks metadata and enables
// keyword search, while EncryptedCache secures the actual content on disk.
type EncryptedCache struct {
	cacheDir      string
	encryptionKey []byte // 32 bytes, derived from license key
}

// NewEncryptedCache creates a cache backed by the given directory,
// encrypting entries with a key derived from the license key.
func NewEncryptedCache(cacheDir string, licenseKey string) (*EncryptedCache, error) {
	if licenseKey == "" {
		return nil, fmt.Errorf("license key required for encrypted cache")
	}

	key, err := deriveKey(licenseKey)
	if err != nil {
		return nil, fmt.Errorf("deriving encryption key: %w", err)
	}

	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}

	return &EncryptedCache{cacheDir: cacheDir, encryptionKey: key}, nil
}

// deriveKey uses HKDF-SHA256 to derive a 32-byte AES key from a license key.
func deriveKey(licenseKey string) ([]byte, error) {
	salt := []byte("coderank-cache-v1")
	info := []byte("aes-256-gcm-cache-encryption")
	r := hkdf.New(sha256.New, []byte(licenseKey), salt, info)
	key := make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptedCacheKey builds the on-disk path for a cache entry.
// Args are hashed to avoid filesystem issues with special characters.
func EncryptedCacheKey(library, version, command string, args []string) string {
	argsHash := sha256.Sum256([]byte(strings.Join(args, "|")))
	return fmt.Sprintf("%s/%s/%s/%s.enc", library, version, command, hex.EncodeToString(argsHash[:8]))
}

// Get retrieves and decrypts a cache entry. Returns ("", time.Time{}, nil) on
// cache miss. maxAge of 0 disables TTL checking.
func (c *EncryptedCache) Get(key string, maxAge time.Duration) (string, time.Time, error) {
	path := filepath.Join(c.cacheDir, key)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", time.Time{}, nil
	}
	if err != nil {
		return "", time.Time{}, fmt.Errorf("reading cache entry: %w", err)
	}

	content, cachedAt, err := c.decrypt(data)
	if err != nil {
		// Corrupted or wrong key — auto-evict and treat as miss
		os.Remove(path)
		return "", time.Time{}, nil
	}

	if maxAge > 0 && time.Since(cachedAt) > maxAge {
		return "", cachedAt, nil // stale
	}

	return content, cachedAt, nil
}

// Put encrypts content and stores it at the given cache key atomically.
func (c *EncryptedCache) Put(key string, content string) error {
	path := filepath.Join(c.cacheDir, key)

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating cache subdirectory: %w", err)
	}

	encrypted, err := c.encrypt(content, time.Now())
	if err != nil {
		return fmt.Errorf("encrypting cache entry: %w", err)
	}

	// Atomic write: write to temp file then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, encrypted, 0600); err != nil {
		return fmt.Errorf("writing cache entry: %w", err)
	}
	return os.Rename(tmpPath, path)
}

// encrypt produces: [12-byte nonce][GCM ciphertext of (8-byte big-endian unix timestamp + content)]
func (c *EncryptedCache) encrypt(content string, ts time.Time) ([]byte, error) {
	block, err := aes.NewCipher(c.encryptionKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Prepend 8-byte big-endian unix timestamp
	unixSec := ts.Unix()
	payload := make([]byte, 8+len(content))
	for i := 0; i < 8; i++ {
		payload[i] = byte(unixSec >> (56 - 8*i))
	}
	copy(payload[8:], content)

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, payload, nil)
	result := make([]byte, len(nonce)+len(ciphertext))
	copy(result, nonce)
	copy(result[len(nonce):], ciphertext)
	return result, nil
}

// decrypt reverses encrypt: extracts nonce, decrypts, parses timestamp + content.
func (c *EncryptedCache) decrypt(data []byte) (string, time.Time, error) {
	block, err := aes.NewCipher(c.encryptionKey)
	if err != nil {
		return "", time.Time{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", time.Time{}, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize+8 {
		return "", time.Time{}, fmt.Errorf("cache entry too short")
	}

	payload, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("decryption failed (wrong key or corrupted): %w", err)
	}

	var unixSec int64
	for i := 0; i < 8; i++ {
		unixSec |= int64(payload[i]) << (56 - 8*i)
	}
	return string(payload[8:]), time.Unix(unixSec, 0), nil
}

// Evict removes a single cache entry by key.
func (c *EncryptedCache) Evict(key string) error {
	path := filepath.Join(c.cacheDir, key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// EvictLibrary removes all cache entries for a library (e.g. on version bump).
func (c *EncryptedCache) EvictLibrary(library string) error {
	return os.RemoveAll(filepath.Join(c.cacheDir, library))
}
