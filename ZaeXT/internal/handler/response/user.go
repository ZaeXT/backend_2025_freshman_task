package response

import "time"

type UserProfile struct {
	ID         uint      `json:"id"`
	Username   string    `json:"username"`
	Tier       string    `json:"tier"`
	MemoryInfo string    `json:"memory_info"`
	CreatedAt  time.Time `json:"created_at"`
}
