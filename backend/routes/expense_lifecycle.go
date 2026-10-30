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

func loadAccessibleGroup(ctx context.Context, groupID, uid string) (models.Group, error) {
	var group models.Group
	err := config.GetDB().Collection("groups").FindOne(ctx, groupIDAccessFilter(groupID, uid)).Decode(&group)
	return group, err
}

func findExpense(group models.Group, expenseID string) (models.Expense, bool) {
	for _, expense := range group.Expenses {
		if expense.ID == expenseID {
			return expense, true
		}
	}
	return models.Expense{}, false
}

func memberIDForUID(group models.Group, uid string) string {
	for _, member := range group.Members {
		if member.UID == uid && uid != "" {
			return member.ID
		}
	}
	return ""
}

func replaceExpense(ctx context.Context, group models.Group, before, after models.Expense, actorUID, kind, detail string) (bool, error) {
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx, bson.M{
		"_id":      group.ID,
		"expenses": bson.M{"$elemMatch": bson.M{"id": before.ID, "version": before.Version}},
	}, bson.M{
		"$set":  bson.M{"expenses.$": after, "updatedAt": time.Now()},
		"$push": bson.M{"activity": models.NewActivity(kind, actorUID, detail)},
		"$inc":  bson.M{"balanceRevision": 1},
	})
	if err != nil {
		return false, err
	}
	return result.MatchedCount > 0, nil
}

func decideExpense(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		Decision string `json:"decision"`
	}
	if err := c.BodyParser(&body); err != nil || (body.Decision != "approved" && body.Decision != "rejected") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Decision must be approved or rejected"})
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
	memberID := memberIDForUID(group, user.UID)
	if memberID == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Link your account to a member before deciding"})
	}
	participant := false
	for _, id := range expense.Participants {
		participant = participant || id == memberID
	}
	if !participant {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Only included participants can decide"})
	}
	updated, err := expense.Decide(memberID, body.Decision, time.Now())
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	kind, detail := "expense_approved", "Approved "+expense.Description
	if body.Decision == "rejected" {
		kind, detail = "expense_rejected", "Rejected "+expense.Description
	}
	matched, err := replaceExpense(ctx, group, expense, updated, user.UID, kind, detail)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not save decision"})
	}
	if !matched {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Expense changed; refresh and try again"})
	}
	_ = notifyExpenseMembers(ctx, group, expense.ID, user.UID, "expense_decision", "New decision on "+expense.Description, nil)
	return c.JSON(fiber.Map{"success": true, "data": updated.ResponseFor(user.UID, group.UserID == user.UID)})
}

func resubmitExpense(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		Description  string   `json:"description"`
		AmountPaise  int64    `json:"amountPaise"`
		PaidBy       string   `json:"paidBy"`
		Participants []string `json:"participants"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid expense body"})
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
	if expense.CreatedBy != user.UID || expense.Status != models.ExpenseRejected {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Only the proposer can resubmit a rejected expense"})
	}
	replacement, err := models.NewProposedExpense(models.ExpenseInput{
		Description: body.Description, AmountPaise: body.AmountPaise,
		PaidBy: body.PaidBy, Participants: body.Participants,
	}, group.Members, user.UID, time.Now())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	updated, err := expense.Resubmit(replacement)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	if updated.RecurringID != "" {
		average, count := models.RecentRecurringAverage(group.Expenses, updated.RecurringID)
		source := "recent approved average"
		if count == 0 {
			plan, ok := findRecurring(group, updated.RecurringID)
			if ok {
				average = plan.EstimatePaise
			} else {
				average = expense.AmountPaise
			}
			source = "scheduled estimate"
		}
		updated.AnomalyExplanation = models.RecurringAnomaly(updated.AmountPaise, average, source)
	}
	matched, err := replaceExpense(ctx, group, expense, updated, user.UID, "expense_resubmitted", "Resubmitted "+updated.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not resubmit expense"})
	}
	if !matched {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Expense changed; refresh and try again"})
	}
	_ = notifyExpenseMembers(ctx, group, expense.ID, user.UID, "expense_proposal", "Approval requested for "+updated.Description, pendingMemberIDs(updated))
	return c.JSON(fiber.Map{"success": true, "data": updated.ResponseFor(user.UID, group.UserID == user.UID)})
}

func withdrawExpense(c *fiber.Ctx) error {
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
	expense, ok := findExpense(group, c.Params("expenseId"))
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Expense not found"})
	}
	updated, err := expense.Withdraw(user.UID, time.Now())
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	matched, err := replaceExpense(ctx, group, expense, updated, user.UID, "expense_withdrawn", "Withdrew "+expense.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not withdraw expense"})
	}
	if !matched {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Expense changed; refresh and try again"})
	}
	return c.JSON(fiber.Map{"success": true, "data": updated.ResponseFor(user.UID, group.UserID == user.UID)})
}
