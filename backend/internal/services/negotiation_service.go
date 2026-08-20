package services

import (
	"errors"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
)

const (
	MaxNegotiationRounds = 15
	NegotiationWindow    = 24 * time.Hour
)

// NegotiationService handles all business logic related to negotiations.
//
// Responsibility:
// - Start negotiations.
// - Send offers and counter-offers.
// - Accept individual offers.
// - Reject individual offers.
// - Keep the negotiation conversation open when an offer is rejected.
// - Move accepted deals into the buyer's cart.
type NegotiationService struct {
	repo        *repository.NegotiationRepository
	msgRepo     *repository.NegotiationMessageRepository
	cropRepo    *repository.CropRepository
	cartService *CartService
}

// NewNegotiationService creates a new NegotiationService.
func NewNegotiationService(
	repo *repository.NegotiationRepository,
	msgRepo *repository.NegotiationMessageRepository,
	cropRepo *repository.CropRepository,
	cartService *CartService,
) *NegotiationService {
	return &NegotiationService{
		repo:        repo,
		msgRepo:     msgRepo,
		cropRepo:    cropRepo,
		cartService: cartService,
	}
}

// StartNegotiation starts a new negotiation.
//
// The first offer is automatically stored as a pending offer.
func (s *NegotiationService) StartNegotiation(
	buyerID,
	cropID int,
	quantity,
	offerPrice float64,
	message string,
) (*models.Negotiation, error) {

	crop, err := s.cropRepo.GetByID(cropID)
	if err != nil {
		return nil, err
	}

	if crop == nil || !crop.ListedForSale {
		return nil, errors.New(
			"this produce is not available for negotiation",
		)
	}

	if crop.FarmerID == buyerID {
		return nil, errors.New(
			"you can't negotiate on your own produce",
		)
	}

	if quantity <= 0 {
		return nil, errors.New(
			"quantity must be greater than zero",
		)
	}

	if quantity > crop.Quantity {
		return nil, errors.New(
			"not enough quantity available",
		)
	}

	if offerPrice <= 0 {
		return nil, errors.New(
			"offer price must be greater than zero",
		)
	}

	negotiation := &models.Negotiation{
		CropID:    cropID,
		BuyerID:   buyerID,
		FarmerID:  crop.FarmerID,
		Quantity:  quantity,
		Status:    "open",
		MaxRounds: MaxNegotiationRounds,
		ExpiresAt: time.Now().Add(NegotiationWindow),
	}

	if err := s.repo.Create(negotiation); err != nil {
		return nil, err
	}

	firstOffer := &models.NegotiationMessage{
		NegotiationID: negotiation.ID,
		SenderID:      buyerID,
		OfferPrice:    offerPrice,
		Message:       message,
		OfferStatus:   "pending",
	}

	if err := s.msgRepo.Create(firstOffer); err != nil {
		return nil, err
	}

	if err := s.repo.IncrementRound(negotiation.ID); err != nil {
		return nil, err
	}

	return negotiation, nil
}

// SendOffer sends a new offer or counter-offer.
//
// Rejecting an offer does NOT close the negotiation.
// The conversation remains open until:
// - an offer is accepted,
// - the negotiation expires,
// - or the negotiation reaches its round limit.
func (s *NegotiationService) SendOffer(
	negotiationID,
	senderID int,
	offerPrice float64,
	message string,
) error {

	negotiation, err := s.repo.GetByID(negotiationID)
	if err != nil {
		return err
	}

	if negotiation == nil {
		return errors.New("negotiation not found")
	}

	if senderID != negotiation.BuyerID &&
		senderID != negotiation.FarmerID {
		return errors.New(
			"you're not part of this negotiation",
		)
	}

	if negotiation.Status != "open" {
		return errors.New(
			"this negotiation is no longer open",
		)
	}

	if negotiation.IsExpired() {
		_ = s.repo.UpdateStatus(
			negotiationID,
			"expired",
		)

		return errors.New(
			"this negotiation has expired",
		)
	}

	if negotiation.RoundCount >= negotiation.MaxRounds {
		return errors.New(
			"you've reached the negotiation round limit — no more offers can be made",
		)
	}

	if offerPrice <= 0 {
		return errors.New(
			"offer price must be greater than zero",
		)
	}

	newOffer := &models.NegotiationMessage{
		NegotiationID: negotiationID,
		SenderID:      senderID,
		OfferPrice:    offerPrice,
		Message:       message,
		OfferStatus:   "pending",
	}

	if err := s.msgRepo.Create(newOffer); err != nil {
		return err
	}

	return s.repo.IncrementRound(negotiationID)
}

