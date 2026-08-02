// Package conversation persists ConversationSession and ConversationMessage
// (docs/data-model.md §ConversationSession/§ConversationMessage) — the
// AI-assisted trip planning conversation history. This package is
// repository-only (001-T034): it has no service or HTTP layer of its own,
// and it deliberately owns no business rules (e.g. "only one in_progress
// session per trip at a time") — those belong to the future Conversation
// service that consumes this Repository (001-T036).
package conversation

import "time"

// Status values accepted by Session.Status, mirroring the CHECK constraint on
// the conversation_sessions table
// (migrations/017_create_conversation_sessions_table.sql).
const (
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusAbandoned  = "abandoned"
)

// Role values accepted by Message.Role, mirroring the CHECK constraint on the
// conversation_messages table
// (migrations/018_create_conversation_messages_table.sql).
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Session is an AI-assisted trip planning session, tracking a multi-turn
// conversation between a user and an AI provider (docs/data-model.md
// §ConversationSession). CompletedAt is nil while the session is in
// progress. TotalTokens is the cumulative token count across the session's
// messages, used for cost tracking.
type Session struct {
	ID          string     `json:"id" db:"id"`
	TripID      string     `json:"trip_id" db:"trip_id"`
	StartedAt   time.Time  `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	Status      string     `json:"status" db:"status"`
	TotalTokens int        `json:"total_tokens" db:"total_tokens"`
	AIProvider  string     `json:"ai_provider" db:"ai_provider"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Message is a single message within a Session — a system prompt, user
// input, or AI response (docs/data-model.md §ConversationMessage). Content
// sanitization (bluemonday for assistant output) and validation (
// PromptValidator for user input) are [Logic] rules owned by the future
// Conversation service, not enforced at this persistence layer.
type Message struct {
	ID         string    `json:"id" db:"id"`
	SessionID  string    `json:"session_id" db:"session_id"`
	Role       string    `json:"role" db:"role"`
	Content    string    `json:"content" db:"content"`
	TokenCount int       `json:"token_count" db:"token_count"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
}
