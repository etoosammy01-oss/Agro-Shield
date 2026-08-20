package handlers

import (
	"log"
	"net/http"
	"strconv"

	"backend/internal/models"
	"backend/internal/services"
	"backend/middleware"
	"backend/render"
)

type Negotiation struct {
	service *services.NegotiationService
}

func NewNegotiationHandler(service *services.NegotiationService) *Negotiation {
	return &Negotiation{
		service: service,
	}
}

type NegotiationListPageData struct {
	Negotiations []models.Negotiation
	UserID       int
}

func (h *Negotiation) ListHandler(w http.ResponseWriter, r *http.Request) {
	farmer, ok := middleware.FarmerFromContext(r)
	if !ok || farmer == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	negotiations, err := h.service.MyNegotiations(farmer.ID)
	if err != nil {
		log.Println("failed to load negotiations:", err)
	}

	data := NegotiationListPageData{
		Negotiations: negotiations,
		UserID:       farmer.ID,
	}

	if err := render.RenderTemplates(
		w,
		"negotiations.html",
		data,
	); err != nil {
		log.Println("render error:", err)
		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)
	}
}

// StartHandler starts a new negotiation from the marketplace.
func (h *Negotiation) StartHandler(w http.ResponseWriter, r *http.Request) {
	farmer, ok := middleware.FarmerFromContext(r)
	if !ok || farmer == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	cropID, err := strconv.Atoi(r.FormValue("crop_id"))
	if err != nil || cropID <= 0 {
		http.Redirect(w, r, "/marketplace", http.StatusSeeOther)
		return
	}

	quantity, err := strconv.ParseFloat(
		r.FormValue("quantity"),
		64,
	)
	if err != nil || quantity <= 0 {
		http.Redirect(w, r, "/marketplace", http.StatusSeeOther)
		return
	}

	price, err := strconv.ParseFloat(
		r.FormValue("offer_price"),
		64,
	)
	if err != nil || price <= 0 {
		http.Redirect(w, r, "/marketplace", http.StatusSeeOther)
		return
	}

	message := r.FormValue("message")

	negotiation, err := h.service.StartNegotiation(
		farmer.ID,
		cropID,
		quantity,
		price,
		message,
	)

	if err != nil {
		log.Println("failed to start negotiation:", err)

		http.Redirect(
			w,
			r,
			"/marketplace",
			http.StatusSeeOther,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/negotiation?id="+strconv.Itoa(negotiation.ID),
		http.StatusSeeOther,
	)
}

type NegotiationThreadPageData struct {
	Negotiation *models.Negotiation
	Messages    []models.NegotiationMessage
	UserID      int
	Error       string
	TimeLeft    string
}

// ThreadHandler displays the negotiation chat.
//
// It handles:
// - Sending new offers.
// - Accepting one specific offer.
// - Rejecting one specific offer.
//
// Rejecting an offer does NOT close the negotiation.
func (h *Negotiation) ThreadHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	farmer, ok := middleware.FarmerFromContext(r)

	if !ok || farmer == nil {
		http.Redirect(
			w,
			r,
			"/login",
			http.StatusSeeOther,
		)
		return
	}

	negotiationID, err := strconv.Atoi(
		r.URL.Query().Get("id"),
	)

	if err != nil || negotiationID <= 0 {
		http.Error(
			w,
			"Invalid negotiation ID",
			http.StatusBadRequest,
		)
		return
	}

	switch r.Method {

	case http.MethodGet:

		h.render(
			w,
			negotiationID,
			farmer.ID,
			"",
		)

	case http.MethodPost:

		action := r.FormValue("action")

		switch action {

		case "offer":

			price, err := strconv.ParseFloat(
				r.FormValue("offer_price"),
				64,
			)

			if err != nil || price <= 0 {
				h.render(
					w,
					negotiationID,
					farmer.ID,
					"invalid offer price",
				)
				return
			}

			message := r.FormValue("message")

			err = h.service.SendOffer(
				negotiationID,
				farmer.ID,
				price,
				message,
			)

			if err != nil {
				h.render(
					w,
					negotiationID,
					farmer.ID,
					err.Error(),
				)
				return
			}

		case "accept":

			offerID, err := strconv.Atoi(
				r.FormValue("offer_id"),
			)

			if err != nil || offerID <= 0 {
				h.render(
					w,
					negotiationID,
					farmer.ID,
					"invalid offer ID",
				)
				return
			}

			err = h.service.Accept(
				negotiationID,
				offerID,
				farmer.ID,
			)

			if err != nil {
				h.render(
					w,
					negotiationID,
					farmer.ID,
					err.Error(),
				)
				return
			}

		case "reject":

			offerID, err := strconv.Atoi(
				r.FormValue("offer_id"),
			)

			if err != nil || offerID <= 0 {
				h.render(
					w,
					negotiationID,
					farmer.ID,
					"invalid offer ID",
				)
				return
			}

			err = h.service.Reject(
				negotiationID,
				offerID,
				farmer.ID,
			)

			if err != nil {
				h.render(
					w,
					negotiationID,
					farmer.ID,
					err.Error(),
				)
				return
			}

		default:

			h.render(
				w,
				negotiationID,
				farmer.ID,
				"invalid negotiation action",
			)
			return
		}

		http.Redirect(
			w,
			r,
			"/negotiation?id="+strconv.Itoa(negotiationID),
			http.StatusSeeOther,
		)

	default:

		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)
	}
}

func (h *Negotiation) render(
	w http.ResponseWriter,
	negotiationID,
	userID int,
	errMsg string,
) {

	negotiation, messages, err := h.service.Thread(
		negotiationID,
	)

	if err != nil || negotiation == nil {
		http.Error(
			w,
			"Negotiation not found",
			http.StatusNotFound,
		)
		return
	}

	timeLeft := negotiation.TimeLeft()

	timeLeftStr := "Expired"

	if timeLeft > 0 {
		hours := int(timeLeft.Hours())
		minutes := int(timeLeft.Minutes()) % 60

		timeLeftStr =
			strconv.Itoa(hours) +
				"h " +
				strconv.Itoa(minutes) +
				"m left"
	}

	data := NegotiationThreadPageData{
		Negotiation: negotiation,
		Messages:    messages,
		UserID:      userID,
		Error:       errMsg,
		TimeLeft:    timeLeftStr,
	}

	if err := render.RenderTemplates(
		w,
		"negotiation.html",
		data,
	); err != nil {

		log.Println("render error:", err)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)
	}
}