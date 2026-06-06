package channels

type CreateChannelRequest struct {
	Name       string `json:"name"`
	IsReadOnly bool   `json:"is_read_only"`
}