package routes

import "testing"

func TestInviteTokenHash(t *testing.T) {
	token := "example-secret-token"
	first := hashInviteToken(token)
	if first == token || first != hashInviteToken(token) || first == hashInviteToken(token+"x") {
		t.Fatal("invite token hash must be stable, distinct, and not reveal token")
	}
}
