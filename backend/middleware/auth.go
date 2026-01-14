package middleware

import (
	"context"
	"split-it/backend/config"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// UserContext represents the authenticated user information
type UserContext struct {
	UID           string
	Email         string
	Name          string
	EmailVerified bool
}

// AuthenticateUser middleware verifies Firebase ID token
func AuthenticateUser(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "No token provided",
		})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	firebaseAuth := config.GetFirebaseAuth()
	if firebaseAuth == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Firebase authentication not initialized",
		})
	}

	decodedToken, err := firebaseAuth.VerifyIDToken(context.Background(), token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid or expired token",
		})
	}

	// Extract user information
	email := ""
	if val, ok := decodedToken.Claims["email"].(string); ok {
		email = val
	}

	name := email
	if val, ok := decodedToken.Claims["name"].(string); ok {
		name = val
	}

	// Extract email verification status
	emailVerified := false
	if val, ok := decodedToken.Claims["email_verified"].(bool); ok {
		emailVerified = val
	}

	// Store user context in locals
	c.Locals("user", &UserContext{
		UID:           decodedToken.UID,
		Email:         email,
		Name:          name,
		EmailVerified: emailVerified,
	})

	return c.Next()
}

// GetUserFromContext retrieves user context from fiber context
func GetUserFromContext(c *fiber.Ctx) *UserContext {
	user := c.Locals("user")
	if user == nil {
		return nil
	}
	return user.(*UserContext)
}
