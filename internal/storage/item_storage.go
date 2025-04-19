package storage

import (
	"context"

	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/domain"
)

type ItemStore interface {
	GetItemByID(ctx context.Context, id uint64) (*domain.Item, error)
	ListItems(ctx context.Context, params pagination.ListParams[domain.ItemFilters]) ([]domain.Item, int64, error)
}
