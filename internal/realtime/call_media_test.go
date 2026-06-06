package realtime

import "testing"

func TestIsCallMediaEvent(t *testing.T) {
	mediaEvents := []string{
		"call.video.on", "call.video.off",
		"call.screen.start", "call.screen.stop",
		"call.speaker.changed",
		"call.recording.start", "call.recording.stop",
		"call.preferences",
	}
	for _, eventType := range mediaEvents {
		if !isCallMediaEvent(eventType) {
			t.Fatalf("expected %s to be a call media event", eventType)
		}
	}
	if isCallMediaEvent("call.offer") {
		t.Fatal("call.offer should not be a media event")
	}
}
