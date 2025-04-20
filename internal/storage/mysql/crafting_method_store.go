package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/dubbie/calculator-api/internal/app/pagination"
	"github.com/dubbie/calculator-api/internal/domain"
	"github.com/dubbie/calculator-api/internal/storage"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var _ storage.CraftingMethodStore = (*mysqlCraftingMethodStore)(nil)

type mysqlCraftingMethodStore struct {
	db *sqlx.DB
}

func NewMySQLCraftingMethodStore(db *sqlx.DB) *mysqlCraftingMethodStore {
	if db == nil {
		panic("sqlx.DB instance is required")
	}
	return &mysqlCraftingMethodStore{db: db}
}

func (s *mysqlCraftingMethodStore) CreateCraftingMethod(
	ctx context.Context,
	craftingMethod *domain.CraftingMethod,
) error {
	now := time.Now()
	craftingMethod.CreatedAt = now
	craftingMethod.UpdatedAt = now

	query := `
        INSERT INTO crafting_methods (name, slug, description, created_at, updated_at)
        VALUES (:name, :slug, :description, :created_at, :updated_at);
	`

	res, err := s.db.NamedExecContext(ctx, query, craftingMethod)
	if err != nil {
		// Debug the crafting method
		fmt.Printf("Crafting Method: %+v\n", craftingMethod)
		// Check for duplicate entry (MySQL specific error number 1062)
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return fmt.Errorf("crafting method creation failed: %w: %s", storage.ErrDuplicateEntry, err.Error())
		}
		return fmt.Errorf("error creating crafting method: %w", err)
	}

	// Get the ID of the newly created item
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID after creating crafting method: %w", err)
	}
	craftingMethod.ID = uint64(id)

	return nil
}

// GetCraftingMethodByID retrieves a crafting method by its ID.
func (s *mysqlCraftingMethodStore) GetCraftingMethodByID(
	ctx context.Context,
	id uint64,
) (*domain.CraftingMethod, error) {
	query := `
        SELECT id, name, slug, description, created_at, updated_at
        FROM crafting_methods
        WHERE id = :id;
	`
	var craftingMethod domain.CraftingMethod

	err := s.db.GetContext(ctx, &craftingMethod, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("error fetching crafting method with id: %w", err)
	}

	return &craftingMethod, nil
}

// UpdateCraftingMethod updates a crafting method.
func (s *mysqlCraftingMethodStore) UpdateCraftingMethod(
	ctx context.Context,
	craftingMethod *domain.CraftingMethod,
) error {
	craftingMethod.UpdatedAt = time.Now()

	query := `
        UPDATE crafting_methods SET
        	name = :name,
         	slug = :slug,
        	description = :description,
        	updated_at = :updated_at
        WHERE id = :id;
	`

	res, err := s.db.NamedExecContext(ctx, query, craftingMethod)
	if err != nil {
		// Check for duplicate entry error (MySQL specific error number 1062)
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("crafting method update failed: %w", err)
		}
		return fmt.Errorf("error updating crafting method: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected after updating crafting method %d: %w", craftingMethod.ID, err)
	}

	if rowsAffected == 0 {
		return storage.ErrNotFound
	}

	return nil
}

// DeleteCraftingMethod deletes a crafting method.
func (s *mysqlCraftingMethodStore) DeleteCraftingMethod(
	ctx context.Context,
	id uint64,
) error {
	query := `
        DELETE FROM crafting_methods
        WHERE id = ?;
	`

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting crafting method with id %d: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected after deleting crafting method: %w", err)
	}

	if rowsAffected == 0 {
		return storage.ErrNotFound
	}

	return nil
}

// ListCraftingMethods retrieves a paginated and filtered list of crafting methods.
func (s *mysqlCraftingMethodStore) ListCraftingMethods(
	ctx context.Context,
	params pagination.ListParams[domain.CraftingMethodFilters],
) ([]domain.CraftingMethod, int64, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	// Base select query for crafting methods
	selectBuilder := psql.Select(
		"id", "name", "slug", "description", "created_at", "updated_at",
	).From("crafting_methods")

	// Base count query
	countBuilder := psql.Select("COUNT(*)").From("crafting_methods")

	// Apply filters
	if params.Filters.Name != nil && *params.Filters.Name != "" {
		// Use LIKE for partial matching, adjust if exact match needed
		namePattern := "%" + *params.Filters.Name + "%"
		selectBuilder = selectBuilder.Where(squirrel.Like{"name": namePattern})
		countBuilder = countBuilder.Where(squirrel.Like{"name": namePattern})
	}

	// Get total count matching filters before applying limit/offset
	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("error building count query for crafting methods: %w", err)
	}

	var total int64
	err = s.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing count query for crafting methods: %w", err)
	}

	if total == 0 {
		return []domain.CraftingMethod{}, 0, nil
	}

	// Apply sorting
	sortField, sortOrder := "created_at", "DESC"
	if params.Sort != "" {
		parts := strings.Split(params.Sort, "_")
		if len(parts) == 2 {
			allowedSortFields := map[string]bool{"name": true, "slug": true, "created_at": true, "updated_at": true}
			if allowedSortFields[parts[0]] {
				sortField = parts[0]
				if strings.ToLower(parts[1]) == "asc" {
					sortOrder = "ASC"
				} else if strings.ToLower(parts[1]) == "desc" {
					sortOrder = "DESC"
				}
			}
		}
	}
	selectBuilder = selectBuilder.OrderBy(fmt.Sprintf("%s %s", sortField, sortOrder))

	// Apply pagination (Limit and Offset)
	offset := uint64((params.Page - 1) * params.PerPage)
	selectBuilder = selectBuilder.Limit(uint64(params.PerPage)).Offset(offset)

	// Build the final select query
	methodsQuery, methodsArgs, err := selectBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("error building select query for crafting methods: %w", err)
	}

	// Execute the query to get the crafting methods for the current page
	craftingMethods := []domain.CraftingMethod{}
	err = s.db.SelectContext(ctx, &craftingMethods, methodsQuery, methodsArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing select query for crafting methods: %w", err)
	}

	return craftingMethods, total, nil
}
