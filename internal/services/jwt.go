package services

import (
	"fmt"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// ValidateToken Returns userID and error. Returns error if token is invalid.
func (s *Services) ValidateToken(token string) (string, error) {
	var claims jwt.RegisteredClaims
	accessToken, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", errs.ErrUnexpectedSigningMethod, token.Header["alg"])
		}

		return s.key, nil
	})

	if err != nil {
		return "", err
	}

	if !accessToken.Valid {
		return "", errs.ErrNotAuthorized
	}

	return claims.Subject, nil
}

// CreateToken returns a new token and an error
func (s *Services) CreateToken(userID string) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(3600 * time.Second)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	accessTokenSigned, err := accessToken.SignedString(s.key)
	if err != nil {
		return "", err
	}

	return accessTokenSigned, nil
}

// RefreshToken returns a new token with same user_id and an error
func (s *Services) RefreshToken(userID string) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(3600 * time.Second)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	accessTokenSigned, err := accessToken.SignedString(s.key)
	if err != nil {
		return "", err
	}

	return accessTokenSigned, nil
}
