package middleware

import (
	"errors"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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
