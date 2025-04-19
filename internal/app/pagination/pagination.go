package pagination

import (
	// Import errors
	"fmt"
	"math"
	"net/url"

	"github.com/gorilla/schema"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 15
	MaxPerPage     = 100
)

// BaseListParams contains common pagination and sorting parameters.
type BaseListParams struct {
	Page    int    `schema:"page"`
	PerPage int    `schema:"per_page"`
	Sort    string `schema:"sort"` // e.g., "name_asc", "created_at_desc"
}

// ListParams embeds BaseListParams and adds specific filters.
type ListParams[F any] struct {
	BaseListParams   // Embed directly
	Filters        F `schema:",inline"` // Embed filters directly
}

// PaginatedResponse defines the standard structure for paginated list responses.
type PaginatedResponse[T any] struct {
	Total       int64 `json:"total"`
	PerPage     int   `json:"per_page"`
	CurrentPage int   `json:"current_page"`
	LastPage    int   `json:"last_page"`
	From        int   `json:"from"`
	To          int   `json:"to"`
	Data        []T   `json:"data"`
}

// NewPaginatedResponse creates a PaginatedResponse instance.
func NewPaginatedResponse[T any](data []T, total int64, page, perPage int) PaginatedResponse[T] {
	// Keep existing NewPaginatedResponse logic
	if page <= 0 {
		page = DefaultPage
	}
	if perPage <= 0 {
		perPage = DefaultPerPage
	} // Apply DefaultPerPage here first
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	} // Then clamp to MaxPerPage

	lastPage := 0
	if total > 0 && perPage > 0 {
		lastPage = int(math.Ceil(float64(total) / float64(perPage)))
	} else if page == 1 { // Handle empty results on page 1
		lastPage = 1
	}
	// Don't let page exceed last page if results exist
	if page > lastPage && lastPage > 0 {
		page = lastPage
	}

	from := 0
	to := 0
	if total > 0 && len(data) > 0 {
		from = (page-1)*perPage + 1
		to = from + len(data) - 1
	} // If total is 0, from/to remain 0

	return PaginatedResponse[T]{
		Total:       total,
		PerPage:     perPage,
		CurrentPage: page,
		LastPage:    lastPage,
		From:        from,
		To:          to,
		Data:        data,
	}
}

// decoder is used to parse query parameters into structs.
var decoder = schema.NewDecoder()

// ParseListParams extracts base params and specific filters using generics.
// It now handles defaults and validation directly.
func ParseListParams[F any](queryParams url.Values) (ListParams[F], error) {
	// Initialize with defaults BEFORE decoding
	params := ListParams[F]{
		BaseListParams: BaseListParams{
			Page:    DefaultPage,
			PerPage: DefaultPerPage,
			Sort:    "", // Default sort is handled by service/storage layer
		},
		// Filters field F will have its zero value here
	}

	// Single decode call populates everything based on schema tags:
	// BaseListParams fields (Page, PerPage, Sort) AND Filters fields (due to inline tag)
	err := decoder.Decode(&params, queryParams)
	if err != nil {
		return params, fmt.Errorf("error decoding query parameters: %w", err)
	}

	// Apply validation/clamping AFTER decoding query param values
	if params.Page <= 0 {
		params.Page = DefaultPage
	}
	if params.PerPage <= 0 {
		params.PerPage = DefaultPerPage
	} else if params.PerPage > MaxPerPage {
		params.PerPage = MaxPerPage
	}
	// Optional: Add validation for allowed sort values here if desired

	return params, nil
}
