package conversation

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// Repository defines data access for Session and Message (001-T034).
// PostgresRepository is the only implementation; tests use the generated
// mock (mocks/repository_mock.go). Method names follow
// specs/001-product-vision-scope/tasks.md's T034 literal list.
type Repository interface {
	// CreateSession inserts session, which must already have ID/TripID/
	// Status populated by the caller. Populates StartedAt/TotalTokens/
	// AIProvider/CreatedAt from the database defaults. Plain insert: the
	// "only one in_progress session per trip" business rule is owned by the
	// future Conversation service, not this repository.
	CreateSession(ctx context.Context, session *Session) error
	// GetSessionByTrip returns the most recently started session for
	// tripID, or a domain NotFound error if the trip has no sessions.
	GetSessionByTrip(ctx context.Context, tripID string) (*Session, error)
	// AppendMessage inserts message, which must already have SessionID/Role/
	// Content populated by the caller. Populates ID/Timestamp from the
	// database defaults. Plain insert: it does not update the parent
	// Session's TotalTokens — that aggregation belongs to the future
	// Conversation service.
	AppendMessage(ctx context.Context, message *Message) error
	// ListMessages returns every message for sessionID in chronological
	// order (docs/data-model.md §ConversationMessage's "Business Rules").
	// Returns an empty (non-nil) slice, not an error, when the session has
	// no messages.
	ListMessages(ctx context.Context, sessionID string) ([]*Message, error)
}

// PostgresRepository implements Repository using the shared pgx pool exposed
// by database.Client.
type PostgresRepository struct {
	db database.Client
}

// NewPostgresRepository builds a Repository backed by PostgreSQL.
func NewPostgresRepository(db database.Client) Repository {
	return &PostgresRepository{db: db}
}

// CreateSession inserts session; see the Repository interface doc comment.
func (r *PostgresRepository) CreateSession(ctx context.Context, session *Session) error {
	const query = `
		INSERT INTO conversation_sessions (id, trip_id, status)
		VALUES ($1, $2, $3)
		RETURNING started_at, total_tokens, ai_provider, created_at`

	err := r.db.Pool().QueryRow(ctx, query, session.ID, session.TripID, session.Status).
		Scan(&session.StartedAt, &session.TotalTokens, &session.AIProvider, &session.CreatedAt)
	if err != nil {
		return fmt.Errorf("create conversation session: %w", err)
	}

	return nil
}

// GetSessionByTrip returns tripID's most recent session; see the Repository
// interface doc comment.
func (r *PostgresRepository) GetSessionByTrip(ctx context.Context, tripID string) (*Session, error) {
	const query = `
		SELECT id, trip_id, started_at, completed_at, status, total_tokens, ai_provider, created_at
		FROM conversation_sessions
		WHERE trip_id = $1
		ORDER BY started_at DESC, id DESC
		LIMIT 1`

	var session Session
	err := r.db.Pool().QueryRow(ctx, query, tripID).Scan(
		&session.ID, &session.TripID, &session.StartedAt, &session.CompletedAt,
		&session.Status, &session.TotalTokens, &session.AIProvider, &session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("conversation session for trip", tripID)
		}
		return nil, fmt.Errorf("get conversation session by trip: %w", err)
	}

	return &session, nil
}

// AppendMessage inserts message; see the Repository interface doc comment.
func (r *PostgresRepository) AppendMessage(ctx context.Context, message *Message) error {
	const query = `
		INSERT INTO conversation_messages (session_id, role, content)
		VALUES ($1, $2, $3)
		RETURNING id, token_count, timestamp`

	err := r.db.Pool().QueryRow(ctx, query, message.SessionID, message.Role, message.Content).
		Scan(&message.ID, &message.TokenCount, &message.Timestamp)
	if err != nil {
		return fmt.Errorf("append conversation message: %w", err)
	}

	return nil
}

// ListMessages returns sessionID's messages in chronological order; see the
// Repository interface doc comment.
func (r *PostgresRepository) ListMessages(ctx context.Context, sessionID string) ([]*Message, error) {
	const query = `
		SELECT id, session_id, role, content, token_count, timestamp
		FROM conversation_messages
		WHERE session_id = $1
		ORDER BY timestamp ASC, id ASC`

	rows, err := r.db.Pool().Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list conversation messages: %w", err)
	}
	defer rows.Close()

	messages := []*Message{}
	for rows.Next() {
		var msg Message
		if err := rows.Scan(
			&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.TokenCount, &msg.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scan conversation message row: %w", err)
		}
		messages = append(messages, &msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list conversation messages: %w", err)
	}

	return messages, nil
}
