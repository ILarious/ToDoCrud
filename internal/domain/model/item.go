package model

type TodoItem struct {
	ID          int64  `json:"id"`
	ListID      int64  `json:"list_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}
