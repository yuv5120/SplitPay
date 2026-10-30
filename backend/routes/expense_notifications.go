package routes

import (
	"context"
	"fmt"
	"sort"
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

type expenseNotification struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	GroupID      primitive.ObjectID `bson:"groupId" json:"-"`
	RecipientUID string             `bson:"recipientUid" json:"-"`
	ExpenseID    string             `bson:"expenseId" json:"expenseId"`
	Kind         string             `bson:"kind" json:"kind"`
	Message      string             `bson:"message" json:"message"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	ReadAt       *time.Time         `bson:"readAt,omitempty" json:"readAt,omitempty"`
}

func pendingMemberIDs(expense models.Expense) []string {
	ids := make([]string, 0, len(expense.Approvals))
	for _, approval := range expense.Approvals {
		if approval.Status == "pending" {
			ids = append(ids, approval.MemberID)
		}
	}
	return ids
}

func (n expenseNotification) response() fiber.Map {
	return fiber.Map{"id": n.ID.Hex(), "expenseId": n.ExpenseID, "kind": n.Kind,
		"message": n.Message, "createdAt": n.CreatedAt, "readAt": n.ReadAt}
}

func notifyExpenseMembers(ctx context.Context, group models.Group, expenseID, actorUID, kind, message string, recipientMemberIDs []string) error {
	seen := map[string]bool{actorUID: true}
	allowed := map[string]bool{}
	for _, id := range recipientMemberIDs {
		allowed[id] = true
	}
	docs := make([]interface{}, 0, len(group.Members))
	for _, member := range group.Members {
		if member.UID == "" || seen[member.UID] || (recipientMemberIDs != nil && !allowed[member.ID]) {
			continue
		}
		seen[member.UID] = true
		docs = append(docs, expenseNotification{ID: primitive.NewObjectID(), GroupID: group.ID,
			RecipientUID: member.UID, ExpenseID: expenseID, Kind: kind, Message: message, CreatedAt: time.Now()})
	}
	if len(docs) == 0 {
		return nil
	}
	_, err := config.GetDB().Collection("expenseNotifications").InsertMany(ctx, docs)
	return err
}

func getExpenseNotifications(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	if err == mongo.ErrNoDocuments {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	cursor, err := config.GetDB().Collection("expenseNotifications").Find(ctx,
		bson.M{"groupId": group.ID, "recipientUid": user.UID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(100))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer cursor.Close(ctx)
	var notices []expenseNotification
	if err := cursor.All(ctx, &notices); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	response := make([]fiber.Map, len(notices))
	for i, notice := range notices {
		response[i] = notice.response()
	}
	for _, payment := range group.Payments {
		if payment.Status != models.PaymentSucceeded || paymentMember(group, payment.To).UID != user.UID {
			continue
		}
		payer := paymentMember(group, payment.From)
		message := fmt.Sprintf("%s recorded a simulated settlement of ₹%.2f to you", payer.Name, float64(payment.AmountPaise)/100)
		if payment.Method == "razorpay_test" {
			message = fmt.Sprintf("%s completed a Razorpay test payment of ₹%.2f; no real money was transferred to you", payer.Name, float64(payment.AmountPaise)/100)
		}
		response = append(response, fiber.Map{"id": "payment:" + payment.ID, "expenseId": "", "kind": "payment_succeeded",
			"message":   message,
			"createdAt": payment.CompletedAt, "readAt": payment.RecipientReadAt})
	}
	for _, reminder := range group.Reminders {
		if paymentMember(group, reminder.From).UID != user.UID {
			continue
		}
		creditor := paymentMember(group, reminder.To)
		response = append(response, fiber.Map{"id": "reminder:" + reminder.ID, "expenseId": "", "kind": "settlement_reminder",
			"message":   fmt.Sprintf("%s reminded you about a simulated settlement of ₹%.2f", creditor.Name, float64(reminder.AmountPaise)/100),
			"createdAt": reminder.CreatedAt, "readAt": reminder.ReadAt})
	}
	sort.Slice(response, func(i, j int) bool {
		date := func(item fiber.Map) time.Time {
			switch value := item["createdAt"].(type) {
			case time.Time:
				return value
			case *time.Time:
				if value != nil {
					return *value
				}
			}
			return time.Time{}
		}
		return date(response[i]).After(date(response[j]))
	})
	return c.JSON(fiber.Map{"success": true, "data": response})
}

func readExpenseNotification(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, err := loadAccessibleGroup(ctx, c.Params("groupId"), user.UID)
	if err == mongo.ErrNoDocuments {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if paymentID, ok := strings.CutPrefix(c.Params("notificationId"), "payment:"); ok {
		result, err := config.GetDB().Collection("groups").UpdateOne(ctx,
			bson.M{"_id": group.ID, "payments": bson.M{"$elemMatch": bson.M{"id": paymentID, "to": memberIDForUID(group, user.UID), "status": models.PaymentSucceeded}}},
			bson.M{"$set": bson.M{"payments.$.recipientReadAt": time.Now()}})
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if result.MatchedCount == 0 {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return c.JSON(fiber.Map{"success": true})
	}
	if reminderID, ok := strings.CutPrefix(c.Params("notificationId"), "reminder:"); ok {
		result, err := config.GetDB().Collection("groups").UpdateOne(ctx,
			bson.M{"_id": group.ID, "reminders": bson.M{"$elemMatch": bson.M{"id": reminderID, "from": memberIDForUID(group, user.UID)}}},
			bson.M{"$set": bson.M{"reminders.$.readAt": time.Now()}})
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if result.MatchedCount == 0 {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return c.JSON(fiber.Map{"success": true})
	}
	id, err := primitive.ObjectIDFromHex(c.Params("notificationId"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	result, err := config.GetDB().Collection("expenseNotifications").UpdateOne(ctx,
		bson.M{"_id": id, "groupId": group.ID, "recipientUid": user.UID, "readAt": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"readAt": time.Now()}})
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if result.MatchedCount == 0 {
		count, err := config.GetDB().Collection("expenseNotifications").CountDocuments(ctx, bson.M{"_id": id, "groupId": group.ID, "recipientUid": user.UID})
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if count == 0 {
			return c.SendStatus(fiber.StatusNotFound)
		}
	}
	return c.JSON(fiber.Map{"success": true})
}
