package db

import "database/sql"

// NullString converts a string to sql.NullString.
// An empty string is treated as NULL.
func NullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

// StringPtr extracts a string pointer from sql.NullString.
// Returns nil if the value is NULL.
func StringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// NullInt64 converts an int64 to sql.NullInt64.
// Zero is treated as a valid value; use NullInt64Ptr for zero-as-null.
func NullInt64(n int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: n,
		Valid: true,
	}
}

// NullInt64Ptr converts an *int64 to sql.NullInt64.
// A nil pointer is treated as NULL.
func NullInt64Ptr(n *int64) sql.NullInt64 {
	if n == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Int64: *n,
		Valid: true,
	}
}

// Int64Ptr extracts an int64 pointer from sql.NullInt64.
// Returns nil if the value is NULL.
func Int64Ptr(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

// NullFloat64 converts a float64 to sql.NullFloat64.
func NullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: f,
		Valid:   true,
	}
}

// Float64Ptr extracts a float64 pointer from sql.NullFloat64.
// Returns nil if the value is NULL.
func Float64Ptr(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

// NullBool converts a bool to sql.NullBool.
func NullBool(b bool) sql.NullBool {
	return sql.NullBool{
		Bool:  b,
		Valid: true,
	}
}

// BoolPtr extracts a bool pointer from sql.NullBool.
// Returns nil if the value is NULL.
func BoolPtr(nb sql.NullBool) *bool {
	if !nb.Valid {
		return nil
	}
	return &nb.Bool
}
