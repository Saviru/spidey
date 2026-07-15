package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SpideyClaims struct {
	UserID string                 `json:"user_id"`
	Data   map[string]interface{} `json:"data,omitempty"`
	jwt.RegisteredClaims
}

// creates new signed JWT.
func GenerateToken(userID string, data map[string]interface{}, secretKey []byte, duration time.Duration) (string, error) {
	claims := SpideyClaims{
		UserID: userID,
		Data:   data,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "spidey-framework",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}

// parse and verify token
func ValidateToken(tokenString string, secretKey []byte) (*SpideyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &SpideyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*SpideyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
