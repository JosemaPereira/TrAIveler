// Package ai defines the AI client contract and shared request/response types
// used to generate travel itineraries from a conversation history, plus the
// prompt-validation and output-sanitization stubs that will gain real
// deny-list/HTML-stripping logic in a future spec 002 (NFR) ticket.
package ai

// Message is a single turn in a conversation exchanged with the AI provider.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// TravelPreferences captures the trip constraints supplied by the traveler,
// matching the seed data and entity shapes documented in docs/data-model.md.
type TravelPreferences struct {
	Budget       string   `json:"budget"`
	Pace         string   `json:"pace"`
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	TravelStyles []string `json:"travel_styles"`
	GroupSize    int      `json:"group_size"`
}

// ItineraryRequest carries everything needed to generate or continue an
// itinerary: the trip being planned, the conversation so far, and the
// traveler's preferences.
type ItineraryRequest struct {
	TripID              string            `json:"trip_id"`
	ConversationHistory []Message         `json:"conversation_history"`
	Preferences         TravelPreferences `json:"preferences"`
}

// Destination identifies a place within an itinerary, matching
// docs/data-model.md's Destination entity (Country is ISO 3166-1 alpha-2).
type Destination struct {
	Name    string `json:"name"`
	Country string `json:"country"`
}

// Activity is a single planned item within a Day, matching docs/data-model.md's
// Activity entity.
type Activity struct {
	Title         string `json:"title"`
	Type          string `json:"type"`
	Description   string `json:"description"`
	SequenceOrder int    `json:"sequence_order"`
}

// Day groups the activities planned for a single day of the trip at a given
// destination, matching docs/data-model.md's Day entity.
type Day struct {
	Destination Destination `json:"destination"`
	Activities  []Activity  `json:"activities"`
	DayNumber   int         `json:"day_number"`
}

// ResponseMetadata describes provenance of a generated itinerary: which
// model/provider produced it and how many tokens it consumed.
type ResponseMetadata struct {
	Model      string `json:"model"`
	Provider   string `json:"provider"`
	TokensUsed int    `json:"tokens_used"`
}

// ItineraryResponse is the structured result of generating a trip itinerary.
type ItineraryResponse struct {
	Destinations []Destination    `json:"destinations"`
	Days         []Day            `json:"days"`
	Activities   []Activity       `json:"activities"`
	Metadata     ResponseMetadata `json:"metadata"`
}

// StreamChunk is one fragment of a streamed itinerary generation. Done is
// true on the final chunk, at which point Content is empty and the
// producing channels are closed.
type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}
