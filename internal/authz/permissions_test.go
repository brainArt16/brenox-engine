package authz

import "testing"

func TestAllowed_CreateChannel(t *testing.T) {
	tests := []struct {
		role   Role
		allow  bool
	}{
		{RoleOwner, true},
		{RoleAdmin, true},
		{RoleModerator, true},
		{RoleMember, false},
	}

	for _, tt := range tests {
		got := Allowed(tt.role, ActionCreateChannel, Options{})
		if got != tt.allow {
			t.Fatalf("role %s: got %v want %v", tt.role, got, tt.allow)
		}
	}
}

func TestAllowed_ManageMembers(t *testing.T) {
	for _, role := range []Role{RoleOwner, RoleAdmin} {
		if !Allowed(role, ActionInviteMember, Options{}) {
			t.Fatalf("%s should invite members", role)
		}
	}
	if Allowed(RoleModerator, ActionInviteMember, Options{}) {
		t.Fatal("moderator should not invite members")
	}
}

func TestAllowed_SendMessage_ReadOnlyChannel(t *testing.T) {
	if !Allowed(RoleMember, ActionSendMessage, Options{ReadOnlyChannel: false}) {
		t.Fatal("member should send in normal channel")
	}
	if Allowed(RoleMember, ActionSendMessage, Options{ReadOnlyChannel: true}) {
		t.Fatal("member should not send in read-only channel")
	}
	if !Allowed(RoleModerator, ActionSendMessage, Options{ReadOnlyChannel: true}) {
		t.Fatal("moderator should send in read-only channel")
	}
	if !Allowed(RoleMember, ActionSendMessage, Options{
		ReadOnlyChannel: true,
		ChannelRole:     RoleModerator,
	}) {
		t.Fatal("channel moderator override should allow send")
	}
}

func TestCanAssignRole(t *testing.T) {
	if !canAssignRole(RoleOwner, RoleAdmin) {
		t.Fatal("owner should assign admin")
	}
	if canAssignRole(RoleAdmin, RoleOwner) {
		t.Fatal("admin should not assign owner")
	}
	if !canAssignRole(RoleAdmin, RoleMember) {
		t.Fatal("admin should assign member")
	}
}
