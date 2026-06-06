package authz

import "errors"

var ErrForbidden = errors.New("permission denied")

// Role is a workspace membership role.
type Role string

const (
	RoleOwner     Role = "owner"
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleMember    Role = "member"
)

// Action is an authorization check target.
type Action string

const (
	ActionCreateChannel   Action = "channel.create"
	ActionInviteMember    Action = "workspace.member.invite"
	ActionRemoveMember    Action = "workspace.member.remove"
	ActionChangeMemberRole Action = "workspace.member.change_role"
	ActionSendMessage     Action = "message.send"
)

// Options carries optional context for a permission check.
type Options struct {
	ReadOnlyChannel bool
	ChannelRole     Role // optional channel-level override (moderator elevates send in read-only)
	channelIDValue  int64
}

// Allowed reports whether role may perform action with the given options.
func Allowed(role Role, action Action, opts Options) bool {
	switch action {
	case ActionCreateChannel:
		return role == RoleOwner || role == RoleAdmin || role == RoleModerator
	case ActionInviteMember, ActionRemoveMember, ActionChangeMemberRole:
		return role == RoleOwner || role == RoleAdmin
	case ActionSendMessage:
		if !opts.ReadOnlyChannel {
			return true
		}
		if role == RoleOwner || role == RoleAdmin || role == RoleModerator {
			return true
		}
		return opts.ChannelRole == RoleModerator
	default:
		return false
	}
}

func ParseRole(raw string) (Role, bool) {
	switch Role(raw) {
	case RoleOwner, RoleAdmin, RoleModerator, RoleMember:
		return Role(raw), true
	default:
		return "", false
	}
}

func AssignableRoles(actor Role) []Role {
	switch actor {
	case RoleOwner:
		return []Role{RoleAdmin, RoleModerator, RoleMember}
	case RoleAdmin:
		return []Role{RoleModerator, RoleMember}
	default:
		return nil
	}
}

func canAssignRole(actor, target Role) bool {
	for _, allowed := range AssignableRoles(actor) {
		if allowed == target {
			return true
		}
	}
	return false
}
