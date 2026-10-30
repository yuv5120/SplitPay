package routes

import (
	"context"
	"encoding/json"
	"errors"
	"splitpay/backend/config"
	"splitpay/backend/middleware"
	"splitpay/backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func paymentConfig(c *fiber.Ctx) error {
	r := config.GetRazorpayTest()
	if r == nil {
		return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"razorpayTestEnabled": false}})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"razorpayTestEnabled": true, "razorpayTestKeyId": r.KeyID}})
}

func balanceRevisionFilter(revision int64) bson.M {
	if revision == 0 {
		return bson.M{"$or": []bson.M{{"balanceRevision": 0}, {"balanceRevision": bson.M{"$exists": false}}}}
	}
	return bson.M{"balanceRevision": revision}
}

func paymentMember(group models.Group, memberID string) models.Member {
	for _, member := range group.Members {
		if member.ID == memberID {
			return member
		}
	}
	return models.Member{}
}

func startPayment(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		To          string `json:"to"`
		AmountPaise int64  `json:"amountPaise"`
		Method      string `json:"method"`
	}
	if err := c.BodyParser(&body); err != nil || (body.Method != "simulated" && body.Method != "razorpay_test") || body.AmountPaise <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Choose a payment method and a positive amount"})
	}
	if body.Method == "razorpay_test" && config.GetRazorpayTest() == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"success": false, "message": "Razorpay Test Mode is not configured"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	if err == mongo.ErrNoDocuments {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	from := memberIDForUID(group, user.UID)
	to := paymentMember(group, body.To)
	if from == "" || to.UID == "" || from == to.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Both payer and recipient must be joined group members"})
	}
	outstanding := models.OutstandingTo(models.CalculateBalances(group), from, to.ID)
	if outstanding == 0 || body.AmountPaise > outstanding {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Amount exceeds the current settlement suggestion"})
	}
	attempt := models.PaymentAttempt{ID: primitive.NewObjectID().Hex(), From: from, To: to.ID,
		AmountPaise: body.AmountPaise, Method: body.Method, Status: models.PaymentPending,
		BaseRevision: group.BalanceRevision, CreatedAt: time.Now()}
	if body.Method == "razorpay_test" {
		order, err := config.GetRazorpayTest().CreateOrder(ctx, body.AmountPaise, attempt.ID)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "message": "Could not create Razorpay test order"})
		}
		attempt.GatewayOrderID = order.ID
	}
	filter := bson.M{"_id": group.ID}
	for key, value := range balanceRevisionFilter(group.BalanceRevision) {
		filter[key] = value
	}
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx, filter, bson.M{"$push": bson.M{
		"payments": attempt, "activity": models.NewActivity("payment_started", user.UID, "Started a "+body.Method+" payment"),
	}})
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Balances changed; refresh and try again"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": attempt})
}

func confirmPayment(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		Outcome string `json:"outcome"`
	}
	if err := c.BodyParser(&body); err != nil || (body.Outcome != "success" && body.Outcome != "failure" && body.Outcome != "cancel") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Outcome must be success, failure, or cancel"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	if err == mongo.ErrNoDocuments {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	var attempt models.PaymentAttempt
	found := false
	for _, item := range group.Payments {
		if item.ID == c.Params("paymentId") {
			attempt, found = item, true
			break
		}
	}
	if !found {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if memberIDForUID(group, user.UID) != attempt.From {
		return c.SendStatus(fiber.StatusForbidden)
	}
	if attempt.Method != "simulated" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Razorpay payments require gateway verification"})
	}
	if attempt.Status == models.PaymentSucceeded && body.Outcome == "success" {
		return c.JSON(fiber.Map{"success": true, "data": attempt})
	}
	if attempt.Status != models.PaymentPending {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Payment attempt is already final"})
	}
	status := models.PaymentSucceeded
	switch body.Outcome {
	case "failure":
		status = models.PaymentFailed
	case "cancel":
		status = models.PaymentCancelled
	}
	if status == models.PaymentSucceeded {
		if group.BalanceRevision != attempt.BaseRevision || models.OutstandingTo(models.CalculateBalances(group), attempt.From, attempt.To) < attempt.AmountPaise {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Balances changed; cancel this attempt and start a new one"})
		}
	}
	now := time.Now()
	filter := bson.M{"_id": group.ID, "payments": bson.M{"$elemMatch": bson.M{"id": attempt.ID, "status": models.PaymentPending}}}
	update := bson.M{"$set": bson.M{"payments.$.status": status, "payments.$.completedAt": now},
		"$push": bson.M{"activity": models.NewActivity("payment_"+status, user.UID, "Simulated payment "+status)}}
	if status == models.PaymentSucceeded {
		for key, value := range balanceRevisionFilter(attempt.BaseRevision) {
			filter[key] = value
		}
		update["$inc"] = bson.M{"balanceRevision": 1}
	}
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx, filter, update)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Payment changed; refresh and try again"})
	}
	attempt.Status, attempt.CompletedAt = status, &now
	return c.JSON(fiber.Map{"success": true, "data": attempt})
}

