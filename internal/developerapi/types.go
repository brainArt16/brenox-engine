package developerapi

type ProvisionUserRequest struct {
	ExternalID string `json:"external_id"`
	Email      string `json:"email"`
	Username   string `json:"username"`
}

type CreateChannelRequest struct {
	Name       string `json:"name"`
	IsReadOnly bool   `json:"is_read_only"`
}

type SendMessageRequest struct {
	ChannelID  int64  `json:"channel_id"`
	UserID     int64  `json:"user_id"`
	ExternalID string `json:"external_id"`
	Content    string `json:"content"`
}

type UserResponse struct {
	ID         int64  `json:"id"`
	ExternalID string `json:"external_id"`
	Email      string `json:"email"`
	Username   string `json:"username"`
}

type ChannelResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	WorkspaceID int64  `json:"workspace_id"`
	IsReadOnly  bool   `json:"is_read_only"`
}

type MessageResponse struct {
	ID        int64  `json:"id"`
	ChannelID int64  `json:"channel_id"`
	SenderID  int64  `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type MessageListItem struct {
	ID        int64  `json:"id"`
	ChannelID int64  `json:"channel_id"`
	SenderID  int64  `json:"sender_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// CreateSessionRequest issues a user JWT for an app-provisioned user (embed flow).
type CreateSessionRequest struct {
	ExternalID string `json:"external_id"`
	ChannelID  int64  `json:"channel_id,omitempty"`
}

type SessionResponse struct {
	Token       string       `json:"token"`
	WorkspaceID int64        `json:"workspace_id"`
	Environment string       `json:"environment"`
	ChannelID   int64        `json:"channel_id,omitempty"`
	User        UserResponse `json:"user"`
}
