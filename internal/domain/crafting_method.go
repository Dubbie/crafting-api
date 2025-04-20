package domain

import "time"

// CraftingMethod represents a crafting method.
type CraftingMethod struct {
	ID          uint64         `db:"id" json:"id"`
	Name        string         `db:"name" json:"name"`
	Slug        string         `db:"slug" json:"slug"`
	Description JSONNullString `db:"description" json:"description"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

// CraftingMethodFilters define parameters for listing crafting methods.
type CraftingMethodFilters struct {
	Name *string `schema:"name"` // Pointer allows checking if filter was provided
}
