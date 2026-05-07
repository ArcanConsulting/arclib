package db

import (
	_ "github.com/go-sql-driver/mysql" // MySQL driver registration
	_ "github.com/lib/pq"              // PostgreSQL driver registration
	_ "github.com/mattn/go-sqlite3"    // SQLite driver registration
)
