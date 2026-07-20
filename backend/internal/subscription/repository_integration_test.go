package subscription

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
)

// basicPlanID is the fixed UUID of the seeded 'basic' plan
// (migrations/006_create_plans_table.sql).
const basicPlanID = "00000000-0000-0000-0000-000000000001"

// setupRepositoryTestDB starts a PostgreSQL testcontainer, applies goose
// migrations, and returns a ready database.Client. Mirrors
// internal/example/repository_integration_test.go's helper and its
// testing.Short() skip convention.
func setupRepositoryTestDB(t *testing.T, ctx context.Context) database.Client {
	t.Helper()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
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

// seedUser inserts a minimal users row (subscriptions.user_id FKs users.id)
// and returns its generated ID. Only the NOT NULL columns without a default
// are supplied; role/has_subscription/version fall back to their defaults.
func seedUser(t *testing.T, ctx context.Context, db database.Client, emailSuffix string) string {
	t.Helper()
	id := uuid.New().String()
	_, err := db.Pool().Exec(ctx,
		`INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`,
		id, "sub-"+emailSuffix+"@example.com", "hash",
	)
	require.NoError(t, err, "failed to seed user")
	return id
}

// newActiveSubscription builds an active Subscription for userID with a fresh
// ID, mirroring the service layer's UUID generation before Create.
func newActiveSubscription(userID string) *Subscription {
	ref := "tok_" + userID[:8]
	return &Subscription{
		ID:             uuid.New().String(),
		UserID:         userID,
		PlanID:         basicPlanID,
		Status:         StatusActive,
		StubPaymentRef: &ref,
	}
}

func TestIntegrationCreate_ValidSubscription_PopulatesCreatedAt(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	sub := newActiveSubscription(seedUser(t, ctx, db, "create-1"))
	err := repo.Create(ctx, sub)

	require.NoError(t, err)
	assert.False(t, sub.CreatedAt.IsZero())
}

func TestIntegrationCreate_DuplicateUser_ReturnsConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	userID := seedUser(t, ctx, db, "dup")
	require.NoError(t, repo.Create(ctx, newActiveSubscription(userID)))

	err := repo.Create(ctx, newActiveSubscription(userID))

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "conflict", domainErr.Code)
}

func TestIntegrationGetByUserID_Existing_ReturnsSubscription(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	userID := seedUser(t, ctx, db, "get-1")
	created := newActiveSubscription(userID)
	require.NoError(t, repo.Create(ctx, created))

	found, err := repo.GetByUserID(ctx, userID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, userID, found.UserID)
	assert.Equal(t, StatusActive, found.Status)
	require.NotNil(t, found.StubPaymentRef)
	assert.Equal(t, *created.StubPaymentRef, *found.StubPaymentRef)
}

func TestIntegrationGetByUserID_Missing_ReturnsNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	_, err := repo.GetByUserID(ctx, "00000000-0000-0000-0000-000000000000")

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationUpdate_ExistingRow_PersistsChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	userID := seedUser(t, ctx, db, "update-1")
	sub := newActiveSubscription(userID)
	require.NoError(t, repo.Create(ctx, sub))

	graceEnd := time.Now().Add(30 * 24 * time.Hour).UTC()
	sub.Status = StatusCancelled
	sub.GracePeriodEndsAt = &graceEnd
	err := repo.Update(ctx, sub)

	require.NoError(t, err)
	reloaded, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, StatusCancelled, reloaded.Status)
	require.NotNil(t, reloaded.GracePeriodEndsAt)
	assert.WithinDuration(t, graceEnd, *reloaded.GracePeriodEndsAt, time.Second)
}

func TestIntegrationUpdate_MissingRow_ReturnsNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	sub := newActiveSubscription(seedUser(t, ctx, db, "update-missing"))
	// sub.ID never inserted.
	err := repo.Update(ctx, sub)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationCancel_ActiveSubscription_SetsCancelledAndGrace(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	userID := seedUser(t, ctx, db, "cancel-1")
	sub := newActiveSubscription(userID)
	require.NoError(t, repo.Create(ctx, sub))

	err := repo.Cancel(ctx, sub.ID)
	require.NoError(t, err)

	reloaded, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, StatusCancelled, reloaded.Status)
	require.NotNil(t, reloaded.CancelledAt)
	require.NotNil(t, reloaded.GracePeriodEndsAt)
	// Grace period is 30 days out from cancellation (BR-001).
	assert.WithinDuration(t, time.Now().Add(gracePeriodDays*24*time.Hour), *reloaded.GracePeriodEndsAt, time.Minute)
	assert.False(t, reloaded.IsActive(reloaded.GracePeriodEndsAt.Add(time.Hour)))
	assert.True(t, reloaded.IsActive(time.Now()))
}

func TestIntegrationCancel_NonActiveSubscription_ReturnsNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	userID := seedUser(t, ctx, db, "cancel-twice")
	sub := newActiveSubscription(userID)
	require.NoError(t, repo.Create(ctx, sub))
	require.NoError(t, repo.Cancel(ctx, sub.ID))

	// Second cancel finds no *active* row with that ID.
	err := repo.Cancel(ctx, sub.ID)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}
