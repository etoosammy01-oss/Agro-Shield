package main

import (
	"log"
	"net/http"
	"os"

	app "backend/internal"
	"backend/internal/database"
	"backend/internal/repository"
	"backend/internal/services"
	"backend/routes"
)

func main() {

	// ============================================================
	// 1. CONNECT TO DATABASE
	// Establishes a connection between Agro-Shield and PostgreSQL.
	// ============================================================

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ============================================================
	// 2. RUN DATABASE MIGRATIONS
	// Makes sure all required Agro-Shield tables exist.
	// ============================================================

	if err := database.RunMigration(db); err != nil {
		log.Fatal(err)
	}

	// ============================================================
	// 3. CREATE REPOSITORIES
	//
	// Repositories are responsible for communicating directly
	// with the PostgreSQL database.
	// ============================================================

	// Farmer repository
	farmerRepo := repository.NewFarmerRepository(db)

	// Crop repository
	cropRepo := repository.NewCropRepository(db)

	// Order repository
	orderRepo := repository.NewOrderRepository(db)

	// Diagnosis repository
	diagnosisRepo := repository.NewDiagnosisRepository(db)

	// Negotiation repository
	negotiationRepo := repository.NewNegotiationRepository(db)

	// Negotiation message repository
	negotiationMsgRepo := repository.NewNegotiationMessageRepository(db)

	// Cart repository
	// This was the missing repository causing your current error.
	cartRepo := repository.NewCartRepository(db)

	// ============================================================
	// 4. CREATE AI PROVIDER
	//
	// GeminiProvider handles communication between Agro-Shield
	// and the Gemini AI service.
	// ============================================================

	aiProvider, err := services.NewGeminiProvider(
		os.Getenv("GEMINI_API_KEY"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// ============================================================
	// 5. CREATE APPLICATION CONTAINER
	//
	// The container connects all repositories and services
	// together so the rest of the application can use them.
	//
	// IMPORTANT:
	// cartRepo must be passed here because NewContainer()
	// now requires it.
	// ============================================================

	container := app.NewContainer(
		farmerRepo,
		cropRepo,
		orderRepo,
		cartRepo,
		diagnosisRepo,
		negotiationRepo,
		negotiationMsgRepo,
		aiProvider,
	)

	// ============================================================
	// 6. REGISTER ROUTES
	//
	// Connects URLs such as:
	// /login
	// /storage
	// /marketplace
	// /negotiation
	// /cart
	// to their handlers.
	// ============================================================

	routes.RegisterRoutes(container)

	// ============================================================
	// 7. START HTTP SERVER
	// ============================================================

	log.Println("Server Starting on: http://localhost:8080...")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println(err)
	}
}