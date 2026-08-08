package handlers

import "encoding/json"

// toastTrigger builds the HX-Trigger header value for the base layout's
// showToast listener. Messages are JSON-marshaled — never string-concatenated
// — so user/sheet-derived text (ability names, CSV values, error messages)
// cannot break the header or inject arbitrary JSON.
func toastTrigger(message, typ string) string {
	payload, err := json.Marshal(map[string]any{
		"showToast": map[string]string{"message": message, "type": typ},
	})
	if err != nil {
		// Cannot happen for string values, but never emit a broken header.
		return `{"showToast":{"message":"Unexpected error","type":"error"}}`
	}
	return string(payload)
}
