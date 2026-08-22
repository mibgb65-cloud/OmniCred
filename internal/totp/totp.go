package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
)

const (
	Digits = 6
	Period = 30
)

var ErrInvalidSecret = errors.New("invalid Base32 TOTP secret")

var base32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

func NormalizeSecret(value string) (string, error) {
	var cleaned strings.Builder
	for _, character := range strings.TrimSpace(value) {
		if unicode.IsSpace(character) || character == '-' {
			continue
		}
		cleaned.WriteRune(unicode.ToUpper(character))
	}
	normalized := strings.TrimRight(cleaned.String(), "=")
	decoded, err := base32NoPadding.DecodeString(normalized)
	if err != nil || len(decoded) == 0 {
		return "", ErrInvalidSecret
	}
	return normalized, nil
}

func Generate(secret string, timestamp time.Time) (string, error) {
	normalized, err := NormalizeSecret(secret)
	if err != nil {
		return "", err
	}
	key, err := base32NoPadding.DecodeString(normalized)
	if err != nil {
		return "", ErrInvalidSecret
	}

	counter := uint64(timestamp.Unix() / Period)
	message := make([]byte, 8)
	binary.BigEndian.PutUint64(message, counter)
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(message)
	digest := mac.Sum(nil)
	offset := digest[len(digest)-1] & 0x0f
	value := (uint32(digest[offset])&0x7f)<<24 |
		uint32(digest[offset+1])<<16 |
		uint32(digest[offset+2])<<8 |
		uint32(digest[offset+3])
	return fmt.Sprintf("%0*d", Digits, value%1_000_000), nil
}

func SecondsRemaining(timestamp time.Time) int {
	elapsed := timestamp.Unix() % Period
	if elapsed < 0 {
		elapsed += Period
	}
	return int(Period - elapsed)
}
