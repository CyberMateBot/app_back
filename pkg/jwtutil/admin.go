package jwtutil

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AdminClaims struct {
	AdminID int64  `json:"admin_id"`
	Email   string `json:"email"`
	jwt.RegisteredClaims
}

func SignAdminToken(secret string, ttl time.Duration, adminID int64, email string) (string, error) {
	now := time.Now()
	claims := AdminClaims{
		AdminID: adminID,
		Email:   email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("admin:%d", adminID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

func ParseAdminToken(secret string, token string) (AdminClaims, error) {
	var out AdminClaims
	parsed, err := jwt.ParseWithClaims(token, &out, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return []byte(secret), nil
	})
	if err != nil {
		return AdminClaims{}, err
	}
	if !parsed.Valid || out.AdminID <= 0 {
		return AdminClaims{}, fmt.Errorf("invalid token")
	}
	return out, nil
}
