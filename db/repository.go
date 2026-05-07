package db

import (
	"context"
	"fmt"
)

// BaseRepository provides common database operations for a single table.
type BaseRepository struct {
	conn  *Conn
	table string
}

// NewBaseRepository creates a new BaseRepository for the given table.
func NewBaseRepository(conn *Conn, table string) *BaseRepository {
	return &BaseRepository{conn: conn, table: table}
}

// Conn returns the underlying database connection.
func (r *BaseRepository) Conn() *Conn {
	return r.conn
}

// Table returns the table name.
func (r *BaseRepository) Table() string {
	return r.table
}

// ExistsBy checks if at least one row matches the given column value.
func (r *BaseRepository) ExistsBy(ctx context.Context, column string, value any) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ?)", r.table, column) //nolint:gosec // table/column from trusted code
	var exists bool
	err := r.conn.QueryRowContext(ctx, query, value).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("db: exists by %s: %w", column, err)
	}
	return exists, nil
}

// CountBy counts rows matching the given column value.
func (r *BaseRepository) CountBy(ctx context.Context, column string, value any) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", r.table, column) //nolint:gosec // table/column from trusted code
	var count int64
	err := r.conn.QueryRowContext(ctx, query, value).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db: count by %s: %w", column, err)
	}
	return count, nil
}

// DeleteBy deletes rows matching the given column value and returns
// the number of rows affected.
func (r *BaseRepository) DeleteBy(ctx context.Context, column string, value any) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", r.table, column) //nolint:gosec // table/column from trusted code
	result, err := r.conn.ExecContext(ctx, query, value)
	if err != nil {
		return 0, fmt.Errorf("db: delete by %s: %w", column, err)
	}
	return result.RowsAffected()
}
