// Package db provides a thin database abstraction layer for the Arc ecosystem.
//
// It wraps database/sql with convenience methods for connection management,
// transaction handling, query building, schema migrations, and NULL helpers.
//
// Supported drivers:
//   - PostgreSQL (lib/pq)
//   - SQLite (go-sqlite3)
//   - MySQL (go-sql-driver/mysql)
//
// Usage:
//
//	cfg := db.Config{
//	    Driver: "sqlite3",
//	    DSN:    ":memory:",
//	}
//	conn, err := db.Open(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer conn.Close()
package db
