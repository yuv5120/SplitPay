package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGroupResponseForMemberAndOwner(t *testing.T) {
	g := Group{
		GroupID: "group", UserID: "owner",
		Members:  []Member{{ID: "a", Name: "Owner", UID: "owner"}, {ID: "b", Name: "Guest", UID: "guest"}},
		Expenses: []Expense{{ID: "expense", AmountPaise: 100, PaidBy: "a", Participants: []string{"a", "b"}, CreatedBy: "guest"}},
		Activity: []Activity{{ID: "event", Kind: "member_joined", ActorUID: "guest", Detail: "Joined the group"}},
	}
	owner := g.ResponseFor("owner")
	guest := g.ResponseFor("guest")
	if !owner.CanManage || owner.CurrentMemberID != "a" || !owner.Expenses[0].CanDelete {
		t.Fatalf("owner permissions incorrect: %#v", owner)
	}
	if guest.CanManage || guest.CurrentMemberID != "b" || guest.Expenses[0].CanDelete {
		t.Fatalf("member permissions incorrect: %#v", guest)
	}
	if guest.RemindersEnabled {
		t.Fatal("reminders should require opt-in")
	}
	g.ReminderOptIn = []string{"guest"}
	if !g.ResponseFor("guest").RemindersEnabled {
		t.Fatal("opted-in member should receive reminders")
	}
	if len(guest.Activity) != 1 || guest.Activity[0].ActorName != "Guest" {
		t.Fatalf("activity actor mapping incorrect: %#v", guest.Activity)
	}
	encoded, err := json.Marshal(guest)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "\"uid\"") || strings.Contains(string(encoded), "createdBy") {
		t.Fatalf("private identifiers leaked in group response: %s", encoded)
	}
}
