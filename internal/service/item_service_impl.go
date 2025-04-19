package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/domain"
	"github.com/dubbie/calculator-api/internal/storage"
)

// Ensure itemServiceImpl implements ItemService
var _ ItemService = (*itemServiceImpl)(nil)

// Ensure itemServiceImpl implements the generic ListService for items
var _ ListService[domain.Item, domain.ItemFilters] = (*itemServiceImpl)(nil)

type itemServiceImpl struct {
	itemStore storage.ItemStore
	// Add other dependencies like a RecipeStore if needed later
}

// NewItemService creates a new ItemService implementation.
// Dependencies (like ItemStore) are injected via the constructor.
func NewItemService(itemStore storage.ItemStore) ItemService {
	return &itemServiceImpl{
		itemStore: itemStore,
	}
}

// GetItemByID retrieves an item using the storage layer.
func (s *itemServiceImpl) GetItemByID(ctx context.Context, id uint64) (*domain.Item, error) {
	item, err := s.itemStore.GetItemByID(ctx, id)
	if err != nil {
		// Map storage errors to service-level errors if needed, or just wrap
		if errors.Is(err, storage.ErrNotFound) {
			// Consider defining service-level errors too, e.g., service.ErrItemNotFound
			return nil, fmt.Errorf("item with id %d not found: %w", id, err)
		}
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	// Add any business logic here if needed (e.g., checking permissions)
	return item, nil
}

// ListItems retrieves a paginated list of items using the storage layer
// and constructs the PaginatedResponse.
func (s *itemServiceImpl) ListItems(ctx context.Context, params pagination.ListParams[domain.ItemFilters]) (pagination.PaginatedResponse[domain.Item], error) {
	// Add any service-level validation or default setting for params if needed
	// e.g., sanitize sort parameters, enforce max per_page again

	items, total, err := s.itemStore.ListItems(ctx, params)
	if err != nil {
		// Wrap error for context
		return pagination.PaginatedResponse[domain.Item]{}, fmt.Errorf("failed to list items: %w", err)
	}

	// Construct the paginated response using the generic helper
	response := pagination.NewPaginatedResponse(items, total, params.Page, params.PerPage)

	return response, nil
}

// Implement the List method for the generic ListService interface
// This simply delegates to ListItems for this specific implementation.
func (s *itemServiceImpl) List(ctx context.Context, params pagination.ListParams[domain.ItemFilters]) (pagination.PaginatedResponse[domain.Item], error) {
	return s.ListItems(ctx, params)
}
