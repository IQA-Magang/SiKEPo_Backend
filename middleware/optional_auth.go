package middleware

import (
	"errors"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// OptionalAuth mengisi data user bila request membawa token yang valid, tanpa
// menolak request yang tidak membawa token atau tokennya salah.
//
// Dipakai pada endpoint publik yang menampilkan data tambahan hanya untuk user
// yang login. Pengecekannya: c.Locals("user_id") != nil.
func OptionalAuth(c *fiber.Ctx) error {
	auth := c.Get("Authorization")

	if len(auth) <= 7 || auth[:7] != "Bearer " {
		return c.Next()
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return c.Next()
	}

	token, err := jwt.Parse(auth[7:], func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return c.Next()
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Next()
	}

	if v, ok := claims["user_id"]; ok {
		c.Locals("user_id", v)
	}

	if v, ok := claims["role"]; ok {
		c.Locals("role", v)
	}

	return c.Next()
}
