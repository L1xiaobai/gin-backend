package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

type UpdateUserRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

type ListUsersRequest struct {
	Page     int `json:"page" binding:"min=1"`
	PageSize int `json:"page_size" binding:"min=1,max=100"`
}