package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/panoptescloud/api/pkg/dto"
	"github.com/panoptescloud/api/pkg/util/point"
)

type HMACSHA256Hasher struct {
	key []byte
}

func NewHMACSHA256Hasher(key []byte) *HMACSHA256Hasher {
	if len(key) == 0 {
		panic("HMAC key cannot be empty")
	}

	return &HMACSHA256Hasher{key: key}
}

func (h *HMACSHA256Hasher) Hash(input string) (dto.HashedValue, error) {
	mac := hmac.New(sha256.New, h.key)
	_, err := mac.Write([]byte(input))

	if err != nil {
		return dto.HashedValue{}, fmt.Errorf("failed to write to HMAC: %w", err)
	}

	sum := mac.Sum(nil)

	return dto.HashedValue{
		Original: point.To(input),
		Value: base64.RawURLEncoding.EncodeToString(sum),
	}, nil
}
