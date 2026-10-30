package models

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PlanOpen      = "open"
	PlanClosed    = "closed"
	PlanCancelled = "cancelled"
)

type PurchaseOption struct {
	ID          string `bson:"id" json:"id"`
	Label       string `bson:"label" json:"label"`
	AmountPaise int64  `bson:"amountPaise" json:"amountPaise"`
}

type PurchaseVote struct {
	MemberID string    `bson:"memberId" json:"memberId"`
	OptionID string    `bson:"optionId" json:"optionId"`
	VotedAt  time.Time `bson:"votedAt" json:"votedAt"`
}

type PlanComment struct {
	ID        string    `bson:"id" json:"id"`
	AuthorUID string    `bson:"authorUid" json:"-"`
	Text      string    `bson:"text" json:"text"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

type PurchasePlan struct {
	ID                 string           `bson:"id" json:"id"`
	Title              string           `bson:"title" json:"title"`
	Details            string           `bson:"details" json:"details"`
	Options            []PurchaseOption `bson:"options" json:"options"`
	Votes              []PurchaseVote   `bson:"votes" json:"-"`
	Comments           []PlanComment    `bson:"comments" json:"-"`
	Status             string           `bson:"status" json:"status"`
	SelectedOptionID   string           `bson:"selectedOptionId,omitempty" json:"selectedOptionId,omitempty"`
	ConvertedExpenseID string           `bson:"convertedExpenseId,omitempty" json:"convertedExpenseId,omitempty"`
	CreatedBy          string           `bson:"createdBy" json:"-"`
	DeadlineAt         time.Time        `bson:"deadlineAt" json:"deadlineAt"`
	CreatedAt          time.Time        `bson:"createdAt" json:"createdAt"`
	ClosedAt           *time.Time       `bson:"closedAt,omitempty" json:"closedAt,omitempty"`
	Version            int64            `bson:"version" json:"-"`
}

type PurchasePlanInput struct {
	Title      string                `json:"title"`
	Details    string                `json:"details"`
	Options    []PurchaseOptionInput `json:"options"`
	VotingDays int                   `json:"votingDays"`
}

type PurchaseOptionInput struct {
	Label       string `json:"label"`
	AmountPaise int64  `json:"amountPaise"`
}

type PlanCommentResponse struct {
	ID         string    `json:"id"`
	AuthorName string    `json:"authorName"`
	Text       string    `json:"text"`
	CreatedAt  time.Time `json:"createdAt"`
}

type PurchasePlanResponse struct {
	ID                 string                `json:"id"`
	Title              string                `json:"title"`
	Details            string                `json:"details"`
	Options            []PurchaseOption      `json:"options"`
	VoteCounts         map[string]int        `json:"voteCounts"`
	MyVote             string                `json:"myVote"`
	Comments           []PlanCommentResponse `json:"comments"`
	Status             string                `json:"status"`
	SelectedOptionID   string                `json:"selectedOptionId,omitempty"`
	ConvertedExpenseID string                `json:"convertedExpenseId,omitempty"`
	CanVote            bool                  `json:"canVote"`
	CanClose           bool                  `json:"canClose"`
	CanCancel          bool                  `json:"canCancel"`
	CanConvert         bool                  `json:"canConvert"`
	DeadlineAt         time.Time             `json:"deadlineAt"`
	CreatedAt          time.Time             `json:"createdAt"`
}

func NewPurchasePlan(input PurchasePlanInput, members []Member, actorUID string, now time.Time) (PurchasePlan, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Details = strings.TrimSpace(input.Details)
	if input.Title == "" || len(input.Title) > 100 || len(input.Details) > 500 || len(input.Options) < 2 || len(input.Options) > 5 || input.VotingDays < 1 || input.VotingDays > 14 {
		return PurchasePlan{}, errors.New("Provide a title, 2–5 options, and a voting period of 1–14 days")
	}
	joined, actorJoined := 0, false
	for _, member := range members {
		if member.UID != "" {
			joined++
			actorJoined = actorJoined || member.UID == actorUID
		}
	}
	if !actorJoined || joined < 2 {
		return PurchasePlan{}, errors.New("At least two joined group members are needed to plan a purchase")
	}
	options := make([]PurchaseOption, len(input.Options))
	labels := map[string]bool{}
	for i, item := range input.Options {
		label := strings.TrimSpace(item.Label)
		key := strings.ToLower(label)
		if label == "" || len(label) > 100 || item.AmountPaise <= 0 || item.AmountPaise > 100_000_000_000 || labels[key] {
			return PurchasePlan{}, errors.New("Each option needs a distinct name and a positive amount")
		}
		labels[key] = true
		options[i] = PurchaseOption{ID: primitive.NewObjectID().Hex(), Label: label, AmountPaise: item.AmountPaise}
	}
	return PurchasePlan{ID: primitive.NewObjectID().Hex(), Title: input.Title, Details: input.Details, Options: options,
		Votes: []PurchaseVote{}, Comments: []PlanComment{}, Status: PlanOpen, CreatedBy: actorUID,
		DeadlineAt: now.AddDate(0, 0, input.VotingDays), CreatedAt: now, Version: 1}, nil
}

func (plan PurchasePlan) Vote(memberID, optionID string, now time.Time) (PurchasePlan, error) {
	if plan.Status != PlanOpen || !now.Before(plan.DeadlineAt) {
		return PurchasePlan{}, errors.New("Voting has closed")
	}
	valid := false
	for _, option := range plan.Options {
		valid = valid || option.ID == optionID
	}
	if !valid || memberID == "" {
		return PurchasePlan{}, errors.New("Choose a valid option")
	}
	updated := plan
	updated.Votes = append([]PurchaseVote{}, plan.Votes...)
	for i := range updated.Votes {
		if updated.Votes[i].MemberID == memberID {
			updated.Votes[i] = PurchaseVote{MemberID: memberID, OptionID: optionID, VotedAt: now}
			updated.Version++
			return updated, nil
		}
	}
	updated.Votes = append(updated.Votes, PurchaseVote{MemberID: memberID, OptionID: optionID, VotedAt: now})
	updated.Version++
	return updated, nil
}

func (plan PurchasePlan) CanClose(members []Member, actorUID, ownerUID string, now time.Time) bool {
	if plan.Status != PlanOpen || (actorUID != plan.CreatedBy && actorUID != ownerUID) {
		return false
	}
	joined := map[string]bool{}
	for _, member := range members {
		if member.UID != "" {
			joined[member.ID] = true
		}
	}
	voted := map[string]bool{}
	for _, vote := range plan.Votes {
		if joined[vote.MemberID] {
			voted[vote.MemberID] = true
		}
	}
	return !now.Before(plan.DeadlineAt) || (len(joined) > 0 && len(voted) == len(joined))
}

func (plan PurchasePlan) Close(members []Member, actorUID, ownerUID string, now time.Time) (PurchasePlan, error) {
	if !plan.CanClose(members, actorUID, ownerUID, now) {
		return PurchasePlan{}, errors.New("Close voting after the deadline or when all joined members have voted")
	}
	counts := map[string]int{}
	for _, option := range plan.Options {
		counts[option.ID] = 0
	}
	joined := map[string]bool{}
	for _, member := range members {
		if member.UID != "" {
			joined[member.ID] = true
		}
	}
	for _, vote := range plan.Votes {
		if joined[vote.MemberID] {
			counts[vote.OptionID]++
		}
	}
	winner, highest, lowest := "", 0, int64(0)
	for _, option := range plan.Options {
		count := counts[option.ID]
		if count > highest || (count == highest && count > 0 && option.AmountPaise < lowest) {
			winner, highest, lowest = option.ID, count, option.AmountPaise
		}
	}
	if highest == 0 {
		return PurchasePlan{}, errors.New("At least one vote is needed before closing")
	}
	updated := plan
	updated.Status, updated.SelectedOptionID, updated.ClosedAt = PlanClosed, winner, &now
	updated.Version++
	return updated, nil
}

func (plan PurchasePlan) Cancel(actorUID, ownerUID string) (PurchasePlan, error) {
	if plan.Status != PlanOpen || (actorUID != plan.CreatedBy && actorUID != ownerUID) {
		return PurchasePlan{}, errors.New("Only the creator or owner can cancel an open plan")
	}
	updated := plan
	updated.Status = PlanCancelled
	updated.Version++
	return updated, nil
}

func (plan PurchasePlan) SelectedOption() (PurchaseOption, bool) {
	for _, option := range plan.Options {
		if option.ID == plan.SelectedOptionID {
			return option, true
		}
	}
	return PurchaseOption{}, false
}

func (plan PurchasePlan) ResponseFor(group Group, uid string, now time.Time) PurchasePlanResponse {
	memberID := ""
	names := map[string]string{}
	for _, member := range group.Members {
		if member.UID != "" {
			names[member.UID] = member.Name
		}
		if uid != "" && member.UID == uid {
			memberID = member.ID
		}
	}
	counts := map[string]int{}
	for _, option := range plan.Options {
		counts[option.ID] = 0
	}
	myVote := ""
	joinedIDs := map[string]bool{}
	for _, member := range group.Members {
		if member.UID != "" {
			joinedIDs[member.ID] = true
		}
	}
	for _, vote := range plan.Votes {
		if joinedIDs[vote.MemberID] {
			if _, ok := counts[vote.OptionID]; ok {
				counts[vote.OptionID]++
			}
		}
		if vote.MemberID == memberID {
			myVote = vote.OptionID
		}
	}
	comments := make([]PlanCommentResponse, len(plan.Comments))
	for i, comment := range plan.Comments {
		name := names[comment.AuthorUID]
		if name == "" {
			name = "Former member"
		}
		comments[i] = PlanCommentResponse{ID: comment.ID, AuthorName: name, Text: comment.Text, CreatedAt: comment.CreatedAt}
	}
	return PurchasePlanResponse{ID: plan.ID, Title: plan.Title, Details: plan.Details, Options: append([]PurchaseOption{}, plan.Options...), VoteCounts: counts,
		MyVote: myVote, Comments: comments, Status: plan.Status, SelectedOptionID: plan.SelectedOptionID, ConvertedExpenseID: plan.ConvertedExpenseID,
		CanVote:  memberID != "" && plan.Status == PlanOpen && now.Before(plan.DeadlineAt),
		CanClose: plan.CanClose(group.Members, uid, group.UserID, now), CanCancel: plan.Status == PlanOpen && (uid == plan.CreatedBy || uid == group.UserID),
		CanConvert: memberID != "" && plan.Status == PlanClosed && plan.ConvertedExpenseID == "",
		DeadlineAt: plan.DeadlineAt, CreatedAt: plan.CreatedAt}
}
