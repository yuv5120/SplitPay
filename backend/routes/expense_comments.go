package routes

import (
	"context"
	"splitpay/backend/config"
	"splitpay/backend/middleware"
	"splitpay/backend/models"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type expenseComment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	GroupID   primitive.ObjectID `bson:"groupId"`
	ExpenseID string             `bson:"expenseId"`
	AuthorUID string             `bson:"authorUid"`
	Text      string             `bson:"text"`
	CreatedAt time.Time          `bson:"createdAt"`
}

type commentResponse struct {
	ID         string    `json:"id"`
	AuthorName string    `json:"authorName"`
	Text       string    `json:"text"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (comment expenseComment) response(group models.Group) commentResponse {
	name := "Former member"
	if comment.AuthorUID == group.UserID {
		name = "Group owner"
	}
	for _, member := range group.Members {
		if member.UID == comment.AuthorUID && comment.AuthorUID != "" {
			name = member.Name
			break
		}
	}
	return commentResponse{ID: comment.ID.Hex(), AuthorName: name, Text: comment.Text, CreatedAt: comment.CreatedAt}
}

func getExpenseComments(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Group not found"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not load group"})
	}
	if _, ok := findExpense(group, c.Params("expenseId")); !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Expense not found"})
	}
	cursor, err := config.GetDB().Collection("expenseComments").Find(ctx, bson.M{
		"groupId": group.ID, "expenseId": c.Params("expenseId"),
	}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}).SetLimit(200))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not load comments"})
	}
	defer cursor.Close(ctx)
	var comments []expenseComment
	if err := cursor.All(ctx, &comments); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not decode comments"})
	}
	response := make([]commentResponse, len(comments))
	for i, comment := range comments {
		response[i] = comment.response(group)
	}
	return c.JSON(fiber.Map{"success": true, "data": response})
}

func addExpenseComment(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		Text string `json:"text"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid comment body"})
	}
	body.Text = strings.TrimSpace(body.Text)
	if body.Text == "" || len(body.Text) > 1000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Comment must be 1 to 1000 characters"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Group not found"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not load group"})
	}
	expense, ok := findExpense(group, c.Params("expenseId"))
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Expense not found"})
	}
	comment := expenseComment{
		ID: primitive.NewObjectID(), GroupID: group.ID, ExpenseID: c.Params("expenseId"),
		AuthorUID: user.UID, Text: body.Text, CreatedAt: time.Now(),
	}
	if _, err := config.GetDB().Collection("expenseComments").InsertOne(ctx, comment); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not add comment"})
	}
	_, _ = config.GetDB().Collection("groups").UpdateOne(ctx, bson.M{"_id": group.ID}, bson.M{
		"$push": bson.M{"activity": models.NewActivity("expense_commented", user.UID, "Commented on an expense")},
		"$set":  bson.M{"updatedAt": time.Now()},
	})
	_ = notifyExpenseMembers(ctx, group, comment.ExpenseID, user.UID, "expense_comment", "New comment on "+expense.Description, nil)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": comment.response(group)})
}
