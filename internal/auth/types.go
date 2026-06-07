package auth

// Structs tag: Provide metadata -- parse JSON and serialize responses

type RegisterRequest struct {

	Email string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`

}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	Token string `json:"token"`
}