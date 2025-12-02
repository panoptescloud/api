package random

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateRefreshToken returns a securely generated random string of n bytes,
// base64 URL encoded for safe transport/storage.
func GenerateString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure random bytes: %w", err)
	}

	// Base64 URL encoding avoids '+' and '/' characters
	return base64.RawURLEncoding.EncodeToString(b), nil
}
