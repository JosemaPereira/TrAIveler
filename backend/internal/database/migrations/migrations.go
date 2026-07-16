// Package migrations resolves the shared goose migrations directory and
// dialect config, replacing the caller-relative "../../migrations" consts
// and goose.SetDialect("postgres") calls previously duplicated across three
// integration test helpers.
package migrations

import (
	"path/filepath"
	"runtime"

	"github.com/pressly/goose/v3"
)

// Dir is the absolute path to backend/migrations, resolved from this file's
// own location so it is correct regardless of the caller's working
// directory or package depth.
var Dir = resolveDir()

func resolveDir() string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		// Only fails in pathological environments (e.g. no debug info);
		// panic at init rather than a confusing lookup failure mid-test.
		panic("migrations: unable to resolve caller for migrations directory")
	}

	// thisFile is backend/internal/database/migrations/migrations.go; three
	// levels up is the backend/ module root.
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "migrations")
}

// SetDialect configures goose for PostgreSQL.
func SetDialect() error {
	return goose.SetDialect("postgres")
}
