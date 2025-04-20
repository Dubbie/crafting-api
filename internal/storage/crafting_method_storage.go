package storage

import (
	"context"

	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/domain"
)

// CraftingMethodStore defines the interface for data storage operations.
type CraftingMethodStore interface {
	CreateCraftingMethod(ctx context.Context, method *domain.CraftingMethod) error
	GetCraftingMethodByID(ctx context.Context, id uint64) (*domain.CraftingMethod, error)
	UpdateCraftingMethod(ctx context.Context, method *domain.CraftingMethod) error
	DeleteCraftingMethod(ctx context.Context, id uint64) error
	ListCraftingMethods(ctx context.Context, params pagination.ListParams[domain.CraftingMethodFilters]) ([]domain.CraftingMethod, int64, error)
}
