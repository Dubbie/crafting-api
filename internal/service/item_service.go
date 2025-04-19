package service

import (
	"context"
	"database/sql"

	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/domain"
)

// CreateItemRequest defines the payload for creating a new item.
// We don't include ID, Slug, CreatedAt, UpdatedAt as they are generated/set by the system.
type CreateItemRequest struct {
	Name          string         `json:"name" validate:"required,min=2,max=255"` // Example validation tags
	IsRawMaterial bool           `json:"is_raw_material"`                        // Use value type for boolean
	Description   sql.NullString `json:"description"`                            // Use NullString for nullable fields
	ImageURL      sql.NullString `json:"image_url" validate:"omitempty,url"`     // Example validation
}

// UpdateItemRequest defines the payload for updating an existing item.
// Use pointers for fields that are optional to update.
// This allows distinguishing between providing an empty value ("") vs. not providing the field at all.
type UpdateItemRequest struct {
	Name          *string        `json:"name" validate:"omitempty,min=2,max=255"` // Pointer, omitempty if not provided
	IsRawMaterial *bool          `json:"is_raw_material"`                         // Pointer
	Description   sql.NullString `json:"description"`                             // NullString handles nullability
	ImageURL      sql.NullString `json:"image_url" validate:"omitempty,url"`      // NullString handles nullability
}

// ItemService defines the interface for item-related business logic.
type ItemService interface {
	CreateItem(ctx context.Context, req CreateItemRequest) (*domain.Item, error)
	GetItemByID(ctx context.Context, id uint64) (*domain.Item, error)
	UpdateItem(ctx context.Context, id uint64, req UpdateItemRequest) (*domain.Item, error)
	DeleteItem(ctx context.Context, id uint64) error
	ListItems(ctx context.Context, params pagination.ListParams[domain.ItemFilters]) (pagination.PaginatedResponse[domain.Item], error)
}
