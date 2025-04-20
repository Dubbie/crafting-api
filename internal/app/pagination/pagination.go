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

type BaseListParams struct {
	Page    int    `schema:"page"`
	PerPage int    `schema:"per_page"`
	Sort    string `schema:"sort"`
}

// ListParams embeds BaseListParams and adds specific filters.
type ListParams[F any] struct {
	Page    int
	PerPage int
	Sort    string

	// Filters inside
	Filters F
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

func init() {
	decoder.IgnoreUnknownKeys(true)
}

// ParseListParams extracts base params and specific filters using generics.
// It now handles defaults and validation directly.
func ParseListParams[F any](queryParams url.Values) (ListParams[F], error) {
	var baseParams BaseListParams // Temporary struct for base params
	var filters F                 // Use the generic type F directly for filters

	// --- Decode Base Params ---
	// Set defaults BEFORE decoding this part
	baseParams.Page = DefaultPage
	baseParams.PerPage = DefaultPerPage
	baseParams.Sort = ""

	// Decode query into baseParams. decoder will ignore 'name', 'is_raw_material', etc.
	err := decoder.Decode(&baseParams, queryParams)
	if err != nil {
		return ListParams[F]{}, fmt.Errorf("error decoding base parameters: %w", err)
	}

	// --- Decode Filter Params ---
	err = decoder.Decode(&filters, queryParams)
	if err != nil {
		return ListParams[F]{}, fmt.Errorf("error decoding filter parameters: %w", err)
	}

	// --- Apply validation/clamping to decoded base params ---
	if baseParams.Page <= 0 {
		baseParams.Page = DefaultPage
	}
	if baseParams.PerPage <= 0 {
		baseParams.PerPage = DefaultPerPage
	} else if baseParams.PerPage > MaxPerPage {
		baseParams.PerPage = MaxPerPage
	}

	// --- Combine results manually ---
	finalParams := ListParams[F]{
		Page:    baseParams.Page,
		PerPage: baseParams.PerPage,
		Sort:    baseParams.Sort,
		Filters: filters,
	}

	return finalParams, nil
}
