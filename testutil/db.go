package testutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupTestDB creates an isolated PostgreSQL schema for integration tests and
// returns a pool pre-configured to use it. Skips the test if TEST_DATABASE_URL
// is not set. Drops the schema via t.Cleanup, so tests can run in parallel
// without state leaking between them.
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
	if _, err = adminConn.Exec(context.Background(), "CREATE SCHEMA "+schema); err != nil {
		_ = adminConn.Close(context.Background())
		t.Fatalf("testutil.SetupTestDB: create schema %s: %v", schema, err)
	}
	_ = adminConn.Close(context.Background())

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
