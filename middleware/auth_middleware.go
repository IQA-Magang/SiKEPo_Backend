package middleware

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"backend/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// RequireAuth verifies JWT and sets user info in context locals.
func RequireAuth(c *fiber.Ctx) error {
	auth := c.Get("Authorization")

	if auth == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Missing Authorization header",
		})
	}

	// Expect Bearer <token>
	var tokenString string

	if len(auth) > 7 && auth[:7] == "Bearer " {
		tokenString = auth[7:]
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid Authorization header format",
		})
	}

	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "JWT secret not configured",
		})
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		errMsg := ""

		if err != nil {
			errMsg = err.Error()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid or expired token",
			"error":   errMsg,
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid token claims",
		})
	}

	// Store JWT claims in Fiber Locals
	if v, ok := claims["user_id"]; ok {
		c.Locals("user_id", v)
	}

	if v, ok := claims["email"]; ok {
		c.Locals("email", v)
	}

	if v, ok := claims["role"]; ok {
		c.Locals("role", v)
	}

	if pengelolaValue := claims["pengelola"]; pengelolaValue != nil {
		switch value := pengelolaValue.(type) {
		case bool:
			c.Locals("pengelola", value)
		case string:
			if parsed, err := strconv.ParseBool(value); err == nil {
				c.Locals("pengelola", parsed)
			}
		}
	}

	return c.Next()
}

// RequireRoles ensures the authenticated user has one of the allowed roles.
func RequireRoles(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)

		if !ok || role == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Role not found",
			})
		}

		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "You do not have permission to access this resource",
		})
	}
}

func RequireDatabaseRoles(db *gorm.DB, allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		email, ok := c.Locals("email").(string)
		email = strings.TrimSpace(email)
		if !ok || email == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Identitas pengguna tidak valid",
			})
		}

		var user models.User
		if err := db.Select("user_id", "email", "role").
			Where("email = ?", email).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"message": "Pengguna tidak ditemukan atau tidak memiliki akses",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal memverifikasi role pengguna",
			})
		}

		role := strings.ToLower(strings.TrimSpace(user.Role))
		c.Locals("user_id", user.UserID)
		c.Locals("role", role)
		c.Locals("email", user.Email)

		for _, allowedRole := range allowedRoles {
			if role == strings.ToLower(strings.TrimSpace(allowedRole)) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Akun ini tidak memiliki role Manager di backend",
		})
	}
}

// RequireAdminOrStaffPengelola allows admin or staff marked as pengelola.
func RequireAdminOrStaffPengelola() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Role not found",
			})
		}

		if role == "admin" {
			return c.Next()
		}

		if role == "staff" {
			pengelolaValue := c.Locals("pengelola")
			if pengelolaValue != nil {
				if pengelola, ok := pengelolaValue.(bool); ok && pengelola {
					return c.Next()
				}
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Hanya admin atau staff pengelola yang dapat mengakses endpoint ini",
		})
	}
}
