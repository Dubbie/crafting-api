package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dubbie/calculator-api/internal/domain"
	"github.com/dubbie/calculator-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(
	itemService service.ItemService,
	itemListService service.ListService[domain.Item, domain.ItemFilters],
) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	// Add CORS, authenticate here later

	// Health Check Endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// --- Item Routes ---
	itemHandler := NewItemHandler(itemService)
	// Create the specific list handler for items using the generic factory
	itemListHandler := service.MakeListHandler(itemListService)
	r.Route("/items", func(r chi.Router) {
		itemHandler.RegisterItemRoutes(r, itemListHandler)
	})

	return r
}
