-- +goose Up
-- conversation_sessions (001-T034): AI-assisted trip planning session,
-- tracking multi-turn conversation between user and AI provider.
-- Authority: docs/data-model.md §ConversationSession.
--   - No `version`: ConversationSession is not a versioned entity
--     (Invariant 7 scopes optimistic locking to Trip and Activity).
--   - The "only one in_progress session per trip" business rule is a
--     [Logic] rule enforced by the future Conversation service, not a DB
--     constraint — repositories in this codebase are dumb persistence.
CREATE TABLE conversation_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    status VARCHAR(20) NOT NULL CHECK (status IN ('in_progress', 'completed', 'abandoned')),
    total_tokens INT NOT NULL DEFAULT 0 CHECK (total_tokens >= 0),
    ai_provider VARCHAR(50) NOT NULL DEFAULT 'anthropic-claude',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversation_sessions_trip_id ON conversation_sessions(trip_id);
CREATE INDEX idx_conversation_sessions_status ON conversation_sessions(status);

-- +goose Down
DROP INDEX IF EXISTS idx_conversation_sessions_status;
DROP INDEX IF EXISTS idx_conversation_sessions_trip_id;
DROP TABLE conversation_sessions;
