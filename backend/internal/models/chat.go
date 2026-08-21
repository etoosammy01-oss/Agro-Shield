package models

import "time"

// ============================================================
// CONVERSATION MODEL
//
// Responsibility:
// - Represents one private General Chat conversation
//   between two Agro-Shield users.
//
// IMPORTANT:
// This is NOT a negotiation.
//
// Negotiations use:
//
// negotiations
// negotiation_messages
//
// General Chat uses:
//
// conversations
// chat_messages
// ============================================================

type Conversation struct {
	ID        int       `json:"id"`
	UserOneID int       `json:"user_one_id"`
	UserTwoID int       `json:"user_two_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================
// CHAT MESSAGE MODEL
//
// Responsibility:
// - Represents one normal message inside a General Chat
//   conversation.
//
// Example:
//
// User A:
// "Hello, are you selling yam?"
//
// User B:
// "Yes, I have 50 bags available."
//
// These are ChatMessage records.
// ============================================================

type ChatMessage struct {
	ID             int       `json:"id"`
	ConversationID int       `json:"conversation_id"`
	SenderID       int       `json:"sender_id"`
	Message        string    `json:"message"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}
