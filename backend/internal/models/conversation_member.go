package models

import "time"

// ============================================================
// CONVERSATION MEMBER MODEL
//
// Responsibility:
// - Represents a user belonging to a conversation.
//
// This is the model that makes GROUP CHAT possible.
//
// Example:
//
// Maize Farmers Association
//
//     conversation_id = 5
//
//     user_id = 1
//     user_id = 7
//     user_id = 12
//     user_id = 25
//
// Every row represents ONE member.
// ============================================================

type ConversationMember struct {
	ID int `json:"id"`

	ConversationID int `json:"conversation_id"`

	UserID int `json:"user_id"`

	JoinedAt time.Time `json:"joined_at"`
}