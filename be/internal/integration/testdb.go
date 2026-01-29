package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// projectRoot returns the repo root by walking up from this file.
func projectRoot() string {
	_, thisFile, _, _ := runtime.Caller(0)
	// .../internal/integration/testdb.go -> .../ (repo root)
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
}

func resetSchema(ctx context.Context, pool *pgxpool.Pool) error {
	// Recreate public schema for deterministic tests.
	_, err := pool.Exec(ctx, `
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO public;
`)
	return err
}

func ensureExtensions(ctx context.Context, pool *pgxpool.Pool) error {
	// gen_random_uuid() requires pgcrypto
	_, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto;`)
	return err
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationsDir := filepath.Join(projectRoot(), "internal", "db", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		files = append(files, filepath.Join(migrationsDir, name))
	}
	sort.Strings(files)

	for _, path := range files {
		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", path, err)
		}
		upSQL, err := extractGooseUp(string(b))
		if err != nil {
			return fmt.Errorf("parse migration %s: %w", path, err)
		}
		if strings.TrimSpace(upSQL) == "" {
			continue
		}
		if _, err := pool.Exec(ctx, upSQL); err != nil {
			return fmt.Errorf("exec migration %s: %w", filepath.Base(path), err)
		}
	}
	return nil
}

func extractGooseUp(contents string) (string, error) {
	// Very small parser for our migration format:
	// -- +goose Up
	// -- +goose StatementBegin
	// <SQL>
	// -- +goose StatementEnd
	// -- +goose Down
	lines := strings.Split(contents, "\n")
	inUp := false
	var out []string

	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "-- +goose Up") {
			inUp = true
			continue
		}
		if strings.HasPrefix(trim, "-- +goose Down") {
			inUp = false
			break
		}
		if !inUp {
			continue
		}
		if strings.HasPrefix(trim, "-- +goose ") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n"), nil
}
