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
)

func loadPurchasePlan(c *fiber.Ctx) (models.Group, models.PurchasePlan, context.Context, context.CancelFunc, error) {
	user := middleware.GetUserFromContext(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	if err != nil {
		return group, models.PurchasePlan{}, ctx, cancel, err
	}
	for _, plan := range group.PurchasePlans {
		if plan.ID == c.Params("planId") {
			return group, plan, ctx, cancel, nil
		}
	}
	return group, models.PurchasePlan{}, ctx, cancel, mongo.ErrNoDocuments
}

func purchasePlanLoadError(c *fiber.Ctx, err error) error {
	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Plan not found"})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not load plan"})
}

func savePurchasePlan(ctx context.Context, group models.Group, before, after models.PurchasePlan, actorUID, kind, detail string) (bool, error) {
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx,
		bson.M{"_id": group.ID, "purchasePlans": bson.M{"$elemMatch": bson.M{"id": before.ID, "version": before.Version}}},
		bson.M{"$set": bson.M{"purchasePlans.$": after, "updatedAt": time.Now()}, "$push": bson.M{"activity": models.NewActivity(kind, actorUID, detail)}})
	if err != nil {
		return false, err
	}
	return result.MatchedCount == 1, nil
}

func createPurchasePlan(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var input models.PurchasePlanInput
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	if err != nil {
		return purchasePlanLoadError(c, err)
	}
	plan, err := models.NewPurchasePlan(input, group.Members, user.UID, time.Now())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx,
		bson.M{"_id": group.ID, "purchasePlans.49": bson.M{"$exists": false}},
		bson.M{"$push": bson.M{"purchasePlans": plan, "activity": models.NewActivity("plan_created", user.UID, "Proposed "+plan.Title)}, "$set": bson.M{"updatedAt": time.Now()}})
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "This group has reached its plan limit"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": plan.ResponseFor(group, user.UID, time.Now())})
}

func votePurchasePlan(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		OptionID string `json:"optionId"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	group, plan, ctx, cancel, err := loadPurchasePlan(c)
	defer cancel()
	if err != nil {
		return purchasePlanLoadError(c, err)
	}
	memberID := memberIDForUID(group, user.UID)
	updated, err := plan.Vote(memberID, body.OptionID, time.Now())
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	matched, err := savePurchasePlan(ctx, group, plan, updated, user.UID, "plan_voted", "Voted on "+plan.Title)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !matched {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Plan changed; refresh and try again"})
	}
	return c.JSON(fiber.Map{"success": true, "data": updated.ResponseFor(group, user.UID, time.Now())})
}

func commentPurchasePlan(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		Text string `json:"text"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	body.Text = strings.TrimSpace(body.Text)
	if body.Text == "" || len(body.Text) > 1000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Comment must be 1 to 1000 characters"})
	}
	group, plan, ctx, cancel, err := loadPurchasePlan(c)
	defer cancel()
	if err != nil {
		return purchasePlanLoadError(c, err)
	}
	if memberIDForUID(group, user.UID) == "" {
		return c.SendStatus(fiber.StatusForbidden)
	}
	if plan.Status == models.PlanCancelled {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Cancelled plans cannot be discussed"})
	}
	if len(plan.Comments) >= 100 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "This plan has reached its comment limit"})
	}
	updated := plan
	updated.Comments = append(append([]models.PlanComment{}, plan.Comments...), models.PlanComment{ID: primitive.NewObjectID().Hex(), AuthorUID: user.UID, Text: body.Text, CreatedAt: time.Now()})
	updated.Version++
	matched, err := savePurchasePlan(ctx, group, plan, updated, user.UID, "plan_commented", "Commented on "+plan.Title)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !matched {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Plan changed; refresh and try again"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": updated.ResponseFor(group, user.UID, time.Now())})
}

func closePurchasePlan(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	group, plan, ctx, cancel, err := loadPurchasePlan(c)
	defer cancel()
	if err != nil {
		return purchasePlanLoadError(c, err)
	}
	updated, err := plan.Close(group.Members, user.UID, group.UserID, time.Now())
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	matched, err := savePurchasePlan(ctx, group, plan, updated, user.UID, "plan_closed", "Selected an option for "+plan.Title)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !matched {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Plan changed; refresh and try again"})
	}
	return c.JSON(fiber.Map{"success": true, "data": updated.ResponseFor(group, user.UID, time.Now())})
}

func cancelPurchasePlan(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	group, plan, ctx, cancel, err := loadPurchasePlan(c)
	defer cancel()
	if err != nil {
		return purchasePlanLoadError(c, err)
	}
	updated, err := plan.Cancel(user.UID, group.UserID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	matched, err := savePurchasePlan(ctx, group, plan, updated, user.UID, "plan_cancelled", "Cancelled "+plan.Title)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !matched {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Plan changed; refresh and try again"})
	}
	return c.JSON(fiber.Map{"success": true, "data": updated.ResponseFor(group, user.UID, time.Now())})
}

func convertPurchasePlan(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		PaidBy       string   `json:"paidBy"`
		Participants []string `json:"participants"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	group, plan, ctx, cancel, err := loadPurchasePlan(c)
	defer cancel()
	if err != nil {
		return purchasePlanLoadError(c, err)
	}
	if plan.Status != models.PlanClosed || plan.ConvertedExpenseID != "" {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "This plan is not ready or already converted"})
	}
	option, ok := plan.SelectedOption()
	if !ok {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Selected option is missing"})
	}
	expense, err := models.NewProposedExpense(models.ExpenseInput{Description: option.Label, AmountPaise: option.AmountPaise, PaidBy: body.PaidBy, Participants: body.Participants}, group.Members, user.UID, time.Now())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	expense.PurchasePlanID = plan.ID
	updated := plan
	updated.ConvertedExpenseID = expense.ID
	updated.Version++
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx,
		bson.M{"_id": group.ID, "purchasePlans": bson.M{"$elemMatch": bson.M{"id": plan.ID, "version": plan.Version, "status": models.PlanClosed, "convertedExpenseId": bson.M{"$exists": false}}}},
		bson.M{"$set": bson.M{"purchasePlans.$": updated, "updatedAt": time.Now()}, "$push": bson.M{"expenses": expense, "activity": models.NewActivity("plan_converted", user.UID, "Proposed purchase from "+plan.Title)}, "$inc": bson.M{"balanceRevision": 1}})
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Plan changed or was already converted; refresh and try again"})
	}
	if expense.Status == models.ExpenseProposed {
		_ = notifyExpenseMembers(ctx, group, expense.ID, user.UID, "expense_proposal", "Approval requested for "+expense.Description, pendingMemberIDs(expense))
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": expense.ResponseFor(user.UID, group.UserID == user.UID)})
}
