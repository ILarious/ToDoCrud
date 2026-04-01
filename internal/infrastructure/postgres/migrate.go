package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ApplyMigrations executes all *.up.sql files from dir in lexicographical order.
// Migrations should be idempotent (for example, CREATE TABLE IF NOT EXISTS),
// because this runner does not track schema versions.
func ApplyMigrations(ctx context.Context, db *sql.DB, dir string) error {
	if db == nil {
		return fmt.Errorf("apply migrations: nil db")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	files := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)

	for _, path := range files {
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %q: %w", path, err)
		}

		query := strings.TrimSpace(string(sqlBytes))
		if query == "" {
			continue
		}

		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("apply migration %q: %w", path, err)
		}
	}

	return nil
}
