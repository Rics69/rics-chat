package core_jwt

import (
	"errors"
	"testing"
	"time"

	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
	"github.com/golang-jwt/jwt/v5"
)

func TestTokenManagerRoundTrip(t *testing.T) {
	manager := NewTokenManager(Config{Secret: "secret", TTL: time.Hour})

	token, err := manager.NewToken(42)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	userID, expiresAt, err := manager.ParseTokenWithExpiry(token)
	if err != nil {
		t.Fatalf("ParseTokenWithExpiry: %v", err)
	}

	if userID != 42 {
		t.Fatalf("userID = %d, want 42", userID)
	}

	if until := time.Until(expiresAt); until < 59*time.Minute || until > time.Hour {
		t.Fatalf("expiresAt in %s, want ~1h", until)
	}
}

func TestTokenManagerRejects(t *testing.T) {
	manager := NewTokenManager(Config{Secret: "secret", TTL: time.Hour})

	expired, _ := NewTokenManager(Config{Secret: "secret", TTL: -time.Minute}).NewToken(1)
	otherSecret, _ := NewTokenManager(Config{Secret: "other", TTL: time.Hour}).NewToken(1)
	algNone, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		Subject:   "1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)

	tests := map[string]string{
		"garbage":      "abc.def.ghi",
		"expired":      expired,
		"other secret": otherSecret,
		"alg none":     algNone,
	}

	for name, token := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := manager.ParseToken(token)
			if !errors.Is(err, core_errors.ErrUnauthenticated) {
				t.Fatalf("error = %v, want ErrUnauthenticated", err)
			}
		})
	}
}
