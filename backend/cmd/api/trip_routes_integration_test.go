//go:build test

package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
	"github.com/JosemaPereira/TrAIveler/backend/internal/testdb"
	"github.com/JosemaPereira/TrAIveler/backend/internal/trip"
)

// basicPlanID is the fixed UUID of the seeded 'basic' plan
// (migrations/006_create_plans_table.sql), mirroring
// internal/subscription/repository_integration_test.go's identical constant.
const basicPlanID = "00000000-0000-0000-0000-000000000001"

// setupIntegrationTestDB starts a PostgreSQL testcontainer, applies goose
// migrations, and returns a ready database.Client. Mirrors the same helper
// duplicated across every repository package in this codebase (trip,
// conversation, subscription, ...) and its testing.Short() skip convention
// (docs/testing-guidelines.md's "Repository Integration Tests with Optional
// Testcontainers"). This is cmd/api's first use of it: it replaces the
// removed extraProtectedRoutes probe with a real end-to-end assertion that
// the subscription gate on PUT/DELETE /trips/{id} actually works when wired
// through the real HTTPServer, real trip.Service, and a real database —
// something a mocked database.Client (routes_test.go) cannot exercise, since
// trip.Service.Update/Delete call the repository before the subscription
// check ever runs.
func setupIntegrationTestDB(t *testing.T, ctx context.Context) database.Client {
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

// seedGatedUser inserts a users row whose id is the fixed gatedUserID
// constant (routes_test.go), so the seeded row's id matches the fake
// TokenValidator's claimed user id — mirrors
// internal/trip/repository_integration_test.go's seedUser, hardcoded to
// gatedUserID rather than generating one, since every test in this file
// needs the two to agree.
func seedGatedUser(t *testing.T, ctx context.Context, db database.Client, emailSuffix string) {
	t.Helper()
	_, err := db.Pool().Exec(ctx,
		`INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`,
		gatedUserID, "gated-"+emailSuffix+"@example.com", "hash",
	)
	require.NoError(t, err, "failed to seed user")
}

// seedTripForGatedUser persists a draft trip owned by gatedUserID via the
// real trip.Repository, mirroring
// internal/trip/repository_integration_test.go's newTestTrip/CreateTrip
// usage, and returns it with its DB-assigned version/timestamps populated.
func seedTripForGatedUser(t *testing.T, ctx context.Context, db database.Client) *trip.Trip {
	t.Helper()
	tr := &trip.Trip{
		ID:        uuid.New().String(),
		CreatorID: gatedUserID,
		Title:     "Original Title",
		Status:    trip.TripStatusDraft,
	}
	require.NoError(t, trip.NewPostgresRepository(db).CreateTrip(ctx, tr))
	return tr
}

// seedActiveSubscription persists an active subscription for userID via the
// real subscription.Repository, mirroring
// internal/subscription/repository_integration_test.go's
// newActiveSubscription/Create usage.
func seedActiveSubscription(t *testing.T, ctx context.Context, db database.Client, userID string) {
	t.Helper()
	ref := "tok_integration"
	sub := &subscription.Subscription{
		ID:             uuid.New().String(),
		UserID:         userID,
		PlanID:         basicPlanID,
		Status:         subscription.StatusActive,
		StubPaymentRef: &ref,
	}
	require.NoError(t, subscription.NewPostgresRepository(db).Create(ctx, sub))
}

func doGatedTripRequest(
	t *testing.T, srv *HTTPServer, method, path, body, ifMatch string,
) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: anyAccessTokenVal})
	if ifMatch != "" {
		req.Header.Set("If-Match", ifMatch)
	}
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)
	return rec
}

