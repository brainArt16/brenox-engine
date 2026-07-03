package workspaces

import "errors"

var (
	ErrNotFound          = errors.New("workspace not found")
	ErrNotMember         = errors.New("not a workspace member")
	ErrSlugTaken         = errors.New("workspace slug already exists")
	ErrInvalidSlug       = errors.New("invalid workspace slug")
	ErrForbidden         = errors.New("permission denied")
	ErrInvalidRole       = errors.New("invalid role")
	ErrCannotModifyOwner = errors.New("cannot modify workspace owner")
	ErrUserNotFound      = errors.New("user not found")
	ErrAlreadyMember     = errors.New("user is already a workspace member")
)

type CreateWorkspaceRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type WorkspaceResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	OwnerID   int64  `json:"owner_id"`
	Role      string `json:"role,omitempty"`
	CreatedAt string `json:"created_at"`
}

type AddMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

type MemberResponse struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}
