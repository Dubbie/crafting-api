package domain

import (
	"database/sql"
	"encoding/json"
	"errors"
)

// JSONNullString wraps sql.NullString to customize JSON marshaling.
type JSONNullString struct {
	sql.NullString
}

// MarshalJSON implements the json.Marshaler interface.
// It marshals the String value if Valid is true, otherwise marshals null.
func (ns JSONNullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		// If it's not valid, marshal it as JSON null.
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It unmarshals the JSON value into the String field if it's a valid string, otherwise sets Valid to false.
func (ns *JSONNullString) UnmarshalJSON(data []byte) error {
	// Check if the input is JSON null
	if string(data) == "null" {
		ns.Valid = false
		ns.String = ""
		return nil
	}

	// Try to unmarshal as a regular string
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		// If it's not null and not a valid JSON string, return an error
		return errors.New("JSONNullString: value must be a string or null")
	}

	ns.String = str
	ns.Valid = true
	return nil
}
