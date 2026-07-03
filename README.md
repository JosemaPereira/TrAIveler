# TrAIveler ✈️

> **Your AI travel companion for planning unforgettable trips**

_Capstone project for the AI Bootcamp at Slalom_

---

## Why TrAIveler?

Remember the last time you planned a trip? Hours lost in browser tabs, spreadsheets filling up with half-researched destinations, the nagging feeling that you're missing the _real_ local spots everyone raves about once they return. Planning shouldn't feel like work.

**TrAIveler transforms travel planning into a conversation.** Tell it where you want to go and for how long. It asks the right questions — your travel style, what excites you, whether you prefer hidden gems or iconic landmarks — then crafts a complete day-by-day itinerary with visits, local food recommendations, and logistics. No overwhelming research. No missed opportunities.

---

## The Experience

### 🗣️ **Start with a conversation**
"I want to explore Japan for 10 days." That's all you need. TrAIveler's AI guides you through a natural dialogue to understand your preferences, pace, and priorities.

### 🎯 **Get a personalized itinerary**
Receive a complete plan tailored to your style — whether you're a gastronomy enthusiast hunting for authentic ramen shops, a sports lover seeking local stadiums, or an art devotee chasing hidden galleries. Every day includes must-see attractions _and_ off-the-beaten-path discoveries.

### 🤝 **Travel together**
Invite a travel companion to view your itinerary and suggest changes. Review their ideas, approve what fits, and keep everything in one place. No more group chat chaos or lost messages.

### ✏️ **Refine as you go**
Already have a list of places you want to visit? Share it. TrAIveler anchors your itinerary around them and fills the gaps with complementary suggestions. Edit, reorder, or add activities anytime.

---

## What's Inside (MVP)

This first version focuses on the core planning experience:

- **Smart authentication** — secure sign-up with subscription management
- **Conversational AI planning** — multi-turn dialogue that understands context and refines suggestions
- **Travel style personalization** — gastronomy, sports, tech, museums, film & audiovisual
- **Flexible itineraries** — add, edit, remove, and reorder destinations and activities
- **Collaborative planning** — invite one companion to suggest changes; you approve or pass
- **Anchor-based enrichment** — bring your own list of must-visit places; we build around them

---

## Built With Care

TrAIveler is crafted following modern development best practices:

- **Go backend** powering a robust REST API with PostgreSQL
- **React + TypeScript frontend** for a smooth, type-safe user experience
- **Anthropic Claude** as the AI engine, delivering natural conversations and intelligent recommendations
- **Test-driven development** ensuring every feature works reliably
- **Accessibility-first design** meeting WCAG 2.1 AA standards
- **Security by design** with prompt injection defense, secret scanning, and sanitized outputs

The entire codebase follows strict quality gates — automated linting, testing, and security checks run on every change.

---

## Current Status

🔨 **In active development** — Planning and architecture complete; implementation in progress.

The project has comprehensive specifications covering product vision, technical design, cloud infrastructure, security model, data models, API contracts, and a detailed roadmap with 199 prioritized tasks across three foundational specs. All foundational documentation is complete and ready to guide the build.

---

## The Vision

TrAIveler starts as a web application for individual and small group travel planning. The long-term vision? A platform where travelers discover not just where to go, but _how_ to experience a destination authentically — through AI that learns from real journeys, local insights, and the hidden stories that make travel meaningful.

---

## Project Areas

TrAIveler is organized into distinct areas, each with its own README and development workflow:

- **[backend/](backend/)** — Go REST API powering authentication, trip management, AI integration, and collaboration
- **[frontend/](frontend/)** — React + TypeScript SPA delivering the user interface and experience
- **[e2e/](e2e/)** — Playwright end-to-end test suite validating complete user flows
- **infra/** — Terraform infrastructure as code for AWS deployment _(directory will be created in Phase 1)_

---

## Want to Know More?

This repository contains the full specification and implementation plan. If you're interested in the technical details, architecture decisions, or want to contribute, explore the `docs/` and `specs/` directories.

**Project planning & process:**
- [Project Roadmap](docs/roadmap.md) — 199 tasks organized into clear phases with sprint planning
- [Project Workflow](docs/project-workflow.md) — complete development process from spec to shipped feature

**Product & requirements:**
- [Product Vision](docs/product-vision.md) — core value proposition, personas, MVP scope, and roles
- [Functional Requirements](docs/functional-requirements.md) — what the system must do
- [Non-Functional Requirements](docs/nfrs.md) — performance, security, accessibility, and quality standards

**Technical architecture:**
- [Architecture Overview](docs/architecture.md) — system components, integration rules, and boundaries
- [Cloud & Environments](docs/cloud-and-environments.md) — AWS infrastructure, IaC strategy, and CI/CD
- [Security Model](docs/security.md) — authentication, authorization, and security practices

**Development standards:**
- [Coding Guidelines](docs/coding-guidelines.md) — formatting, naming, and code organization
- [Testing Guidelines](docs/testing-guidelines.md) — three-layer testing strategy and coverage targets
- [UI Guidelines](docs/ui-guidelines.md) — design tokens, component structure, and accessibility rules

---

**Let's make travel planning joyful again.** 🌍