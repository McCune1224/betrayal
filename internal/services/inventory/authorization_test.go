package inventory

import "testing"

func TestInventoryAuthorizationRejectsOtherPlayerInConfessional(t *testing.T) {
	if inventoryAuthorizationDecision("200", []string{}, 100, "confessional", "confessional", nil) {
		t.Fatal("non-admin must not mutate another player's inventory in that player's confessional")
	}
}

func TestInventoryAuthorizationAllowsOwnerInOwnConfessional(t *testing.T) {
	if !inventoryAuthorizationDecision("100", nil, 100, "confessional", "confessional", nil) {
		t.Fatal("owner must be allowed in their own confessional")
	}
}

func TestInventoryAuthorizationAllowsAdminInPlayerConfessional(t *testing.T) {
	if !inventoryAuthorizationDecision("200", []string{"Host"}, 100, "confessional", "confessional", nil) {
		t.Fatal("admin must be allowed in a player's confessional")
	}
}

func TestInventoryAuthorizationRequiresAdminInWhitelistedChannel(t *testing.T) {
	if inventoryAuthorizationDecision("200", nil, 100, "confessional", "admin-channel", []string{"admin-channel"}) {
		t.Fatal("non-admin must not mutate inventory from a whitelisted channel")
	}
	if !inventoryAuthorizationDecision("200", []string{"Host"}, 100, "confessional", "admin-channel", []string{"admin-channel"}) {
		t.Fatal("admin must be allowed in a whitelisted channel")
	}
}
