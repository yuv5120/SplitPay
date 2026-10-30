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

// SetupGroupRoutes configures group-related routes
func SetupGroupRoutes(app *fiber.App) {
	app.Post("/api/webhooks/razorpay-test", razorpayTestWebhook)
	app.Get("/api/payment-config", middleware.AuthenticateUser, paymentConfig)
	groups := app.Group("/api/groups", middleware.AuthenticateUser)

	// Group CRUD operations
	groups.Get("/", getAllGroups)
	groups.Get("/:groupId", getGroup)
	groups.Post("/", createGroup)
	groups.Put("/:groupId", updateGroup)
	groups.Put("/:groupId/budget", updateGroupBudget)
	groups.Delete("/:groupId", deleteGroup)
	groups.Post("/:groupId/members/:memberId/claim", claimOwnerMember)
	groups.Post("/:groupId/members/:memberId/invite", createGroupInvite)
	app.Post("/api/invitations/:token/accept", middleware.AuthenticateUser, acceptGroupInvite)

	// Expense operations
	groups.Post("/:groupId/expenses", addExpense)
	groups.Delete("/:groupId/expenses/:expenseId", deleteExpense)
	groups.Post("/:groupId/expenses/:expenseId/decision", decideExpense)
	groups.Put("/:groupId/expenses/:expenseId", resubmitExpense)
	groups.Post("/:groupId/expenses/:expenseId/withdraw", withdrawExpense)
	groups.Get("/:groupId/expenses/:expenseId/comments", getExpenseComments)
	groups.Post("/:groupId/expenses/:expenseId/comments", addExpenseComment)
	groups.Get("/:groupId/notifications", getExpenseNotifications)
	groups.Post("/:groupId/notifications/:notificationId/read", readExpenseNotification)
	groups.Post("/:groupId/payments", startPayment)
	groups.Post("/:groupId/payments/:paymentId/confirm", confirmPayment)
	groups.Post("/:groupId/payments/:paymentId/verify", verifyGatewayPayment)
	groups.Post("/:groupId/payments/:paymentId/reconcile", reconcileGatewayPayment)
	groups.Put("/:groupId/reminders/preference", setReminderPreference)
	groups.Post("/:groupId/reminders", sendReminder)
	groups.Post("/:groupId/recurring", createRecurring)
	groups.Put("/:groupId/recurring/:recurringId", updateRecurring)
	groups.Post("/:groupId/recurring/:recurringId/archive", archiveRecurring)
	groups.Post("/:groupId/recurring/:recurringId/proposals", proposeRecurring)
	groups.Post("/:groupId/purchase-plans", createPurchasePlan)
	groups.Post("/:groupId/purchase-plans/:planId/vote", votePurchasePlan)
	groups.Post("/:groupId/purchase-plans/:planId/comments", commentPurchasePlan)
	groups.Post("/:groupId/purchase-plans/:planId/close", closePurchasePlan)
	groups.Post("/:groupId/purchase-plans/:planId/cancel", cancelPurchasePlan)
	groups.Post("/:groupId/purchase-plans/:planId/convert", convertPurchasePlan)
}

func groupAccessFilter(uid string) bson.M {
	return bson.M{"$or": []bson.M{{"userId": uid}, {"members.uid": uid}}}
}

