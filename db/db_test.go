package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func openTestDB(t *testing.T) *Conn {
	t.Helper()
	conn, err := Open(Config{
		Driver: "sqlite3",
		DSN:    ":memory:",
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestOpenClose(t *testing.T) {
	conn, err := Open(Config{
		Driver: "sqlite3",
		DSN:    ":memory:",
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestOpenInvalidDriver(t *testing.T) {
	_, err := Open(Config{
		Driver: "nonexistent",
		DSN:    "foo",
	})
	if err == nil {
		t.Fatal("expected error for invalid driver")
	}
}

func TestPing(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()
	if err := conn.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestDriverDetection(t *testing.T) {
	conn := openTestDB(t)

	if !conn.IsSQLite() {
		t.Error("expected IsSQLite to be true")
	}
	if conn.IsPostgres() {
		t.Error("expected IsPostgres to be false")
	}
	if conn.IsMySQL() {
		t.Error("expected IsMySQL to be false")
	}
	if conn.Driver() != "sqlite3" {
		t.Errorf("expected driver sqlite3, got %s", conn.Driver())
	}
}

func TestDB(t *testing.T) {
	conn := openTestDB(t)
	if conn.DB() == nil {
		t.Error("expected non-nil *sql.DB")
	}
}

func TestConfigOptions(t *testing.T) {
	conn, err := Open(Config{
		Driver:                 "sqlite3",
		DSN:                    ":memory:",
		MaxOpenConns:           5,
		MaxIdleConns:           2,
		ConnMaxLifetimeSeconds: 300,
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = conn.Close() }()

	ctx := context.Background()
	if err := conn.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestWithTxCommit(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	_, err := conn.ExecContext(ctx, "CREATE TABLE test_tx (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	err = conn.WithTx(ctx, nil, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO test_tx (id, val) VALUES (1, 'hello')")
		return err
	})
	if err != nil {
		t.Fatalf("with tx: %v", err)
	}

	var val string
	err = conn.QueryRowContext(ctx, "SELECT val FROM test_tx WHERE id = 1").Scan(&val)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if val != "hello" {
		t.Errorf("expected 'hello', got %q", val)
	}
}

func TestWithTxRollback(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	_, err := conn.ExecContext(ctx, "CREATE TABLE test_rollback (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	err = conn.WithTx(ctx, nil, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO test_rollback (id, val) VALUES (1, 'rollme')")
		if err != nil {
			return err
		}
		return sql.ErrNoRows // simulate error to trigger rollback
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}

	var count int
	err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_rollback").Scan(&count)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows after rollback, got %d", count)
	}
}

func TestBeginTx(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
}

func TestQueryContext(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	_, err := conn.ExecContext(ctx, "CREATE TABLE test_query (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	_, err = conn.ExecContext(ctx, "INSERT INTO test_query (id) VALUES (1), (2), (3)")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	rows, err := conn.QueryContext(ctx, "SELECT id FROM test_query ORDER BY id")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan: %v", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}
	if len(ids) != 3 {
		t.Errorf("expected 3 rows, got %d", len(ids))
	}
}

// --- Placeholder conversion tests ---

func TestConvertPlaceholders(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"SELECT * FROM t WHERE id = $1", "SELECT * FROM t WHERE id = ?"},
		{"INSERT INTO t (a, b) VALUES ($1, $2)", "INSERT INTO t (a, b) VALUES (?, ?)"},
		{"$1 $2 $3 $10 $99", "? ? ? ? ?"},
		{"no placeholders here", "no placeholders here"},
		{"price is $100", "price is ?"},
		{"", ""},
	}
	for _, tt := range tests {
		got := ConvertPlaceholders(tt.input)
		if got != tt.want {
			t.Errorf("ConvertPlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPlaceholders(t *testing.T) {
	tests := []struct {
		n        int
		postgres bool
		want     string
	}{
		{0, false, ""},
		{1, false, "?"},
		{3, false, "?, ?, ?"},
		{1, true, "$1"},
		{3, true, "$1, $2, $3"},
	}
	for _, tt := range tests {
		got := Placeholders(tt.n, tt.postgres)
		if got != tt.want {
			t.Errorf("Placeholders(%d, %v) = %q, want %q", tt.n, tt.postgres, got, tt.want)
		}
	}
}

func TestSetClause(t *testing.T) {
	tests := []struct {
		columns  []string
		postgres bool
		offset   int
		want     string
	}{
		{nil, false, 0, ""},
		{[]string{"name"}, false, 0, "name = ?"},
		{[]string{"name", "email"}, false, 0, "name = ?, email = ?"},
		{[]string{"name"}, true, 0, "name = $1"},
		{[]string{"name", "email"}, true, 2, "name = $3, email = $4"},
	}
	for _, tt := range tests {
		got := SetClause(tt.columns, tt.postgres, tt.offset)
		if got != tt.want {
			t.Errorf("SetClause(%v, %v, %d) = %q, want %q", tt.columns, tt.postgres, tt.offset, got, tt.want)
		}
	}
}

func TestQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder("SELECT * FROM users", false)
	qb.Where("active = ?").Where("age > ?").OrderBy("name ASC").Limit(10).Offset(20)

	want := "SELECT * FROM users WHERE active = ? AND age > ? ORDER BY name ASC LIMIT 10 OFFSET 20"
	got := qb.Build()
	if got != want {
		t.Errorf("Build() = %q, want %q", got, want)
	}
}

func TestQueryBuilderNoConditions(t *testing.T) {
	qb := NewQueryBuilder("SELECT * FROM users", false)
	want := "SELECT * FROM users"
	got := qb.Build()
	if got != want {
		t.Errorf("Build() = %q, want %q", got, want)
	}
}

// --- BaseRepository tests ---

func setupRepoTable(t *testing.T, conn *Conn) {
	t.Helper()
	ctx := context.Background()
	_, err := conn.ExecContext(ctx, `CREATE TABLE items (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		category TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	_, err = conn.ExecContext(ctx, `INSERT INTO items (id, name, category) VALUES
		(1, 'apple', 'fruit'),
		(2, 'banana', 'fruit'),
		(3, 'carrot', 'vegetable')`)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
}

func TestBaseRepositoryExistsBy(t *testing.T) {
	conn := openTestDB(t)
	setupRepoTable(t, conn)
	repo := NewBaseRepository(conn, "items")
	ctx := context.Background()

	exists, err := repo.ExistsBy(ctx, "name", "apple")
	if err != nil {
		t.Fatalf("exists: %v", err)
	}
	if !exists {
		t.Error("expected apple to exist")
	}

	exists, err = repo.ExistsBy(ctx, "name", "pizza")
	if err != nil {
		t.Fatalf("exists: %v", err)
	}
	if exists {
		t.Error("expected pizza to not exist")
	}
}

func TestBaseRepositoryCountBy(t *testing.T) {
	conn := openTestDB(t)
	setupRepoTable(t, conn)
	repo := NewBaseRepository(conn, "items")
	ctx := context.Background()

	count, err := repo.CountBy(ctx, "category", "fruit")
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 fruits, got %d", count)
	}

	count, err = repo.CountBy(ctx, "category", "unknown")
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 unknowns, got %d", count)
	}
}

func TestBaseRepositoryDeleteBy(t *testing.T) {
	conn := openTestDB(t)
	setupRepoTable(t, conn)
	repo := NewBaseRepository(conn, "items")
	ctx := context.Background()

	affected, err := repo.DeleteBy(ctx, "category", "fruit")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if affected != 2 {
		t.Errorf("expected 2 deleted, got %d", affected)
	}

	count, err := repo.CountBy(ctx, "category", "fruit")
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 after delete, got %d", count)
	}
}

func TestBaseRepositoryTableAndConn(t *testing.T) {
	conn := openTestDB(t)
	repo := NewBaseRepository(conn, "test_table")

	if repo.Table() != "test_table" {
		t.Errorf("expected table 'test_table', got %q", repo.Table())
	}
	if repo.Conn() != conn {
		t.Error("expected Conn() to return the same connection")
	}
}

// --- Migration tests ---

func TestMigratorUpDown(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	migrator := NewMigrator(conn, "schema_migrations")
	migrator.Add(Migration{
		Version: "001",
		Name:    "create_users",
		Up: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, `CREATE TABLE users (
				id INTEGER PRIMARY KEY,
				name TEXT NOT NULL
			)`)
			return err
		},
		Down: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "DROP TABLE users")
			return err
		},
	})
	migrator.Add(Migration{
		Version: "002",
		Name:    "create_posts",
		Up: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, `CREATE TABLE posts (
				id INTEGER PRIMARY KEY,
				user_id INTEGER NOT NULL,
				title TEXT NOT NULL
			)`)
			return err
		},
		Down: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "DROP TABLE posts")
			return err
		},
	})

	// Apply all migrations.
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("up: %v", err)
	}

	// Verify tables exist.
	_, err := conn.ExecContext(ctx, "INSERT INTO users (id, name) VALUES (1, 'Alice')")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	_, err = conn.ExecContext(ctx, "INSERT INTO posts (id, user_id, title) VALUES (1, 1, 'Hello')")
	if err != nil {
		t.Fatalf("insert post: %v", err)
	}

	// Running Up again should be idempotent.
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("second up: %v", err)
	}

	// Revert last migration.
	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("down: %v", err)
	}

	// Posts table should be gone.
	_, err = conn.ExecContext(ctx, "INSERT INTO posts (id, user_id, title) VALUES (2, 1, 'World')")
	if err == nil {
		t.Fatal("expected error after dropping posts table")
	}

	// Users table should still exist.
	_, err = conn.ExecContext(ctx, "INSERT INTO users (id, name) VALUES (2, 'Bob')")
	if err != nil {
		t.Fatalf("insert user after partial down: %v", err)
	}

	// Revert again.
	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("second down: %v", err)
	}

	// Users table should be gone.
	_, err = conn.ExecContext(ctx, "INSERT INTO users (id, name) VALUES (3, 'Carol')")
	if err == nil {
		t.Fatal("expected error after dropping users table")
	}

	// Down with nothing applied should be a no-op.
	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("down on empty: %v", err)
	}
}

func TestMigratorStatus(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	migrator := NewMigrator(conn, "schema_migrations")
	migrator.Add(Migration{
		Version: "001",
		Name:    "first",
		Up: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "CREATE TABLE t1 (id INTEGER)")
			return err
		},
		Down: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "DROP TABLE t1")
			return err
		},
	})
	migrator.Add(Migration{
		Version: "002",
		Name:    "second",
		Up: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "CREATE TABLE t2 (id INTEGER)")
			return err
		},
		Down: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "DROP TABLE t2")
			return err
		},
	})

	// Apply first migration only by running Up then Down on second.
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("up: %v", err)
	}
	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("down: %v", err)
	}

	statuses, err := migrator.Status(ctx)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	if !statuses[0].Applied {
		t.Error("expected first migration to be applied")
	}
	if statuses[1].Applied {
		t.Error("expected second migration to not be applied")
	}
}

func TestMigratorNoUpFunction(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	migrator := NewMigrator(conn, "schema_migrations")
	migrator.Add(Migration{
		Version: "001",
		Name:    "broken",
	})

	err := migrator.Up(ctx)
	if err == nil {
		t.Fatal("expected error for nil Up function")
	}
}

func TestMigratorNoDownFunction(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	migrator := NewMigrator(conn, "schema_migrations")
	migrator.Add(Migration{
		Version: "001",
		Name:    "no_down",
		Up: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "CREATE TABLE nd (id INTEGER)")
			return err
		},
	})

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("up: %v", err)
	}
	err := migrator.Down(ctx)
	if err == nil {
		t.Fatal("expected error for nil Down function")
	}
}

// --- SQL file migration tests ---

func TestMigratorAddSQLWithDialect(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	// Create temp directory with SQL files.
	dir := t.TempDir()
	// Write dialect-specific up file.
	if err := os.WriteFile(filepath.Join(dir, "001_create_table.up.sqlite3.sql"),
		[]byte("CREATE TABLE dialect_test (id INTEGER PRIMARY KEY, name TEXT)"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Write generic down file.
	if err := os.WriteFile(filepath.Join(dir, "001_create_table.down.sql"),
		[]byte("DROP TABLE dialect_test"), 0o600); err != nil {
		t.Fatal(err)
	}

	migrator := NewMigrator(conn, "sql_migrations")
	migrator.SetSQLDir(dir)
	migrator.AddSQL("001", "create_table")

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("up: %v", err)
	}

	// Verify table was created.
	_, err := conn.ExecContext(ctx, "INSERT INTO dialect_test (id, name) VALUES (1, 'test')")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Revert.
	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("down: %v", err)
	}

	// Table should be gone.
	_, err = conn.ExecContext(ctx, "INSERT INTO dialect_test (id, name) VALUES (2, 'gone')")
	if err == nil {
		t.Fatal("expected error after drop")
	}
}

func TestMigratorAddSQLGenericFallback(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	dir := t.TempDir()
	// Only write generic files (no dialect-specific).
	if err := os.WriteFile(filepath.Join(dir, "001_items.up.sql"),
		[]byte("CREATE TABLE generic_items (id INTEGER)"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "001_items.down.sql"),
		[]byte("DROP TABLE generic_items"), 0o600); err != nil {
		t.Fatal(err)
	}

	migrator := NewMigrator(conn, "sql_migrations")
	migrator.SetSQLDir(dir)
	migrator.AddSQL("001", "items")

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("up: %v", err)
	}

	_, err := conn.ExecContext(ctx, "INSERT INTO generic_items (id) VALUES (1)")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
}

func TestMigratorAddSQLNoDir(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	migrator := NewMigrator(conn, "sql_migrations")
	// Deliberately not calling SetSQLDir.
	migrator.AddSQL("001", "missing")

	err := migrator.Up(ctx)
	if err == nil {
		t.Fatal("expected error when sql dir not set")
	}
}

func TestMigratorAddSQLMissingFile(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	dir := t.TempDir()
	migrator := NewMigrator(conn, "sql_migrations")
	migrator.SetSQLDir(dir)
	migrator.AddSQL("001", "nonexistent")

	err := migrator.Up(ctx)
	if err == nil {
		t.Fatal("expected error when sql file is missing")
	}
}

func TestMigratorAddSQLDown(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "001_tbl.up.sql"),
		[]byte("CREATE TABLE sql_down_test (id INTEGER)"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "001_tbl.down.sql"),
		[]byte("DROP TABLE sql_down_test"), 0o600); err != nil {
		t.Fatal(err)
	}

	migrator := NewMigrator(conn, "sql_migrations")
	migrator.SetSQLDir(dir)
	migrator.AddSQL("001", "tbl")

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("up: %v", err)
	}
	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("down: %v", err)
	}

	// Table should be gone.
	_, err := conn.ExecContext(ctx, "INSERT INTO sql_down_test (id) VALUES (1)")
	if err == nil {
		t.Fatal("expected error after down")
	}
}

func TestBaseRepositoryExistsByError(t *testing.T) {
	conn := openTestDB(t)
	repo := NewBaseRepository(conn, "nonexistent_table")
	ctx := context.Background()

	_, err := repo.ExistsBy(ctx, "id", 1)
	if err == nil {
		t.Fatal("expected error for nonexistent table")
	}
}

func TestBaseRepositoryCountByError(t *testing.T) {
	conn := openTestDB(t)
	repo := NewBaseRepository(conn, "nonexistent_table")
	ctx := context.Background()

	_, err := repo.CountBy(ctx, "id", 1)
	if err == nil {
		t.Fatal("expected error for nonexistent table")
	}
}

func TestBaseRepositoryDeleteByError(t *testing.T) {
	conn := openTestDB(t)
	repo := NewBaseRepository(conn, "nonexistent_table")
	ctx := context.Background()

	_, err := repo.DeleteBy(ctx, "id", 1)
	if err == nil {
		t.Fatal("expected error for nonexistent table")
	}
}

func TestMigratorUpError(t *testing.T) {
	conn := openTestDB(t)
	ctx := context.Background()

	migrator := NewMigrator(conn, "schema_migrations")
	migrator.Add(Migration{
		Version: "001",
		Name:    "bad_sql",
		Up: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "INVALID SQL SYNTAX HERE!!!")
			return err
		},
		Down: func(ctx context.Context, tx *sql.Tx) error {
			return nil
		},
	})

	err := migrator.Up(ctx)
	if err == nil {
		t.Fatal("expected error from bad SQL")
	}
}

// --- NULL helper tests ---

func TestNullString(t *testing.T) {
	ns := NullString("hello")
	if !ns.Valid || ns.String != "hello" {
		t.Errorf("NullString(hello) = %+v", ns)
	}

	ns = NullString("")
	if ns.Valid {
		t.Error("NullString('') should be invalid")
	}
}

func TestStringPtr(t *testing.T) {
	ns := sql.NullString{String: "world", Valid: true}
	ptr := StringPtr(ns)
	if ptr == nil || *ptr != "world" {
		t.Errorf("StringPtr valid = %v", ptr)
	}

	ns = sql.NullString{Valid: false}
	ptr = StringPtr(ns)
	if ptr != nil {
		t.Error("StringPtr invalid should be nil")
	}
}

func TestNullInt64(t *testing.T) {
	ni := NullInt64(42)
	if !ni.Valid || ni.Int64 != 42 {
		t.Errorf("NullInt64(42) = %+v", ni)
	}

	ni = NullInt64(0)
	if !ni.Valid {
		t.Error("NullInt64(0) should be valid")
	}
}

func TestNullInt64Ptr(t *testing.T) {
	v := int64(99)
	ni := NullInt64Ptr(&v)
	if !ni.Valid || ni.Int64 != 99 {
		t.Errorf("NullInt64Ptr(&99) = %+v", ni)
	}

	ni = NullInt64Ptr(nil)
	if ni.Valid {
		t.Error("NullInt64Ptr(nil) should be invalid")
	}
}

func TestInt64Ptr(t *testing.T) {
	ni := sql.NullInt64{Int64: 7, Valid: true}
	ptr := Int64Ptr(ni)
	if ptr == nil || *ptr != 7 {
		t.Errorf("Int64Ptr valid = %v", ptr)
	}

	ni = sql.NullInt64{Valid: false}
	ptr = Int64Ptr(ni)
	if ptr != nil {
		t.Error("Int64Ptr invalid should be nil")
	}
}

func TestNullFloat64(t *testing.T) {
	nf := NullFloat64(3.14)
	if !nf.Valid || nf.Float64 != 3.14 {
		t.Errorf("NullFloat64(3.14) = %+v", nf)
	}
}

func TestFloat64Ptr(t *testing.T) {
	nf := sql.NullFloat64{Float64: 2.71, Valid: true}
	ptr := Float64Ptr(nf)
	if ptr == nil || *ptr != 2.71 {
		t.Errorf("Float64Ptr valid = %v", ptr)
	}

	nf = sql.NullFloat64{Valid: false}
	ptr = Float64Ptr(nf)
	if ptr != nil {
		t.Error("Float64Ptr invalid should be nil")
	}
}

func TestNullBool(t *testing.T) {
	nb := NullBool(true)
	if !nb.Valid || !nb.Bool {
		t.Errorf("NullBool(true) = %+v", nb)
	}

	nb = NullBool(false)
	if !nb.Valid || nb.Bool {
		t.Errorf("NullBool(false) = %+v", nb)
	}
}

func TestBoolPtr(t *testing.T) {
	nb := sql.NullBool{Bool: true, Valid: true}
	ptr := BoolPtr(nb)
	if ptr == nil || !*ptr {
		t.Errorf("BoolPtr valid = %v", ptr)
	}

	nb = sql.NullBool{Valid: false}
	ptr = BoolPtr(nb)
	if ptr != nil {
		t.Error("BoolPtr invalid should be nil")
	}
}
