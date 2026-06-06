package presence_test

import (
	"context"
	"testing"

	"github.com/brainart16/brenox/internal/presence"
)

type stubBroadcaster struct {
	online  int
	offline int
	status  int
}

func (s *stubBroadcaster) PublishPresenceOnline(_, _, _ int64)  { s.online++ }
func (s *stubBroadcaster) PublishPresenceOffline(_, _, _ int64) { s.offline++ }
func (s *stubBroadcaster) PublishPresenceStatus(_, _, _ int64, _, _ string) {
	s.status++
}

func TestMemoryPresenceMultiTab(t *testing.T) {
	store := presence.NewMemoryStore(presence.Config{})
	ctx := context.Background()

	first, err := store.Connect(ctx, 1, 10, 100)
	if err != nil {
		t.Fatal(err)
	}
	if !first.BecameOnline {
		t.Fatal("expected first connection to become online")
	}

	second, err := store.Connect(ctx, 1, 10, 101)
	if err != nil {
		t.Fatal(err)
	}
	if second.BecameOnline {
		t.Fatal("second tab should not trigger became online")
	}
	if second.GlobalCount != 2 {
		t.Fatalf("expected 2 connections, got %d", second.GlobalCount)
	}

	oneLeft, err := store.Disconnect(ctx, 1, 10, 101)
	if err != nil {
		t.Fatal(err)
	}
	if oneLeft.BecameOffline {
		t.Fatal("one tab left should remain online")
	}

	last, err := store.Disconnect(ctx, 1, 10, 100)
	if err != nil {
		t.Fatal(err)
	}
	if !last.BecameOffline {
		t.Fatal("expected offline after last tab disconnect")
	}

	p, err := store.Get(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != presence.StatusOffline {
		t.Fatalf("expected offline status, got %s", p.Status)
	}
}

func TestUpdateStatusBroadcast(t *testing.T) {
	broadcaster := &stubBroadcaster{}
	store := presence.NewMemoryStore(presence.Config{})
	ctx := context.Background()

	_, err := store.Connect(ctx, 7, 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	channels, err := store.ActiveChannels(ctx, 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 active channel, got %d", len(channels))
	}

	p, err := store.SetStatus(ctx, 7, presence.StatusAway)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != presence.StatusAway {
		t.Fatalf("expected away, got %s", p.Status)
	}

	for _, ch := range channels {
		broadcaster.PublishPresenceStatus(ch.WorkspaceID, ch.ChannelID, 7, p.Status, p.LastSeen)
	}
	if broadcaster.status != 1 {
		t.Fatalf("expected status broadcast, got %d", broadcaster.status)
	}
}
