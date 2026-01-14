package routes

import (
	"context"
	"split-it/backend/config"
	"split-it/backend/middleware"
	"split-it/backend/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetupGroupRoutes configures group-related routes
func SetupGroupRoutes(app *fiber.App) {
	groups := app.Group("/api/groups", middleware.AuthenticateUser)

	// Group CRUD operations
	groups.Get("/", getAllGroups)
	groups.Get("/:groupId", getGroup)
	groups.Post("/", createGroup)
	groups.Put("/:groupId", updateGroup)
	groups.Delete("/:groupId", deleteGroup)

	// Expense operations
	groups.Post("/:groupId/expenses", addExpense)
	groups.Delete("/:groupId/expenses/:expenseId", deleteExpense)
}

func getAllGroups(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	db := config.GetDB()
	collection := db.Collection("groups")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := collection.Find(ctx, bson.M{"userId": user.UID}, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching groups",
		})
	}
	defer cursor.Close(ctx)

	var groups []models.Group
	if err = cursor.All(ctx, &groups); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error decoding groups",
		})
	}

	// Transform to response format
	response := make([]models.GroupResponse, len(groups))
	for i, g := range groups {
		response[i] = models.GroupResponse{
			ID:        g.GroupID,
			Name:      g.Name,
			Members:   g.Members,
			Expenses:  g.Expenses,
			CreatedAt: g.CreatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

func getGroup(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	groupId := c.Params("groupId")

	db := config.GetDB()
	collection := db.Collection("groups")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var group models.Group
	err := collection.FindOne(ctx, bson.M{
		"id":     groupId,
		"userId": user.UID,
	}).Decode(&group)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Group not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching group",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": models.GroupResponse{
			ID:        group.GroupID,
			Name:      group.Name,
			Members:   group.Members,
			Expenses:  group.Expenses,
			CreatedAt: group.CreatedAt,
		},
	})
}

func createGroup(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	var body struct {
		ID      string          `json:"id"`
		Name    string          `json:"name"`
		Members []models.Member `json:"members"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	if body.Name == "" || len(body.Members) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Group name and at least 2 members are required",
		})
	}

	// Generate ID if not provided
	groupID := body.ID
	if groupID == "" {
		groupID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	newGroup := models.Group{
		ID:        primitive.NewObjectID(),
		GroupID:   groupID,
		Name:      body.Name,
		Members:   body.Members,
		Expenses:  []models.Expense{},
		UserID:    user.UID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	db := config.GetDB()
	collection := db.Collection("groups")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, newGroup)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error creating group",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": models.GroupResponse{
			ID:        newGroup.GroupID,
			Name:      newGroup.Name,
			Members:   newGroup.Members,
			Expenses:  newGroup.Expenses,
			CreatedAt: newGroup.CreatedAt,
		},
	})
}

func updateGroup(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	groupId := c.Params("groupId")

	var body struct {
		Name     string           `json:"name"`
		Members  []models.Member  `json:"members"`
		Expenses []models.Expense `json:"expenses"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	db := config.GetDB()
	collection := db.Collection("groups")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"name":      body.Name,
			"members":   body.Members,
			"expenses":  body.Expenses,
			"updatedAt": time.Now(),
		},
	}

	var updatedGroup models.Group
	err := collection.FindOneAndUpdate(
		ctx,
		bson.M{"id": groupId, "userId": user.UID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedGroup)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Group not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error updating group",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": models.GroupResponse{
			ID:        updatedGroup.GroupID,
			Name:      updatedGroup.Name,
			Members:   updatedGroup.Members,
			Expenses:  updatedGroup.Expenses,
			CreatedAt: updatedGroup.CreatedAt,
		},
	})
}

func deleteGroup(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	groupId := c.Params("groupId")

	db := config.GetDB()
	collection := db.Collection("groups")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{
		"id":     groupId,
		"userId": user.UID,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error deleting group",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Group not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Group deleted successfully",
	})
}

func addExpense(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	groupId := c.Params("groupId")

	var body struct {
		ID           string   `json:"id"`
		Description  string   `json:"description"`
		Amount       float64  `json:"amount"`
		PaidBy       string   `json:"paidBy"`
		Participants []string `json:"participants"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	if body.Description == "" || body.Amount == 0 || body.PaidBy == "" || len(body.Participants) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "All expense fields are required",
		})
	}

	// Generate ID if not provided
	expenseID := body.ID
	if expenseID == "" {
		expenseID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	newExpense := models.Expense{
		ID:           expenseID,
		Description:  body.Description,
		Amount:       body.Amount,
		PaidBy:       body.PaidBy,
		Participants: body.Participants,
		Date:         time.Now(),
	}

	db := config.GetDB()
	collection := db.Collection("groups")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$push": bson.M{"expenses": newExpense},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	var updatedGroup models.Group
	err := collection.FindOneAndUpdate(
		ctx,
		bson.M{"id": groupId, "userId": user.UID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedGroup)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Group not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error adding expense",
		})
	}

	// Return the newly added expense
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    newExpense,
	})
}

func deleteExpense(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	groupId := c.Params("groupId")
	expenseId := c.Params("expenseId")

	db := config.GetDB()
	collection := db.Collection("groups")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$pull": bson.M{"expenses": bson.M{"id": expenseId}},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := collection.UpdateOne(
		ctx,
		bson.M{"id": groupId, "userId": user.UID},
		update,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error deleting expense",
		})
	}

	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Group not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Expense deleted successfully",
	})
}
