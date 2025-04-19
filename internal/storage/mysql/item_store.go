package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/domain"
	"github.com/dubbie/calculator-api/internal/storage"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Ensure mysqlItemStore implements ItemStore interface
var _ storage.ItemStore = (*mysqlItemStore)(nil)

type mysqlItemStore struct {
	db *sqlx.DB
}

func NewMySQLItemStore(db *sqlx.DB) *mysqlItemStore {
	if db == nil {
		panic("sqlx.DB instance is required")
	}
	return &mysqlItemStore{db: db}
}

// CreateItem creates a new item in the database.
func (s *mysqlItemStore) CreateItem(ctx context.Context, item *domain.Item) error {
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	query := `
		INSERT INTO items (name, slug, is_raw_material, description, image_url, created_at, updated_at)
		VALUES (:name, :slug, :is_raw_material, :description, :image_url, :created_at, :updated_at);
	`

	res, err := s.db.NamedExecContext(ctx, query, item)
	if err != nil {
		// Debug the item
		fmt.Printf("Item: %+v\n", item)
		// Check for duplicate entry error (MySQL specific error number 1062)
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return fmt.Errorf("item creation failed: %w: %s", storage.ErrDuplicateEntry, err.Error())
		}
		return fmt.Errorf("error creating item: %w", err)
	}

	// Get the ID of the newly created item
	id, err := res.LastInsertId()
	if err != nil {
		// This is less likely but possible
		return fmt.Errorf("error getting last insert ID after creating item: %w", err)
	}
	item.ID = uint64(id) // Update the item struct with the new ID

	return nil
}

// GetItemByID retrieves a single item by its ID.
func (s *mysqlItemStore) GetItemByID(ctx context.Context, id uint64) (*domain.Item, error) {
	query := "SELECT id, name, slug, is_raw_material, description, image_url, created_at, updated_at FROM items WHERE id = ?"
	var item domain.Item

	err := s.db.GetContext(ctx, &item, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNotFound // Define ErrNotFound in storage package
		}
		// Wrap error for context
		return nil, fmt.Errorf("error fetching item with id %d: %w", id, err)
	}
	return &item, nil
}

// --- UpdateItem ---
func (s *mysqlItemStore) UpdateItem(ctx context.Context, item *domain.Item) error {
	// Update the UpdatedAt timestamp before saving
	item.UpdatedAt = time.Now()

	query := `
        UPDATE items SET
            name = :name,
            slug = :slug,
            is_raw_material = :is_raw_material,
            description = :description,
            image_url = :image_url,
            updated_at = :updated_at
        WHERE id = :id
    `
	res, err := s.db.NamedExecContext(ctx, query, item)
	if err != nil {
		// Check for duplicate entry error on update (e.g., changing name/slug to one that exists)
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return fmt.Errorf("item update failed: %w: %s", storage.ErrDuplicateEntry, err.Error())
		}
		return fmt.Errorf("error updating item with id %d: %w", item.ID, err)
	}

	// Check if any row was actually updated
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// Error getting rows affected, but the query might have succeeded
		return fmt.Errorf("error checking rows affected after updating item %d: %w", item.ID, err)
	}
	if rowsAffected == 0 {
		// No rows updated, likely means the item ID didn't exist
		return storage.ErrNotFound
	}

	return nil
}

// --- DeleteItem ---
func (s *mysqlItemStore) DeleteItem(ctx context.Context, id uint64) error {
	query := "DELETE FROM items WHERE id = ?"
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		// Foreign key constraint errors might occur here if not handled by ON DELETE CASCADE/SET NULL etc.
		return fmt.Errorf("error deleting item with id %d: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected after deleting item %d: %w", id, err)
	}
	if rowsAffected == 0 {
		// No rows deleted, likely means the item ID didn't exist
		return storage.ErrNotFound
	}

	return nil
}

// ListItems retrieves a paginated and filtered list of items.
func (s *mysqlItemStore) ListItems(ctx context.Context, params pagination.ListParams[domain.ItemFilters]) ([]domain.Item, int64, error) {
	// Use squirrel for building the query to handle filters and pagination dynamically
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)

	// Base select query for items
	selectBuilder := psql.Select(
		"id", "name", "slug", "is_raw_material",
		"description", "image_url", "created_at", "updated_at",
	).From("items")

	// Base count query
	countBuilder := psql.Select("COUNT(*)").From("items")

	// Apply filters
	if params.Filters.Name != nil && *params.Filters.Name != "" {
		// Use LIKE for partial matching, adjust if exact match needed
		namePattern := "%" + *params.Filters.Name + "%"
		selectBuilder = selectBuilder.Where(sq.Like{"name": namePattern})
		countBuilder = countBuilder.Where(sq.Like{"name": namePattern})
	}
	if params.Filters.IsRawMaterial != nil {
		selectBuilder = selectBuilder.Where(sq.Eq{"is_raw_material": *params.Filters.IsRawMaterial})
		countBuilder = countBuilder.Where(sq.Eq{"is_raw_material": *params.Filters.IsRawMaterial})
	}
	// Add more filters here...

	// Get total count matching filters *before* applying limit/offset
	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("error building count query for items: %w", err)
	}

	var total int64
	err = s.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing count query for items: %w", err)
	}

	if total == 0 {
		// No need to query for items if count is zero
		return []domain.Item{}, 0, nil
	}

	// Apply sorting
	sortField := "created_at" // Default sort
	sortOrder := "DESC"       // Default order
	if params.Sort != "" {
		parts := strings.Split(params.Sort, "_")
		if len(parts) == 2 {
			// Basic validation: check if field is allowed (e.g., "name", "created_at")
			allowedSortFields := map[string]bool{"name": true, "slug": true, "created_at": true, "updated_at": true}
			if allowedSortFields[parts[0]] {
				sortField = parts[0]
				if strings.ToLower(parts[1]) == "asc" {
					sortOrder = "ASC"
				} else if strings.ToLower(parts[1]) == "desc" {
					sortOrder = "DESC"
				}
				// else stick to default DESC
			}
		}
	}
	selectBuilder = selectBuilder.OrderBy(fmt.Sprintf("%s %s", sortField, sortOrder))

	// Apply pagination (Limit and Offset)
	offset := uint64((params.Page - 1) * params.PerPage)
	selectBuilder = selectBuilder.Limit(uint64(params.PerPage)).Offset(offset)

	// Build the final select query
	itemsQuery, itemsArgs, err := selectBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("error building select query for items: %w", err)
	}

	// Execute the query to get the items for the current page
	items := []domain.Item{}
	err = s.db.SelectContext(ctx, &items, itemsQuery, itemsArgs...)
	if err != nil {
		// No need to check for sql.ErrNoRows here, an empty slice is fine
		return nil, 0, fmt.Errorf("error executing select query for items: %w", err)
	}

	return items, total, nil
}
