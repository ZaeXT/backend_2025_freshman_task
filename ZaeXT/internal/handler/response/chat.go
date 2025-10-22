package response

import "time"

type ConversationInfo struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	IsTemporary bool       `json:"is_temporary"`
	CategoryID  *uint      `json:"category_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
