/**
 * 功能：null.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package database

import (
	"database/sql/driver"
	"encoding/json"
)

// NullString represents a string that may be NULL in the database.
// It scans NULL into an empty string and serializes as empty string in JSON,
// so the rest of the code never needs to deal with nil pointers.
type NullString string

// Scan implements sql.Scanner. NULL becomes empty string.
func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		*ns = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		// pgx may return []byte
		if b, ok := value.([]byte); ok {
			s = string(b)
		} else {
			s = ""
		}
	}
	*ns = NullString(s)
	return nil
}

// Value implements driver.Valuer for write operations.
func (ns NullString) Value() (driver.Value, error) {
	if ns == "" {
		return nil, nil
	}
	return string(ns), nil
}

// MarshalJSON implements json.Marshaler. Empty string stays empty string (not null).
func (ns NullString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ns))
}

// UnmarshalJSON implements json.Unmarshaler.
func (ns *NullString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*ns = NullString(s)
	return nil
}

// String returns the underlying string value.
func (ns NullString) String() string {
	return string(ns)
}
