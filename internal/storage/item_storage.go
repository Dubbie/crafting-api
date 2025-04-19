package storage

import (
	"context"

	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/domain"
)

// ItemStore defines the interface for data storage operations.
type ItemStore interface {
	CreateItem(ctx context.Context, item *domain.Item) error
	GetItemByID(ctx context.Context, id uint64) (*domain.Item, error)
	UpdateItem(ctx context.Context, item *domain.Item) error
	DeleteItem(ctx context.Context, id uint64) error
	ListItems(ctx context.Context, params pagination.ListParams[domain.ItemFilters]) ([]domain.Item, int64, error)
}
