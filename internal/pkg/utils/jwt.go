/**
 * 功能：jwt.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

// Package utils documentation
package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redgreat/teweicun/internal/config"
)

type CustomClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token for a user
func GenerateToken(userID int64, username string) (string, error) {
	expireHours := config.GlobalConfig.JWT.ExpireHours
	if expireHours <= 0 {
		expireHours = 24
	}

	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "teweicun",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(config.GlobalConfig.JWT.Secret)
	
	return token.SignedString(secret)
}

// ParseToken parses and validates a JWT token
func ParseToken(tokenStr string) (*CustomClaims, error) {
	secret := []byte(config.GlobalConfig.JWT.Secret)

	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
