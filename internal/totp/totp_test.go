package totp

import (
	"errors"
	"testing"
	"time"
)

const rfcSHA1Secret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

func TestGenerateMatchesRFC6238SHA1Vectors(t *testing.T) {
	tests := []struct {
		timestamp int64
		want      string
	}{
		{59, "287082"},
		{1111111109, "081804"},
		{1234567890, "005924"},
	}
	for _, test := range tests {
		code, err := Generate(rfcSHA1Secret, time.Unix(test.timestamp, 0))
		if err != nil || code != test.want {
			t.Fatalf("Generate(%d) = %q, %v; want %q", test.timestamp, code, err, test.want)
		}
	}
}

func TestNormalizeSecretAcceptsGroupedLowercaseAndRejectsInvalidInput(t *testing.T) {
	normalized, err := NormalizeSecret("gezd gnbv-gy3t qojq")
	if err != nil || normalized != "GEZDGNBVGY3TQOJQ" {
		t.Fatalf("NormalizeSecret() = %q, %v", normalized, err)
	}
	if _, err := NormalizeSecret("not-a-base32-secret!"); !errors.Is(err, ErrInvalidSecret) {
		t.Fatalf("invalid secret error = %v", err)
	}
}

func TestSecondsRemainingUsesThirtySecondWindow(t *testing.T) {
	if got := SecondsRemaining(time.Unix(60, 0)); got != 30 {
		t.Fatalf("remaining at boundary = %d", got)
	}
	if got := SecondsRemaining(time.Unix(89, 0)); got != 1 {
		t.Fatalf("remaining before boundary = %d", got)
	}
}
