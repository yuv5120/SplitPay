package models

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RecurringCost struct {
	ID                 string    `bson:"id"`
	Description        string    `bson:"description"`
	EstimatePaise      int64     `bson:"estimatePaise"`
	DueDay             int       `bson:"dueDay"`
	PaidBy             string    `bson:"paidBy"`
	Participants       []string  `bson:"participants"`
	CreatedBy          string    `bson:"createdBy"`
	LastProposedPeriod string    `bson:"lastProposedPeriod,omitempty"`
	Archived           bool      `bson:"archived,omitempty"`
	Version            int64     `bson:"version"`
	CreatedAt          time.Time `bson:"createdAt"`
}

// Recurring dates use the product's current INR calendar.
var RecurringLocation = time.FixedZone("IST", 5*3600+30*60)

type RecurringInput struct {
	Description   string   `json:"description"`
	EstimatePaise int64    `json:"estimatePaise"`
	DueDay        int      `json:"dueDay"`
	PaidBy        string   `json:"paidBy"`
	Participants  []string `json:"participants"`
}

type RecurringResponse struct {
	ID                   string           `json:"id"`
	Description          string           `json:"description"`
	EstimatePaise        int64            `json:"estimatePaise"`
	DueDay               int              `json:"dueDay"`
	PaidBy               string           `json:"paidBy"`
	Participants         []string         `json:"participants"`
	NextDueDate          time.Time        `json:"nextDueDate"`
	EstimatedSharesPaise map[string]int64 `json:"estimatedSharesPaise"`
	RecentAveragePaise   int64            `json:"recentAveragePaise"`
	RecentCount          int              `json:"recentCount"`
	LastProposedPeriod   string           `json:"lastProposedPeriod"`
	CanPropose           bool             `json:"canPropose"`
	CanManage            bool             `json:"canManage"`
}

func NewRecurringCost(input RecurringInput, members []Member, actorUID string, now time.Time) (RecurringCost, error) {
	if input.DueDay < 1 || input.DueDay > 28 {
		return RecurringCost{}, errors.New("Due day must be between 1 and 28")
	}
	validated, err := NewProposedExpense(ExpenseInput{Description: input.Description, AmountPaise: input.EstimatePaise,
		PaidBy: input.PaidBy, Participants: input.Participants}, members, actorUID, now)
	if err != nil {
		return RecurringCost{}, err
	}
	return RecurringCost{ID: primitive.NewObjectID().Hex(), Description: validated.Description, EstimatePaise: validated.AmountPaise,
		DueDay: input.DueDay, PaidBy: input.PaidBy, Participants: validated.Participants,
		CreatedBy: actorUID, Version: 1, CreatedAt: now}, nil
}

func NewRecurringProposal(plan RecurringCost, amountPaise int64, members []Member, actorUID string, now time.Time) (Expense, error) {
	expense, err := NewProposedExpense(ExpenseInput{Description: plan.Description, AmountPaise: amountPaise,
		PaidBy: plan.PaidBy, Participants: plan.Participants}, members, actorUID, now)
	if err != nil {
		return Expense{}, err
	}
	expense.RecurringID = plan.ID
	expense.PeriodKey = plan.UpcomingDue(now).Format("2006-01")
	expense.Status = ExpenseProposed
	expense.DecisionHistory = []ExpenseDecision{}
	for i := range expense.Approvals {
		expense.Approvals[i].Status = "pending"
		expense.Approvals[i].DecidedAt = nil
	}
	return expense, nil
}

func (r RecurringCost) UpcomingDue(now time.Time) time.Time {
	loc := now.Location()
	candidate := time.Date(now.Year(), now.Month(), r.DueDay, 0, 0, 0, 0, loc)
	if candidate.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)) {
		candidate = candidate.AddDate(0, 1, 0)
	}
	return candidate
}

func (r RecurringCost) NextDue(now time.Time) time.Time {
	candidate := r.UpcomingDue(now)
	for candidate.Format("2006-01") <= r.LastProposedPeriod {
		candidate = candidate.AddDate(0, 1, 0)
	}
	return candidate
}

func (r RecurringCost) EstimatedShares() map[string]int64 {
	shares := make(map[string]int64, len(r.Participants))
	if len(r.Participants) == 0 {
		return shares
	}
	base := r.EstimatePaise / int64(len(r.Participants))
	remainder := r.EstimatePaise % int64(len(r.Participants))
	for i, id := range r.Participants {
		shares[id] = base
		if int64(i) < remainder {
			shares[id]++
		}
	}
	return shares
}

func RecentRecurringAverage(expenses []Expense, recurringID string) (int64, int) {
	matching := make([]Expense, 0, 3)
	for _, expense := range expenses {
		if expense.RecurringID == recurringID && expense.EffectiveStatus() == ExpenseActive && expense.Paise() > 0 {
			matching = append(matching, expense)
		}
	}
	sort.Slice(matching, func(i, j int) bool { return matching[i].Date.After(matching[j].Date) })
	if len(matching) > 3 {
		matching = matching[:3]
	}
	if len(matching) == 0 {
		return 0, 0
	}
	var total int64
	for _, expense := range matching {
		total += expense.Paise()
	}
	return total / int64(len(matching)), len(matching)
}

func RecurringAnomaly(actual, baseline int64, source string) string {
	if baseline <= 0 || actual <= 0 {
		return ""
	}
	delta := actual - baseline
	if delta < 0 {
		delta = -delta
	}
	if delta < 1000 || delta*100 < baseline*25 {
		return ""
	}
	direction := "higher"
	if actual < baseline {
		direction = "lower"
	}
	return fmt.Sprintf("₹%.2f is %d%% %s than the %s of ₹%.2f. Review before approving.", float64(actual)/100, delta*100/baseline, direction, source, float64(baseline)/100)
}

func RecurringOverview(g Group, uid string, now time.Time) ([]RecurringResponse, int64) {
	now = now.In(RecurringLocation)
	items := make([]RecurringResponse, 0, len(g.Recurring))
	var forecast int64
	for _, plan := range g.Recurring {
		if plan.Archived {
			continue
		}
		due := plan.NextDue(now)
		average, count := RecentRecurringAverage(g.Expenses, plan.ID)
		items = append(items, RecurringResponse{ID: plan.ID, Description: plan.Description, EstimatePaise: plan.EstimatePaise,
			DueDay: plan.DueDay, PaidBy: plan.PaidBy, Participants: append([]string{}, plan.Participants...),
			NextDueDate: due, EstimatedSharesPaise: plan.EstimatedShares(), RecentAveragePaise: average,
			RecentCount: count, LastProposedPeriod: plan.LastProposedPeriod,
			CanPropose: plan.UpcomingDue(now).Format("2006-01") > plan.LastProposedPeriod,
			CanManage:  uid != "" && (uid == plan.CreatedBy || uid == g.UserID)})
		if !due.After(now.Add(30 * 24 * time.Hour)) {
			forecast += plan.EstimatePaise
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].NextDueDate.Before(items[j].NextDueDate) })
	return items, forecast
}
