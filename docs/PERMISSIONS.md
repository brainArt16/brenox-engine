# Brenox Permissions

Workspace roles (stored on `workspace_members.role`):

| Role | Description |
|------|-------------|
| `owner` | Full workspace control; created with workspace |
| `admin` | Manage members and channels |
| `moderator` | Create channels; post in read-only channels |
| `member` | Join channels; send in non-read-only channels |

Optional `channel_roles` grant channel-level `moderator` for read-only posting overrides.

## Permission matrix

| Action | owner | admin | moderator | member |
|--------|:-----:|:-----:|:---------:|:------:|
| Create channel | yes | yes | yes | no |
| Set channel read-only on create | yes | yes | no | no |
| Invite workspace member | yes | yes | no | no |
| Remove workspace member | yes | yes | no | no |
| Change member role | yes | yes | no | no |
| Join channel | yes | yes | yes | yes |
| Send message (normal channel) | yes | yes | yes | yes |
| Send message (read-only channel) | yes | yes | yes | no* |

\* Members with channel-level `moderator` role may post in read-only channels.

## API enforcement

- `internal/authz` — `Can(workspaceID, userID, action, options)`
- Channel create/list/join — workspace membership + role checks
- Messages — read-only flag on channel
- Member admin routes — owner/admin only

## Member management routes

```http
POST   /api/workspaces/:workspace_id/members
DELETE /api/workspaces/:workspace_id/members/:user_id
PATCH  /api/workspaces/:workspace_id/members/:user_id
GET    /api/workspaces/:workspace_id/members
```

Owner cannot be removed or demoted via API (ownership transfer deferred).
