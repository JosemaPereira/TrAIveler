package example

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

// setupRepositoryTestDB starts a PostgreSQL testcontainer, applies goose
// migrations against it, and returns a ready-to-use database.Client. Mirrors
// internal/database/client_test.go's setupPostgresContainer helper and skip
// convention (testing.Short()) so `make test` stays fast and `make test-all`
// (Colima) covers this.
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

// newTestExample builds an Example with a caller-supplied ID, mirroring the
// service layer's UUID generation (repository.Create requires ID to already
// be set — see repository.go's Create doc comment).
func newTestExample(t *testing.T, suffix string) *Example {
	t.Helper()
	return &Example{
		ID:     uuid.New().String(),
		Name:   "Test Example " + suffix,
		Email:  "test-" + suffix + "@example.com",
		Status: StatusActive,
		Count:  0,
	}
}

func TestIntegrationCreate_ValidExample_PopulatesGeneratedFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	ex := newTestExample(t, "create-1")
	err := repo.Create(ctx, ex)

	require.NoError(t, err)
	assert.NotEmpty(t, ex.ID)
	assert.False(t, ex.CreatedAt.IsZero())
	assert.False(t, ex.UpdatedAt.IsZero())
	assert.Equal(t, 1, ex.Version)
}

func TestIntegrationCreate_DuplicateEmail_ReturnsConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	first := newTestExample(t, "dup")
	require.NoError(t, repo.Create(ctx, first))

	second := newTestExample(t, "dup")
	err := repo.Create(ctx, second)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "conflict", domainErr.Code)
}

func TestIntegrationFindByID_ExistingID_ReturnsExample(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	created := newTestExample(t, "find-1")
	require.NoError(t, repo.Create(ctx, created))

	found, err := repo.FindByID(ctx, created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Name, found.Name)
	assert.Equal(t, created.Email, found.Email)
}

func TestIntegrationFindByID_MissingID_ReturnsNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	_, err := repo.FindByID(ctx, "00000000-0000-0000-0000-000000000000")

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationFindByEmail_MissingEmail_ReturnsNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	_, err := repo.FindByEmail(ctx, "nobody@example.com")

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationUpdate_MatchingVersion_AppliesChangesAndIncrementsVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	created := newTestExample(t, "update-1")
	require.NoError(t, repo.Create(ctx, created))

	created.Name = "Updated Name"
	created.Count = 42
	err := repo.Update(ctx, created)

	require.NoError(t, err)
	assert.Equal(t, 2, created.Version)

	reloaded, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", reloaded.Name)
	assert.Equal(t, 42, reloaded.Count)
}

func TestIntegrationUpdate_StaleVersion_ReturnsConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	created := newTestExample(t, "update-conflict")
	require.NoError(t, repo.Create(ctx, created))

	stale := *created
	stale.Version = created.Version + 99 // simulate a version that no longer matches

	err := repo.Update(ctx, &stale)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "conflict", domainErr.Code)
}

func TestIntegrationDelete_ExistingID_RemovesRow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	created := newTestExample(t, "delete-1")
	require.NoError(t, repo.Create(ctx, created))

	err := repo.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, created.ID)
	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationDelete_MissingID_ReturnsNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	err := repo.Delete(ctx, "00000000-0000-0000-0000-000000000000")

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationList_PaginatesAndFiltersByStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	for i := 0; i < 3; i++ {
		ex := newTestExample(t, "list-active-"+string(rune('a'+i)))
		require.NoError(t, repo.Create(ctx, ex))
	}
	inactive := newTestExample(t, "list-inactive")
	inactive.Status = StatusInactive
	require.NoError(t, repo.Create(ctx, inactive))

	t.Run("when filtering by status it should return only matching rows", func(t *testing.T) {
		result, err := repo.List(ctx, ListFilters{Status: StatusInactive, Page: 1, PerPage: 20})
		require.NoError(t, err)
		assert.Equal(t, 1, result.Total)
		require.Len(t, result.Items, 1)
		assert.Equal(t, inactive.ID, result.Items[0].ID)
	})

	t.Run("when paginating it should split rows across pages with a stable total", func(t *testing.T) {
		firstPage, err := repo.List(ctx, ListFilters{Page: 1, PerPage: 2})
		require.NoError(t, err)
		assert.Equal(t, 4, firstPage.Total)
		assert.Len(t, firstPage.Items, 2)

		secondPage, err := repo.List(ctx, ListFilters{Page: 2, PerPage: 2})
		require.NoError(t, err)
		assert.Equal(t, 4, secondPage.Total)
		assert.Len(t, secondPage.Items, 2)
	})

	t.Run("when the page is out of range it should return an empty result set", func(t *testing.T) {
		result, err := repo.List(ctx, ListFilters{Page: 99, PerPage: 20})
		require.NoError(t, err)
		assert.Equal(t, 4, result.Total)
		assert.Empty(t, result.Items)
	})
}
