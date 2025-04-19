package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

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

// --- Slug Generation Helper ---
// Maybe put this in a separate file or package if it's used elsewhere
var (
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)
	whitespaceRegex      = regexp.MustCompile(`\s+`)
)

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = whitespaceRegex.ReplaceAllString(slug, "-")
	slug = nonAlphanumericRegex.ReplaceAllString(slug, "")
	slug = strings.Trim(slug, "-")
	return slug
}

// --- CreateItem ---
func (s *itemServiceImpl) CreateItem(
	ctx context.Context,
	req CreateItemRequest,
) (*domain.Item, error) {
	slug := generateSlug(req.Name)

	// Map request to domain model
	newItem := &domain.Item{
		Name:          req.Name,
		Slug:          slug,
		IsRawMaterial: req.IsRawMaterial,
		Description:   req.Description,
		ImageURL:      req.ImageURL,
	}

	err := s.itemStore.CreateItem(ctx, newItem)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicateEntry) {
			return nil, fmt.Errorf("failed to create item: %w", err)
		}
		return nil, fmt.Errorf("failed to store new item: %w", err)
	}

	// The newItem struct should have its ID populated by the CreateItem store method
	if newItem.ID == 0 {
		// This would indicate an issue with the storage layer implementation
		return nil, errors.New("failed to retrieve ID after item creation")
	}

	// We need CreatedAt/UpdatedAt which were set by DB, fetch the full item
	// Alternatively, the storage CreateItem could return these.
	createdItem, err := s.itemStore.GetItemByID(ctx, newItem.ID)
	if err != nil {
		// Log this inconsistency but maybe return the newItem with ID anyway? Or fail?
		fmt.Printf("WARNING: Failed to fetch item %d immediately after creation: %v\n", newItem.ID, err)
		// Let's return what we have, the ID is the most critical part populated.
		// The caller might make a separate GET request if they need fresh timestamps immediately.
		return newItem, nil
	}

	return createdItem, nil
}

// --- UpdateItem ---
func (s *itemServiceImpl) UpdateItem(
	ctx context.Context,
	id uint64,
	req UpdateItemRequest,
) (*domain.Item, error) {
	// TODO: Add validation for the request struct `req`

	// 1. Get the existing item
	existingItem, err := s.itemStore.GetItemByID(ctx, id)
	if err != nil {
		// Handles ErrNotFound already
		return nil, fmt.Errorf("cannot update item: %w", err)
	}

	// 2. Merge changes from request into existing item
	updated := false
	if req.Name != nil && *req.Name != existingItem.Name {
		existingItem.Name = *req.Name
		existingItem.Slug = generateSlug(existingItem.Name) // Regenerate slug if name changes
		updated = true
	}
	if req.IsRawMaterial != nil && *req.IsRawMaterial != existingItem.IsRawMaterial {
		existingItem.IsRawMaterial = *req.IsRawMaterial
		updated = true
	}
	// For sql.NullString, check if the request field itself is different *or* its validity changes
	if req.Description != existingItem.Description {
		existingItem.Description = req.Description
		updated = true
	}
	if req.ImageURL != existingItem.ImageURL {
		existingItem.ImageURL = req.ImageURL
		updated = true
	}

	// Only call update if something actually changed
	if !updated {
		return existingItem, nil // No changes, return existing item
	}

	// UpdatedAt is set by the storage layer or DB trigger

	// 3. Store the updated item
	err = s.itemStore.UpdateItem(ctx, existingItem)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicateEntry) {
			// Possible if slug/name changed to an existing one
			return nil, fmt.Errorf("failed to update item: %w", err)
		}
		if errors.Is(err, storage.ErrNotFound) {
			// Should not happen if GetByID succeeded, but possible race condition or DB issue
			return nil, fmt.Errorf("failed to update item, inconsistency detected: %w", err)
		}
		return nil, fmt.Errorf("failed to store updated item: %w", err)
	}

	// Fetch again to get DB-generated UpdatedAt timestamp? Or assume store updated it?
	// Let's fetch again for consistency, like in Create.
	updatedItem, fetchErr := s.itemStore.GetItemByID(ctx, id)
	if fetchErr != nil {
		fmt.Printf("WARNING: Failed to fetch item %d immediately after update: %v\n", id, fetchErr)
		// Return the item as it was before the failed fetch
		return existingItem, nil
	}

	return updatedItem, nil
}

// --- DeleteItem ---
func (s *itemServiceImpl) DeleteItem(ctx context.Context, id uint64) error {
	err := s.itemStore.DeleteItem(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("cannot delete item: %w", err) // Wrap ErrNotFound
		}
		return fmt.Errorf("failed to delete item: %w", err) // Wrap other errors
	}
	return nil
}

// GetItemByID retrieves an item using the storage layer.
func (s *itemServiceImpl) GetItemByID(
	ctx context.Context,
	id uint64,
) (*domain.Item, error) {
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
func (s *itemServiceImpl) ListItems(
	ctx context.Context,
	params pagination.ListParams[domain.ItemFilters],
) (pagination.PaginatedResponse[domain.Item], error) {
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

// List (Generic Interface)
func (s *itemServiceImpl) List(ctx context.Context, params pagination.ListParams[domain.ItemFilters]) (pagination.PaginatedResponse[domain.Item], error) {
	// Keep existing implementation
	return s.ListItems(ctx, params)
}
