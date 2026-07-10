package example

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// defaultPerPage/maxPerPage/defaultPage mirror the pagination defaults
// mandated by docs/api-design-standards.md §9.
const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

// uniqueViolationCode is the Postgres error code for a UNIQUE constraint
// violation (23505), used to translate a racing concurrent Create into a
// domain Conflict instead of a raw driver error.
const uniqueViolationCode = "23505"

// ListFilters narrows and paginates a List query. Page/PerPage are raw,
// caller-supplied values; normalizePagination applies defaults/bounds.
type ListFilters struct {
	Status  string
	Page    int
	PerPage int
}

// ListResult carries a page of Examples plus the total row count matching
// the filters (ignoring pagination), so callers can build the standard
// data+pagination envelope (docs/api-design-standards.md §6).
type ListResult struct {
	Items []*Example
	Total int
}

// Repository defines data access for Example. PostgresRepository is the only
// implementation; tests use the generated mock (mocks/repository_mock.go).
type Repository interface {
	Create(ctx context.Context, ex *Example) error
	FindByID(ctx context.Context, id string) (*Example, error)
	FindByEmail(ctx context.Context, email string) (*Example, error)
	Update(ctx context.Context, ex *Example) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters ListFilters) (*ListResult, error)
}

// PostgresRepository implements Repository using the shared pgx pool
// exposed by database.Client.
type PostgresRepository struct {
	db database.Client
}

// NewPostgresRepository builds a Repository backed by PostgreSQL.
func NewPostgresRepository(db database.Client) Repository {
	return &PostgresRepository{db: db}
}

// Create inserts ex, which must already have ID populated by the caller
// (the service layer generates it via google/uuid — see service.go — rather
// than relying on the table's gen_random_uuid() column default, so the ID
// is known before the row is committed). Populates CreatedAt/UpdatedAt/
// Version from the database defaults. A racing duplicate email is reported
// as a domain Conflict rather than a raw pgconn error.
func (r *PostgresRepository) Create(ctx context.Context, ex *Example) error {
	const query = `
		INSERT INTO examples (id, name, email, status, count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at, version`

	err := r.db.Pool().QueryRow(ctx, query, ex.ID, ex.Name, ex.Email, ex.Status, ex.Count).
		Scan(&ex.CreatedAt, &ex.UpdatedAt, &ex.Version)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return domainerrors.Conflict(fmt.Sprintf("an example with email %q already exists", ex.Email))
		}
		return fmt.Errorf("create example: %w", err)
	}

	return nil
}

// FindByID returns the example with the given ID, or a domain NotFound
// error if no row matches.
func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*Example, error) {
	const query = `
		SELECT id, name, email, status, count, created_at, updated_at, version
		FROM examples
		WHERE id = $1`

	return r.scanOne(ctx, query, id)
}

// FindByEmail returns the example with the given email, or a domain
// NotFound error if no row matches. Callers (the service layer) treat
// NotFound as "email is available" during uniqueness checks.
func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (*Example, error) {
	const query = `
		SELECT id, name, email, status, count, created_at, updated_at, version
		FROM examples
		WHERE email = $1`

	return r.scanOne(ctx, query, email)
}

// scanOne runs a single-row query and maps pgx.ErrNoRows to a domain
// NotFound, keeping that translation in one place for FindByID/FindByEmail.
func (r *PostgresRepository) scanOne(ctx context.Context, query string, arg any) (*Example, error) {
	var ex Example
	err := r.db.Pool().QueryRow(ctx, query, arg).Scan(
		&ex.ID, &ex.Name, &ex.Email, &ex.Status, &ex.Count, &ex.CreatedAt, &ex.UpdatedAt, &ex.Version,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("example", fmt.Sprintf("%v", arg))
		}
		return nil, fmt.Errorf("find example: %w", err)
	}
	return &ex, nil
}

// Update applies optimistic locking: it only updates the row when both id
// and the caller-supplied ex.Version match, incrementing the version on
// success. Zero rows affected (pgx.ErrNoRows on the RETURNING clause) is
// reported as a domain Conflict — the caller already confirmed the row
// exists via FindByID, so a no-op here means a concurrent writer won the
// race, not that the row is missing.
func (r *PostgresRepository) Update(ctx context.Context, ex *Example) error {
	const query = `
		UPDATE examples
		SET name = $1, status = $2, count = $3, updated_at = NOW(), version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING updated_at, version`

	err := r.db.Pool().QueryRow(ctx, query, ex.Name, ex.Status, ex.Count, ex.ID, ex.Version).
		Scan(&ex.UpdatedAt, &ex.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainerrors.Conflict("example was modified by another request")
		}
		return fmt.Errorf("update example: %w", err)
	}

	return nil
}

// Delete removes the example with the given ID. Deleting an ID that
// doesn't exist is reported as a domain NotFound, matching FindByID's
// convention so handlers don't need special-case logic per method.
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM examples WHERE id = $1`

	tag, err := r.db.Pool().Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete example: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.NotFound("example", id)
	}

	return nil
}

// List returns a page of examples matching filters plus the total matching
// row count (ignoring pagination). Total is computed via a separate COUNT(*)
// query rather than a COUNT(*) OVER() window function deliberately: a window
// function only yields a value on rows actually returned, so an out-of-range
// page (LIMIT/OFFSET beyond the result set) would silently report total=0 —
// breaking the documented contract that empty pages still carry a correct
// pagination envelope (docs/api-design-standards.md §9).
func (r *PostgresRepository) List(ctx context.Context, filters ListFilters) (*ListResult, error) {
	_, _, limit, offset := normalizePagination(filters.Page, filters.PerPage)

	const countQuery = `SELECT COUNT(*) FROM examples WHERE ($1 = '' OR status = $1)`

	var total int
	if err := r.db.Pool().QueryRow(ctx, countQuery, filters.Status).Scan(&total); err != nil {
		return nil, fmt.Errorf("count examples: %w", err)
	}

	result := &ListResult{Items: []*Example{}, Total: total}
	if total == 0 {
		return result, nil
	}

	const listQuery = `
		SELECT id, name, email, status, count, created_at, updated_at, version
		FROM examples
		WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Pool().Query(ctx, listQuery, filters.Status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list examples: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ex Example
		if err := rows.Scan(
			&ex.ID, &ex.Name, &ex.Email, &ex.Status, &ex.Count, &ex.CreatedAt, &ex.UpdatedAt, &ex.Version,
		); err != nil {
			return nil, fmt.Errorf("scan example row: %w", err)
		}
		result.Items = append(result.Items, &ex)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list examples: %w", err)
	}

	return result, nil
}

// normalizePagination applies docs/api-design-standards.md §9 defaults/
// bounds (page default 1, per_page default 20 / max 100) and converts the
// result into SQL LIMIT/OFFSET values.
func normalizePagination(page, perPage int) (normPage, normPerPage, limit, offset int) {
	if page < 1 {
		page = defaultPage
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return page, perPage, perPage, (page - 1) * perPage
}

// totalPages computes the pagination envelope's total_pages as a ceiling
// division of total/perPage, returning 0 (not 1) when there are no results.
func totalPages(total, perPage int) int {
	if total <= 0 || perPage <= 0 {
		return 0
	}
	return (total + perPage - 1) / perPage
}
