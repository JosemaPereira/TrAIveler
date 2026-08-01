package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver, used only to run goose migrations
	"github.com/pressly/goose/v3"

	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
	"github.com/JosemaPereira/TrAIveler/backend/internal/database/migrations"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/testdb"
)

// setupRepositoryTestDB starts a PostgreSQL testcontainer, applies goose
// migrations, and returns a ready database.Client. Mirrors
// internal/subscription/repository_integration_test.go's helper and its
// testing.Short() skip convention.
func setupRepositoryTestDB(t *testing.T, ctx context.Context) database.Client {
	t.Helper()

	pgContainer, err := postgres.Run(ctx,
		testdb.PostgresImage,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	sqlDB, err := sql.Open("pgx", connStr)
	require.NoError(t, err, "failed to open database/sql connection for migrations")
	defer sqlDB.Close()

	require.NoError(t, migrations.SetDialect())
	require.NoError(t, goose.Up(sqlDB, migrations.Dir), "failed to apply goose migrations")

	client, err := database.NewClient(ctx, connStr, 5, 25)
	require.NoError(t, err, "failed to create database client")
	t.Cleanup(func() {
		_ = client.Close()
	})

	return client
}

// newUser builds a valid User with a fresh ID for insertion, mirroring the
// service layer's UUID generation before CreateUser.
func newUser(emailSuffix string) *User {
	return &User{
		ID:           uuid.New().String(),
		Email:        "user-" + emailSuffix + "@example.com",
		PasswordHash: "bcrypt-hash",
		Role:         RoleAdmin,
		FullName:     "Test User",
	}
}

func requireConflict(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "conflict", domainErr.Code)
}

func requireNotFound(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationCreateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresUserRepository(db)

	t.Run("when the email is unique", func(t *testing.T) {
		t.Run("should insert the row and populate managed fields", func(t *testing.T) {
			user := newUser("create-1")

			err := repo.CreateUser(ctx, user)

			require.NoError(t, err)
			assert.False(t, user.CreatedAt.IsZero())
			assert.False(t, user.UpdatedAt.IsZero())
			assert.Equal(t, int64(1), user.Version, "version defaults to 1")
		})
	})

	t.Run("when the email already exists", func(t *testing.T) {
		t.Run("should return a domain Conflict", func(t *testing.T) {
			first := newUser("dup")
			require.NoError(t, repo.CreateUser(ctx, first))

			second := newUser("dup") // same email suffix -> same email
			err := repo.CreateUser(ctx, second)

			requireConflict(t, err)
		})
	})
}

func TestIntegrationGetUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresUserRepository(db)

	user := newUser("get-1")
	require.NoError(t, repo.CreateUser(ctx, user))

	t.Run("when looking up by email", func(t *testing.T) {
		t.Run("should return the stored user", func(t *testing.T) {
			found, err := repo.GetUserByEmail(ctx, user.Email)

			require.NoError(t, err)
			assert.Equal(t, user.ID, found.ID)
			assert.Equal(t, user.Email, found.Email)
			assert.Equal(t, user.PasswordHash, found.PasswordHash)
			assert.Equal(t, RoleAdmin, found.Role)
			assert.Equal(t, "Test User", found.FullName)
			assert.Equal(t, int64(1), found.Version)
		})
	})

	t.Run("when looking up by id", func(t *testing.T) {
		t.Run("should return the stored user", func(t *testing.T) {
			found, err := repo.GetUserByID(ctx, user.ID)

			require.NoError(t, err)
			assert.Equal(t, user.Email, found.Email)
		})
	})

	t.Run("when the email does not exist", func(t *testing.T) {
		t.Run("should return a domain NotFound", func(t *testing.T) {
			_, err := repo.GetUserByEmail(ctx, "missing@example.com")

			requireNotFound(t, err)
		})
	})

	t.Run("when the id does not exist", func(t *testing.T) {
		t.Run("should return a domain NotFound", func(t *testing.T) {
			_, err := repo.GetUserByID(ctx, uuid.New().String())

			requireNotFound(t, err)
		})
	})
}

func TestIntegrationUpdateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresUserRepository(db)

	t.Run("when the version matches", func(t *testing.T) {
		t.Run("should persist changes and bump the version", func(t *testing.T) {
			user := newUser("update-ok")
			require.NoError(t, repo.CreateUser(ctx, user))

			user.HasSubscription = true
			user.FailedLoginAttempts = 3
			err := repo.UpdateUser(ctx, user)

			require.NoError(t, err)
			assert.Equal(t, int64(2), user.Version, "version is incremented and reflected back")

			reloaded, err := repo.GetUserByID(ctx, user.ID)
			require.NoError(t, err)
			assert.True(t, reloaded.HasSubscription)
			assert.Equal(t, 3, reloaded.FailedLoginAttempts)
			assert.Equal(t, int64(2), reloaded.Version)
		})
	})

	t.Run("when the version is stale", func(t *testing.T) {
		t.Run("should return a domain Conflict and not modify the row", func(t *testing.T) {
			user := newUser("update-stale")
			require.NoError(t, repo.CreateUser(ctx, user))

			// Simulate a concurrent update that already advanced the version.
			concurrent := *user
			concurrent.FullName = "Winner"
			require.NoError(t, repo.UpdateUser(ctx, &concurrent))

			// The caller still holds version 1 and tries to write.
			user.FullName = "Loser"
			err := repo.UpdateUser(ctx, user)

			requireConflict(t, err)

			reloaded, err := repo.GetUserByID(ctx, user.ID)
			require.NoError(t, err)
			assert.Equal(t, "Winner", reloaded.FullName, "the stale write must not overwrite the row")
		})
	})

	t.Run("when the id does not exist", func(t *testing.T) {
		t.Run("should return a domain NotFound (distinct from a version conflict)", func(t *testing.T) {
			ghost := newUser("update-ghost")
			ghost.Version = 1 // never inserted

			err := repo.UpdateUser(ctx, ghost)

			requireNotFound(t, err)
		})
	})
}

