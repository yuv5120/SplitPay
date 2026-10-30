package models

import (
	"testing"
	"time"
)

func testPlan(t *testing.T, now time.Time) (PurchasePlan, []Member) {
	t.Helper()
	members := []Member{{ID: "a", Name: "A", UID: "uid-a"}, {ID: "b", Name: "B", UID: "uid-b"}, {ID: "pending", Name: "C"}}
	plan, err := NewPurchasePlan(PurchasePlanInput{Title: "Weekend stay", VotingDays: 2,
		Options: []PurchaseOptionInput{{Label: "Hotel", AmountPaise: 70000}, {Label: "Apartment", AmountPaise: 60000}}}, members, "uid-a", now)
	if err != nil {
		t.Fatal(err)
	}
	return plan, members
}

func TestPurchasePlanVoteChangeAndTieBreaker(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	plan, members := testPlan(t, now)
	if plan.CanClose(members, "uid-a", "uid-a", now) {
		t.Fatal("closed before members voted")
	}
	var err error
	plan, err = plan.Vote("a", plan.Options[0].ID, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	plan, err = plan.Vote("a", plan.Options[1].ID, now.Add(2*time.Minute))
	if err != nil || len(plan.Votes) != 1 || plan.Votes[0].OptionID != plan.Options[1].ID {
		t.Fatalf("vote change failed: %+v, %v", plan.Votes, err)
	}
	plan, err = plan.Vote("b", plan.Options[0].ID, now.Add(3*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if !plan.CanClose(members, "uid-a", "uid-a", now.Add(4*time.Minute)) {
		t.Fatal("creator could not close after all joined members voted")
	}
	if plan.CanClose(members, "uid-b", "uid-a", now.Add(4*time.Minute)) {
		t.Fatal("non-owner non-creator could close")
	}
	closed, err := plan.Close(members, "uid-a", "uid-a", now.Add(4*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if closed.Status != PlanClosed || closed.SelectedOptionID != plan.Options[1].ID {
		t.Fatalf("tied vote should select cheaper option: %+v", closed)
	}
	if _, err := closed.Vote("a", plan.Options[0].ID, now.Add(5*time.Minute)); err == nil {
		t.Fatal("accepted vote after close")
	}
	if _, err := closed.Cancel("uid-a", "uid-a"); err == nil {
		t.Fatal("cancelled a closed plan")
	}
}

func TestPurchasePlanDeadlineAndInputRules(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	plan, members := testPlan(t, now)
	if _, err := plan.Vote("a", "missing", now); err == nil {
		t.Fatal("accepted unknown option")
	}
	if _, err := plan.Vote("a", plan.Options[0].ID, plan.DeadlineAt); err == nil {
		t.Fatal("accepted late vote")
	}
	if !plan.CanClose(members, "uid-a", "uid-a", plan.DeadlineAt) {
		t.Fatal("creator cannot close at deadline")
	}
	if _, err := plan.Close(members, "uid-a", "uid-a", plan.DeadlineAt); err == nil {
		t.Fatal("closed without votes")
	}
	cancelled, err := plan.Cancel("uid-a", "uid-a")
	if err != nil || cancelled.Status != PlanCancelled {
		t.Fatalf("could not cancel expired empty plan: %v", err)
	}
	_, err = NewPurchasePlan(PurchasePlanInput{Title: "Duplicate", VotingDays: 1,
		Options: []PurchaseOptionInput{{Label: "Hotel", AmountPaise: 100}, {Label: "hotel", AmountPaise: 200}}}, members, "uid-a", now)
	if err == nil {
		t.Fatal("accepted duplicate option labels")
	}
	_, err = NewPurchasePlan(PurchasePlanInput{Title: "Solo", VotingDays: 1,
		Options: []PurchaseOptionInput{{Label: "A", AmountPaise: 100}, {Label: "B", AmountPaise: 200}}}, members, "uid-unknown", now)
	if err == nil {
		t.Fatal("non-member created plan")
	}
}
