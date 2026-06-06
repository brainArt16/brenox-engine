package workspaces

import "errors"

var (
	ErrNotFound   = errors.New("workspace not found")
	ErrNotMember  = errors.New("not a workspace member")
	ErrSlugTaken  = errors.New("workspace slug already exists")
	ErrInvalidSlug = errors.New("invalid workspace slug")
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
	CreatedAt string `json:"created_at"`
}
