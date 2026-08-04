package trip

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
// codebase (subscription, auth, ...) — see
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

// seedUser inserts a minimal users row (trips.creator_id FKs users.id) and
// returns its generated ID. Only the NOT NULL columns without a default are
// supplied; role/has_subscription/version fall back to their defaults.
func seedUser(t *testing.T, ctx context.Context, db database.Client, suffix string) string {
	t.Helper()
	id := uuid.New().String()
	_, err := db.Pool().Exec(ctx,
		`INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`,
		id, "trip-"+suffix+"@example.com", "hash",
	)
	require.NoError(t, err, "failed to seed user")
	return id
}

// newTestTrip builds a Trip for creatorID with a fresh ID, mirroring
// Service.Create's UUID generation (repository.CreateTrip requires ID to
// already be set).
func newTestTrip(creatorID, suffix string) *Trip {
	return &Trip{
		ID:        uuid.New().String(),
		CreatorID: creatorID,
		Title:     "Test Trip " + suffix,
		Status:    TripStatusDraft,
	}
}

// newTestDestination builds a Destination with a fresh ID; CreateDestination
// overwrites ID/CreatedAt from the database defaults, so the caller-supplied
// ID here only needs to be non-empty for pre-assignment convenience.
func newTestDestination(suffix string) *Destination {
	return &Destination{
		Name:      "Test City " + suffix,
		Country:   "US",
		Latitude:  40.7128,
		Longitude: -74.0060,
	}
}

func TestIntegrationCreateTrip_ValidTrip_PopulatesGeneratedFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "create-1")

	tr := newTestTrip(creatorID, "create-1")
	err := repo.CreateTrip(ctx, tr)

	require.NoError(t, err)
	assert.False(t, tr.CreatedAt.IsZero())
	assert.False(t, tr.UpdatedAt.IsZero())
	assert.Equal(t, 1, tr.Version)
	assert.False(t, tr.Archived)
}

func TestIntegrationFindTripByID_ExistingID_ReturnsTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "find-1")

	created := newTestTrip(creatorID, "find-1")
	require.NoError(t, repo.CreateTrip(ctx, created))

	found, err := repo.FindTripByID(ctx, created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Title, found.Title)
	assert.Equal(t, created.CreatorID, found.CreatorID)
	assert.Equal(t, TripStatusDraft, found.Status)
}

func TestIntegrationFindTripByID_MissingID_ReturnsNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	_, err := repo.FindTripByID(ctx, "00000000-0000-0000-0000-000000000000")

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationListTripsByUser_ReturnsOnlyThatUsersTrips(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	ownerID := seedUser(t, ctx, db, "list-owner")
	otherID := seedUser(t, ctx, db, "list-other")

	first := newTestTrip(ownerID, "list-1")
	require.NoError(t, repo.CreateTrip(ctx, first))
	second := newTestTrip(ownerID, "list-2")
	require.NoError(t, repo.CreateTrip(ctx, second))
	other := newTestTrip(otherID, "list-other")
	require.NoError(t, repo.CreateTrip(ctx, other))

	trips, err := repo.ListTripsByUser(ctx, ownerID)

	require.NoError(t, err)
	require.Len(t, trips, 2)
	for _, tr := range trips {
		assert.Equal(t, ownerID, tr.CreatorID)
	}
}

func TestIntegrationListTripsByUser_NoTrips_ReturnsEmptySlice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	userID := seedUser(t, ctx, db, "list-empty")

	trips, err := repo.ListTripsByUser(ctx, userID)

	require.NoError(t, err)
	assert.Empty(t, trips)
}

func TestIntegrationUpdateTrip_MatchingVersion_UpdatesAndIncrementsVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "update-1")
	created := newTestTrip(creatorID, "update-1")
	require.NoError(t, repo.CreateTrip(ctx, created))

	created.Title = "Updated Title"
	created.Status = TripStatusPublished
	err := repo.UpdateTrip(ctx, created)

	require.NoError(t, err)
	assert.Equal(t, 2, created.Version)

	reloaded, err := repo.FindTripByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", reloaded.Title)
	assert.Equal(t, TripStatusPublished, reloaded.Status)
	assert.Equal(t, 2, reloaded.Version)
}