// TestHTTPServer_TripRoutes_UpdateWithoutActiveSubscription_Returns403Forbidden
// is the real-database counterpart of the removed
// extraProtectedRoutes/withProtectedRoutes probe: it proves the Authenticate
// gate really propagates the caller's identity all the way through the real
// HTTPServer -> trip.Handler -> trip.Service -> PostgresRepository stack,
// using PUT /trips/{id} (a genuine, production subscription-gated route)
// instead of a test-only probe route.
func TestHTTPServer_TripRoutes_UpdateWithoutActiveSubscription_Returns403Forbidden(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupIntegrationTestDB(t, ctx)
	seedGatedUser(t, ctx, db, "update-no-sub")
	tr := seedTripForGatedUser(t, ctx, db)
	// Deliberately no subscription row seeded for gatedUserID.

	srv, validator := newGatedServer(t, db)

	rec := doGatedTripRequest(t, srv, http.MethodPut, "/api/v1/trips/"+tr.ID,
		`{"title":"Updated Title","status":"draft"}`, "1")

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "forbidden", decodeErrorCode(t, rec))
	assert.Equal(t, 1, validator.callCount, "a presented cookie must be validated")
}

// TestHTTPServer_TripRoutes_UpdateWithActiveSubscription_ReachesServiceAndPersists
// is the positive counterpart: an active subscription lets the request reach
// trip.Service.Update for real, and the update is actually persisted.
func TestHTTPServer_TripRoutes_UpdateWithActiveSubscription_ReachesServiceAndPersists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupIntegrationTestDB(t, ctx)
	seedGatedUser(t, ctx, db, "update-active-sub")
	tr := seedTripForGatedUser(t, ctx, db)
	seedActiveSubscription(t, ctx, db, gatedUserID)

	srv, _ := newGatedServer(t, db)

	rec := doGatedTripRequest(t, srv, http.MethodPut, "/api/v1/trips/"+tr.ID,
		`{"title":"Updated Title","status":"published"}`, "1")

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	updated := body["trip"].(map[string]any)
	assert.Equal(t, "Updated Title", updated["title"])
	assert.Equal(t, "published", updated["status"])
	assert.Equal(t, float64(2), updated["version"], "a successful update must increment the version")

	// Confirm the write really reached PostgreSQL, not just the response body.
	persisted, err := trip.NewPostgresRepository(db).FindTripByID(ctx, tr.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", persisted.Title)
}

// TestHTTPServer_TripRoutes_DeleteWithoutActiveSubscription_Returns403Forbidden
// mirrors the Update test above for DELETE, the other subscription-gated
// write.
func TestHTTPServer_TripRoutes_DeleteWithoutActiveSubscription_Returns403Forbidden(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupIntegrationTestDB(t, ctx)
	seedGatedUser(t, ctx, db, "delete-no-sub")
	tr := seedTripForGatedUser(t, ctx, db)

	srv, _ := newGatedServer(t, db)

	rec := doGatedTripRequest(t, srv, http.MethodDelete, "/api/v1/trips/"+tr.ID, "", "")

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "forbidden", decodeErrorCode(t, rec))

	// The trip must still exist: Forbidden must short-circuit before DeleteTrip.
	_, err := trip.NewPostgresRepository(db).FindTripByID(ctx, tr.ID)
	require.NoError(t, err)
}

// TestHTTPServer_TripRoutes_DeleteWithActiveSubscription_Returns204AndDeletes
// is the positive DELETE counterpart.
func TestHTTPServer_TripRoutes_DeleteWithActiveSubscription_Returns204AndDeletes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupIntegrationTestDB(t, ctx)
	seedGatedUser(t, ctx, db, "delete-active-sub")
	tr := seedTripForGatedUser(t, ctx, db)
	seedActiveSubscription(t, ctx, db, gatedUserID)

	srv, _ := newGatedServer(t, db)

	rec := doGatedTripRequest(t, srv, http.MethodDelete, "/api/v1/trips/"+tr.ID, "", "")

	require.Equal(t, http.StatusNoContent, rec.Code)

	_, err := trip.NewPostgresRepository(db).FindTripByID(ctx, tr.ID)
	require.Error(t, err, "the trip must actually be gone from PostgreSQL")
}
