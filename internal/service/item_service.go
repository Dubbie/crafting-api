package service

import (
	"context"

	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/domain"
)

// ItemService defines the interface for item-related business logic.
type ItemService interface {
	GetItemByID(ctx context.Context, id uint64) (*domain.Item, error)
	ListItems(ctx context.Context, params pagination.ListParams[domain.ItemFilters]) (pagination.PaginatedResponse[domain.Item], error)
}
