package routes

import (
	"context"
	"splitpay/backend/config"
	"splitpay/backend/middleware"
	"splitpay/backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func updateGroupBudget(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		LimitPaise *int64 `json:"limitPaise"`
	}
	if err := c.BodyParser(&body); err != nil || body.LimitPaise == nil || *body.LimitPaise < 0 || *body.LimitPaise > 100_000_000_000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Budget must be 0 to ₹1 billion; 0 removes the target"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var group models.Group
	err := config.GetDB().Collection("groups").FindOneAndUpdate(ctx,
		bson.M{"id": c.Params("groupId"), "userId": user.UID},
		bson.M{"$set": bson.M{"monthlyBudgetPaise": *body.LimitPaise, "updatedAt": time.Now()},
			"$push": bson.M{"activity": models.NewActivity("budget_updated", user.UID, "Updated the monthly spending target")}},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&group)
	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Group not found or you cannot edit its budget"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not save budget"})
	}
	return c.JSON(fiber.Map{"success": true, "data": group.ResponseFor(user.UID).Budget})
}
