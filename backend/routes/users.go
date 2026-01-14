package routes

import (
	"context"
	"split-it/backend/config"
	"split-it/backend/middleware"
	"split-it/backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupUserRoutes configures user-related routes
func SetupUserRoutes(app *fiber.App) {
	users := app.Group("/api/users")

	// Get or create user profile
	users.Post("/profile", middleware.AuthenticateUser, getOrCreateProfile)

	// Update user profile
	users.Put("/profile", middleware.AuthenticateUser, updateProfile)
}

func getOrCreateProfile(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Parse request body for phone number
	var body struct {
		Phone string `json:"phone"`
	}
	c.BodyParser(&body) // Ignore error, phone is optional

	db := config.GetDB()
	collection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to find existing user
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"firebaseUid": user.UID}).Decode(&existingUser)

	if err == mongo.ErrNoDocuments {
		// Create new user
		phone := body.Phone

		newUser := models.User{
			ID:            primitive.NewObjectID(),
			FirebaseUID:   user.UID,
			Email:         user.Email,
			Name:          user.Name,
			Phone:         phone,
			EmailVerified: user.EmailVerified,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		_, err := collection.InsertOne(ctx, newUser)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error creating user profile",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data": models.UserResponse{
				ID:            newUser.ID.Hex(),
				FirebaseUID:   newUser.FirebaseUID,
				Email:         newUser.Email,
				Name:          newUser.Name,
				Phone:         newUser.Phone,
				EmailVerified: newUser.EmailVerified,
				CreatedAt:     newUser.CreatedAt,
			},
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching user profile",
		})
	}

	// Return existing user
	return c.JSON(fiber.Map{
		"success": true,
		"data": models.UserResponse{
			ID:            existingUser.ID.Hex(),
			FirebaseUID:   existingUser.FirebaseUID,
			Email:         existingUser.Email,
			Name:          existingUser.Name,
			Phone:         existingUser.Phone,
			EmailVerified: existingUser.EmailVerified,
			CreatedAt:     existingUser.CreatedAt,
		},
	})
}

func updateProfile(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	var body struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	db := config.GetDB()
	collection := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"name":      body.Name,
			"phone":     body.Phone,
			"updatedAt": time.Now(),
		},
	}

	var updatedUser models.User
	err := collection.FindOneAndUpdate(
		ctx,
		bson.M{"firebaseUid": user.UID},
		update,
		// mongo.options.FindOneAndUpdate().SetReturnDocument(mongo.options.After),
	).Decode(&updatedUser)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": models.UserResponse{
			ID:            updatedUser.ID.Hex(),
			FirebaseUID:   updatedUser.FirebaseUID,
			Email:         updatedUser.Email,
			Name:          body.Name,  // Use the updated name from body
			Phone:         body.Phone, // Use the updated phone from body
			EmailVerified: updatedUser.EmailVerified,
		},
	})
}
