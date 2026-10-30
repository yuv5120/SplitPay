package models

import "time"

// BudgetOverview is a planning projection, separate from balances and settlements.
type BudgetOverview struct {
	Period          string `json:"period"`
	LimitPaise      int64  `json:"limitPaise"`
	ApprovedPaise   int64  `json:"approvedPaise"`
	PendingPaise    int64  `json:"pendingPaise"`
	ScheduledPaise  int64  `json:"scheduledPaise"`
	ProjectedPaise  int64  `json:"projectedPaise"`
	RemainingPaise  int64  `json:"remainingPaise"`
	OverBudgetPaise int64  `json:"overBudgetPaise"`
}

func MonthlyBudget(g Group, now time.Time) BudgetOverview {
	period := now.In(RecurringLocation).Format("2006-01")
	view := BudgetOverview{Period: period, LimitPaise: g.MonthlyBudgetPaise}
	recurringPeriods := make(map[string]bool)
	for _, expense := range g.Expenses {
		expensePeriod := expense.Date.In(RecurringLocation).Format("2006-01")
		if expense.RecurringID != "" && expense.PeriodKey != "" {
			expensePeriod = expense.PeriodKey
		}
		if expensePeriod != period || expense.Paise() <= 0 {
			continue
		}
		switch expense.EffectiveStatus() {
		case ExpenseActive:
			view.ApprovedPaise += expense.Paise()
		case ExpenseProposed:
			view.PendingPaise += expense.Paise()
		default:
			continue
		}
		if expense.RecurringID != "" && expense.PeriodKey == period {
			recurringPeriods[expense.RecurringID] = true
		}
	}
	for _, plan := range g.Recurring {
		if plan.Archived || plan.EstimatePaise <= 0 || recurringPeriods[plan.ID] {
			continue
		}
		// The monthly plan is counted once even when its due date has passed.
		view.ScheduledPaise += plan.EstimatePaise
	}
	view.ProjectedPaise = view.ApprovedPaise + view.PendingPaise + view.ScheduledPaise
	if view.LimitPaise > 0 {
		view.RemainingPaise = view.LimitPaise - view.ProjectedPaise
		if view.RemainingPaise < 0 {
			view.OverBudgetPaise = -view.RemainingPaise
		}
	}
	return view
}
