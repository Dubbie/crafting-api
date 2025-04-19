package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/dubbie/calculator-api/internal/service"
	"github.com/dubbie/calculator-api/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
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
	ctx := r.Context() // Get context early
	var req service.CreateItemRequest

	// Decode request body first
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid JSON request body", err)
		return
	}
	defer r.Body.Close()

	// Validate the decoded request struct
	if err := validate.StructCtx(ctx, req); err != nil {
		// Check if it's validation errors
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			// Format the validation errors nicely
			errDetails := formatValidationErrors(validationErrs)
			respondWithError(w, r, http.StatusUnprocessableEntity, "Validation failed", err, errDetails) // 422 for validation errors
		} else {
			// Handle other potential errors from validate.StructCtx (unlikely)
			respondWithError(w, r, http.StatusBadRequest, "Failed to validate request", err)
		}
		return
	}

	// Call the service
	newItem, err := h.itemService.CreateItem(ctx, req)
	if err != nil {
		// Map service/storage errors to HTTP status codes
		if errors.Is(err, storage.ErrDuplicateEntry) {
			respondWithError(w, r, http.StatusConflict, "Item name or slug already exists", err) // 409 Conflict
		} else {
			// Logged within respondWithError
			respondWithError(w, r, http.StatusInternalServerError, "Failed to create item", err)
		}
		return
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newItem); err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to encode successful response", err)
	}
}

// --- GetItemByID ---
func (h *ItemHandler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid item ID format", err)
		return
	}

	item, err := h.itemService.GetItemByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			respondWithError(w, r, http.StatusNotFound, "Item not found", err) // Use service/storage error message if preferred: err.Error()
		} else {
			respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve item", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(item); err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to encode successful response", err)
	}
}

// --- UpdateItem ---
func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid item ID format", err)
		return
	}

	var req service.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid JSON request body", err)
		return
	}
	defer r.Body.Close()

	// Validate the decoded request struct
	if err := validate.StructCtx(ctx, req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			errDetails := formatValidationErrors(validationErrs)
			respondWithError(w, r, http.StatusUnprocessableEntity, "Validation failed", err, errDetails)
		} else {
			respondWithError(w, r, http.StatusBadRequest, "Failed to validate request", err)
		}
		return
	}

	// Call the service
	updatedItem, err := h.itemService.UpdateItem(ctx, itemID, req)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			respondWithError(w, r, http.StatusNotFound, "Item not found", err)
		} else if errors.Is(err, storage.ErrDuplicateEntry) {
			respondWithError(w, r, http.StatusConflict, "Item name or slug conflicts with an existing item", err)
		} else {
			respondWithError(w, r, http.StatusInternalServerError, "Failed to update item", err)
		}
		return
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedItem); err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to encode successful response", err)
	}
}

// --- DeleteItem ---
func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid item ID format", err)
		return
	}

	err = h.itemService.DeleteItem(ctx, itemID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			respondWithError(w, r, http.StatusNotFound, "Item not found", err)
		} else {
			// Handle other potential errors (e.g., FK constraints if not CASCADE)
			respondWithError(w, r, http.StatusInternalServerError, "Failed to delete item", err)
		}
		return
	}

	// Successful deletion
	w.WriteHeader(http.StatusNoContent)
}
