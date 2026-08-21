package handlers

import (
	"log"
	"net/http"

	"backend/middleware"
	"backend/render"
	"backend/internal/models"
	"backend/internal/services"
)

type AIDiagnosisHistory struct {
	ai *services.AIService
}

func NewAIDiagnosisHistoryHandler(
	ai *services.AIService,
) *AIDiagnosisHistory {

	return &AIDiagnosisHistory{
		ai: ai,
	}
}

type AIDiagnosisHistoryPageData struct {
	History []models.Diagnosis
}

func (h *AIDiagnosisHistory) Handler(
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


	if r.Method != http.MethodGet {

		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}


	log.Println("User visited AI diagnosis history")


	history, err := h.ai.History(farmer.ID)

	if err != nil {

		log.Println(
			"failed to load diagnosis history:",
			err,
		)

		http.Error(
			w,
			"Unable to load diagnosis history",
			http.StatusInternalServerError,
		)

		return
	}


	data := AIDiagnosisHistoryPageData{
		History: history,
	}


	if err := render.RenderTemplates(
		w,
		"ai-diagnosis-history.html",
		data,
	); err != nil {

		log.Println(
			"render error:",
			err,
		)

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)

		return
	}

}