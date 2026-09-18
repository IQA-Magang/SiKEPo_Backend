package utils

import (
	"errors"
	"os"
	"strconv"
	"time"

	"backend/models"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(user *models.User) (string, error) {

	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		return "", errors.New("jwt secret not configured")
	}

	expiryMinutes := 60

	if value := os.Getenv("JWT_EXPIRY_MINUTES"); value != "" {

		if minutes, err := strconv.Atoi(value); err == nil && minutes > 0 {
			expiryMinutes = minutes
		}
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"user_id": user.UserID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     now.Add(
			time.Duration(expiryMinutes) * time.Minute,
		).Unix(),
		"iat": now.Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString(
		[]byte(secret),
	)

	if err != nil {
		return "", err
	}

	return signedToken, nil
}