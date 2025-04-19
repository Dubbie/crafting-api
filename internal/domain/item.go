package domain

import (
	"database/sql"
	"time"
)

// Item represents an item in the game.
type Item struct {
	ID            uint64         `db:"id" json:"id"`
	Name          string         `db:"name" json:"name"`
	Slug          string         `db:"slug" json:"slug"`
	IsRawMaterial bool           `db:"is_raw_material" json:"is_raw_material"`
	Description   sql.NullString `db:"description" json:"description,omitempty"`
	ImageURL      sql.NullString `db:"image_url" json:"image_url,omitempty"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updated_at"`
}

// ItemFilters define parameters for listing items.
type ItemFilters struct {
	Name          *string `schema:"name"` // Pointer allows checking if filter was provided
	IsRawMaterial *bool   `schema:"is_raw_material"`
}
