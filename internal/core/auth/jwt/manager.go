package core_jwt

import (
	"fmt"
	"strconv"
	"time"

	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(config Config) *TokenManager {
	return &TokenManager{
		secret: []byte(config.Secret),
		ttl:    config.TTL,
	}
}

func (m *TokenManager) NewToken(userID int64) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return token, nil
}

func (m *TokenManager) ParseToken(token string) (int64, error) {
	var claims jwt.RegisteredClaims

	// WithValidMethods обязателен: иначе можно подсунуть токен с alg=none
	// или другим алгоритмом и обойти проверку подписи
	_, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return 0, fmt.Errorf("parse token: %v: %w", err, core_errors.ErrUnauthenticated)
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid token subject='%s': %w", claims.Subject, core_errors.ErrUnauthenticated)
	}

	return userID, nil
}