func gatewayAttempt(c *fiber.Ctx, ctx context.Context) (models.Group, models.PaymentAttempt, error) {
	user := middleware.GetUserFromContext(c)
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return group, models.PaymentAttempt{}, fiber.ErrNotFound
		}
		return group, models.PaymentAttempt{}, err
	}
	for _, attempt := range group.Payments {
		if attempt.ID == c.Params("paymentId") {
			if memberIDForUID(group, user.UID) != attempt.From {
				return group, attempt, fiber.ErrForbidden
			}
			if attempt.Method != "razorpay_test" || attempt.GatewayOrderID == "" {
				return group, attempt, fiber.ErrBadRequest
			}
			return group, attempt, nil
		}
	}
	return group, models.PaymentAttempt{}, fiber.ErrNotFound
}

func gatewayPaymentMatches(attempt models.PaymentAttempt, payment config.RazorpayPayment) bool {
	return payment.ID != "" && payment.OrderID == attempt.GatewayOrderID && payment.Amount == attempt.AmountPaise && payment.Currency == "INR" && payment.Status == "captured" && payment.Captured
}

// finalizeGatewayPayment is idempotent and only changes balances when the original suggestion remains valid.
func finalizeGatewayPayment(ctx context.Context, group models.Group, attempt models.PaymentAttempt, payment config.RazorpayPayment, actorUID string) (models.PaymentAttempt, error) {
	if !gatewayPaymentMatches(attempt, payment) {
		return attempt, errors.New("Razorpay payment is not captured or does not match this order")
	}
	if attempt.Status == models.PaymentSucceeded || attempt.Status == models.PaymentReviewRequired {
		if attempt.GatewayPaymentID == payment.ID {
			return attempt, nil
		}
		return attempt, errors.New("payment attempt already has a different gateway result")
	}
	if attempt.Status != models.PaymentPending {
		return attempt, errors.New("payment attempt is already final")
	}
	status := models.PaymentSucceeded
	if group.BalanceRevision != attempt.BaseRevision || models.OutstandingTo(models.CalculateBalances(group), attempt.From, attempt.To) < attempt.AmountPaise {
		status = models.PaymentReviewRequired
	}
	now := time.Now()
	filter := bson.M{"_id": group.ID, "payments": bson.M{"$elemMatch": bson.M{"id": attempt.ID, "status": models.PaymentPending, "gatewayOrderId": attempt.GatewayOrderID}}}
	for key, value := range balanceRevisionFilter(group.BalanceRevision) {
		filter[key] = value
	}
	update := bson.M{"$set": bson.M{"payments.$.status": status, "payments.$.gatewayPaymentId": payment.ID, "payments.$.completedAt": now},
		"$push": bson.M{"activity": models.NewActivity("payment_"+status, actorUID, "Razorpay test payment "+status)}}
	if status == models.PaymentSucceeded {
		update["$inc"] = bson.M{"balanceRevision": 1}
	}
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx, filter, update)
	if err != nil {
		return attempt, err
	}
	if result.MatchedCount == 0 {
		return attempt, errors.New("balances changed while recording payment; check status again")
	}
	attempt.Status, attempt.CompletedAt, attempt.GatewayPaymentID = status, &now, payment.ID
	return attempt, nil
}

