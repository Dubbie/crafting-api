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

var _ CraftingMethodService = (*craftingMethodServiceImpl)(nil)
var _ ListService[domain.CraftingMethod, domain.CraftingMethodFilters] = (*craftingMethodServiceImpl)(nil)

type craftingMethodServiceImpl struct {
	craftingMethodStore storage.CraftingMethodStore
}

func NewCraftingMethodService(craftingMethodStore storage.CraftingMethodStore) CraftingMethodService {
	return &craftingMethodServiceImpl{
		craftingMethodStore: craftingMethodStore,
	}
}

// --- Slug generation helper ---
var (
	widgetNonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)
	widgetWhitespaceRegex      = regexp.MustCompile(`\s+`)
)

func generateCraftingMethodSlug(name string) string {
	slug := strings.ToLower(name)
	slug = widgetWhitespaceRegex.ReplaceAllString(slug, "-")
	slug = widgetNonAlphanumericRegex.ReplaceAllString(slug, "")
	slug = strings.Trim(slug, "-")
	// Add uniqueness check/suffix if needed by querying storage
	return slug
}

// CreateCraftingMethod
func (s *craftingMethodServiceImpl) CreateCraftingMethod(
	ctx context.Context,
	req CreateCraftingMethodRequest,
) (*domain.CraftingMethod, error) {
	slug := generateCraftingMethodSlug(req.Name)

	// Map request to domain model
	newMethod := &domain.CraftingMethod{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
	}
	err := s.craftingMethodStore.CreateCraftingMethod(ctx, newMethod)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicateEntry) {
			return nil, fmt.Errorf("failed to create crafting method: %w", err)
		}
		return nil, fmt.Errorf("failed to store new crafting method: %w", err)
	}

	// The newMethod struct should have its ID populated by the CreateCraftingMethod store method
	if newMethod.ID == 0 {
		// This would indicate an issue with the storage layer implementation
		return nil, errors.New("failed to retrieve ID after crafting method creation")
	}

	// We need CreatedAt/UpdatedAt which were set by DB, fetch the full item
	// Alternatively, the storage CreateItem could return these.
	createdItem, err := s.craftingMethodStore.GetCraftingMethodByID(ctx, newMethod.ID)
	if err != nil {
		fmt.Printf("WARNING: Failed to fetch crafting method %d immediately after creation: %v\n", newMethod.ID, err)
		return newMethod, nil
	}

	return createdItem, nil
}

// UpdateCraftingMethod
func (s *craftingMethodServiceImpl) UpdateCraftingMethod(
	ctx context.Context,
	id uint64,
	req UpdateCraftingMethodRequest,
) (*domain.CraftingMethod, error) {
	// 1. Get the existing crafting method
	existingMethod, err := s.craftingMethodStore.GetCraftingMethodByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crafting method %d: %w", id, err)
	}

	// 2. Merge changes from request into existing crafting method
	updated := false
	if req.Name != nil && *req.Name != existingMethod.Name {
		existingMethod.Name = *req.Name
		existingMethod.Slug = generateSlug(existingMethod.Name)
		updated = true
	}
	if req.Description != existingMethod.Description {
		existingMethod.Description = req.Description
		updated = true
	}

	if !updated {
		return existingMethod, nil
	}

	// 3. Store the updated item
	err = s.craftingMethodStore.UpdateCraftingMethod(ctx, existingMethod)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicateEntry) {
			return nil, fmt.Errorf("failed to update crafting method: %w", err)
		}
		if errors.Is(err, storage.ErrNotFound) {
			return nil, fmt.Errorf("failed to update crafting method, inconsistency detected: %w", err)
		}
		return nil, fmt.Errorf("failed to store updated crafting method: %w", err)
	}

	// Fetch again to get db generated timestamps
	updatedMethod, fetchErr := s.craftingMethodStore.GetCraftingMethodByID(ctx, id)
	if fetchErr != nil {
		fmt.Printf("WARNING: Failed to fetch crafting method %d immediately after update: %v\n", id, fetchErr)
		return existingMethod, nil
	}

	return updatedMethod, nil
}

// --- DeleteCraftingMethod ---
func (s *craftingMethodServiceImpl) DeleteCraftingMethod(ctx context.Context, id uint64) error {
	err := s.craftingMethodStore.DeleteCraftingMethod(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("cannot delete crafting method: %w", err)
		}
		return fmt.Errorf("failed to delete crafting method: %w", err)
	}
	return nil
}

// GetCraftingMethodByID retrieves a crafting method using the storage layer.
func (s *craftingMethodServiceImpl) GetCraftingMethodByID(
	ctx context.Context,
	id uint64,
) (*domain.CraftingMethod, error) {
	method, err := s.craftingMethodStore.GetCraftingMethodByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, fmt.Errorf("crafting method with id %d not found: %w", id, err)
		}
		return nil, fmt.Errorf("failed to get crafting method: %w", err)
	}
	return method, nil
}

// ListCraftingMethods retrieves a paginated list of crafting methods using the storage layer
// and constructs the PaginatedResponse.
func (s *craftingMethodServiceImpl) ListCraftingMethods(
	ctx context.Context,
	params pagination.ListParams[domain.CraftingMethodFilters],
) (pagination.PaginatedResponse[domain.CraftingMethod], error) {
	// Add any service-level validation or default setting for params if needed
	// e.g., sanitize sort parameters, enforce max per_page again

	methods, total, err := s.craftingMethodStore.ListCraftingMethods(ctx, params)
	if err != nil {
		// Wrap error for context
		return pagination.PaginatedResponse[domain.CraftingMethod]{}, fmt.Errorf("failed to list crafting methods: %w", err)
	}

	// Construct the paginated response using the generic helper
	response := pagination.NewPaginatedResponse(methods, total, params.Page, params.PerPage)

	return response, nil
}

func (s *craftingMethodServiceImpl) List(ctx context.Context, params pagination.ListParams[domain.CraftingMethodFilters]) (pagination.PaginatedResponse[domain.CraftingMethod], error) {
	return s.ListCraftingMethods(ctx, params)
}
