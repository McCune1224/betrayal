package handlers

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// FormatCommandArguments converts the stored JSONB command arguments into
// a Discord-like command syntax string.
//
// Input JSON example:
//
//	{
//	  "inv": "subcommand",
//	  "inv.add": "subcommand",
//	  "inv.add.player": {"id": "123456", "username": "PlayerName"},
//	  "inv.add.item": "Healing Potion",
//	  "inv.add.quantity": 5
//	}
//
// Output: "add player:PlayerName item:\"Healing Potion\" quantity:5"
func FormatCommandArguments(argsJSON []byte) string {
	if len(argsJSON) == 0 {
		return ""
	}

	var args map[string]interface{}
	if err := json.Unmarshal(argsJSON, &args); err != nil {
		return ""
	}

	if len(args) == 0 {
		return ""
	}

	// Collect subcommands and actual arguments
	var subcommands []string
	params := make(map[string]string)

	// Sort keys to get consistent ordering
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := args[key]

		// Check if this is a subcommand marker
		if strVal, ok := value.(string); ok && strVal == "subcommand" {
			// Extract the last part of the key as the subcommand name
			parts := strings.Split(key, ".")
			subcommands = append(subcommands, parts[len(parts)-1])
			continue
		}

		// Extract the parameter name (last part after the dot)
		parts := strings.Split(key, ".")
		paramName := parts[len(parts)-1]

		// Format the value
		paramValue := formatValue(value)
		if paramValue != "" {
			params[paramName] = paramValue
		}
	}

	// Build the result string
	var result strings.Builder

	// Add subcommands (skip the first one as it's usually the main command)
	if len(subcommands) > 1 {
		for i := 1; i < len(subcommands); i++ {
			if result.Len() > 0 {
				result.WriteString(" ")
			}
			result.WriteString(subcommands[i])
		}
	} else if len(subcommands) == 1 {
		// Single subcommand - include it
		result.WriteString(subcommands[0])
	}

	// Add parameters in sorted order
	paramKeys := make([]string, 0, len(params))
	for k := range params {
		paramKeys = append(paramKeys, k)
	}
	sort.Strings(paramKeys)

	for _, key := range paramKeys {
		if result.Len() > 0 {
			result.WriteString(" ")
		}
		result.WriteString(key)
		result.WriteString(":")
		result.WriteString(params[key])
	}

	return result.String()
}

// formatValue converts a value to a display string
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		if v == "subcommand" {
			return ""
		}
		// Quote strings with spaces
		if strings.Contains(v, " ") {
			return fmt.Sprintf("\"%s\"", v)
		}
		return v

	case float64:
		// JSON numbers are float64, but if it's a whole number, show as int
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%.2f", v)

	case bool:
		if v {
			return "true"
		}
		return "false"

	case map[string]interface{}:
		// This is likely a user/channel/role object with id and name
		// Try to get the display name
		if username, ok := v["username"].(string); ok {
			return username
		}
		if name, ok := v["name"].(string); ok {
			return name
		}
		if id, ok := v["id"].(string); ok {
			return id
		}
		return ""

	case nil:
		return ""

	default:
		return fmt.Sprintf("%v", v)
	}
}
