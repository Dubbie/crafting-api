package service // Or internal/app/web

import (
	"context" // Import errors
	"net/url"

	"github.com/dubbie/calculator-api/internal/app/pagination"
)

// ListService defines a generic interface for listing resources.
type ListService[T any, F any] interface {
	List(ctx context.Context, params pagination.ListParams[F]) (pagination.PaginatedResponse[T], error)
}

// FilterParser defines a function type for parsing specific filter structs from query params.
type FilterParser[F any] func(values url.Values) (F, error)