func TestIntegrationUpdateTrip_StaleVersion_ReturnsConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "update-conflict")
	created := newTestTrip(creatorID, "update-conflict")
	require.NoError(t, repo.CreateTrip(ctx, created))

	stale := *created
	stale.Version = created.Version + 99 // simulate a version that no longer matches
	stale.Title = "Stale Update"

	err := repo.UpdateTrip(ctx, &stale)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "conflict", domainErr.Code)
}

func TestIntegrationUpdateTrip_MissingID_ReturnsNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	missing := &Trip{
		ID:      "00000000-0000-0000-0000-000000000000",
		Title:   "Ghost Trip",
		Status:  TripStatusDraft,
		Version: 1,
	}
	err := repo.UpdateTrip(ctx, missing)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationDeleteTrip_ExistingID_RemovesRow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "delete-1")

	created := newTestTrip(creatorID, "delete-1")
	require.NoError(t, repo.CreateTrip(ctx, created))

	err := repo.DeleteTrip(ctx, created.ID)
	require.NoError(t, err)

	_, err = repo.FindTripByID(ctx, created.ID)
	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationDeleteTrip_MissingID_ReturnsNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	err := repo.DeleteTrip(ctx, "00000000-0000-0000-0000-000000000000")

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationCreateDestination_ValidDestination_PopulatesGeneratedFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	dest := newTestDestination("dest-1")
	err := repo.CreateDestination(ctx, dest)

	require.NoError(t, err)
	assert.NotEmpty(t, dest.ID)
	assert.False(t, dest.CreatedAt.IsZero())
}

func TestIntegrationCreateDestination_DuplicateCoordinates_AllowedByDesign(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	first := newTestDestination("dup")
	require.NoError(t, repo.CreateDestination(ctx, first))
	second := newTestDestination("dup")
	err := repo.CreateDestination(ctx, second)

	require.NoError(t, err, "destinations table has no uniqueness constraint by design")
	assert.NotEqual(t, first.ID, second.ID)
}

func TestIntegrationUpsertDay_NewDayNumber_InsertsRow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "day-insert")
	tr := newTestTrip(creatorID, "day-insert")
	require.NoError(t, repo.CreateTrip(ctx, tr))

	day := &Day{TripID: tr.ID, DayNumber: 1}
	err := repo.UpsertDay(ctx, day)

	require.NoError(t, err)
	assert.NotEmpty(t, day.ID)
	assert.False(t, day.CreatedAt.IsZero())
}

