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
	r.MethodFunc(http.MethodPost, "/", h.CreateItem)
	r.MethodFunc(http.MethodGet, "/{itemID}", h.GetItemByID)
	r.MethodFunc(http.MethodPut, "/{itemID}", h.UpdateItem)
	r.MethodFunc(http.MethodDelete, "/{itemID}", h.DeleteItem)
}

// --- CreateItem ---
func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req service.CreateItemRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// TODO: Add validation using a library

	newItem, err := h.itemService.CreateItem(ctx, req)
	if err != nil {
		// Specific error mapping
		if errors.Is(err, storage.ErrDuplicateEntry) {
			// Be more specific if service returns refined duplicate errors (name vs slug)
			http.Error(w, "Failed to create item: name or slug may already exist.", http.StatusConflict) // 409 Conflict
		} else {
			// Log the internal error
			fmt.Printf("Internal error creating item: %v\n", err) // Replace with proper logging
			http.Error(w, "Failed to create item", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	if err := json.NewEncoder(w).Encode(newItem); err != nil {
		fmt.Printf("Error encoding JSON response for created item %d: %v\n", newItem.ID, err)
	}
}

// --- GetItemByID ---
func (h *ItemHandler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	// Keep existing implementation
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.itemService.GetItemByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			fmt.Printf("Error getting item by ID %d: %v\n", itemID, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(item); err != nil {
		fmt.Printf("Error encoding JSON response for item %d: %v\n", itemID, err)
	}
}

// --- UpdateItem ---
func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var req service.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// TODO: Add validation using a library

	updatedItem, err := h.itemService.UpdateItem(ctx, itemID, req)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "Item not found", http.StatusNotFound)
		} else if errors.Is(err, storage.ErrDuplicateEntry) {
			http.Error(w, "Failed to update item: name or slug may conflict with an existing item.", http.StatusConflict)
		} else {
			// Log internal error
			fmt.Printf("Internal error updating item %d: %v\n", itemID, err)
			http.Error(w, "Failed to update item", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	if err := json.NewEncoder(w).Encode(updatedItem); err != nil {
		fmt.Printf("Error encoding JSON response for updated item %d: %v\n", itemID, err)
	}
}

// --- DeleteItem ---
func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	err = h.itemService.DeleteItem(ctx, itemID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "Item not found", http.StatusNotFound)
		} else {
			// Log internal error
			// Consider foreign key constraints - might need specific error handling if not using CASCADE
			fmt.Printf("Internal error deleting item %d: %v\n", itemID, err)
			http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}
