package routes

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"splitpay/backend/config"
	"splitpay/backend/middleware"
	"splitpay/backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type groupInvitation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	GroupID   primitive.ObjectID `bson:"groupId"`
	MemberID  string             `bson:"memberId"`
	TokenHash string             `bson:"tokenHash"`
	ExpiresAt time.Time          `bson:"expiresAt"`
	CreatedAt time.Time          `bson:"createdAt"`
	UsedAt    *time.Time         `bson:"usedAt,omitempty"`
}

func hashInviteToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func claimOwnerMember(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	groupID, memberID := c.Params("groupId"), c.Params("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := config.GetDB().Collection("groups").UpdateOne(ctx, bson.M{
		"id": groupID, "userId": user.UID,
		"members.uid": bson.M{"$ne": user.UID},
		"members":     bson.M{"$elemMatch": bson.M{"id": memberID, "uid": bson.M{"$in": []any{nil, ""}}}},
	}, bson.M{
		"$set":  bson.M{"members.$.uid": user.UID, "updatedAt": time.Now()},
		"$push": bson.M{"activity": models.NewActivity("member_claimed", user.UID, "Linked their account to a member slot")},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not claim member"})
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Member is unavailable or your account already has a member slot"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Member linked to your account"})
}

func createGroupInvite(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	groupID, memberID := c.Params("groupId"), c.Params("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var group models.Group
	err := config.GetDB().Collection("groups").FindOne(ctx, bson.M{"id": groupID, "userId": user.UID}).Decode(&group)
	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Group not found"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not load group"})
	}
	available := false
	for _, member := range group.Members {
		if member.ID == memberID && member.UID == "" {
			available = true
			break
		}
	}
	if !available {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Member slot is already claimed or does not exist"})
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not generate invite"})
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)
	_, err = config.GetDB().Collection("groupInvitations").InsertOne(ctx, groupInvitation{
		ID: primitive.NewObjectID(), GroupID: group.ID, MemberID: memberID,
		TokenHash: hashInviteToken(token), ExpiresAt: expiresAt, CreatedAt: now,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not save invite"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": fiber.Map{"token": token, "expiresAt": expiresAt}})
}

func acceptGroupInvite(c *fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	token := c.Params("token")
	if len(token) != 43 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid invite link"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var invite groupInvitation
	err := config.GetDB().Collection("groupInvitations").FindOne(ctx, bson.M{
		"tokenHash": hashInviteToken(token), "expiresAt": bson.M{"$gt": time.Now()}, "usedAt": bson.M{"$exists": false},
	}).Decode(&invite)
	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Invite has expired or was already used"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not verify invite"})
	}
	groups := config.GetDB().Collection("groups")
	var group models.Group
	if err := groups.FindOne(ctx, bson.M{"_id": invite.GroupID}).Decode(&group); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Group no longer exists"})
	}
	if group.UserID == user.UID {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "You already own this group"})
	}
	for _, member := range group.Members {
		if member.UID == user.UID {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "You are already in this group"})
		}
	}
	result, err := groups.UpdateOne(ctx, bson.M{
		"_id":         invite.GroupID,
		"members.uid": bson.M{"$ne": user.UID},
		"members":     bson.M{"$elemMatch": bson.M{"id": invite.MemberID, "uid": bson.M{"$in": []any{nil, ""}}}},
	}, bson.M{
		"$set":  bson.M{"members.$.uid": user.UID, "updatedAt": time.Now()},
		"$push": bson.M{"activity": models.NewActivity("member_joined", user.UID, "Joined the group")},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Could not join group"})
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "This member slot was already claimed"})
	}
	now := time.Now()
	_, _ = config.GetDB().Collection("groupInvitations").UpdateOne(ctx, bson.M{"_id": invite.ID}, bson.M{"$set": bson.M{"usedAt": now}})
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"groupId": group.GroupID}})
}
