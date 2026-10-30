package models

import (
	"strings"
	"testing"
	"time"
)

func TestRecurringForecastAndDuplicatePeriodAdvance(t *testing.T) {
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, RecurringLocation)
	plan := RecurringCost{ID: "rent", Description: "Rent", EstimatePaise: 10001,
		DueDay: 5, Participants: []string{"a", "b", "c"}}
	if got := plan.NextDue(now).Format("2006-01-02"); got != "2026-11-05" {
		t.Fatalf("next due: %s", got)
	}
	shares := plan.EstimatedShares()
	if shares["a"] != 3334 || shares["b"] != 3334 || shares["c"] != 3333 {
		t.Fatalf("shares do not sum exactly: %#v", shares)
	}
	group := Group{Recurring: []RecurringCost{plan}}
	items, forecast := RecurringOverview(group, "alice", now)
	if len(items) != 1 || forecast != 10001 {
		t.Fatalf("forecast: %#v %d", items, forecast)
	}
	plan.LastProposedPeriod = "2026-11"
	if plan.UpcomingDue(now).Format("2006-01") != plan.LastProposedPeriod {
		t.Fatal("current upcoming period should remain the same after proposal")
	}
	if got := plan.NextDue(now).Format("2006-01-02"); got != "2026-12-05" {
		t.Fatalf("proposed period repeated: %s", got)
	}
}

func TestRecurringAnomalyUsesRecentActiveBills(t *testing.T) {
	now := time.Now()
	expenses := []Expense{
		{RecurringID: "utility", AmountPaise: 10000, Status: ExpenseActive, Date: now.AddDate(0, -3, 0)},
		{RecurringID: "utility", AmountPaise: 12000, Status: ExpenseActive, Date: now.AddDate(0, -2, 0)},
		{RecurringID: "utility", AmountPaise: 11000, Status: ExpenseActive, Date: now.AddDate(0, -1, 0)},
		{RecurringID: "utility", AmountPaise: 99000, Status: ExpenseProposed, Date: now},
	}
	average, count := RecentRecurringAverage(expenses, "utility")
	if average != 11000 || count != 3 {
		t.Fatalf("average: %d, count: %d", average, count)
	}
	if RecurringAnomaly(13000, average, "recent approved average") != "" {
		t.Fatal("small change incorrectly flagged")
	}
	message := RecurringAnomaly(15000, average, "recent approved average")
	if !strings.Contains(message, "36% higher") || !strings.Contains(message, "₹110.00") {
		t.Fatalf("unclear anomaly explanation: %q", message)
	}
}

func TestRecurringScheduleValidation(t *testing.T) {
	members := []Member{{ID: "a", UID: "alice"}, {ID: "b"}}
	input := RecurringInput{Description: "Rent", EstimatePaise: 10000, DueDay: 29, PaidBy: "a", Participants: []string{"a"}}
	if _, err := NewRecurringCost(input, members, "alice", time.Now()); err == nil {
		t.Fatal("invalid due day accepted")
	}
	input.DueDay = 1
	input.Participants = []string{"a", "b"}
	if _, err := NewRecurringCost(input, members, "alice", time.Now()); err == nil {
		t.Fatal("unjoined participant accepted")
	}
}

func TestRecurringBillNeedsExplicitApproval(t *testing.T) {
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, RecurringLocation)
	members := []Member{{ID: "a", UID: "alice"}}
	plan := RecurringCost{ID: "rent", Description: "Rent", EstimatePaise: 10000, DueDay: 5,
		PaidBy: "a", Participants: []string{"a"}}
	expense, err := NewRecurringProposal(plan, 12000, members, "alice", now)
	if err != nil || expense.Status != ExpenseProposed || expense.Approvals[0].Status != "pending" {
		t.Fatalf("bill activated before approval: %#v %v", expense, err)
	}
	group := Group{Members: members, Expenses: []Expense{expense}}
	if CalculateBalances(group)["a"] != 0 {
		t.Fatal("unapproved bill changed balances")
	}
	approved, err := expense.Decide("a", "approved", now)
	if err != nil || approved.Status != ExpenseActive {
		t.Fatalf("approval failed: %#v %v", approved, err)
	}
	rejected, err := expense.Decide("a", "rejected", now)
	if err != nil {
		t.Fatal(err)
	}
	replacement, err := NewProposedExpense(ExpenseInput{Description: "Rent", AmountPaise: 13000,
		PaidBy: "a", Participants: []string{"a"}}, members, "alice", now)
	if err != nil {
		t.Fatal(err)
	}
	resubmitted, err := rejected.Resubmit(replacement)
	if err != nil || resubmitted.Status != ExpenseProposed || resubmitted.Approvals[0].Status != "pending" || len(resubmitted.DecisionHistory) != 1 {
		t.Fatalf("resubmission bypassed approval or lost history: %#v %v", resubmitted, err)
	}
}