func TestIntegrationUpsertDay_ExistingDayNumber_UpdatesRowInPlace(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "day-update")
	tr := newTestTrip(creatorID, "day-update")
	require.NoError(t, repo.CreateTrip(ctx, tr))

	first := &Day{TripID: tr.ID, DayNumber: 1}
	require.NoError(t, repo.UpsertDay(ctx, first))

	label := "Updated Label"
	second := &Day{TripID: tr.ID, DayNumber: 1, Label: &label}
	err := repo.UpsertDay(ctx, second)

	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID, "upserting the same (trip_id, day_number) must update, not duplicate")

	var count int
	require.NoError(t, db.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM days WHERE trip_id = $1 AND day_number = 1`, tr.ID,
	).Scan(&count))
	assert.Equal(t, 1, count, "expected exactly one row for (trip_id, day_number)")
}

func TestIntegrationUpsertActivity_NoID_InsertsRowWithVersionOne(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "activity-insert")
	tr := newTestTrip(creatorID, "activity-insert")
	require.NoError(t, repo.CreateTrip(ctx, tr))
	day := &Day{TripID: tr.ID, DayNumber: 1}
	require.NoError(t, repo.UpsertDay(ctx, day))

	activity := &Activity{
		DayID:         day.ID,
		Title:         "Visit Museum",
		Type:          ActivityTypeVisit,
		SequenceOrder: 1,
	}
	err := repo.UpsertActivity(ctx, activity)

	require.NoError(t, err)
	assert.NotEmpty(t, activity.ID)
	assert.Equal(t, 1, activity.Version)
	assert.False(t, activity.CreatedAt.IsZero())
	assert.False(t, activity.UpdatedAt.IsZero())
}

func TestIntegrationUpsertActivity_ExistingIDMatchingVersion_UpdatesAndIncrementsVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "activity-update")
	tr := newTestTrip(creatorID, "activity-update")
	require.NoError(t, repo.CreateTrip(ctx, tr))
	day := &Day{TripID: tr.ID, DayNumber: 1}
	require.NoError(t, repo.UpsertDay(ctx, day))

	created := &Activity{
		DayID:         day.ID,
		Title:         "Visit Museum",
		Type:          ActivityTypeVisit,
		SequenceOrder: 1,
	}
	require.NoError(t, repo.UpsertActivity(ctx, created))

	created.Title = "Visit Updated Museum"
	created.SequenceOrder = 2
	err := repo.UpsertActivity(ctx, created)

	require.NoError(t, err)
	assert.Equal(t, 2, created.Version)

	reloaded := &Activity{}
	require.NoError(t, db.Pool().QueryRow(ctx,
		`SELECT title, sequence_order, version FROM activities WHERE id = $1`, created.ID,
	).Scan(&reloaded.Title, &reloaded.SequenceOrder, &reloaded.Version))
	assert.Equal(t, "Visit Updated Museum", reloaded.Title)
	assert.Equal(t, 2, reloaded.SequenceOrder)
	assert.Equal(t, 2, reloaded.Version)
}

func TestIntegrationUpsertActivity_ExistingIDStaleVersion_ReturnsConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "activity-conflict")
	tr := newTestTrip(creatorID, "activity-conflict")
	require.NoError(t, repo.CreateTrip(ctx, tr))
	day := &Day{TripID: tr.ID, DayNumber: 1}
	require.NoError(t, repo.UpsertDay(ctx, day))

	created := &Activity{
		DayID:         day.ID,
		Title:         "Visit Museum",
		Type:          ActivityTypeVisit,
		SequenceOrder: 1,
	}
	require.NoError(t, repo.UpsertActivity(ctx, created))

	stale := *created
	stale.Version = created.Version + 99 // simulate a version that no longer matches

	err := repo.UpsertActivity(ctx, &stale)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "conflict", domainErr.Code)
}

func TestIntegrationDeleteActivity_ExistingID_RemovesRow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "activity-delete")
	tr := newTestTrip(creatorID, "activity-delete")
	require.NoError(t, repo.CreateTrip(ctx, tr))
	day := &Day{TripID: tr.ID, DayNumber: 1}
	require.NoError(t, repo.UpsertDay(ctx, day))
	activity := &Activity{DayID: day.ID, Title: "Visit Museum", Type: ActivityTypeVisit, SequenceOrder: 1}
	require.NoError(t, repo.UpsertActivity(ctx, activity))

	err := repo.DeleteActivity(ctx, activity.ID)
	require.NoError(t, err)

	var count int
	require.NoError(t, db.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM activities WHERE id = $1`, activity.ID,
	).Scan(&count))
	assert.Zero(t, count)
}

func TestIntegrationDeleteActivity_MissingID_ReturnsNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	err := repo.DeleteActivity(ctx, "00000000-0000-0000-0000-000000000000")

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestIntegrationListDaysByTrip_NoDays_ReturnsEmptySlice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "days-empty")
	tr := newTestTrip(creatorID, "days-empty")
	require.NoError(t, repo.CreateTrip(ctx, tr))

	days, err := repo.ListDaysByTrip(ctx, tr.ID)

	require.NoError(t, err)
	assert.Empty(t, days)
}

func TestIntegrationListDaysByTrip_MultipleDays_ReturnsOrderedByDayNumber(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "days-order")
	tr := newTestTrip(creatorID, "days-order")
	require.NoError(t, repo.CreateTrip(ctx, tr))
	third := &Day{TripID: tr.ID, DayNumber: 3}
	require.NoError(t, repo.UpsertDay(ctx, third))
	first := &Day{TripID: tr.ID, DayNumber: 1}
	require.NoError(t, repo.UpsertDay(ctx, first))
	second := &Day{TripID: tr.ID, DayNumber: 2}
	require.NoError(t, repo.UpsertDay(ctx, second))

	days, err := repo.ListDaysByTrip(ctx, tr.ID)

	require.NoError(t, err)
	require.Len(t, days, 3)
	assert.Equal(t, []int{1, 2, 3}, []int{days[0].DayNumber, days[1].DayNumber, days[2].DayNumber})
}

