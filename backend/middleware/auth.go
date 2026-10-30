package middleware

import (
	"context"
	"errors"
	"os"
	"regexp"
	"splitpay/backend/config"
	"splitpay/backend/models"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	if os.Getenv("AUTH_PROVIDER") == "clerk" {
		return authenticateClerk(c, token)
	}

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

func authenticateClerk(c *fiber.Ctx, token string) error {
	identity, err := config.VerifyClerkToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Invalid or expired session"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	users := config.GetDB().Collection("users")
	var user models.User
	err = users.FindOne(ctx, bson.M{"clerkUserId": identity.UserID}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		user, err = provisionClerkUser(ctx, identity.UserID)
	}
	if err != nil {
		if errors.Is(err, errLegacyConflict) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "This email belongs to an existing account. Ask the administrator to migrate it before signing in."})
		}
		if errors.Is(err, errUnverifiedEmail) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Verify your email before using SplitPay."})
		}
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"success": false, "message": "Account lookup is temporarily unavailable"})
	}
	uid := user.InternalUID
	if uid == "" {
		uid = user.FirebaseUID
	}
	if uid == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Account migration is incomplete"})
	}
	c.Locals("user", &UserContext{UID: uid, Email: user.Email, Name: user.Name, EmailVerified: user.EmailVerified})
	return c.Next()
}

var errLegacyConflict = errors.New("legacy identity conflict")
var errUnverifiedEmail = errors.New("unverified email")

func provisionClerkUser(ctx context.Context, clerkID string) (models.User, error) {
	identity, err := config.FetchClerkUser(ctx, clerkID)
	if err != nil {
		return models.User{}, err
	}
	email, verified := identity.VerifiedPrimaryEmail()
	if !verified {
		return models.User{}, errUnverifiedEmail
	}
	users := config.GetDB().Collection("users")
	if identity.ExternalID != "" {
		var legacy models.User
		err := users.FindOne(ctx, bson.M{"firebaseUid": identity.ExternalID}).Decode(&legacy)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return models.User{}, errLegacyConflict
			}
			return models.User{}, err
		}
		if !legacy.EmailVerified || !strings.EqualFold(legacy.Email, email) || legacy.ClerkUserID != "" {
			return models.User{}, errLegacyConflict
		}
		result := users.FindOneAndUpdate(ctx, bson.M{"_id": legacy.ID, "clerkUserId": bson.M{"$exists": false}, "emailVerified": true}, bson.M{"$set": bson.M{"clerkUserId": clerkID, "internalUid": identity.ExternalID, "updatedAt": time.Now()}}, options.FindOneAndUpdate().SetReturnDocument(options.After))
		var linked models.User
		if err := result.Decode(&linked); err != nil {
			return models.User{}, errLegacyConflict
		}
		return linked, nil
	}
	// Email alone never authorizes a link to a Firebase account.
	var existing models.User
	err = users.FindOne(ctx, bson.M{"email": bson.M{"$regex": "^" + regexp.QuoteMeta(email) + "$", "$options": "i"}}).Decode(&existing)
	if err == nil {
		return models.User{}, errLegacyConflict
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return models.User{}, err
	}
	id := primitive.NewObjectID()
	name := strings.TrimSpace(identity.FirstName + " " + identity.LastName)
	if name == "" {
		name = strings.Split(email, "@")[0]
	}
	now := time.Now()
	user := models.User{ID: id, InternalUID: id.Hex(), ClerkUserID: clerkID, Email: email, Name: name, EmailVerified: true, CreatedAt: now, UpdatedAt: now}
	_, err = users.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		if lookupErr := users.FindOne(ctx, bson.M{"clerkUserId": clerkID}).Decode(&existing); lookupErr == nil {
			return existing, nil
		}
	}
	return user, err
}

// GetUserFromContext retrieves user context from fiber context
func GetUserFromContext(c *fiber.Ctx) *UserContext {
	user := c.Locals("user")
	if user == nil {
		return nil
	}
	return user.(*UserContext)
}
