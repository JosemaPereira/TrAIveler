# Product Vision

<!-- PROMOTED:product-vision START -->
<!-- Generated from specs/001-product-vision-scope/spec.md -->
<!-- Last promoted: 2026-07-03 -->

## Product Identity

**Product Name**: TrAIveler

**Tagline**: Plan smarter trips. Discover hidden gems. Travel together.

**Release Status**: Proof of Concept (POC) / MVP demo. This release is a working demonstration of the full product vision. The platform is designed as a future paid-subscription product; for this demo, subscription payment collection is intentionally mocked.

## Core Problem

Planning a trip from scratch is time-consuming and requires research across multiple sources. Travelers either over-plan (spending hours in spreadsheets) or under-plan (missing local gems). Both ends of the spectrum lead to frustration.

## Value Proposition

A traveler can describe where they want to go and for how long, and receive a complete, personalized itinerary in seconds — covering must-see attractions, off-the-beaten-path discoveries, local gastronomy, and inter-city logistics — which they can then refine and share with companions.

## What It Is

TrAIveler is a web application that uses artificial intelligence to help travelers generate, customize, and collaboratively refine travel itineraries. It transforms an often overwhelming planning process — researching destinations, sequencing visits, finding local food, and coordinating with companions — into a guided, personalized, and shareable experience.

## Target Users

### Persona 1 — The Overwhelmed First-Timer

A traveler with little or no prior knowledge of their destination. May be visiting a new country or continent for the first time. Likely spends hours researching and still feels uncertain about their plan.

**Core need**: A trustworthy, complete starting point that removes decision paralysis and covers the essentials without requiring deep research.

### Persona 2 — The Repeat Explorer

A seasoned traveler revisiting a familiar destination or one they know well from research. Wants to go beyond the standard tourist circuit and discover what locals experience.

**Core need**: Fresh, non-obvious recommendations that respect their existing knowledge and do not repeat what they already know.

### Persona 3 — The Detail Planner

An organized traveler who has already compiled a list of places, restaurants, or activities they want to include. Treats trip planning as a project and may have a partial spreadsheet or list ready before using the tool.

**Core need**: Intelligent organization of their existing preferences — sequenced by day, enriched with gaps filled, and formatted so it can be shared and used during the trip.

### Persona 4 — The Group Organizer

The de-facto trip coordinator within a friend group or family. Takes responsibility for planning and wants a way to get buy-in from companions without managing a shared spreadsheet.

**Core need**: A collaborative space where all travelers can see, react to, and contribute to the itinerary — without one person having to relay every change.

## MVP Scope

The following capabilities are within scope for the MVP:

1. **User authentication and subscription-based accounts** — required for all trip management; the basic plan supports one admin user and one partner collaborator per subscription. Registration includes a visible stub checkout screen that always succeeds, simulating the full future paid-subscription onboarding flow.

2. **AI itinerary generation** via a conversational, multi-turn input flow — the user submits a free-form natural language request (destination, timeframe, traveler profiles, interests); the AI may ask targeted follow-up questions to refine preferences before producing a day-by-day plan covering visit suggestions (well-known landmarks AND lesser-known local points of interest), local food specialties, arrival logistics, and inter-city transfers.

3. **Travel style personalization** — adapting generated content to at least: gastronomy, sports, technology, museums and art, film and audiovisual media. A single trip may carry multiple travel styles to reflect different preference profiles within the travel group.

4. **Itinerary customization** — allowing users to edit, add, remove, and reorder destinations and activities after generation.

5. **Existing plan enrichment** — accepting a user-provided free-form natural language description of desired places, ideas, or constraints; the AI treats these as anchors and builds a complete itinerary around them. All user-provided places MUST be included in the final itinerary without omission or replacement.

6. **Trip sharing and collaboration** — allowing an admin user to invite one partner collaborator who can view the full itinerary and submit suggestions for modifications. Suggestions have "pending" status until the admin approves (suggestion applied to itinerary) or rejects (itinerary unchanged) them. Direct edits by partners without admin approval are rejected with permission errors.

## Explicit Out of Scope (MVP)

The following capabilities are explicitly excluded from the current scope:

- **Third-party integrations**: No Google Maps, Google Drive, booking platforms, review sites, or any external API beyond the AI provider (Anthropic Claude)
- **Native mobile applications**: No iOS or Android apps; web application only
- **Real-time booking**: No pricing, availability checking, or booking for transportation, accommodation, or activities
- **Offline access**: No downloadable itinerary formats (PDF, spreadsheet export), no offline mode
- **Real payment processing**: Subscription plan flows and limit enforcement ARE in scope as application logic with a mocked payment stub; only actual payment collection and billing provider integration (Stripe, PayPal) are deferred to post-MVP
- **Social features**: No public itinerary discovery, no social feeds, no community features beyond invite-based one-on-one collaboration (admin + one partner)
- **Localization (i18n)**: English only at MVP; no multi-language UI infrastructure required. AI-generated content responds naturally in the language of the user's prompt.
- **Advanced session management**: No "remember me", device tracking, or concurrent session limits beyond JWT token expiration
- **Multi-factor authentication (MFA)**: Basic email/password authentication only; no MFA, no OAuth providers (Google, Facebook)
- **Compliance certifications**: No SOC 2, ISO 27001; GDPR-aware design only (minimal PII collection, right-to-deletion support, privacy policy page)

## Roles and Permissions

The system enforces two distinct roles:

- **Admin**: Full CRUD on their own trips; exclusive authority to approve or reject partner suggestions
- **Partner**: Can view trips they are invited to and submit suggestions for modifications; cannot directly apply changes to any itinerary item

**Basic Plan Limit**: One admin user and a maximum of one partner collaborator per subscription.

<!-- PROMOTED:product-vision END -->
