// Package itinerary implements conversation.ItineraryGenerator: it drives
// the AI provider through a trip-planning conversation and, once the
// conversation has gathered enough information, persists the resulting
// day-by-day itinerary as Destination/Day/Activity rows (docs/data-model.md
// §Destination/§Day/§Activity). Service.Continue is the sole entry point
// (001-T037, issue #236); prompt.go holds the system prompt instructing the
// AI provider how to conduct the conversation and shape its JSON reply.
// This package has no HTTP layer or cmd/api wiring of its own — that is a
// separate, not-yet-built ticket (001-T038/T039/T040, issue #237).
package itinerary