// Accept accepts ONE specific offer.
//
// IMPORTANT:
// - Only the selected offer becomes accepted.
// - The negotiation becomes accepted.
// - The accepted deal is moved into the buyer's cart.
// - The conversation history remains in the database.
// - Previous rejected offers remain visible.
func (s *NegotiationService) Accept(
	negotiationID,
	offerID,
	accepterID int,
) error {

	negotiation, err := s.repo.GetByID(negotiationID)
	if err != nil {
		return err
	}

	if negotiation == nil {
		return errors.New("negotiation not found")
	}

	if accepterID != negotiation.BuyerID &&
		accepterID != negotiation.FarmerID {
		return errors.New(
			"you're not part of this negotiation",
		)
	}

	if negotiation.Status != "open" {
		return errors.New(
			"this negotiation is already finalized",
		)
	}

	if negotiation.IsExpired() {
		_ = s.repo.UpdateStatus(
			negotiationID,
			"expired",
		)

		return errors.New(
			"this negotiation has expired",
		)
	}

	offer, err := s.msgRepo.GetByID(offerID)
	if err != nil {
		return err
	}

	if offer == nil {
		return errors.New("offer not found")
	}

	if offer.NegotiationID != negotiationID {
		return errors.New(
			"this offer does not belong to this negotiation",
		)
	}

	if offer.OfferStatus != "pending" {
		return errors.New(
			"this offer is no longer pending",
		)
	}

	// Make sure the agreed quantity is still available.
	crop, err := s.cropRepo.GetByID(negotiation.CropID)
	if err != nil {
		return err
	}

	if crop == nil {
		return errors.New("product no longer exists")
	}

	if !crop.ListedForSale {
		return errors.New(
			"this product is no longer available for sale",
		)
	}

	if negotiation.Quantity > crop.Quantity {
		return errors.New(
			"the requested quantity is no longer available",
		)
	}

	// Mark this specific offer as accepted.
	if err := s.msgRepo.UpdateOfferStatus(
		offerID,
		"accepted",
	); err != nil {
		return err
	}

	// Mark the negotiation as accepted.
	if err := s.repo.UpdateStatus(
		negotiationID,
		"accepted",
	); err != nil {
		return err
	}

	// Move the accepted negotiation into the buyer's cart.
	if err := s.cartService.AddFromNegotiation(
		negotiation,
		offer.OfferPrice,
	); err != nil {
		return err
	}

	return nil
}

// Reject rejects ONE specific offer.
//
// IMPORTANT:
// - Only the selected offer becomes rejected.
// - The negotiation remains OPEN.
// - The conversation remains available.
// - Either party can send another offer.
func (s *NegotiationService) Reject(
	negotiationID,
	offerID,
	rejecterID int,
) error {

	negotiation, err := s.repo.GetByID(negotiationID)
	if err != nil {
		return err
	}

	if negotiation == nil {
		return errors.New("negotiation not found")
	}

	if rejecterID != negotiation.BuyerID &&
		rejecterID != negotiation.FarmerID {
		return errors.New(
			"you're not part of this negotiation",
		)
	}

	if negotiation.Status != "open" {
		return errors.New(
			"this negotiation is no longer open",
		)
	}

	if negotiation.IsExpired() {
		_ = s.repo.UpdateStatus(
			negotiationID,
			"expired",
		)

		return errors.New(
			"this negotiation has expired",
		)
	}

	offer, err := s.msgRepo.GetByID(offerID)
	if err != nil {
		return err
	}

	if offer == nil {
		return errors.New("offer not found")
	}

	if offer.NegotiationID != negotiationID {
		return errors.New(
			"this offer does not belong to this negotiation",
		)
	}

	if offer.OfferStatus != "pending" {
		return errors.New(
			"this offer is no longer pending",
		)
	}

	// Reject ONLY this offer.
	// The negotiation remains open.
	return s.msgRepo.UpdateOfferStatus(
		offerID,
		"rejected",
	)
}

// Thread retrieves the negotiation and its complete message history.
//
// Responsibility:
// - Keep the conversation available.
// - Return accepted, rejected, and pending offers.
// - Never delete negotiation history.
func (s *NegotiationService) Thread(
	negotiationID int,
) (*models.Negotiation, []models.NegotiationMessage, error) {

	negotiation, err := s.repo.GetByID(negotiationID)
	if err != nil || negotiation == nil {
		return negotiation, nil, err
	}

	messages, err := s.msgRepo.ListByNegotiation(
		negotiationID,
	)

	return negotiation, messages, err
}

// MyNegotiations retrieves all negotiations involving a user.
func (s *NegotiationService) MyNegotiations(
	userID int,
) ([]models.Negotiation, error) {

	return s.repo.ListForUser(userID)
}