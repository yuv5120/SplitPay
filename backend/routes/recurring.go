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
)

func findRecurring(group models.Group, id string) (models.RecurringCost, bool) {
	for _, plan := range group.Recurring {
		if plan.ID == id {
			return plan, true
		}
	}
	return models.RecurringCost{}, false
}

func recurringGroup(c *fiber.Ctx) (models.Group, context.Context, context.CancelFunc, error) {
	user := middleware.GetUserFromContext(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	return group, ctx, cancel, err
}

func recurringLoadError(c *fiber.Ctx, err error) error {
	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Group not found"})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not load group"})
}

func createRecurring(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var input models.RecurringInput
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	group, ctx, cancel, err := recurringGroup(c)
	defer cancel()
	if err != nil {
		return recurringLoadError(c, err)
	}
	activeCount := 0
	for _, existing := range group.Recurring {
		if !existing.Archived {
			activeCount++
		}
	}
	if activeCount >= 50 || len(group.Recurring) >= 200 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "This group has reached its schedule limit"})
	}
	plan, err := models.NewRecurringCost(input, group.Members, user.UID, time.Now())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx,
		bson.M{"_id": group.ID, "recurring.199": bson.M{"$exists": false}},
		bson.M{"$push": bson.M{"recurring": plan, "activity": models.NewActivity("recurring_created", user.UID, "Scheduled "+plan.Description)}})
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Schedule limit reached; refresh and try again"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": plan})
}

func updateRecurring(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var input models.RecurringInput
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	group, ctx, cancel, err := recurringGroup(c)
	defer cancel()
	if err != nil {
		return recurringLoadError(c, err)
	}
	old, ok := findRecurring(group, c.Params("recurringId"))
	if !ok || old.Archived {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if old.CreatedBy != user.UID && group.UserID != user.UID {
		return c.SendStatus(fiber.StatusForbidden)
	}
	updated, err := models.NewRecurringCost(input, group.Members, user.UID, time.Now())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	updated.ID = old.ID
	updated.CreatedBy = old.CreatedBy
	updated.CreatedAt = old.CreatedAt
	updated.LastProposedPeriod = old.LastProposedPeriod
	updated.Version = old.Version + 1
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx,
		bson.M{"_id": group.ID, "recurring": bson.M{"$elemMatch": bson.M{"id": old.ID, "version": old.Version, "archived": bson.M{"$ne": true}}}},
		bson.M{"$set": bson.M{"recurring.$": updated}, "$push": bson.M{"activity": models.NewActivity("recurring_updated", user.UID, "Updated "+updated.Description)}})
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Schedule changed; refresh and try again"})
	}
	return c.JSON(fiber.Map{"success": true, "data": updated})
}

func archiveRecurring(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	group, ctx, cancel, err := recurringGroup(c)
	defer cancel()
	if err != nil {
		return recurringLoadError(c, err)
	}
	plan, ok := findRecurring(group, c.Params("recurringId"))
	if !ok || plan.Archived {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if plan.CreatedBy != user.UID && group.UserID != user.UID {
		return c.SendStatus(fiber.StatusForbidden)
	}
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx,
		bson.M{"_id": group.ID, "recurring": bson.M{"$elemMatch": bson.M{"id": plan.ID, "version": plan.Version, "archived": bson.M{"$ne": true}}}},
		bson.M{"$set": bson.M{"recurring.$.archived": true}, "$inc": bson.M{"recurring.$.version": 1},
			"$push": bson.M{"activity": models.NewActivity("recurring_archived", user.UID, "Archived "+plan.Description)}})
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Schedule changed; refresh and try again"})
	}
	return c.JSON(fiber.Map{"success": true})
}

func proposeRecurring(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		AmountPaise int64 `json:"amountPaise"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	group, ctx, cancel, err := recurringGroup(c)
	defer cancel()
	if err != nil {
		return recurringLoadError(c, err)
	}
	plan, ok := findRecurring(group, c.Params("recurringId"))
	if !ok || plan.Archived {
		return c.SendStatus(fiber.StatusNotFound)
	}
	now := time.Now().In(models.RecurringLocation)
	period := plan.UpcomingDue(now).Format("2006-01")
	if period <= plan.LastProposedPeriod {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "This period already has a proposal; wait until its due date passes"})
	}
	expense, err := models.NewRecurringProposal(plan, body.AmountPaise, group.Members, user.UID, now)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	average, count := models.RecentRecurringAverage(group.Expenses, plan.ID)
	source := "scheduled estimate"
	if count > 0 {
		source = "recent approved average"
	} else {
		average = plan.EstimatePaise
	}
	expense.AnomalyExplanation = models.RecurringAnomaly(expense.AmountPaise, average, source)
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx,
		bson.M{"_id": group.ID, "recurring": bson.M{"$elemMatch": bson.M{"id": plan.ID, "version": plan.Version, "archived": bson.M{"$ne": true}, "lastProposedPeriod": bson.M{"$ne": period}}}},
		bson.M{"$set": bson.M{"recurring.$.lastProposedPeriod": period},
			"$inc":  bson.M{"recurring.$.version": 1, "balanceRevision": 1},
			"$push": bson.M{"expenses": expense, "activity": models.NewActivity("recurring_proposed", user.UID, "Proposed scheduled "+plan.Description)}})
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "This schedule changed or its period was already proposed; refresh and try again"})
	}
	if expense.Status == models.ExpenseProposed {
		_ = notifyExpenseMembers(ctx, group, expense.ID, user.UID, "expense_proposal", "Approval requested for "+expense.Description, pendingMemberIDs(expense))
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": expense.ResponseFor(user.UID, group.UserID == user.UID)})
}
