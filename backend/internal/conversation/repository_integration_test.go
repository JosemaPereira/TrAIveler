package conversation

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
// migrations against it, and returns a ready-to-use database.Client. Mirrors
// the same helper duplicated across every repository package in this
// codebase (trip, subscription, auth, ...) — see
// internal/subscription/repository_integration_test.go's identical version
// and its doc comment for why it is not shared via import.
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

// seedTrip inserts a minimal users row and a minimal trips row
// (conversation_sessions.trip_id FKs trips.id) and returns the trip's
// generated ID.
func seedTrip(t *testing.T, ctx context.Context, db database.Client, suffix string) string {
	t.Helper()

	userID := uuid.New().String()
	_, err := db.Pool().Exec(ctx,
		`INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`,
		userID, "conv-"+suffix+"@example.com", "hash",
	)
	require.NoError(t, err, "failed to seed user")

	tripID := uuid.New().String()
	_, err = db.Pool().Exec(ctx,
		`INSERT INTO trips (id, creator_id, title, status) VALUES ($1, $2, $3, 'draft')`,
		tripID, userID, "Test Trip "+suffix,
	)
	require.NoError(t, err, "failed to seed trip")

	return tripID
}

// newTestSession builds a Session for tripID with a fresh ID, mirroring the
// future Conversation service's UUID generation (repository.CreateSession
// requires ID to already be set).
func newTestSession(tripID string) *Session {
	return &Session{
		ID:     uuid.New().String(),
		TripID: tripID,
		Status: StatusInProgress,
	}
}

func TestIntegrationCreateSession_ValidSession_PopulatesGeneratedFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	tripID := seedTrip(t, ctx, db, "create-1")

	session := newTestSession(tripID)
	err := repo.CreateSession(ctx, session)

	require.NoError(t, err)
	assert.False(t, session.StartedAt.IsZero())
	assert.False(t, session.CreatedAt.IsZero())
	assert.Equal(t, "anthropic-claude", session.AIProvider,
		"expected the database default ai_provider to be populated back onto the struct")
	assert.Zero(t, session.TotalTokens)
	assert.Nil(t, session.CompletedAt)
}

func TestIntegrationGetSessionByTrip_ExistingTrip_ReturnsMostRecentSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	tripID := seedTrip(t, ctx, db, "get-1")

	first := newTestSession(tripID)
	require.NoError(t, repo.CreateSession(ctx, first))
	time.Sleep(10 * time.Millisecond) // ensure a distinguishable started_at ordering
	second := newTestSession(tripID)
	require.NoError(t, repo.CreateSession(ctx, second))

	found, err := repo.GetSessionByTrip(ctx, tripID)

	require.NoError(t, err)
	assert.Equal(t, second.ID, found.ID, "expected the most recently started session")
}

func TestIntegrationGetSessionByTrip_NoSessions_ReturnsNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	tripID := seedTrip(t, ctx, db, "get-empty")

	_, err := repo.GetSessionByTrip(ctx, tripID)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationAppendMessage_ValidMessage_PopulatesGeneratedFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	tripID := seedTrip(t, ctx, db, "append-1")
	session := newTestSession(tripID)
	require.NoError(t, repo.CreateSession(ctx, session))

	msg := &Message{
		SessionID: session.ID,
		Role:      RoleUser,
		Content:   "Plan me a trip to Tokyo",
	}
	err := repo.AppendMessage(ctx, msg)

	require.NoError(t, err)
	assert.NotEmpty(t, msg.ID)
	assert.False(t, msg.Timestamp.IsZero())
}

func TestIntegrationListMessages_MultipleMessages_ReturnsThemInChronologicalOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	tripID := seedTrip(t, ctx, db, "list-1")
	session := newTestSession(tripID)
	require.NoError(t, repo.CreateSession(ctx, session))

	first := &Message{SessionID: session.ID, Role: RoleSystem, Content: "System prompt"}
	require.NoError(t, repo.AppendMessage(ctx, first))
	time.Sleep(10 * time.Millisecond)
	second := &Message{SessionID: session.ID, Role: RoleUser, Content: "Plan me a trip"}
	require.NoError(t, repo.AppendMessage(ctx, second))
	time.Sleep(10 * time.Millisecond)
	third := &Message{SessionID: session.ID, Role: RoleAssistant, Content: "Sure, where to?"}
	require.NoError(t, repo.AppendMessage(ctx, third))

	messages, err := repo.ListMessages(ctx, session.ID)

	require.NoError(t, err)
	require.Len(t, messages, 3)
	assert.Equal(t, first.ID, messages[0].ID)
	assert.Equal(t, second.ID, messages[1].ID)
	assert.Equal(t, third.ID, messages[2].ID)
}

func TestIntegrationListMessages_NoMessages_ReturnsEmptySlice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	tripID := seedTrip(t, ctx, db, "list-empty")
	session := newTestSession(tripID)
	require.NoError(t, repo.CreateSession(ctx, session))

	messages, err := repo.ListMessages(ctx, session.ID)

	require.NoError(t, err)
	assert.Empty(t, messages)
}
