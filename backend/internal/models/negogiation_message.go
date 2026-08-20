package models

import "time"

// NegotiationMessage represents one message or price offer
// inside a negotiation conversation.
type NegotiationMessage struct {
	ID            int
	NegotiationID int
	SenderID      int

	// OfferPrice is the price proposed by the sender.
	OfferPrice float64

	// Message contains the sender's explanation or chat message.
	Message string

	// OfferStatus describes what happened to this particular offer.
	//
	// Possible values:
	// - "pending"  → waiting for the other party
	// - "accepted" → this offer was accepted
	// - "rejected" → this offer was rejected
	OfferStatus string

	CreatedAt time.Time
}