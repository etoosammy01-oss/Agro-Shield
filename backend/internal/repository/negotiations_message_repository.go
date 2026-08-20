package repository

import (
	"backend/internal/models"
	"database/sql"
)

// ============================================================
// NEGOTIATION MESSAGE REPOSITORY
//
// This repository is responsible for all database operations
// related to messages and offers inside a negotiation.
//
// IMPORTANT:
// A negotiation is the conversation.
// A negotiation message is an individual message/offer.
//
// Therefore, accepting or rejecting an offer changes the
// individual message status, NOT the entire conversation.
// ============================================================

type NegotiationMessageRepository struct {
	db *sql.DB
}

// ============================================================
// 1. CREATE NEGOTIATION MESSAGE
//
// Creates a new message/offer inside a negotiation.
//
// Responsibility:
// - Store who sent the offer.
// - Store the negotiation it belongs to.
// - Store the offered price.
// - Store the message.
// - Store whether the offer is pending, accepted, or rejected.
// ============================================================

func NewNegotiationMessageRepository(db *sql.DB) *NegotiationMessageRepository {
	return &NegotiationMessageRepository{
		db: db,
	}
}

// Create stores a new negotiation message/offer.
func (r *NegotiationMessageRepository) Create(
	m *models.NegotiationMessage,
) error {

	query := `
		INSERT INTO negotiation_messages (
			negotiation_id,
			sender_id,
			offer_price,
			message,
			offer_status
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		m.NegotiationID,
		m.SenderID,
		m.OfferPrice,
		m.Message,
		m.OfferStatus,
	).Scan(
		&m.ID,
		&m.CreatedAt,
	)

	return err
}

// ============================================================
// 2. LIST NEGOTIATION MESSAGES
//
// Returns every message/offer belonging to a negotiation.
//
// Responsibility:
// - Load the complete negotiation conversation.
// - Include pending offers.
// - Include rejected offers.
// - Include accepted offers.
// - Keep the original order of the conversation.
//
// IMPORTANT:
// We do NOT filter out rejected or accepted offers because the
// chat history should remain visible to both parties.
// ============================================================

func (r *NegotiationMessageRepository) ListByNegotiation(
	negotiationID int,
) ([]models.NegotiationMessage, error) {

	query := `
		SELECT
			id,
			negotiation_id,
			sender_id,
			offer_price,
			message,
			offer_status,
			created_at
		FROM negotiation_messages
		WHERE negotiation_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(
		query,
		negotiationID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []models.NegotiationMessage

	for rows.Next() {

		var m models.NegotiationMessage

		err := rows.Scan(
			&m.ID,
			&m.NegotiationID,
			&m.SenderID,
			&m.OfferPrice,
			&m.Message,
			&m.OfferStatus,
			&m.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		messages = append(messages, m)
	}

	return messages, rows.Err()
}

// ============================================================
// 3. GET NEGOTIATION MESSAGE BY ID
//
// Retrieves one specific offer/message.
//
// Responsibility:
// - Find a particular offer.
// - Return all information about that offer.
// - Allow the service layer to check its status.
// - Allow Accept() and Reject() to operate on one specific
//   offer instead of closing the entire negotiation.
// ============================================================

func (r *NegotiationMessageRepository) GetByID(
	offerID int,
) (*models.NegotiationMessage, error) {

	query := `
		SELECT
			id,
			negotiation_id,
			sender_id,
			offer_price,
			message,
			offer_status,
			created_at
		FROM negotiation_messages
		WHERE id = $1
	`

	var m models.NegotiationMessage

	err := r.db.QueryRow(
		query,
		offerID,
	).Scan(
		&m.ID,
		&m.NegotiationID,
		&m.SenderID,
		&m.OfferPrice,
		&m.Message,
		&m.OfferStatus,
		&m.CreatedAt,
	)

	// sql.ErrNoRows means the requested offer does not exist.
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &m, nil
}

// ============================================================
// 4. UPDATE OFFER STATUS
//
// Changes the status of ONE individual offer.
//
// Possible statuses:
//
//   pending  → offer is waiting for a response
//   accepted → this particular offer was accepted
//   rejected → this particular offer was rejected
//
// IMPORTANT:
//
// This function DOES NOT change the negotiation status.
//
// Example:
//
// Buyer offers ₦50,000
//        ↓
// Seller rejects ₦50,000
//        ↓
// This offer becomes "rejected"
//        ↓
// Negotiation remains "open"
//        ↓
// Buyer can send another offer
//
// The negotiation only becomes "accepted" when the service
// layer decides that the deal has been finalized.
// ============================================================

func (r *NegotiationMessageRepository) UpdateOfferStatus(
	offerID int,
	status string,
) error {

	query := `
		UPDATE negotiation_messages
		SET offer_status = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(
		query,
		status,
		offerID,
	)

	if err != nil {
		return err
	}

	// Make sure the offer actually existed.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}