package models

import (
	"testing"
	"time"
)

func TestExpenseApprovalCycleAndBalances(t *testing.T) {
	now := time.Date(2026, 10, 29, 10, 0, 0, 0, time.UTC)
	members := []Member{{ID: "a", UID: "alice"}, {ID: "b", UID: "bob"}}
	input := ExpenseInput{Description: "Rent", AmountPaise: 10001, PaidBy: "a", Participants: []string{"a", "b"}}
	expense, err := NewProposedExpense(input, members, "alice", now)
	if err != nil || expense.Status != ExpenseProposed || expense.Approvals[0].Status != "approved" || expense.Approvals[1].Status != "pending" {
		t.Fatalf("proposal invalid: %#v %v", expense, err)
	}
	group := Group{Members: members, Expenses: []Expense{expense}}
	if CalculateBalances(group)["a"] != 0 {
		t.Fatal("proposal affected balances")
	}
	rejected, err := expense.Decide("b", "rejected", now.Add(time.Minute))
	if err != nil || rejected.Status != ExpenseRejected {
		t.Fatalf("rejection failed: %#v %v", rejected, err)
	}
	if _, err := rejected.Decide("b", "approved", now); err == nil {
		t.Fatal("rejected expense accepted a decision")
	}
	replacement, err := NewProposedExpense(input, members, "alice", now.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	resubmitted, err := rejected.Resubmit(replacement)
	if err != nil || resubmitted.ID != expense.ID || resubmitted.Cycle != 2 || len(resubmitted.DecisionHistory) != 3 {
		t.Fatalf("resubmission lost history: %#v %v", resubmitted, err)
	}
	active, err := resubmitted.Decide("b", "approved", now.Add(3*time.Minute))
	if err != nil || active.Status != ExpenseActive {
		t.Fatalf("approval failed: %#v %v", active, err)
	}
	group.Expenses[0] = active
	if CalculateBalances(group)["a"] != 5000 || CalculateBalances(group)["b"] != -5000 {
		t.Fatalf("unexpected balances: %#v", CalculateBalances(group))
	}
	if _, err := active.Withdraw("alice", now); err == nil {
		t.Fatal("active expense was withdrawn")
	}
}

func TestProposalRequiresJoinedDistinctMembers(t *testing.T) {
	members := []Member{{ID: "a", UID: "alice"}, {ID: "b"}}
	base := ExpenseInput{Description: "Food", AmountPaise: 100, PaidBy: "a", Participants: []string{"a", "b"}}
	if _, err := NewProposedExpense(base, members, "alice", time.Now()); err == nil {
		t.Fatal("unjoined participant accepted")
	}
	base.Participants = []string{"a", "a"}
	if _, err := NewProposedExpense(base, members, "alice", time.Now()); err == nil {
		t.Fatal("duplicate participant accepted")
	}
	base.Participants = []string{"a"}
	expense, err := NewProposedExpense(base, members, "alice", time.Now())
	if err != nil || expense.Status != ExpenseActive {
		t.Fatalf("single-person expense: %#v %v", expense, err)
	}
}
