package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupTestDB creates an isolated PostgreSQL schema for integration tests,
// runs all *.up.sql migrations in it, and returns a pool pre-configured to
// use that schema. Skips the test if TEST_DATABASE_URL is not set.
// Drops the schema via t.Cleanup, so tests can run in parallel without leaking.
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	schema := "test_" + strings.ReplaceAll(uuid.New().String(), "-", "_")

	adminConn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		t.Fatalf("testutil.SetupTestDB: connect: %v", err)
	}
	defer func() { _ = adminConn.Close(context.Background()) }()

	if _, err = adminConn.Exec(context.Background(), "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("testutil.SetupTestDB: create schema %s: %v", schema, err)
	}

	if err = runMigrations(t, adminConn, schema); err != nil {
		t.Fatalf("testutil.SetupTestDB: migrate: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("testutil.SetupTestDB: parse config: %v", err)
	}
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", schema))
		return err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("testutil.SetupTestDB: pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		conn, err := pgx.Connect(context.Background(), dsn)
		if err == nil {
			_, _ = conn.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA %s CASCADE", schema))
			_ = conn.Close(context.Background())
		}
	})

	return pool
}

// runMigrations executes all *.up.sql files from the migrations directory
// within the given schema, in lexicographic order.
func runMigrations(t *testing.T, conn *pgx.Conn, schema string) error {
	t.Helper()

	migrationsDir := findMigrationsDir(t)
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", migrationsDir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, filepath.Join(migrationsDir, e.Name()))
		}
	}
	sort.Strings(files)

	ctx := context.Background()
	if _, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", schema)); err != nil {
		return fmt.Errorf("set search_path: %w", err)
	}

	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if _, err = conn.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("exec %s: %w", f, err)
		}
	}
	return nil
}

// findMigrationsDir walks up from the current working directory until it
// finds a go.mod file, then returns the sibling migrations/ directory.
func findMigrationsDir(t *testing.T) string {
	t.Helper()

	if dir := os.Getenv("MIGRATIONS_PATH"); dir != "" {
		return dir
	}

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("testutil: getwd: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "migrations")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("testutil: could not find go.mod from %s; set MIGRATIONS_PATH", dir)
		}
		dir = parent
	}
}
