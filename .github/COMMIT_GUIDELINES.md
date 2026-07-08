# Commit Message Guidelines

## Format (MANDATORY)

All commits MUST follow Conventional Commits specification:

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

## Types (REQUIRED)

- **feat**: New feature for the user
- **fix**: Bug fix
- **docs**: Documentation only changes
- **style**: Code style changes (formatting, missing semicolons, etc.)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **perf**: Performance improvement
- **test**: Adding or updating tests
- **chore**: Changes to build process, tools, or dependencies
- **ci**: Changes to CI configuration files and scripts

## Scope (OPTIONAL)

Scope identifies the affected area. Examples:

- **Backend**: `config`, `middleware`, `ai`, `database`, `errors`, `api`
- **Frontend**: `auth`, `trips`, `ui`, `api-client`, `store`, `router`
- **Infrastructure**: `terraform`, `docker`, `ecs`, `rds`, `vpc`
- **E2E**: `playwright`, `fixtures`, `accessibility`
- **Docs**: `readme`, `architecture`, `api`, `roadmap`

## Subject (REQUIRED)

- Use imperative mood ("add" not "adds" or "added")
- Don't capitalize first letter
- No period at the end
- Maximum 50 characters
- Be concise but descriptive

## Body (OPTIONAL)

- Use imperative mood
- Wrap at 72 characters
- Explain **what** and **why**, not **how**
- Separate from subject with blank line

## Footer (OPTIONAL)

- **Issue references** (use when applicable):
  - `Closes #123` — Use in the **final commit** that completes an issue
  - `Fixes #456` — Use when fixing a specific bug issue
  - `Relates to #789` — Use to reference related work without closing
- **Breaking changes**: `BREAKING CHANGE: description`
- **When to omit**: Intermediate commits, refactorings, or work-in-progress within a PR don't need issue references

## Examples

### Simple feature

```
feat(middleware): add request logging middleware

Implements structured logging for all HTTP requests with request ID,
method, path, and duration tracking.

Closes #45
```

### Bug fix

```
fix(api): handle null values in destination response

Previously would panic if destination had no description. Now returns
empty string as per API design standards.

Fixes #128
```

### Grouped task completion

```
feat(primitives): add core UI primitive components

- Add Button with all variants (primary, secondary, ghost)
- Add Input with validation states
- Add Card with header/footer slots
- Add Spinner with size variants

All components include unit tests and Storybook stories.

Closes #67
```

### Documentation

```
docs(roadmap): update Sprint 2 task consolidation groups

Add Group column values for tasks T024-T060. Consolidates 37 tasks
into 14 work items following new consolidation workflow.
```

### Infrastructure

```
chore(docker): optimize backend image build

- Use multi-stage build with Alpine 3.19
- Reduce image size from 800MB to 45MB
- Add health check endpoint
```

### Refactoring (no issue reference needed)

```
refactor(middleware): extract auth logic to separate package

Move JWT validation and token parsing to internal/auth package for
better reusability across handlers.
```

### Work in progress (within a PR)

```
feat(api): add destination model and repository

Implements PostgreSQL repository pattern for destinations with CRUD
operations. Handler and tests will follow in subsequent commits.
```

### Final commit that closes issue

```
feat(backend): complete HTTP server setup with healthcheck

- Add main.go with Chi router initialization
- Integrate PostgreSQL connection pool
- Add /health endpoint with DB ping
- Add graceful shutdown with signal handling

All components tested and integrated successfully.

Closes #89
```

### Breaking change

```
feat(auth)!: change JWT signing algorithm to RS256

BREAKING CHANGE: Existing JWT tokens will be invalid after deployment.
All clients must re-authenticate.

Migrates from HS256 to RS256 for improved security with key rotation
support.

Closes #102
```

## Language Policy (MANDATORY)

**ALL commits MUST be in English**, including:
- Type, scope, subject
- Body text
- Footer references
- Code comments in commit diffs

**Never** mix English and Spanish in commits.

## Anti-patterns (DO NOT DO)

❌ `update files`  
❌ `fix bug`  
❌ `WIP`  
❌ `Fixed the thing from yesterday`  
❌ `feat: Added new feature` (wrong tense)  
❌ `Fix: something` (wrong capitalization)

## Tips

1. **Atomic commits**: One logical change per commit
2. **Test before commit**: All tests must pass
3. **Lint before commit**: Run `go fmt`, `eslint`, etc.
4. **Issue references**: Use `Closes #X` only when the commit **completes** the issue. Intermediate commits don't need references.
5. **Clear intent**: Subject should be understandable without reading the code

## Verification Checklist

Before committing, verify:

- [ ] Follows `type(scope): subject` format
- [ ] Type is from approved list
- [ ] Subject is imperative mood, lowercase, no period, ≤50 chars
- [ ] Body explains what/why (if needed)
- [ ] References issue number if this commit completes/closes an issue
- [ ] All text is in English
- [ ] All tests passing
- [ ] Linting passes

## Git Configuration (Optional)

To use this as a commit message template:

```bash
git config commit.template .github/COMMIT_TEMPLATE.md
```

Then every `git commit` will open your editor with this template pre-filled.
