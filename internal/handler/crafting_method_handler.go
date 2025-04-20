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

type CraftingMethodHandler struct {
	craftingMethodService service.CraftingMethodService
}

// NewCraftingMethodHandler creates a new CraftingMethodHandler instance.
func NewCraftingMethodHandler(craftingMethodService service.CraftingMethodService) *CraftingMethodHandler {
	return &CraftingMethodHandler{
		craftingMethodService: craftingMethodService,
	}
}

// RegisterCraftingMethodRoutes sets up the routes for crafting methods on the provided router.
func (h *CraftingMethodHandler) RegisterCraftingMethodRoutes(r chi.Router, listHandler http.HandlerFunc) {
	r.MethodFunc(http.MethodGet, "/", listHandler)
	r.MethodFunc(http.MethodPost, "/", h.CreateCraftingMethod)
	r.MethodFunc(http.MethodGet, "/{methodID}", h.GetCraftingMethodByID)
	r.MethodFunc(http.MethodPut, "/{methodID}", h.UpdateCraftingMethod)
	r.MethodFunc(http.MethodDelete, "/{methodID}", h.DeleteCraftingMethod)
}

// --- CreateCraftingMethod ---
func (h *CraftingMethodHandler) CreateCraftingMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req service.CreateCraftingMethodRequest

	// Decode request body first
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid JSON request body", err)
		return
	}
	defer r.Body.Close()

	// Validate
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
	newMethod, err := h.craftingMethodService.CreateCraftingMethod(ctx, req)
	if err != nil {
		// Map service/storage errors to HTTP status codes
		if errors.Is(err, storage.ErrDuplicateEntry) {
			respondWithError(w, r, http.StatusConflict, "Crafting method name or slug already exists", err) // 409 Conflict
		} else {
			// Logged within respondWithError
			respondWithError(w, r, http.StatusInternalServerError, "Failed to create crafting method", err)
		}
		return
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newMethod); err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to encode successful response", err)
	}
}

// --- GetCraftingMethodByID ---
func (h *CraftingMethodHandler) GetCraftingMethodByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	methodIDstr := chi.URLParam(r, "methodID")
	methodID, err := strconv.ParseUint(methodIDstr, 10, 64)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid method ID format", err)
		return
	}

	item, err := h.craftingMethodService.GetCraftingMethodByID(ctx, methodID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			respondWithError(w, r, http.StatusNotFound, "Crafting method not found", err)
		} else {
			respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve crafting method", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(item); err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to encode successful response", err)
	}
}

// --- UpdateCraftingMethod ---
func (h *CraftingMethodHandler) UpdateCraftingMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	methodIDStr := chi.URLParam(r, "methodID")
	methodID, err := strconv.ParseUint(methodIDStr, 10, 64)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid method ID format", err)
		return
	}

	var req service.UpdateCraftingMethodRequest
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
	updatedMethod, err := h.craftingMethodService.UpdateCraftingMethod(ctx, methodID, req)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			respondWithError(w, r, http.StatusNotFound, "Crafting method not found", err)
		} else if errors.Is(err, storage.ErrDuplicateEntry) {
			respondWithError(w, r, http.StatusConflict, "Crafting method name or slug conflicts with an existing crafting method", err)
		} else {
			respondWithError(w, r, http.StatusInternalServerError, "Failed to update crafting method", err)
		}
		return
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedMethod); err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to encode successful response", err)
	}
}

// --- DeleteCraftingMethod ---
func (h *CraftingMethodHandler) DeleteCraftingMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	methodIDStr := chi.URLParam(r, "methodID")
	methodID, err := strconv.ParseUint(methodIDStr, 10, 64)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid method ID format", err)
		return
	}

	err = h.craftingMethodService.DeleteCraftingMethod(ctx, methodID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			respondWithError(w, r, http.StatusNotFound, "Crafting method not found", err)
		} else {
			// Handle other potential errors (e.g., FK constraints if not CASCADE)
			respondWithError(w, r, http.StatusInternalServerError, "Failed to delete crafting method", err)
		}
		return
	}

	// Successful deletion
	w.WriteHeader(http.StatusNoContent)
}
