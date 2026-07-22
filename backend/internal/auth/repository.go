package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	authjwt "github.com/JosemaPereira/TrAIveler/backend/internal/auth/jwt"
	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// uniqueViolationCode is the Postgres error code for a UNIQUE constraint
// violation (23505), translated into a domain Conflict — a duplicate email on
// CreateUser/UpdateUser, or a duplicate token_hash on CreateRefreshToken.
const uniqueViolationCode = "23505"

// userColumns is the SELECT list for scanning a full User row. last_login_at is
// excluded (not on the User model); full_name is COALESCEd because migration 005
// added it as nullable, so legacy rows may be NULL while the model uses a
// non-pointer string.
const userColumns = `id, email, password_hash, role, COALESCE(full_name, '') AS full_name,
	has_subscription, failed_login_attempts, last_failed_login_at, email_verified,
	created_at, updated_at, version`

// UserRepository defines data access for the User account model.
// PostgresUserRepository is the only implementation; tests use the generated
// mock (mocks/user_repository_mock.go).
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	// UpdateUser writes the mutable user fields using optimistic locking on
	// version: it fails with a domain Conflict if user.Version no longer matches
	// the stored row, and with a domain NotFound if the id does not exist.
	UpdateUser(ctx context.Context, user *User) error
}

// RefreshTokenRepository defines data access for refresh tokens. Only the
// SHA-256 hash is persisted. Revocation is idempotent.
type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	// RevokeRefreshToken marks a single token revoked; revoking an already-revoked
	// or unknown token is a no-op success (idempotent).
	RevokeRefreshToken(ctx context.Context, id string) error
	// RevokeAllRefreshTokens revokes every currently-live token for a user; a
	// user with no live tokens is a no-op success.
	RevokeAllRefreshTokens(ctx context.Context, userID string) error
}

// PostgresUserRepository implements UserRepository using the shared pgx pool.
type PostgresUserRepository struct {
	db database.Client
}

// NewPostgresUserRepository builds a UserRepository backed by PostgreSQL.
func NewPostgresUserRepository(db database.Client) UserRepository {
	return &PostgresUserRepository{db: db}
}

