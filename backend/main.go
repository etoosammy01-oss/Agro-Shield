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
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.RunMigration(db); err != nil {
		log.Fatal(err)
	}

	farmerRepo := repository.NewFarmerRepository(db)
	cropRepo := repository.NewCropRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	diagnosisRepo := repository.NewDiagnosisRepository(db)
	negotiationRepo := repository.NewNegotiationRepository(db)
	negotiationMsgRepo := repository.NewNegotiationMessageRepository(db)

	aiProvider, err := services.NewGeminiProvider(
		os.Getenv("GEMINI_API_KEY"),
	)
	if err != nil {
		log.Fatal(err)
	}

	container := app.NewContainer(
		farmerRepo,
		cropRepo,
		orderRepo,
		diagnosisRepo,
		negotiationRepo,
		negotiationMsgRepo,
		aiProvider,
	)

	routes.RegisterRoutes(container)

	log.Println("Server Starting on: http://localhost:8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println(err)
	}
}