func groupIDAccessFilter(groupID, uid string) bson.M {
	return bson.M{"id": groupID, "$or": []bson.M{{"userId": uid}, {"members.uid": uid}}}
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
	cursor, err := collection.Find(ctx, groupAccessFilter(user.UID), opts)
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
		response[i] = g.ResponseFor(user.UID)
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
	err := collection.FindOne(ctx, groupIDAccessFilter(groupId, user.UID)).Decode(&group)

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
		"data":    group.ResponseFor(user.UID),
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
		Name    string          `json:"name"`
		Members []models.Member `json:"members"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" || len(body.Name) > 100 || len(body.Members) < 1 || len(body.Members) > 49 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Group name (up to 100 characters) and 1 to 49 other members are required",
		})
	}
	ownerName := strings.TrimSpace(user.Name)
	if ownerName == "" {
		ownerName = user.Email
	}
	if ownerName == "" {
		ownerName = "Group owner"
	}
	body.Members = append([]models.Member{{ID: primitive.NewObjectID().Hex(), Name: ownerName, UID: user.UID}}, body.Members...)

	for i := 1; i < len(body.Members); i++ {
		name := strings.TrimSpace(body.Members[i].Name)
		if name == "" || len(name) > 100 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false, "message": "Each member needs a name of up to 100 characters",
			})
		}
		body.Members[i] = models.Member{ID: primitive.NewObjectID().Hex(), Name: name}
	}

	newGroup := models.Group{
		ID:        primitive.NewObjectID(),
		GroupID:   primitive.NewObjectID().Hex(),
		Name:      body.Name,
		Members:   body.Members,
		Expenses:  []models.Expense{},
		Activity:  []models.Activity{models.NewActivity("group_created", user.UID, "Created the group")},
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
		"data":    newGroup.ResponseFor(user.UID),
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
		Name string `json:"name"`
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

	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" || len(body.Name) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Group name must be 1 to 100 characters"})
	}

	update := bson.M{
		"$set": bson.M{
			"name":      body.Name,
			"updatedAt": time.Now(),
		},
		"$push": bson.M{"activity": models.NewActivity("group_renamed", user.UID, "Renamed the group")},
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
		"data":    updatedGroup.ResponseFor(user.UID),
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
		Description  string   `json:"description"`
		AmountPaise  int64    `json:"amountPaise"`
		PaidBy       string   `json:"paidBy"`
		Participants []string `json:"participants"`
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
	var group models.Group
	if err := collection.FindOne(ctx, groupIDAccessFilter(groupId, user.UID)).Decode(&group); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Group not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Error fetching group"})
	}
	newExpense, err := models.NewProposedExpense(models.ExpenseInput{
		Description: body.Description, AmountPaise: body.AmountPaise,
		PaidBy: body.PaidBy, Participants: body.Participants,
	}, group.Members, user.UID, time.Now())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	activityKind, activityDetail := "expense_proposed", "Proposed "+newExpense.Description
	if newExpense.Status == models.ExpenseActive {
		activityKind, activityDetail = "expense_added", "Added "+newExpense.Description
	}

	update := bson.M{
		"$push": bson.M{
			"expenses": newExpense,
			"activity": models.NewActivity(activityKind, user.UID, activityDetail),
		},
		"$inc": bson.M{"balanceRevision": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	result, err := collection.UpdateOne(
		ctx,
		groupIDAccessFilter(groupId, user.UID),
		update,
	)

	if err == nil && result.MatchedCount == 0 {
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
	if newExpense.Status == models.ExpenseProposed {
		_ = notifyExpenseMembers(ctx, group, newExpense.ID, user.UID, "expense_proposal", "Approval requested for "+newExpense.Description, pendingMemberIDs(newExpense))
	}

	// Return the newly added expense
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    newExpense.ResponseFor(user.UID, group.UserID == user.UID),
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
	var group models.Group
	if err := collection.FindOne(ctx, groupIDAccessFilter(groupId, user.UID)).Decode(&group); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Group not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Error fetching group"})
	}
	allowed := group.UserID == user.UID
	for _, expense := range group.Expenses {
		if expense.ID == expenseId {
			if expense.Status != "" {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Approved and proposed expenses cannot be deleted; withdraw a proposal or resolve a correction with the group"})
			}
			allowed = allowed || expense.CreatedBy == user.UID
			break
		}
	}
	for _, payment := range group.Payments {
		if payment.Status == models.PaymentSucceeded {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Expenses cannot be deleted after a settlement has been recorded"})
		}
	}
	if !allowed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Only the expense creator or group owner can delete it"})
	}

	update := bson.M{
		"$pull": bson.M{"expenses": bson.M{"id": expenseId}},
		"$push": bson.M{"activity": models.NewActivity("expense_deleted", user.UID, "Deleted an expense")},
		"$inc":  bson.M{"balanceRevision": 1},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	filter := bson.M{"id": groupId, "expenses": bson.M{"$elemMatch": bson.M{"id": expenseId, "status": bson.M{"$exists": false}}},
		"$or":             []bson.M{{"userId": user.UID}, {"members.uid": user.UID}},
		"payments.status": bson.M{"$ne": models.PaymentSucceeded}}
	for key, value := range balanceRevisionFilter(group.BalanceRevision) {
		filter[key] = value
	}
	result, err := collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error deleting expense",
		})
	}

	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Expense not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Expense deleted successfully",
	})
}
