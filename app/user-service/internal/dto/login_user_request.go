package dto

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
