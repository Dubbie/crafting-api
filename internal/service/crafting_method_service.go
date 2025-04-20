package service

import (
	"context"

	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/domain"
)

type CreateCraftingMethodRequest struct {
	Name        string                `json:"name" validate:"required"`
	Description domain.JSONNullString `json:"description"`
}

type UpdateCraftingMethodRequest struct {
	Name        *string               `json:"name" validate:"required"`
	Description domain.JSONNullString `json:"description"`
}

type CraftingMethodService interface {
	CreateCraftingMethod(
		ctx context.Context,
		req CreateCraftingMethodRequest,
	) (*domain.CraftingMethod, error)

	UpdateCraftingMethod(
		ctx context.Context,
		id uint64, req UpdateCraftingMethodRequest,
	) (*domain.CraftingMethod, error)

	DeleteCraftingMethod(ctx context.Context, id uint64) error

	GetCraftingMethodByID(ctx context.Context, id uint64) (*domain.CraftingMethod, error)

	ListCraftingMethods(
		ctx context.Context,
		params pagination.ListParams[domain.CraftingMethodFilters],
	) (pagination.PaginatedResponse[domain.CraftingMethod], error)
}
