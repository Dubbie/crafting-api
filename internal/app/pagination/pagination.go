package pagination

import (
	"fmt"
	"math"
	"net/url"
	"strconv"

	"github.com/gorilla/schema"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 10
	MaxPerPage     = 100
)

// BaseListParams contains common pagination and sorting parameters.
type BaseListParams struct {
	Page    int    `schema:"page"`
	PerPage int    `schema:"per_page"`
	Sort    string `schema:"sort"`
}

// ListParams embeds BaseListParams and adds specific filters.
type ListParams[F any] struct {
	BaseListParams
	Filters F `schema:",inline"` // Embed filters directly
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
	if page <= 0 {
		page = DefaultPage
	}

	if perPage <= 0 || perPage > MaxPerPage {
		perPage = DefaultPerPage
	}

	lastPage := 0
	if total > 0 && perPage > 0 {
		lastPage = int(math.Ceil(float64(total) / float64(perPage)))
	} else {
		if page == 1 {
			lastPage = 1
		}
	}

	if page > lastPage && lastPage > 0 {
		page = lastPage
	}

	from, to := 0, 0
	if total > 0 && len(data) > 0 {
		from = (page-1)*perPage + 1
		to = from + len(data) - 1
	}

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

// ParseBaseListParams extracts and validates common pagination/sorting params.
func ParseBaseListParams(queryParams url.Values) BaseListParams {
	params := BaseListParams{
		Page:    DefaultPage,
		PerPage: DefaultPerPage,
		Sort:    "", // Default sort can be handled by service/storage layer
	}

	_ = decoder.Decode(&params, queryParams)

	// Manual parsing & validation
	if pageStr := queryParams.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			params.Page = page
		}
	}
	if perPageStr := queryParams.Get("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 {
			params.PerPage = min(perPage, MaxPerPage)
		}
	}
	if sortStr := queryParams.Get("sort"); sortStr != "" {
		params.Sort = sortStr
	}

	return params
}

// ParseListParams extracts base params and specific filters using generics.
// It requires a function to create an empty filter struct.
func ParseListParams[F any](queryParams url.Values) (ListParams[F], error) {
	var params ListParams[F] // Creates params with zero-value F
	params.BaseListParams = ParseBaseListParams(queryParams)

	// Use schema decoder to populate the Filters field (F)
	// Ensure your Filter struct fields have `schema` tags
	err := decoder.Decode(&params.Filters, queryParams)
	if err != nil {
		return params, fmt.Errorf("error decoding query parameters: %w", err)
	}

	return params, nil
}
