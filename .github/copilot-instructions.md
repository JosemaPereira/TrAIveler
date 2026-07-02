# Copilot Instructions

## Project Context
- Create a web tool to plan trips to any destination in the world. The tool offers suggestions for places, excursions, gastronomy, tips, and recommendations for exploring the most attractive and hidden treasures of a city or country within a specific timeframe.
- Stack: Backend: Golang 1.24 or higher, Frontend: React 19 or higher
- Actual Phase: Initial setup of the project, including the creation of the repository, initial commit, and basic project structure.

## Documentation References
- docs/functional-requirements.md
- docs/ui-guidelines.md
- docs/testing-guidelines.md
- docs/coding-guidelines.md

## Language Policy (MANDATORY)
- Conversation with the developer may be in English or Spanish; respond in whichever
  language the developer uses.
- ALL generated artifacts MUST be in English, without exception. This includes:
  source code, identifiers, comments, docstrings, commit messages, branch names,
  documentation (docs/, README, specs), test names, and any file content committed
  to the repository.
- Never mix languages inside an artifact. If the developer writes a request in
  Spanish, still produce the code/docs in English.

## Development Principles
- Test-Driven Development: Red-Green-Refactor
- Incremental, small, and testable changes
- Validation before commit: tests pass, no lint errors

## Git Workflow
- Conventional commits (in English): feat:, fix:, chore:, docs:, etc.
- Feature branches: feature/<descriptive-name>  (branch names in English)
- Never commit directly to main
- Versioning: Semantic Versioning (SemVer) 2.0.0