package models

// Message represents a chat message.
type Message struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"` // Sender
	Recipient string `json:"recipient"`
	Content   string `json:"content"` // Ciphertext
	CreatedAt string `json:"created_at"`
}
