package service

import (
	"context"

	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/domain"
)

// CreateItemRequest defines the payload for creating a new item.
type CreateItemRequest struct {
	Name          string                `json:"name" validate:"required,min=2,max=255"` // Required, length constraints
	IsRawMaterial bool                  `json:"is_raw_material"`                        // No specific tag needed unless required=true
	Description   domain.JSONNullString `json:"description"`                            // Validation on NullString needs custom validator or check Valid flag
	ImageURL      domain.JSONNullString `json:"image_url" validate:"omitempty,url"`     // Optional, URL if present
}

// UpdateItemRequest defines the payload for updating an existing item.
type UpdateItemRequest struct {
	Name          *string               `json:"name" validate:"omitempty,min=2,max=255"` // Optional, but length constraints if present
	IsRawMaterial *bool                 `json:"is_raw_material"`                         // Optional
	Description   domain.JSONNullString `json:"description"`                             // Handled by NullString
	ImageURL      domain.JSONNullString `json:"image_url" validate:"omitempty,url"`      // Optional, URL if present
}

// ItemService defines the interface for item-related business logic.
type ItemService interface {
	CreateItem(ctx context.Context, req CreateItemRequest) (*domain.Item, error)
	GetItemByID(ctx context.Context, id uint64) (*domain.Item, error)
	UpdateItem(ctx context.Context, id uint64, req UpdateItemRequest) (*domain.Item, error)
	DeleteItem(ctx context.Context, id uint64) error
	ListItems(ctx context.Context, params pagination.ListParams[domain.ItemFilters]) (pagination.PaginatedResponse[domain.Item], error)
}
