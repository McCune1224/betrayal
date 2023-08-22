package main

import "testing"

func TestParseChargeAndAbilityType(t *testing.T) {
	tests := []struct {
		input               string
		expectedCharge      int
		expectedAbilityType string
	}{
		{
			"[x3]*",
			3,
			"*",
		}, {
			"(x3)*",
			3,
			"*",
		}, {
			"[x3]**",
			3,
			"**",
		},
		{
			"[x3]^",
			3,
			"^",
		},
		{
			"[x∞]^",
			-1,
			"^",
		},
		{
			"(x∞)*",
			-1,
			"*",
		},
	}

	for _, test := range tests {
		charge, abilityType, err := parseChargeAndAbilityType(test.input)
		if err != nil {
			t.Error(err.Error())
		}
		if charge != test.expectedCharge {
			t.Errorf("Expected charge to be %d, got %d", test.expectedCharge, charge)
		}
		if abilityType != test.expectedAbilityType {
			t.Errorf("Expected abilityType to be %s, got %s", test.expectedAbilityType, abilityType)
		}
		t.Log("Test passed", test.input, charge, abilityType)
	}

}
