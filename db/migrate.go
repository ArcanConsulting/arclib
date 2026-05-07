package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// MigrateFunc is a function that performs a migration step.
type MigrateFunc func(ctx context.Context, tx *sql.Tx) error

// Migration represents a single database migration.
type Migration struct {
	// Version is the unique, sortable migration version (e.g., "001", "20240101120000").
	Version string

	// Name is a human-readable description of the migration.
	Name string

	// Up applies the migration.
	Up MigrateFunc

	// Down reverts the migration.
	Down MigrateFunc
}

// MigrationStatus represents the status of a single migration.
type MigrationStatus struct {
	Version   string
	Name      string
	Applied   bool
	AppliedAt time.Time
}

// Migrator manages database schema migrations.
type Migrator struct {
	conn       *Conn
	table      string
	migrations []Migration
	sqlDir     string
}

// NewMigrator creates a new Migrator that tracks applied migrations
// in the given table name.
func NewMigrator(conn *Conn, table string) *Migrator {
	return &Migrator{
		conn:  conn,
		table: table,
	}
}

// SetSQLDir sets the directory to search for SQL migration files.
func (m *Migrator) SetSQLDir(dir string) {
	m.sqlDir = dir
}

// Add registers a migration with the migrator.
func (m *Migrator) Add(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// AddSQL registers a SQL file migration. The files are loaded from the
// configured SQL directory. It looks for dialect-specific files first:
//
//	{version}_{name}.{driver}.sql  (e.g., 001_create_users.sqlite3.sql)
//
// Then falls back to the generic file:
//
//	{version}_{name}.sql           (e.g., 001_create_users.sql)
func (m *Migrator) AddSQL(version, name string) {
	m.migrations = append(m.migrations, Migration{
		Version: version,
		Name:    name,
		Up: func(ctx context.Context, tx *sql.Tx) error {
			content, err := m.loadSQL(version, name, "up")
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx, content)
			return err
		},
		Down: func(ctx context.Context, tx *sql.Tx) error {
			content, err := m.loadSQL(version, name, "down")
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx, content)
			return err
		},
	})
}

// loadSQL finds and reads the SQL migration file for the given direction.
// It tries dialect-specific files first, then falls back to generic ones.
func (m *Migrator) loadSQL(version, name, direction string) (string, error) {
	if m.sqlDir == "" {
		return "", fmt.Errorf("db: migrate: sql directory not set")
	}

	base := fmt.Sprintf("%s_%s.%s", version, name, direction)
	driver := m.conn.Driver()

	// Try dialect-specific file first.
	dialectPath := filepath.Clean(filepath.Join(m.sqlDir, base+"."+driver+".sql"))
	if data, err := os.ReadFile(dialectPath); err == nil { //nolint:gosec // path constructed from trusted migration metadata
		return string(data), nil
	}

	// Fall back to generic file.
	genericPath := filepath.Clean(filepath.Join(m.sqlDir, base+".sql"))
	data, err := os.ReadFile(genericPath) //nolint:gosec // path constructed from trusted migration metadata
	if err != nil {
		return "", fmt.Errorf("db: migrate: read %s: %w", genericPath, err)
	}
	return string(data), nil
}

// ensureTable creates the migrations tracking table if it doesn't exist.
func (m *Migrator) ensureTable(ctx context.Context) error {
	query := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
			version TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`, m.table)
	_, err := m.conn.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("db: migrate: create table: %w", err)
	}
	return nil
}

// appliedVersions returns a set of already-applied migration versions.
func (m *Migrator) appliedVersions(ctx context.Context) (map[string]time.Time, error) {
	query := fmt.Sprintf("SELECT version, applied_at FROM %s", m.table) //nolint:gosec // trusted table name
	rows, err := m.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("db: migrate: query applied: %w", err)
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[string]time.Time)
	for rows.Next() {
		var version string
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, fmt.Errorf("db: migrate: scan: %w", err)
		}
		applied[version] = appliedAt
	}
	return applied, rows.Err()
}

// Up applies all pending migrations in version order.
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	applied, err := m.appliedVersions(ctx)
	if err != nil {
		return err
	}

	sorted := m.sortedMigrations()
	for _, mig := range sorted {
		if _, ok := applied[mig.Version]; ok {
			continue
		}

		if mig.Up == nil {
			return fmt.Errorf("db: migrate: version %s has no Up function", mig.Version)
		}

		if err := m.conn.WithTx(ctx, nil, func(tx *sql.Tx) error {
			if err := mig.Up(ctx, tx); err != nil {
				return fmt.Errorf("db: migrate up %s (%s): %w", mig.Version, mig.Name, err)
			}
			_, err := tx.ExecContext(ctx,
				fmt.Sprintf("INSERT INTO %s (version, name) VALUES (?, ?)", m.table),
				mig.Version, mig.Name)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

// Down reverts the last applied migration.
func (m *Migrator) Down(ctx context.Context) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	applied, err := m.appliedVersions(ctx)
	if err != nil {
		return err
	}

	// Find the last applied migration.
	sorted := m.sortedMigrations()
	var last *Migration
	for i := len(sorted) - 1; i >= 0; i-- {
		if _, ok := applied[sorted[i].Version]; ok {
			last = &sorted[i]
			break
		}
	}

	if last == nil {
		return nil // nothing to revert
	}

	if last.Down == nil {
		return fmt.Errorf("db: migrate: version %s has no Down function", last.Version)
	}

	return m.conn.WithTx(ctx, nil, func(tx *sql.Tx) error {
		if err := last.Down(ctx, tx); err != nil {
			return fmt.Errorf("db: migrate down %s (%s): %w", last.Version, last.Name, err)
		}
		_, err := tx.ExecContext(ctx,
			fmt.Sprintf("DELETE FROM %s WHERE version = ?", m.table),
			last.Version)
		return err
	})
}

// Status returns the status of all registered migrations.
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	applied, err := m.appliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	sorted := m.sortedMigrations()
	statuses := make([]MigrationStatus, len(sorted))
	for i, mig := range sorted {
		s := MigrationStatus{
			Version: mig.Version,
			Name:    mig.Name,
		}
		if t, ok := applied[mig.Version]; ok {
			s.Applied = true
			s.AppliedAt = t
		}
		statuses[i] = s
	}
	return statuses, nil
}

// sortedMigrations returns migrations sorted by version.
func (m *Migrator) sortedMigrations() []Migration {
	sorted := make([]Migration, len(m.migrations))
	copy(sorted, m.migrations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version < sorted[j].Version
	})
	return sorted
}