// seedRefreshUser inserts a users row so refresh_tokens.user_id FK is satisfied,
// returning its ID.
func seedRefreshUser(t *testing.T, ctx context.Context, repo UserRepository, suffix string) string {
	t.Helper()
	user := newUser("rt-" + suffix)
	require.NoError(t, repo.CreateUser(ctx, user))
	return user.ID
}

func newRefreshToken(userID, hash string) *RefreshToken {
	return &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour).UTC(),
	}
}

func TestIntegrationRefreshTokenLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	userRepo := NewPostgresUserRepository(db)
	repo := NewPostgresRefreshTokenRepository(db)

	t.Run("when creating and fetching a token", func(t *testing.T) {
		t.Run("should round-trip by hash and populate created_at", func(t *testing.T) {
			userID := seedRefreshUser(t, ctx, userRepo, "roundtrip")
			token := newRefreshToken(userID, "hash-roundtrip")

			require.NoError(t, repo.CreateRefreshToken(ctx, token))
			assert.False(t, token.CreatedAt.IsZero())

			found, err := repo.GetRefreshTokenByHash(ctx, "hash-roundtrip")
			require.NoError(t, err)
			assert.Equal(t, token.ID, found.ID)
			assert.Equal(t, userID, found.UserID)
			assert.Nil(t, found.RevokedAt, "a fresh token is live")
		})
	})

	t.Run("when the hash is unknown", func(t *testing.T) {
		t.Run("should return a domain NotFound", func(t *testing.T) {
			_, err := repo.GetRefreshTokenByHash(ctx, "no-such-hash")

			requireNotFound(t, err)
		})
	})

	t.Run("when the same hash is inserted twice", func(t *testing.T) {
		t.Run("should return a domain Conflict", func(t *testing.T) {
			userID := seedRefreshUser(t, ctx, userRepo, "dup-hash")
			require.NoError(t, repo.CreateRefreshToken(ctx, newRefreshToken(userID, "dup-hash-value")))

			err := repo.CreateRefreshToken(ctx, newRefreshToken(userID, "dup-hash-value"))

			requireConflict(t, err)
		})
	})

	t.Run("when revoking a token", func(t *testing.T) {
		t.Run("should stamp revoked_at and be idempotent on repeat", func(t *testing.T) {
			userID := seedRefreshUser(t, ctx, userRepo, "revoke")
			token := newRefreshToken(userID, "hash-revoke")
			require.NoError(t, repo.CreateRefreshToken(ctx, token))

			require.NoError(t, repo.RevokeRefreshToken(ctx, token.ID))
			revoked, err := repo.GetRefreshTokenByHash(ctx, "hash-revoke")
			require.NoError(t, err)
			require.NotNil(t, revoked.RevokedAt, "revoked_at is stamped")

			// Revoking again is a no-op success.
			require.NoError(t, repo.RevokeRefreshToken(ctx, token.ID))
		})
	})

	t.Run("when revoking an unknown id", func(t *testing.T) {
		t.Run("should be a no-op success (idempotent)", func(t *testing.T) {
			require.NoError(t, repo.RevokeRefreshToken(ctx, uuid.New().String()))
		})
	})

	t.Run("when revoking all tokens for a user", func(t *testing.T) {
		t.Run("should revoke every live token and leave already-revoked ones", func(t *testing.T) {
			userID := seedRefreshUser(t, ctx, userRepo, "revoke-all")
			live1 := newRefreshToken(userID, "hash-all-1")
			live2 := newRefreshToken(userID, "hash-all-2")
			require.NoError(t, repo.CreateRefreshToken(ctx, live1))
			require.NoError(t, repo.CreateRefreshToken(ctx, live2))

			require.NoError(t, repo.RevokeAllRefreshTokens(ctx, userID))

			for _, hash := range []string{"hash-all-1", "hash-all-2"} {
				found, err := repo.GetRefreshTokenByHash(ctx, hash)
				require.NoError(t, err)
				assert.NotNil(t, found.RevokedAt, "token %s should be revoked", hash)
			}

			// A user with no live tokens is a no-op success.
			require.NoError(t, repo.RevokeAllRefreshTokens(ctx, userID))
		})
	})
}
