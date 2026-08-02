-- +goose Up
-- conversation_messages (001-T034): individual message within an AI
-- conversation session (user input, system prompt, AI response).
-- Authority: docs/data-model.md §ConversationMessage.
--   - No `version`: ConversationMessage is not a versioned entity
--     (Invariant 7 scopes optimistic locking to Trip and Activity).
--   - Sanitization/validation of `content` (bluemonday for assistant
--     output, PromptValidator for user input) is a [Logic] rule owned by
--     the future Conversation service, not enforced at the DB layer.
CREATE TABLE conversation_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES conversation_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('system', 'user', 'assistant')),
    content TEXT NOT NULL,
    token_count INT NOT NULL DEFAULT 0 CHECK (token_count >= 0),
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversation_messages_session_id ON conversation_messages(session_id);

-- +goose Down
DROP INDEX IF EXISTS idx_conversation_messages_session_id;
DROP TABLE conversation_messages;
