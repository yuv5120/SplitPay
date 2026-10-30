package models

import (
	"testing"
	"time"
)

func TestMonthlyBudgetCountsCurrentMonthAndAvoidsRecurringDoubleCount(t *testing.T) {
	now := time.Date(2026, time.October, 30, 12, 0, 0, 0, RecurringLocation)
	group := Group{MonthlyBudgetPaise: 300000, Expenses: []Expense{
		{ID: "approved", Date: now.AddDate(0, 0, -2), AmountPaise: 100000, Status: ExpenseActive},
		{ID: "old", Date: now.AddDate(0, -1, 0), AmountPaise: 900000, Status: ExpenseActive},
		{ID: "bill", Date: now, AmountPaise: 75000, Status: ExpenseProposed, RecurringID: "electricity", PeriodKey: "2026-10"},
		{ID: "next-month", Date: now, AmountPaise: 70000, Status: ExpenseProposed, RecurringID: "november", PeriodKey: "2026-11"},
		{ID: "withdrawn", Date: now, AmountPaise: 500000, Status: ExpenseWithdrawn},
	}, Recurring: []RecurringCost{
		{ID: "electricity", EstimatePaise: 60000, DueDay: 10},
		{ID: "rent", EstimatePaise: 180000, DueDay: 2},
		{ID: "archived", EstimatePaise: 50000, Archived: true},
	}}
	view := MonthlyBudget(group, now)
	if view.Period != "2026-10" || view.ApprovedPaise != 100000 || view.PendingPaise != 75000 || view.ScheduledPaise != 180000 || view.ProjectedPaise != 355000 || view.OverBudgetPaise != 55000 || view.RemainingPaise != -55000 {
		t.Fatalf("unexpected budget projection: %+v", view)
	}
	group.Expenses[2].Status = ExpenseActive
	view = MonthlyBudget(group, now)
	if view.ApprovedPaise != 175000 || view.PendingPaise != 0 || view.ScheduledPaise != 180000 {
		t.Fatalf("approved bill counted incorrectly: %+v", view)
	}
}

func TestMonthlyBudgetWithoutTargetStillShowsProjection(t *testing.T) {
	now := time.Date(2026, time.December, 31, 23, 0, 0, 0, RecurringLocation)
	view := MonthlyBudget(Group{Recurring: []RecurringCost{{EstimatePaise: 5000, DueDay: 1}}}, now)
	if view.ProjectedPaise != 5000 || view.RemainingPaise != 0 || view.OverBudgetPaise != 0 || view.Period != "2026-12" {
		t.Fatalf("unexpected unset budget: %+v", view)
	}
}
