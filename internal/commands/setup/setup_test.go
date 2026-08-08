package setup

import "testing"

func TestGetDeceptionistRejectsMissingSession(t *testing.T) {
	users, err := getDeceptionist(nil, "guild-1")
	if err == nil {
		t.Fatal("getDeceptionist error = nil")
	}
	if users != nil {
		t.Fatalf("users = %#v, want nil", users)
	}
}
