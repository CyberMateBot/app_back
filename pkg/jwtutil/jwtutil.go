package jwtutil

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	WebAccountID int64  `json:"web_account_id"`
	Email        string `json:"email"`
	jwt.RegisteredClaims
}

func SignAccessToken(secret string, ttl time.Duration, webAccountID int64, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		WebAccountID: webAccountID,
		Email:        email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", webAccountID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

func ParseAccessToken(secret string, token string) (Claims, error) {
	var out Claims
	parsed, err := jwt.ParseWithClaims(token, &out, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return []byte(secret), nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !parsed.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}
	return out, nil
}

