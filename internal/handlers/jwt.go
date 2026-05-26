package handlers

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

var jwtKey = []byte("secretKey") // TODO

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.StandardClaims
}

// generateJWT создаёт JWT-токен для пользователя
func generateJWT(userID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)

	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("failed token.SignedString %w", err)
	}

	return signedStr, nil
}

// parseJWT парсит и валидирует JWT-токен
func parseJWT(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	if !token.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid token")
	}

	return claims.UserID, nil
}
