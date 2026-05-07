package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Config holds database connection parameters.
type Config struct {
	// Driver is the database driver name: "postgres", "sqlite3", or "mysql".
	Driver string

	// DSN is the data source name (connection string).
	DSN string

	// MaxOpenConns sets the maximum number of open connections.
	// Zero means unlimited.
	MaxOpenConns int

	// MaxIdleConns sets the maximum number of idle connections.
	// Zero means no idle connections are retained.
	MaxIdleConns int

	// ConnMaxLifetimeSeconds sets the maximum lifetime of a connection in seconds.
	// Zero means connections are not closed due to age.
	ConnMaxLifetimeSeconds int
}

// Conn wraps a *sql.DB with driver detection and convenience methods.
type Conn struct {
	db     *sql.DB
	driver string
}

// Open creates a new database connection using the provided configuration.
func Open(cfg Config) (*Conn, error) {
	sqlDB, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetimeSeconds > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSeconds) * time.Second)
	}

	return &Conn{db: sqlDB, driver: cfg.Driver}, nil
}

// Close closes the underlying database connection.
func (c *Conn) Close() error {
	return c.db.Close()
}

// Ping verifies the database connection is alive.
func (c *Conn) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// DB returns the underlying *sql.DB for advanced usage.
func (c *Conn) DB() *sql.DB {
	return c.db
}

// Driver returns the driver name used to open this connection.
func (c *Conn) Driver() string {
	return c.driver
}

// IsPostgres returns true if the connection uses the PostgreSQL driver.
func (c *Conn) IsPostgres() bool {
	return c.driver == "postgres"
}

// IsSQLite returns true if the connection uses the SQLite driver.
func (c *Conn) IsSQLite() bool {
	return c.driver == "sqlite3"
}

// IsMySQL returns true if the connection uses the MySQL driver.
func (c *Conn) IsMySQL() bool {
	return c.driver == "mysql"
}

// Querier abstracts query execution for both *sql.DB and *sql.Tx.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// ExecContext executes a query that doesn't return rows.
func (c *Conn) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows.
func (c *Conn) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns a single row.
func (c *Conn) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a new transaction with the given options.
func (c *Conn) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return c.db.BeginTx(ctx, opts)
}

// WithTx executes fn within a transaction. If fn returns an error, the
// transaction is rolled back. Otherwise, it is committed.
func (c *Conn) WithTx(ctx context.Context, opts *sql.TxOptions, fn func(tx *sql.Tx) error) error {
	tx, err := c.db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("db: begin tx: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("db: rollback failed: %w (original: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db: commit: %w", err)
	}
	return nil
}
