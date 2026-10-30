package models

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ExpenseProposed  = "proposed"
	ExpenseActive    = "active"
	ExpenseRejected  = "rejected"
	ExpenseWithdrawn = "withdrawn"
)

// Older expenses have no status and were active before approvals existed.
func (e Expense) EffectiveStatus() string {
	if e.Status == "" {
		return ExpenseActive
	}
	return e.Status
}

type ExpenseInput struct {
	Description  string
	AmountPaise  int64
	PaidBy       string
	Participants []string
}

func NewProposedExpense(input ExpenseInput, members []Member, actorUID string, now time.Time) (Expense, error) {
	input.Description = strings.TrimSpace(input.Description)
	if input.Description == "" || len(input.Description) > 200 || input.AmountPaise <= 0 || input.AmountPaise > 100000000000 {
		return Expense{}, errors.New("Provide a description of up to 200 characters and a positive amount in paise")
	}
	memberByID := make(map[string]Member, len(members))
	actorMemberID := ""
	for _, member := range members {
		memberByID[member.ID] = member
		if member.UID == actorUID && actorUID != "" {
			actorMemberID = member.ID
		}
	}
	if actorMemberID == "" {
		return Expense{}, errors.New("Link your account to a group member before proposing expenses")
	}
	if payer, ok := memberByID[input.PaidBy]; !ok || payer.UID == "" {
		return Expense{}, errors.New("Payer must be a joined group member")
	}
	if len(input.Participants) == 0 {
		return Expense{}, errors.New("Select at least one participant")
	}
	seen := make(map[string]bool, len(input.Participants))
	approvals := make([]ExpenseApproval, 0, len(input.Participants))
	history := make([]ExpenseDecision, 0, 1)
	pending := 0
	for _, id := range input.Participants {
		member, ok := memberByID[id]
		if !ok || member.UID == "" || seen[id] {
			return Expense{}, errors.New("Participants must be distinct joined group members")
		}
		seen[id] = true
		approval := ExpenseApproval{MemberID: id, Status: "pending"}
		if id == actorMemberID {
			approval.Status = "approved"
			decidedAt := now
			approval.DecidedAt = &decidedAt
			history = append(history, ExpenseDecision{MemberID: id, Decision: "approved", Cycle: 1, At: now})
		} else {
			pending++
		}
		approvals = append(approvals, approval)
	}
	status := ExpenseProposed
	if pending == 0 {
		status = ExpenseActive
	}
	return Expense{
		ID: primitive.NewObjectID().Hex(), Description: input.Description,
		AmountPaise: input.AmountPaise, PaidBy: input.PaidBy,
		Participants: append([]string(nil), input.Participants...),
		Date:         now, CreatedBy: actorUID, Status: status, Version: 1, Cycle: 1,
		Approvals: approvals, DecisionHistory: history,
	}, nil
}

func (e Expense) Decide(memberID, decision string, now time.Time) (Expense, error) {
	if e.Status != ExpenseProposed {
		return Expense{}, errors.New("Only proposed expenses can be approved or rejected")
	}
	if decision != "approved" && decision != "rejected" {
		return Expense{}, errors.New("Decision must be approved or rejected")
	}
	updated := e
	updated.Approvals = append([]ExpenseApproval(nil), e.Approvals...)
	found := false
	for i := range updated.Approvals {
		if updated.Approvals[i].MemberID == memberID {
			if updated.Approvals[i].Status != "pending" {
				return Expense{}, errors.New("You already decided on this proposal")
			}
			decidedAt := now
			updated.Approvals[i].Status = decision
			updated.Approvals[i].DecidedAt = &decidedAt
			found = true
			break
		}
	}
	if !found {
		return Expense{}, errors.New("You are not a participant in this proposal")
	}
	updated.Version++
	updated.DecisionHistory = append(append([]ExpenseDecision(nil), e.DecisionHistory...), ExpenseDecision{
		MemberID: memberID, Decision: decision, Cycle: e.Cycle, At: now,
	})
	if decision == "rejected" {
		updated.Status = ExpenseRejected
	} else {
		allApproved := true
		for _, approval := range updated.Approvals {
			if approval.Status != "approved" {
				allApproved = false
				break
			}
		}
		if allApproved {
			updated.Status = ExpenseActive
		}
	}
	return updated, nil
}

func (e Expense) Resubmit(replacement Expense) (Expense, error) {
	if e.Status != ExpenseRejected {
		return Expense{}, errors.New("Only rejected expenses can be resubmitted")
	}
	if replacement.CreatedBy != e.CreatedBy {
		return Expense{}, errors.New("Only the proposer can resubmit")
	}
	replacement.ID = e.ID
	replacement.RecurringID = e.RecurringID
	replacement.PeriodKey = e.PeriodKey
	replacement.Date = e.Date
	replacement.Version = e.Version + 1
	replacement.Cycle = e.Cycle + 1
	if e.RecurringID != "" {
		replacement.Status = ExpenseProposed
		for i := range replacement.Approvals {
			replacement.Approvals[i].Status = "pending"
			replacement.Approvals[i].DecidedAt = nil
		}
	}
	replacement.DecisionHistory = append([]ExpenseDecision(nil), e.DecisionHistory...)
	for _, approval := range replacement.Approvals {
		if approval.Status == "approved" && approval.DecidedAt != nil {
			replacement.DecisionHistory = append(replacement.DecisionHistory, ExpenseDecision{
				MemberID: approval.MemberID, Decision: "approved", Cycle: replacement.Cycle, At: *approval.DecidedAt,
			})
		}
	}
	return replacement, nil
}

func (e Expense) Withdraw(actorUID string, now time.Time) (Expense, error) {
	if e.CreatedBy == "" || e.CreatedBy != actorUID {
		return Expense{}, errors.New("Only the proposer can withdraw this expense")
	}
	if e.Status != ExpenseProposed && e.Status != ExpenseRejected {
		return Expense{}, errors.New("Only proposed or rejected expenses can be withdrawn")
	}
	updated := e
	updated.Version++
	updated.Status = ExpenseWithdrawn
	updated.DecisionHistory = append(append([]ExpenseDecision(nil), e.DecisionHistory...), ExpenseDecision{
		Decision: "withdrawn", Cycle: e.Cycle, At: now,
	})
	return updated, nil
}
