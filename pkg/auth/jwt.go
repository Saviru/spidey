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

// creates new signed JWT with a Key ID (kid) in the header for dynamic key lookup.
func GenerateTokenWithKID(userID string, data map[string]interface{}, secretKey []byte, duration time.Duration, kid string) (string, error) {
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
	token.Header["kid"] = kid

	return token.SignedString(secretKey)
}

// creates both a short-lived access token and a long-lived refresh token
func GenerateTokenPair(userID string, data map[string]interface{}, secretKey []byte, accessDuration, refreshDuration time.Duration) (string, string, error) {
	accessToken, err := GenerateToken(userID, data, secretKey, accessDuration)
	if err != nil {
		return "", "", err
	}

	refreshClaims := SpideyClaims{
		UserID: userID,
		// Omit custom data in refresh token to keep it small
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "spidey-framework",
			Subject:   "refresh",
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// parses and verifies a token using a dynamic key lookup function (jwt.Keyfunc).
func ValidateTokenDynamic(tokenString string, keyFunc jwt.Keyfunc) (*SpideyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &SpideyClaims{}, keyFunc)

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*SpideyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// parse and verify token using a static secret key
func ValidateToken(tokenString string, secretKey []byte) (*SpideyClaims, error) {
	return ValidateTokenDynamic(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})
}

// strictly validates that the provided token is a refresh token
func ValidateRefreshToken(tokenString string, secretKey []byte) (*SpideyClaims, error) {
	claims, err := ValidateToken(tokenString, secretKey)
	if err != nil {
		return nil, err
	}

	if claims.Subject != "refresh" {
		return nil, errors.New("invalid token type: not a refresh token")
	}

	return claims, nil
}
