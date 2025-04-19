package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dubbie/calculator-api/internal/service"
	"github.com/dubbie/calculator-api/internal/storage"
	"github.com/go-chi/chi/v5"
)

type ItemHandler struct {
	itemService service.ItemService
}

// NewItemHandler creates a handler for item-related HTTP requests.
func NewItemHandler(itemService service.ItemService) *ItemHandler {
	return &ItemHandler{
		itemService: itemService,
	}
}

// RegisterItemRoutes sets up the routes for items on the provided router.
func (h *ItemHandler) RegisterItemRoutes(r chi.Router, listHandler http.HandlerFunc) {
	r.MethodFunc(http.MethodGet, "/", listHandler)
	r.Get("/{itemID}", h.GetItemByID)
}

func (h *ItemHandler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.itemService.GetItemByID(ctx, itemID)
	if err != nil {
		// Map service/storage errors to HTTP status codes
		if errors.Is(err, storage.ErrNotFound) { // Check underlying storage error if service wraps it
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			// Log the full error internally
			fmt.Printf("Error getting item by ID %d: %v\n", itemID, err) // Use proper logging
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(item); err != nil {
		fmt.Printf("Error encoding JSON response for item %d: %v\n", itemID, err)
	}
}
