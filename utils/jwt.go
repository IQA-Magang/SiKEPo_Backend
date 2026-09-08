package utils

import (
	"errors"
	"os"
	"strconv"
	"time"

	"backend/models"

	"github.com/golang-jwt/jwt/v5"
)

// CreateToken creates a signed JWT for the given user.
func CreateToken(user *models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("jwt secret not configured")
	}

	expiryMinutes := 60
	if v := os.Getenv("JWT_EXPIRY_MINUTES"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			expiryMinutes = i
		}
	}

	claims := jwt.MapClaims{
		"user_id": user.UserID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Duration(expiryMinutes) * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signed, nil
}
