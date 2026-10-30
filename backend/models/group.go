package models

import (
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Member represents a member in a group
type Member struct {
	ID   string `bson:"id" json:"id"`
	Name string `bson:"name" json:"name"`
	UID  string `bson:"uid,omitempty" json:"-"`
}

type MemberResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsJoined bool   `json:"isJoined"`
}

type Activity struct {
	ID        string    `bson:"id"`
	Kind      string    `bson:"kind"`
	ActorUID  string    `bson:"actorUid"`
	Detail    string    `bson:"detail"`
	CreatedAt time.Time `bson:"createdAt"`
}

func NewActivity(kind, actorUID, detail string) Activity {
	return Activity{
		ID: primitive.NewObjectID().Hex(), Kind: kind, ActorUID: actorUID,
		Detail: detail, CreatedAt: time.Now(),
	}
}

type ActivityResponse struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	ActorName string    `json:"actorName"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}

// Expense represents an expense in a group
type Expense struct {
	ID                 string            `bson:"id" json:"id"`
	Description        string            `bson:"description" json:"description"`
	AmountPaise        int64             `bson:"amountPaise,omitempty" json:"amountPaise"`
	LegacyAmount       float64           `bson:"amount,omitempty" json:"-"`
	PaidBy             string            `bson:"paidBy" json:"paidBy"`
	Participants       []string          `bson:"participants" json:"participants"`
	Date               time.Time         `bson:"date" json:"date"`
	CreatedBy          string            `bson:"createdBy,omitempty" json:"-"`
	Status             string            `bson:"status,omitempty" json:"-"`
	Version            int64             `bson:"version,omitempty" json:"-"`
	Cycle              int               `bson:"cycle,omitempty" json:"-"`
	Approvals          []ExpenseApproval `bson:"approvals,omitempty" json:"-"`
	DecisionHistory    []ExpenseDecision `bson:"decisionHistory,omitempty" json:"-"`
	RecurringID        string            `bson:"recurringId,omitempty" json:"-"`
	PeriodKey          string            `bson:"periodKey,omitempty" json:"-"`
	AnomalyExplanation string            `bson:"anomalyExplanation,omitempty" json:"-"`
	PurchasePlanID     string            `bson:"purchasePlanId,omitempty" json:"-"`
}

type ExpenseApproval struct {
	MemberID  string     `bson:"memberId" json:"memberId"`
	Status    string     `bson:"status" json:"status"`
	DecidedAt *time.Time `bson:"decidedAt,omitempty" json:"decidedAt,omitempty"`
}

type ExpenseDecision struct {
	MemberID string    `bson:"memberId" json:"memberId"`
	Decision string    `bson:"decision" json:"decision"`
	Cycle    int       `bson:"cycle" json:"cycle"`
	At       time.Time `bson:"at" json:"at"`
}

// Paise keeps existing float-based expenses readable until they are migrated.
func (e Expense) Paise() int64 {
	if e.AmountPaise != 0 {
		return e.AmountPaise
	}
	return int64(math.Round(e.LegacyAmount * 100))
}

type ExpenseResponse struct {
	ID                 string            `json:"id"`
	Description        string            `json:"description"`
	Amount             float64           `json:"amount"`
	AmountPaise        int64             `json:"amountPaise"`
	PaidBy             string            `json:"paidBy"`
	Participants       []string          `json:"participants"`
	Date               time.Time         `json:"date"`
	CanDelete          bool              `json:"canDelete"`
	CanEdit            bool              `json:"canEdit"`
	CanWithdraw        bool              `json:"canWithdraw"`
	Status             string            `json:"status"`
	Cycle              int               `json:"cycle"`
	Approvals          []ExpenseApproval `json:"approvals"`
	DecisionHistory    []ExpenseDecision `json:"decisionHistory"`
	RecurringID        string            `json:"recurringId,omitempty"`
	PeriodKey          string            `json:"periodKey,omitempty"`
	AnomalyExplanation string            `json:"anomalyExplanation,omitempty"`
	PurchasePlanID     string            `json:"purchasePlanId,omitempty"`
}

func (e Expense) ResponseFor(uid string, canManage bool) ExpenseResponse {
	return ExpenseResponse{
		ID: e.ID, Description: e.Description,
		Amount: float64(e.Paise()) / 100, AmountPaise: e.Paise(),
		PaidBy: e.PaidBy, Participants: e.Participants, Date: e.Date,
		CanDelete:   e.Status == "" && canManage,
		CanEdit:     e.Status == "rejected" && uid != "" && e.CreatedBy == uid,
		CanWithdraw: (e.Status == "proposed" || e.Status == "rejected") && uid != "" && e.CreatedBy == uid,
		Status:      e.EffectiveStatus(), Cycle: e.Cycle,
		Approvals:       append([]ExpenseApproval{}, e.Approvals...),
		DecisionHistory: append([]ExpenseDecision{}, e.DecisionHistory...),
		RecurringID:     e.RecurringID, PeriodKey: e.PeriodKey, AnomalyExplanation: e.AnomalyExplanation,
		PurchasePlanID: e.PurchasePlanID,
	}
}

// Group represents a group with members and expenses
type Group struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	GroupID            string             `bson:"id" json:"id"`
	Name               string             `bson:"name" json:"name"`
	Members            []Member           `bson:"members" json:"members"`
	Expenses           []Expense          `bson:"expenses" json:"expenses"`
	Activity           []Activity         `bson:"activity,omitempty" json:"-"`
	Payments           []PaymentAttempt   `bson:"payments,omitempty" json:"-"`
	Recurring          []RecurringCost    `bson:"recurring,omitempty" json:"-"`
	PurchasePlans      []PurchasePlan     `bson:"purchasePlans,omitempty" json:"-"`
	Reminders          []ReminderEvent    `bson:"reminders,omitempty" json:"-"`
	ReminderOptIn      []string           `bson:"reminderOptIn,omitempty" json:"-"`
	BalanceRevision    int64              `bson:"balanceRevision,omitempty" json:"-"`
	MonthlyBudgetPaise int64              `bson:"monthlyBudgetPaise,omitempty" json:"-"`
	UserID             string             `bson:"userId" json:"userId"`
	CreatedAt          time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt          time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// GroupResponse is the response structure for group data
type GroupResponse struct {
	ID                      string                 `json:"id"`
	Name                    string                 `json:"name"`
	Members                 []MemberResponse       `json:"members"`
	Expenses                []ExpenseResponse      `json:"expenses"`
	CanManage               bool                   `json:"canManage"`
	CurrentMemberID         string                 `json:"currentMemberId,omitempty"`
	BalancesPaise           map[string]int64       `json:"balancesPaise"`
	Settlements             []SettlementSuggestion `json:"settlements"`
	Activity                []ActivityResponse     `json:"activity"`
	Payments                []PaymentAttempt       `json:"payments"`
	Recurring               []RecurringResponse    `json:"recurring"`
	PurchasePlans           []PurchasePlanResponse `json:"purchasePlans"`
	MonthAheadForecastPaise int64                  `json:"monthAheadForecastPaise"`
	Budget                  BudgetOverview         `json:"budget"`
	RemindersEnabled        bool                   `json:"remindersEnabled"`
	CreatedAt               time.Time              `json:"createdAt"`
}

func (g Group) Response() GroupResponse {
	return g.ResponseFor("")
}

func (g Group) ResponseFor(uid string) GroupResponse {
	canManage := uid != "" && g.UserID == uid
	members := make([]MemberResponse, len(g.Members))
	currentMemberID := ""
	for i, member := range g.Members {
		members[i] = MemberResponse{ID: member.ID, Name: member.Name, IsJoined: member.UID != ""}
		if member.UID == uid && uid != "" {
			currentMemberID = member.ID
		}
	}
	expenses := make([]ExpenseResponse, len(g.Expenses))
	for i, expense := range g.Expenses {
		expenses[i] = expense.ResponseFor(uid, canManage)
	}
	balances := CalculateBalances(g)
	recurring, forecast := RecurringOverview(g, uid, time.Now())
	plans := make([]PurchasePlanResponse, len(g.PurchasePlans))
	for i, plan := range g.PurchasePlans {
		plans[i] = plan.ResponseFor(g, uid, time.Now())
	}
	activity := make([]ActivityResponse, len(g.Activity))
	for i, event := range g.Activity {
		actorName := "Former member"
		if event.ActorUID == g.UserID {
			actorName = "Group owner"
		}
		for _, member := range g.Members {
			if event.ActorUID != "" && member.UID == event.ActorUID {
				actorName = member.Name
				break
			}
		}
		activity[i] = ActivityResponse{
			ID: event.ID, Kind: event.Kind, ActorName: actorName,
			Detail: event.Detail, CreatedAt: event.CreatedAt,
		}
	}
	return GroupResponse{
		ID: g.GroupID, Name: g.Name, Members: members, Expenses: expenses,
		CanManage: canManage, CurrentMemberID: currentMemberID,
		BalancesPaise: balances, Settlements: SuggestSettlements(balances), CreatedAt: g.CreatedAt,
		Activity:  activity,
		Payments:  append([]PaymentAttempt{}, g.Payments...),
		Recurring: recurring, MonthAheadForecastPaise: forecast,
		PurchasePlans:    plans,
		Budget:           MonthlyBudget(g, time.Now()),
		RemindersEnabled: containsString(g.ReminderOptIn, uid),
	}
}

func containsString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
