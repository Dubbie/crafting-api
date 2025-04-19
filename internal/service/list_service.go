package service // Or internal/app/web

import (
	"context"
	"encoding/json"
	"errors" // Import errors
	"fmt"
	"net/http"
	"net/url"

	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/storage"
)

// ListService defines a generic interface for listing resources.
type ListService[T any, F any] interface {
	List(ctx context.Context, params pagination.ListParams[F]) (pagination.PaginatedResponse[T], error)
}

// FilterParser defines a function type for parsing specific filter structs from query params.
type FilterParser[F any] func(values url.Values) (F, error)

// MakeListHandler creates a generic http.HandlerFunc for listing resources.
// It takes a ListService implementation and a FilterParser function.
func MakeListHandler[T any, F any](
	lister ListService[T, F],
	// filterParser FilterParser[F], // We can use ParseListParams directly if F has schema tags
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queryParams := r.URL.Query()

		// Use the generic ParseListParams function
		params, err := pagination.ParseListParams[F](queryParams)
		if err != nil {
			// Handle parsing errors (e.g., bad query param format for filters)
			http.Error(w, fmt.Sprintf("Bad Request: Invalid query parameters: %v", err), http.StatusBadRequest)
			return
		}

		// Call the specific List method of the injected service
		response, err := lister.List(ctx, params)
		if err != nil {
			// Basic error handling: Map known errors to status codes
			// Production apps often have more sophisticated error handling/logging
			statusCode := http.StatusInternalServerError
			errMsg := "Internal Server Error"

			// Check for specific errors if needed (e.g., storage.ErrNotFound - unlikely for List but possible)
			if errors.Is(err, storage.ErrNotFound) { // Example
				statusCode = http.StatusNotFound
				errMsg = err.Error()
			} else {
				// Log the detailed error for debugging
				fmt.Printf("Error in List handler: %v\n", err) // Replace with proper logging
			}

			http.Error(w, errMsg, statusCode)
			return
		}

		// Send the successful JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			// Log error encoding response
			fmt.Printf("Error encoding JSON response: %v\n", err)
		}
	}
}
