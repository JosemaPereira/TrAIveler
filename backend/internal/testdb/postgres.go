// Package testdb holds shared constants for backend test infrastructure that
// spins up PostgreSQL via testcontainers-go, so that image version cannot
// drift independently across packages.
package testdb

// PostgresImage is the PostgreSQL container image used by every
// testcontainers-backed integration test in this module. It is pinned to
// the same version as the runtime/RDS image in docker-compose.yml — a
// deliberate architectural decision (see patterns-discovered.md's "A
// Version Pin Can Be Deliberate Architecture, Not Drift"), not staleness.
const PostgresImage = "postgres:15.18-alpine"