// CreateUser inserts user, which must already have ID/Email/PasswordHash/Role
// populated by the caller (the service layer generates the ID via google/uuid,
// mirroring the subscription repository). It populates the database-managed
// CreatedAt/UpdatedAt/Version. A duplicate email (unique constraint) is
// reported as a domain Conflict.
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *User) error {
	const query = `
		INSERT INTO users (id, email, password_hash, role, full_name, has_subscription, email_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at, version`

	err := r.db.Pool().QueryRow(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Role, user.FullName,
		user.HasSubscription, user.EmailVerified,
	).Scan(&user.CreatedAt, &user.UpdatedAt, &user.Version)
	if err != nil {
		if isUniqueViolation(err) {
			return domainerrors.Conflict(fmt.Sprintf("a user with email %q already exists", user.Email))
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// GetUserByEmail returns the user with the given email, or a domain NotFound
// error if none exists.
func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const query = "SELECT " + userColumns + " FROM users WHERE email = $1"

	user, err := scanUser(r.db.Pool().QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("user", email)
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

// GetUserByID returns the user with the given id, or a domain NotFound error if
// none exists.
func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	const query = "SELECT " + userColumns + " FROM users WHERE id = $1"

	user, err := scanUser(r.db.Pool().QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("user", id)
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}

// UpdateUser writes the mutable user fields (role is immutable and excluded)
// using optimistic locking on version: the UPDATE matches on id AND the
// caller-held version and bumps version atomically, with RETURNING feeding the
// new version/updated_at back into user. A zero-row result is disambiguated by a
// follow-up existence check — a present id means a stale version (Conflict), an
// absent id means the user is gone (NotFound). A colliding email is also a
// Conflict.
func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *User) error {
	const query = `
		UPDATE users
		SET email = $1,
		    password_hash = $2,
		    full_name = $3,
		    has_subscription = $4,
		    failed_login_attempts = $5,
		    last_failed_login_at = $6,
		    email_verified = $7,
		    updated_at = NOW(),
		    version = version + 1
		WHERE id = $8 AND version = $9
		RETURNING version, updated_at`

	err := r.db.Pool().QueryRow(ctx, query,
		user.Email, user.PasswordHash, user.FullName, user.HasSubscription,
		user.FailedLoginAttempts, user.LastFailedLoginAt, user.EmailVerified,
		user.ID, user.Version,
	).Scan(&user.Version, &user.UpdatedAt)
	if err == nil {
		return nil
	}
	if isUniqueViolation(err) {
		return domainerrors.Conflict(fmt.Sprintf("a user with email %q already exists", user.Email))
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return r.resolveUpdateMiss(ctx, user.ID)
	}
	return fmt.Errorf("update user: %w", err)
}

// resolveUpdateMiss disambiguates a zero-row optimistic update: an existing id
// means the version was stale (Conflict); a missing id means NotFound.
func (r *PostgresUserRepository) resolveUpdateMiss(ctx context.Context, id string) error {
	var exists bool
	err := r.db.Pool().QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check user existence: %w", err)
	}
	if exists {
		return domainerrors.Conflict(fmt.Sprintf("user %q was modified concurrently", id))
	}
	return domainerrors.NotFound("user", id)
}

// scanUser scans one row selected with userColumns into a User.
func scanUser(row pgx.Row) (*User, error) {
	var user User
	err := row.Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.FullName,
		&user.HasSubscription, &user.FailedLoginAttempts, &user.LastFailedLoginAt,
		&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt, &user.Version,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// PostgresRefreshTokenRepository implements RefreshTokenRepository using the
// shared pgx pool.
type PostgresRefreshTokenRepository struct {
	db database.Client
}

// NewPostgresRefreshTokenRepository builds a RefreshTokenRepository backed by
// PostgreSQL.
func NewPostgresRefreshTokenRepository(db database.Client) RefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

// CreateRefreshToken inserts token, which must already have ID/UserID/TokenHash/
// ExpiresAt populated. It populates the database-managed CreatedAt. A duplicate
// token_hash is reported as a domain Conflict.
func (r *PostgresRefreshTokenRepository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	const query = `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	err := r.db.Pool().QueryRow(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt).
		Scan(&token.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domainerrors.Conflict("refresh token already exists")
		}
		return fmt.Errorf("create refresh token: %w", err)
	}

	return nil
}

// GetRefreshTokenByHash returns the token row with the given SHA-256 hash, or a
// domain NotFound error if none exists.
func (r *PostgresRefreshTokenRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1`

	var token RefreshToken
	err := r.db.Pool().QueryRow(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt, &token.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("refresh token", tokenHash)
		}
		return nil, fmt.Errorf("find refresh token: %w", err)
	}

	return &token, nil
}

// RevokeRefreshToken stamps revoked_at on a live token. It is idempotent: the
// WHERE clause matches only a not-yet-revoked row, so revoking an already-revoked
// or unknown token affects zero rows and still succeeds.
func (r *PostgresRefreshTokenRepository) RevokeRefreshToken(ctx context.Context, id string) error {
	const query = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`

	if _, err := r.db.Pool().Exec(ctx, query, id); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}

// RevokeAllRefreshTokens stamps revoked_at on every live token for userID. A
// user with no live tokens is a no-op success.
func (r *PostgresRefreshTokenRepository) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	const query = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`

	if _, err := r.db.Pool().Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}

	return nil
}

// RefreshStore adapts a RefreshTokenRepository to jwt.RefreshTokenStore, the
// contract the jwt.Refresher depends on. It bridges the conventions: the jwt
// package uses uuid.UUID IDs while the repository uses string IDs, and it maps
// the repository's domain NotFound to jwt.ErrRefreshTokenNotFound.
type RefreshStore struct {
	repo RefreshTokenRepository
}

// compile-time assurance that RefreshStore satisfies the jwt contract.
var _ authjwt.RefreshTokenStore = (*RefreshStore)(nil)

// NewRefreshStore builds a RefreshStore over repo.
func NewRefreshStore(repo RefreshTokenRepository) *RefreshStore {
	return &RefreshStore{repo: repo}
}

// ByHash looks up a token by hash, mapping domain NotFound to
// jwt.ErrRefreshTokenNotFound (which the Refresher treats as an auth failure)
// and parsing the string IDs into the uuid.UUID the jwt record uses.
func (s *RefreshStore) ByHash(ctx context.Context, tokenHash string) (authjwt.RefreshTokenRecord, error) {
	token, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if isNotFound(err) {
			return authjwt.RefreshTokenRecord{}, authjwt.ErrRefreshTokenNotFound
		}
		return authjwt.RefreshTokenRecord{}, err
	}

	id, err := uuid.Parse(token.ID)
	if err != nil {
		return authjwt.RefreshTokenRecord{}, fmt.Errorf("parse refresh token id: %w", err)
	}
	userID, err := uuid.Parse(token.UserID)
	if err != nil {
		return authjwt.RefreshTokenRecord{}, fmt.Errorf("parse refresh token user id: %w", err)
	}

	return authjwt.RefreshTokenRecord{
		ID:        id,
		UserID:    userID,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: token.RevokedAt,
	}, nil
}

// Revoke marks the token with the given UUID revoked (idempotent), converting
// the UUID to the string id the repository uses.
func (s *RefreshStore) Revoke(ctx context.Context, id uuid.UUID) error {
	return s.repo.RevokeRefreshToken(ctx, id.String())
}

// Create persists a newly issued refresh token, generating its row ID (the jwt
// payload carries only the user, hash, and expiry).
func (s *RefreshStore) Create(ctx context.Context, token authjwt.NewRefreshToken) error {
	return s.repo.CreateRefreshToken(ctx, &RefreshToken{
		ID:        uuid.NewString(),
		UserID:    token.UserID.String(),
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
	})
}

// isUniqueViolation reports whether err is (or wraps) a Postgres unique
// constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

// isNotFound reports whether err is (or wraps) a domain NotFound error.
func isNotFound(err error) bool {
	var domainErr *domainerrors.DomainError
	return errors.As(err, &domainErr) && domainErr.Code == "not_found"
}
