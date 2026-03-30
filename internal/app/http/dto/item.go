package dto

type CreateTodoItemRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTodoItemRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Done        *bool   `json:"done,omitempty"`
}

type TodoItemResponse struct {
	ID          int64  `json:"id"`
	ListID      int64  `json:"list_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}
