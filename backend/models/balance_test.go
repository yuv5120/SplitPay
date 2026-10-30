package models

import "testing"

func TestCalculateBalancesRemainderAndSettlement(t *testing.T) {
	g := Group{
		Members:  []Member{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		Expenses: []Expense{{AmountPaise: 101, PaidBy: "a", Participants: []string{"a", "b", "c"}}},
	}
	got := CalculateBalances(g)
	want := map[string]int64{"a": 67, "b": -34, "c": -33}
	for id, expected := range want {
		if got[id] != expected {
			t.Fatalf("balance %s: got %d, want %d", id, got[id], expected)
		}
	}
	settlements := SuggestSettlements(got)
	if len(settlements) != 2 || settlements[0].AmountPaise+settlements[1].AmountPaise != 67 {
		t.Fatalf("unexpected settlements: %#v", settlements)
	}
}

func TestCalculateBalancesSkipsMalformedLegacyExpense(t *testing.T) {
	g := Group{
		Members: []Member{{ID: "a"}, {ID: "b"}},
		Expenses: []Expense{
			{LegacyAmount: 1.01, PaidBy: "a", Participants: []string{"a", "b"}},
			{AmountPaise: 500, PaidBy: "unknown", Participants: []string{"a", "b"}},
		},
	}
	got := CalculateBalances(g)
	if got["a"] != 50 || got["b"] != -50 {
		t.Fatalf("unexpected balances: %#v", got)
	}
}

func TestSuccessfulPaymentOnlyChangesBalances(t *testing.T) {
	g := Group{
		Members:  []Member{{ID: "a"}, {ID: "b"}},
		Expenses: []Expense{{AmountPaise: 10001, PaidBy: "a", Participants: []string{"a", "b"}, Status: ExpenseActive}},
		Payments: []PaymentAttempt{
			{From: "b", To: "a", AmountPaise: 2000, Status: PaymentSucceeded},
			{From: "b", To: "a", AmountPaise: 3000, Status: PaymentPending},
			{From: "b", To: "a", AmountPaise: 1000, Status: PaymentFailed},
			{From: "b", To: "a", AmountPaise: 1000, Status: PaymentCancelled},
		},
	}
	balances := CalculateBalances(g)
	if balances["a"] != 3000 || balances["b"] != -3000 {
		t.Fatalf("only succeeded amount should count: %#v", balances)
	}
	if got := OutstandingTo(balances, "b", "a"); got != 3000 {
		t.Fatalf("unexpected outstanding: %d", got)
	}
}