func TestIntegrationListActivitiesByDayIDs_EmptyDayIDs_ReturnsEmptySliceWithoutQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	activities, err := repo.ListActivitiesByDayIDs(ctx, []string{})

	require.NoError(t, err)
	assert.Empty(t, activities)
}

func TestIntegrationListActivitiesByDayIDs_MultipleDays_ReturnsGroupedAndOrdered(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	creatorID := seedUser(t, ctx, db, "activities-bulk")
	tr := newTestTrip(creatorID, "activities-bulk")
	require.NoError(t, repo.CreateTrip(ctx, tr))
	dayOne := &Day{TripID: tr.ID, DayNumber: 1}
	require.NoError(t, repo.UpsertDay(ctx, dayOne))
	dayTwo := &Day{TripID: tr.ID, DayNumber: 2}
	require.NoError(t, repo.UpsertDay(ctx, dayTwo))

	dayOneSecond := &Activity{DayID: dayOne.ID, Title: "Second Stop", Type: ActivityTypeVisit, SequenceOrder: 2}
	require.NoError(t, repo.UpsertActivity(ctx, dayOneSecond))
	dayOneFirst := &Activity{DayID: dayOne.ID, Title: "First Stop", Type: ActivityTypeVisit, SequenceOrder: 1}
	require.NoError(t, repo.UpsertActivity(ctx, dayOneFirst))
	dayTwoFirst := &Activity{DayID: dayTwo.ID, Title: "Day Two Stop", Type: ActivityTypeVisit, SequenceOrder: 1}
	require.NoError(t, repo.UpsertActivity(ctx, dayTwoFirst))

	activities, err := repo.ListActivitiesByDayIDs(ctx, []string{dayOne.ID, dayTwo.ID})

	require.NoError(t, err)
	require.Len(t, activities, 3, "must bulk-fetch across every requested day ID in one call")

	// day_id is a random UUID, so which day's rows sort first in the
	// day_id-then-sequence_order ORDER BY is not deterministic from
	// creation order — only group and check ordering within each group.
	byDay := map[string][]*Activity{}
	for _, a := range activities {
		byDay[a.DayID] = append(byDay[a.DayID], a)
	}
	require.Len(t, byDay[dayOne.ID], 2)
	assert.Equal(t, "First Stop", byDay[dayOne.ID][0].Title, "within a day, activities must be ordered by sequence_order")
	assert.Equal(t, "Second Stop", byDay[dayOne.ID][1].Title)
	require.Len(t, byDay[dayTwo.ID], 1)
	assert.Equal(t, "Day Two Stop", byDay[dayTwo.ID][0].Title)

	// Confirm the query's actual ORDER BY day_id, sequence_order clause: rows
	// for the same day_id must be contiguous and internally sequence_order-
	// ascending, regardless of which day comes first.
	for i := 1; i < len(activities); i++ {
		if activities[i].DayID == activities[i-1].DayID {
			assert.LessOrEqual(t, activities[i-1].SequenceOrder, activities[i].SequenceOrder)
		}
	}
}

func TestIntegrationListDestinationsByIDs_EmptyIDs_ReturnsEmptySliceWithoutQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)

	destinations, err := repo.ListDestinationsByIDs(ctx, []string{})

	require.NoError(t, err)
	assert.Empty(t, destinations)
}

func TestIntegrationListDestinationsByIDs_MultipleIDs_ReturnsAllMatchingRows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := setupRepositoryTestDB(t, ctx)
	repo := NewPostgresRepository(db)
	first := newTestDestination("bulk-1")
	require.NoError(t, repo.CreateDestination(ctx, first))
	second := newTestDestination("bulk-2")
	require.NoError(t, repo.CreateDestination(ctx, second))
	// A third, unrelated destination confirms the query filters by id, not returning everything.
	third := newTestDestination("bulk-3")
	require.NoError(t, repo.CreateDestination(ctx, third))

	destinations, err := repo.ListDestinationsByIDs(ctx, []string{first.ID, second.ID})

	require.NoError(t, err)
	require.Len(t, destinations, 2)
	gotIDs := []string{destinations[0].ID, destinations[1].ID}
	assert.ElementsMatch(t, []string{first.ID, second.ID}, gotIDs)
}
