package app

import (
	"backend/internal/repository"
	"backend/internal/services"
)

// ============================================================
// APPLICATION CONTAINER
//
// Container holds the main services used throughout Agro-Shield.
//
// Instead of creating services separately inside every route,
// we create them once here and make them available through the
// application container.
// ============================================================

type Container struct {
	Auth        *services.AuthService
	Crop        *services.CropService
	Order       *services.OrderService
	Cart        *services.CartService
	AI          *services.AIService
	Negotiation *services.NegotiationService

	// Exposed so route-level authentication middleware can
	// look up the currently logged-in farmer/buyer.
	FarmerRepo *repository.FarmerRepository
}

// ============================================================
// CREATE APPLICATION CONTAINER
//
// This function receives all repositories and providers needed
// by Agro-Shield and uses them to create the application's
// services.
//
// Dependency flow:
//
// Repository
//     ↓
// Service
//     ↓
// Container
//     ↓
// Handlers / Routes
//
// Example:
//
// CartRepository
//      ↓
// CartService
//      ↓
// NegotiationService
//      ↓
// NegotiationHandler
//
// This allows an accepted negotiation to be moved into the
// buyer's cart.
// ============================================================

func NewContainer(
	farmerRepo *repository.FarmerRepository,
	cropRepo *repository.CropRepository,
	orderRepo *repository.OrderRepository,
	cartRepo *repository.CartRepository,
	diagnosisRepo *repository.DiagnosisRepository,
	negotiationRepo *repository.NegotiationRepository,
	negotiationMsgRepo *repository.NegotiationMessageRepository,
	aiProvider services.AIProvider,
) *Container {

	// ========================================================
	// 1. CART SERVICE
	//
	// CartService handles all business logic related to the
	// buyer's shopping cart.
	//
	// It needs:
	//
	// - CartRepository
	//   Handles cart database operations.
	//
	// - CropRepository
	//   Checks the crop/product involved in the cart.
	// ========================================================

	cartService := services.NewCartService(
		cartRepo,
		cropRepo,
	)

	// ========================================================
	// 2. NEGOTIATION SERVICE
	//
	// NegotiationService handles:
	//
	// - Starting negotiations.
	// - Sending offers.
	// - Sending counter-offers.
	// - Accepting individual offers.
	// - Rejecting individual offers.
	// - Keeping rejected negotiations open.
	// - Moving accepted deals into the buyer's cart.
	//
	// It needs:
	//
	// - NegotiationRepository
	// - NegotiationMessageRepository
	// - CropRepository
	// - CartService
	//
	// CartService is passed here because an accepted offer
	// eventually becomes a cart item.
	// ========================================================

	negotiationService := services.NewNegotiationService(
		negotiationRepo,
		negotiationMsgRepo,
		cropRepo,
		cartService,
	)

	// ========================================================
	// 3. RETURN APPLICATION CONTAINER
	//
	// All major Agro-Shield services are now connected.
	// ========================================================

	return &Container{

		// ====================================================
		// AUTH SERVICE
		//
		// Handles farmer/buyer authentication and account
		// related business logic.
		// ====================================================

		Auth: services.NewAuthService(
			farmerRepo,
		),

		// ====================================================
		// CROP SERVICE
		//
		// Handles crop/product operations.
		// ====================================================

		Crop: services.NewCropService(
			cropRepo,
		),

		// ====================================================
		// ORDER SERVICE
		//
		// Handles normal marketplace orders.
		// ====================================================

		Order: services.NewOrderService(
			orderRepo,
			cropRepo,
		),

		// ====================================================
		// CART SERVICE
		//
		// Handles buyer cart operations.
		// ====================================================

		Cart: cartService,

		// ====================================================
		// AI SERVICE
		//
		// Handles AI diagnosis operations.
		// ====================================================

		AI: services.NewAIService(
			diagnosisRepo,
			aiProvider,
		),

		// ====================================================
		// NEGOTIATION SERVICE
		//
		// Handles the negotiation conversation and individual
		// offers.
		// ====================================================

		Negotiation: negotiationService,

		// ====================================================
		// FARMER REPOSITORY
		//
		// Exposed so authentication middleware can look up
		// the logged-in user.
		// ====================================================

		FarmerRepo: farmerRepo,
	}
}