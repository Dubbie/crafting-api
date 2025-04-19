package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/service"
	"github.com/dubbie/calculator-api/internal/storage"
)

// MakeListHandler creates a generic http.HandlerFunc for listing resources.
func MakeListHandler[T any, F any](lister service.ListService[T, F]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queryParams := r.URL.Query()

		params, err := pagination.ParseListParams[F](queryParams)
		if err != nil {
			// Now we can call respondWithError directly (or via handler. prefix if needed)
			respondWithError(w, r, http.StatusBadRequest, "Invalid query parameters", err)
			return
		}

		response, err := lister.List(ctx, params)
		if err != nil {
			// Map errors and respond using the helper
			statusCode := http.StatusInternalServerError
			message := "Failed to list resources"
			if errors.Is(err, storage.ErrNotFound) {
				statusCode = http.StatusNotFound
				message = "Resource not found"
			}

			respondWithError(w, r, statusCode, message, err)
			return
		}

		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			// Log the encoding error using respondWithError (status already sent)
			respondWithError(w, r, http.StatusInternalServerError, "Failed to encode successful response", err)
		}
	}
}
