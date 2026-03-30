package dto

type CreateTodoListRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTodoListRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

type TodoListResponse struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
