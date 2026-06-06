package realtime

func isCallSignal(eventType string) bool {
	switch eventType {
	case "call.offer", "call.answer", "call.ice":
		return true
	default:
		return false
	}
}

func payloadInt64(payload any, key string) int64 {
	payloadMap, ok := payload.(map[string]any)
	if !ok {
		return 0
	}
	switch value := payloadMap[key].(type) {
	case float64:
		return int64(value)
	case int64:
		return value
	case int:
		return int64(value)
	default:
		return 0
	}
}

func withFromUser(payload any, userID int64) map[string]any {
	payloadMap, ok := payload.(map[string]any)
	if !ok {
		return map[string]any{"from_user_id": userID}
	}

	cloned := make(map[string]any, len(payloadMap)+1)
	for key, value := range payloadMap {
		cloned[key] = value
	}
	cloned["from_user_id"] = userID
	return cloned
}
