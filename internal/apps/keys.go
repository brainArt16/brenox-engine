package apps

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	livePrefix  = "bx_live_"
	testPrefix  = "bx_test_"
	prefixLen   = 16
	secretBytes = 24
)

func GenerateAPIKey(sandbox bool) (plainKey, lookupPrefix, keyHash string, err error) {
	secret := make([]byte, secretBytes)
	if _, err := rand.Read(secret); err != nil {
		return "", "", "", err
	}

	prefix := livePrefix
	if sandbox {
		prefix = testPrefix
	}

	plainKey = prefix + base64.RawURLEncoding.EncodeToString(secret)
	if len(plainKey) < prefixLen {
		return "", "", "", fmt.Errorf("generated api key too short")
	}

	lookupPrefix = plainKey[:prefixLen]
	keyHash = HashAPIKey(plainKey)
	return plainKey, lookupPrefix, keyHash, nil
}

func HashAPIKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func IsAPIKeyToken(token string) bool {
	return strings.HasPrefix(token, livePrefix) || strings.HasPrefix(token, testPrefix)
}

func LookupPrefix(key string) string {
	if len(key) < prefixLen {
		return key
	}
	return key[:prefixLen]
}
