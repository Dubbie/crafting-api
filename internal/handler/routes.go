package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dubbie/calculator-api/internal/config"
	"github.com/dubbie/calculator-api/internal/domain"
	"github.com/dubbie/calculator-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func SetupRoutes(
	cfg config.Config,
	// Item related
	itemService service.ItemService,
	itemListService service.ListService[domain.Item, domain.ItemFilters],
	// Crafting Method related
	craftingMethodService service.CraftingMethodService,
	craftingMethodListService service.ListService[domain.CraftingMethod, domain.CraftingMethodFilters],
) http.Handler {
	r := chi.NewRouter()

	// CORS Middleware Setup
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(corsMiddleware.Handler)

	// Health Check Endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// API
	r.Route("/api/v1", func(r chi.Router) {
		// --- Item Routes ---
		itemHandler := NewItemHandler(itemService)
		itemListHandler := MakeListHandler(itemListService)
		r.Route("/items", func(r chi.Router) {
			itemHandler.RegisterItemRoutes(r, itemListHandler)
		})

		// --- Crafting Method Routes ---
		craftingMethodHandler := NewCraftingMethodHandler(craftingMethodService)
		craftingMethodListHandler := MakeListHandler(craftingMethodListService)
		r.Route("/crafting-methods", func(r chi.Router) {
			craftingMethodHandler.RegisterCraftingMethodRoutes(r, craftingMethodListHandler)
		})
	})

	return r
}
