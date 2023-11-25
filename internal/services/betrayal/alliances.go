package betrayal

import "strings"

// TODO: Make these roles configurable via a website or something

// Check if current player is allowed to have 5 members in an alliance
func AllianceMemberLimitBypass(roleName string) bool {
	allowed := []string{
		"hero", "entertainer", "overlord",
	}

	roleName = strings.ToLower(roleName)
	for _, role := range allowed {
		if role == roleName {
			return true
		}
	}
	return false
}

// Check if current player is allowed to be in 2 alliances
func AllianceDoubleBypass(roleName string) bool {
	allowed := []string{
		"backstabber",
	}

	roleName = strings.ToLower(roleName)
	for _, role := range allowed {
		if role == roleName {
			return true
		}
	}
	return false
}

// Check if current player is allowed to be in 2 alliances
func AllianceTripleBypass(roleName string) bool {
	allowed := []string{
		"villager",
	}

	roleName = strings.ToLower(roleName)
	for _, role := range allowed {
		if role == roleName {
			return true
		}
	}
	return false
}
