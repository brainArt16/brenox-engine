package realtime

func isCallMediaEvent(eventType string) bool {
	switch eventType {
	case "call.video.on", "call.video.off",
		"call.screen.start", "call.screen.stop",
		"call.speaker.changed",
		"call.recording.start", "call.recording.stop",
		"call.preferences":
		return true
	default:
		return false
	}
}

func isCallRecordingEvent(eventType string) bool {
	switch eventType {
	case "call.recording.start", "call.recording.stop":
		return true
	default:
		return false
	}
}
