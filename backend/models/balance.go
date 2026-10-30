package models

import "sort"

type SettlementSuggestion struct {
	From        string `json:"from"`
	To          string `json:"to"`
	AmountPaise int64  `json:"amountPaise"`
}

// CalculateBalances uses integer paise and assigns any remainder in participant order.
func CalculateBalances(g Group) map[string]int64 {
	balances := make(map[string]int64, len(g.Members))
	for _, member := range g.Members {
		balances[member.ID] = 0
	}
	for _, expense := range g.Expenses {
		if expense.EffectiveStatus() != "active" {
			continue
		}
		amount := expense.Paise()
		if amount <= 0 || len(expense.Participants) == 0 {
			continue
		}
		if _, ok := balances[expense.PaidBy]; !ok {
			continue
		}
		valid := true
		seen := make(map[string]bool, len(expense.Participants))
		for _, id := range expense.Participants {
			if _, ok := balances[id]; !ok || seen[id] {
				valid = false
				break
			}
			seen[id] = true
		}
		if !valid {
			continue
		}
		balances[expense.PaidBy] += amount
		share := amount / int64(len(expense.Participants))
		remainder := amount % int64(len(expense.Participants))
		for i, id := range expense.Participants {
			balances[id] -= share
			if int64(i) < remainder {
				balances[id]--
			}
		}
	}
	for _, payment := range g.Payments {
		if payment.Status != PaymentSucceeded || payment.AmountPaise <= 0 {
			continue
		}
		if _, ok := balances[payment.From]; !ok {
			continue
		}
		if _, ok := balances[payment.To]; !ok {
			continue
		}
		balances[payment.From] += payment.AmountPaise
		balances[payment.To] -= payment.AmountPaise
	}
	return balances
}

// SuggestSettlements greedily clears balances. It does not claim a global
// minimum number of transfers.
func SuggestSettlements(balances map[string]int64) []SettlementSuggestion {
	type balance struct {
		id     string
		amount int64
	}
	var creditors, debtors []balance
	for id, amount := range balances {
		if amount > 0 {
			creditors = append(creditors, balance{id, amount})
		} else if amount < 0 {
			debtors = append(debtors, balance{id, -amount})
		}
	}
	less := func(items []balance) {
		sort.Slice(items, func(i, j int) bool {
			if items[i].amount == items[j].amount {
				return items[i].id < items[j].id
			}
			return items[i].amount > items[j].amount
		})
	}
	less(creditors)
	less(debtors)
	result := make([]SettlementSuggestion, 0)
	for i, j := 0, 0; i < len(debtors) && j < len(creditors); {
		amount := min(debtors[i].amount, creditors[j].amount)
		result = append(result, SettlementSuggestion{From: debtors[i].id, To: creditors[j].id, AmountPaise: amount})
		debtors[i].amount -= amount
		creditors[j].amount -= amount
		if debtors[i].amount == 0 {
			i++
		}
		if creditors[j].amount == 0 {
			j++
		}
	}
	return result
}