func verifyGatewayPayment(c *fiber.Ctx) error {
	var body struct {
		OrderID   string `json:"razorpay_order_id"`
		PaymentID string `json:"razorpay_payment_id"`
		Signature string `json:"razorpay_signature"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	r := config.GetRazorpayTest()
	if r == nil {
		return c.SendStatus(fiber.StatusServiceUnavailable)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	group, attempt, err := gatewayAttempt(c, ctx)
	if err != nil {
		return err
	}
	if body.OrderID != attempt.GatewayOrderID || !r.VerifySignature(attempt.GatewayOrderID, body.PaymentID, body.Signature) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid Razorpay checkout signature"})
	}
	payment, err := r.FetchPayment(ctx, body.PaymentID)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "message": "Could not verify payment with Razorpay; check status again"})
	}
	updated, err := finalizeGatewayPayment(ctx, group, attempt, payment, middleware.GetUserFromContext(c).UID)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": updated})
}

func reconcileGatewayPayment(c *fiber.Ctx) error {
	r := config.GetRazorpayTest()
	if r == nil {
		return c.SendStatus(fiber.StatusServiceUnavailable)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	group, attempt, err := gatewayAttempt(c, ctx)
	if err != nil {
		return err
	}
	payments, err := r.FetchOrderPayments(ctx, attempt.GatewayOrderID)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "message": "Could not check Razorpay payment status"})
	}
	for _, payment := range payments {
		if gatewayPaymentMatches(attempt, payment) {
			updated, err := finalizeGatewayPayment(ctx, group, attempt, payment, middleware.GetUserFromContext(c).UID)
			if err != nil {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": err.Error()})
			}
			return c.JSON(fiber.Map{"success": true, "data": updated})
		}
	}
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "No captured Razorpay test payment for this order yet"})
}

func razorpayTestWebhook(c *fiber.Ctx) error {
	r := config.GetRazorpayTest()
	if r == nil {
		return c.SendStatus(fiber.StatusServiceUnavailable)
	}
	if !config.VerifyRazorpayWebhook(c.Body(), c.Get("X-Razorpay-Signature")) {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	var event struct {
		Event   string `json:"event"`
		Payload struct {
			Payment struct {
				Entity struct {
					ID      string `json:"id"`
					OrderID string `json:"order_id"`
				} `json:"entity"`
			} `json:"payment"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(c.Body(), &event); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	if event.Event != "payment.captured" {
		return c.SendStatus(fiber.StatusOK)
	}
	if event.Payload.Payment.Entity.OrderID == "" || event.Payload.Payment.Entity.ID == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var group models.Group
	err := config.GetDB().Collection("groups").FindOne(ctx, bson.M{"payments.gatewayOrderId": event.Payload.Payment.Entity.OrderID}).Decode(&group)
	if err == mongo.ErrNoDocuments {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	var attempt models.PaymentAttempt
	for _, item := range group.Payments {
		if item.GatewayOrderID == event.Payload.Payment.Entity.OrderID {
			attempt = item
			break
		}
	}
	if attempt.Method != "razorpay_test" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	payment, err := r.FetchPayment(ctx, event.Payload.Payment.Entity.ID)
	if err != nil {
		return c.SendStatus(fiber.StatusBadGateway)
	}
	actorUID := paymentMember(group, attempt.From).UID
	if _, err = finalizeGatewayPayment(ctx, group, attempt, payment, actorUID); err != nil {
		return c.SendStatus(fiber.StatusConflict)
	}
	return c.SendStatus(fiber.StatusOK)
}

func setReminderPreference(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.BodyParser(&body); err != nil || body.Enabled == nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	if err == mongo.ErrNoDocuments {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	update := bson.M{"$pull": bson.M{"reminderOptIn": user.UID}}
	if *body.Enabled {
		update = bson.M{"$addToSet": bson.M{"reminderOptIn": user.UID}}
	}
	_, err = config.GetDB().Collection("groups").UpdateOne(ctx, bson.M{"_id": group.ID}, update)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"remindersEnabled": *body.Enabled}})
}

func sendReminder(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	var body struct {
		From string `json:"from"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	if err == mongo.ErrNoDocuments {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	to := memberIDForUID(group, user.UID)
	debtor := paymentMember(group, body.From)
	if to == "" || debtor.UID == "" || to == debtor.ID {
		return c.SendStatus(fiber.StatusForbidden)
	}
	if !containsUID(group.ReminderOptIn, debtor.UID) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "This member has not enabled reminders"})
	}
	amount := models.OutstandingTo(models.CalculateBalances(group), debtor.ID, to)
	if amount <= 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "No current settlement to remind about"})
	}
	now := time.Now()
	cutoff := now.Add(-24 * time.Hour)
	event := models.ReminderEvent{ID: primitive.NewObjectID().Hex(), From: debtor.ID, To: to, AmountPaise: amount, CreatedAt: now}
	filter := bson.M{"_id": group.ID, "reminderOptIn": debtor.UID,
		"reminders": bson.M{"$not": bson.M{"$elemMatch": bson.M{"from": debtor.ID, "to": to, "createdAt": bson.M{"$gte": cutoff}}}}}
	for key, value := range balanceRevisionFilter(group.BalanceRevision) {
		filter[key] = value
	}
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx, filter, bson.M{"$push": bson.M{
		"reminders": event, "activity": models.NewActivity("reminder_sent", user.UID, "Sent a settlement reminder"),
	}})
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Reminder was recently sent or balances changed; refresh and try again"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": event})
}

func containsUID(items []string, uid string) bool {
	for _, item := range items {
		if item == uid {
			return true
		}
	}
	return false
}
