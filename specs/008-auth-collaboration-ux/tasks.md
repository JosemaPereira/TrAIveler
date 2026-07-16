# Tasks: Authentication & Collaboration User Experience

**Feature**: Authentication & Collaboration UX  
**Branch**: `008-auth-collaboration-ux`  
**Input**: Design documents from `specs/008-auth-collaboration-ux/`

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US6)
- Tests are OPTIONAL and NOT included per mode instructions
- All tasks include exact file paths

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure per plan.md

- [ ] T001 Create backend project structure: backend/cmd/api/, backend/internal/{auth,subscription,collaboration,security}/, backend/pkg/{database,config}/, backend/tests/{unit,integration,contract}/
- [ ] T002 [P] Initialize Go module with Chi v5, pgx/v5, goose, golang-jwt/jwt v5, bcrypt dependencies in backend/go.mod
- [ ] T003 [P] Create frontend project structure: frontend/src/{features,components,stores,styles,services,routes}/, frontend/tests/{unit,integration,e2e}/
- [ ] T004 [P] Initialize Vite React TypeScript project with TanStack Query v5, Zustand, React Router v7, Vitest, Playwright in frontend/package.json
- [ ] T005 [P] Configure backend linting: golangci-lint.yml with errcheck, govet, staticcheck, revive, gosec in backend/.golangci.yml
- [ ] T006 [P] Configure frontend linting: ESLint + Prettier with TypeScript strict mode in frontend/.eslintrc.json and frontend/.prettierrc
- [ ] T007 [P] Create E2E test structure: e2e/specs/{auth,collaboration,accessibility}/ directories
- [ ] T008 [P] Create infrastructure directory: infra/terraform/modules/secrets/ for JWT key rotation

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T009 Setup PostgreSQL connection pooling with pgx/v5 in backend/pkg/database/connection.go
- [ ] T010 [P] Configure goose migrations framework in backend/pkg/database/migrations/ directory
- [ ] T011 [P] Create migration 001_create_users.sql: users table with id, email (unique), password_hash, full_name, has_subscription (boolean), failed_login_attempts (integer default 0), last_failed_login_at (timestamp), email_verified (boolean default false), created_at, updated_at, version (integer for optimistic locking)
- [ ] T012 [P] Create migration 002_create_subscriptions.sql: subscriptions table with id, user_id (FK), plan_id (FK), status, current_period_start, current_period_end, grace_period_ends_at (nullable), cancelled_at (nullable), created_at, updated_at, version
- [ ] T013 [P] Create migration 003_create_password_reset_tokens.sql: password_reset_tokens table with id, user_id (FK CASCADE), token_hash (unique), expires_at, used_at (nullable), created_at; indexes on user_id, token_hash, expires_at
- [ ] T014 [P] Create migration 004_create_security_events.sql: security_events table with id, correlation_id (index), event_type (check constraint), user_id (FK SET NULL), email, severity, ip_address, user_agent, details (jsonb), created_at; indexes on correlation_id, user_id, email, created_at, event_type
- [ ] T015 [P] Create migration 005_create_trips.sql: trips table with id, creator_id (FK), destination, start_date, end_date, archived (boolean default false), created_at, updated_at, version; index on creator_id, archived
- [ ] T016 [P] Create migration 006_create_collaborators.sql: collaborators table with id, trip_id (FK CASCADE), user_id (FK CASCADE), email, status (pending/accepted/rejected), invited_at, accepted_at, created_at; indexes on trip_id, user_id, status
- [ ] T017 [P] Create migration 007_create_suggestions.sql: suggestions table with id, trip_id (FK CASCADE), collaborator_id (FK CASCADE), suggestion_type, content, details (jsonb), status (pending/approved/rejected), approved_at, rejected_at, created_at
- [ ] T018 Create environment config loader in backend/pkg/config/config.go: load DATABASE_URL, JWT_SIGNING_KEY_SECRET_ARN, ANTHROPIC_API_KEY_SECRET_ARN from environment
- [x] T019 [P] Implement JWT generator with RS256 in backend/internal/auth/jwt/generator.go: GenerateAccessToken(userID, hasSubscription) returns signed JWT with 24h expiration
- [x] T020 [P] Implement JWT validator in backend/internal/auth/jwt/validator.go: ValidateToken(token) returns claims, supports multi-key validation for zero-downtime rotation
- [x] T021 [P] Implement JWT refresher in backend/internal/auth/jwt/refresher.go: RefreshToken(refreshToken) validates, revokes old token, issues new access+refresh tokens
- [ ] T022 [P] Implement bcrypt password hasher in backend/internal/auth/password/hasher.go: HashPassword(password) with cost 12, ComparePassword(hash, password) for validation
- [ ] T023 [P] Implement password validator in backend/internal/auth/password/validator.go: ValidatePassword(password) checks 8-72 chars, uppercase, lowercase, digit
- [ ] T024 [P] Implement progressive delay rate limiter in backend/internal/auth/ratelimit/limiter.go: CheckRateLimit(email) tracks failed attempts, returns delay seconds (exponential backoff after 5 failures)
- [ ] T025 [P] Implement rate limiter storage in backend/internal/auth/ratelimit/store.go: in-memory map with 15-minute TTL for failed attempt tracking (user_id, count, first_attempt_at)
- [ ] T026 [P] Implement security event logger in backend/internal/security/logger.go: LogSecurityEvent(correlationID, eventType, userID, email, severity, ipAddress, userAgent, details) writes to CloudWatch Logs as structured JSON
- [ ] T027 [P] Implement auth middleware in backend/internal/security/middleware.go: ValidateJWTCookie() extracts token from cookie, validates with JWT validator, attaches user context to request
- [ ] T028 [P] Implement request ID middleware in backend/internal/security/middleware.go: GenerateRequestID() creates correlation ID, adds to context and response header X-Request-ID
- [ ] T029 [P] Implement rate limit middleware in backend/internal/security/middleware.go: RateLimitMiddleware() checks X-RateLimit headers, returns 429 with Retry-After if exceeded
- [ ] T030 Setup Chi router with middleware chain in backend/cmd/api/main.go: request ID → logging → CORS → rate limit → recovery
- [ ] T031 [P] Create Axios instance in frontend/src/services/api.ts: base URL, withCredentials=true, request/response interceptors (correlation ID, 401 refresh, error mapping)
- [ ] T032 [P] Create error handler utility in frontend/src/services/errorHandler.ts: mapApiError(error) converts API errors to user-friendly messages per FR-019
- [ ] T033 [P] Create auth store in frontend/src/stores/authStore.ts: Zustand store with user state (id, email, full_name, has_subscription), isAuthenticated boolean, login/logout/setUser actions
- [ ] T034 [P] Create CSS design tokens in frontend/src/styles/tokens.css: CSS custom properties for colors (primary, secondary, error, warning with WCAG AA contrast), spacing scale (4/8/12/16/24/32/48px), typography (font sizes, weights, line heights), focus indicators (outline-width, outline-color, outline-offset)
- [ ] T035 [P] Create global styles in frontend/src/styles/global.css: CSS reset, base typography, box-sizing border-box, accessible focus styles using tokens
- [ ] T036 [P] Create React Router configuration in frontend/src/routes/router.tsx: routes for /register, /login, /dashboard, /trips/:id, /settings, /password-reset, with protected route wrapper checking authStore.isAuthenticated

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Paid User Registration & First Trip Creation (Priority: P1) 🎯 MVP

**Goal**: Enable new users to register with payment, create subscription, and generate first AI-powered trip

**Independent Test**: Navigate to /register, fill form with payment stub, submit, verify redirect to /dashboard, create trip via AI, verify trip appears in "My Trips"

### Backend Implementation for US1

- [ ] T037 [P] [US1] Create User model in backend/internal/auth/models.go: User struct with ID, Email, PasswordHash, FullName, HasSubscription, FailedLoginAttempts, LastFailedLoginAt, EmailVerified, CreatedAt, UpdatedAt, Version
- [ ] T038 [P] [US1] Create RegisterRequest/RegisterResponse models in backend/internal/auth/models.go: RegisterRequest{Email, Password, FullName, PaymentMethodToken}, RegisterResponse{User, Subscription}
- [ ] T039 [P] [US1] Create Subscription model in backend/internal/subscription/models.go: Subscription struct with ID, UserID, PlanID, Status, CurrentPeriodStart, CurrentPeriodEnd, GracePeriodEndsAt, CancelledAt, CreatedAt, UpdatedAt, Version
- [ ] T040 [US1] Create User repository in backend/internal/auth/repository.go: CreateUser(user), GetUserByEmail(email), UpdateUser(user) with optimistic locking check
- [ ] T041 [P] [US1] Create Subscription repository in backend/internal/subscription/repository.go: CreateSubscription(sub), GetSubscriptionByUserID(userID), UpdateSubscription(sub), CancelSubscription(id)
- [ ] T042 [P] [US1] Create stub payment provider interface in backend/internal/subscription/payment/provider.go: PaymentProvider interface with ProcessPayment(token, planID) method
- [ ] T043 [P] [US1] Implement stub payment provider in backend/internal/subscription/payment/stub.go: StubPaymentProvider always returns success, logs to console "[DEMO] Payment processed: [token]"
- [ ] T044 [US1] Implement auth service Register method in backend/internal/auth/service.go: validates email/password, checks email uniqueness (409 if exists), hashes password with bcrypt cost 12, creates User, if paymentMethodToken provided calls subscription service, sets HasSubscription=true, logs auth_registration security event
- [ ] T045 [US1] Implement subscription service CreateSubscription in backend/internal/subscription/service.go: calls payment provider ProcessPayment, creates Subscription with status='active', current_period_end=now+30days, returns subscription
- [ ] T046 [US1] Implement POST /auth/register handler in backend/internal/auth/handler.go: validates request body (email format, password strength), calls authService.Register, generates JWT tokens (access 24h, refresh 30d), sets HTTP-only Secure SameSite=Strict cookies, returns 201 with user+subscription JSON
- [ ] T047 [US1] Register POST /api/v1/auth/register route in backend/cmd/api/main.go: attach register handler to Chi router with rate limit middleware (10/min per IP)
- [ ] T048 [P] [US1] Create Trip model in backend/internal/collaboration/models.go: Trip struct with ID, CreatorID, Destination, StartDate, EndDate, Archived, CreatedAt, UpdatedAt, Version (note: full AI generation logic deferred to future spec, stub returns hardcoded itinerary)
- [ ] T049 [P] [US1] Create Trip repository in backend/internal/collaboration/repository.go: CreateTrip(trip), GetTripsByCreatorID(creatorID), GetTripByID(id), UpdateTrip(trip), DeleteTrip(id)
- [ ] T050 [US1] Implement POST /trips handler in backend/internal/collaboration/handler.go: validates auth, extracts user from context, validates request (destination, dates), creates Trip with CreatorID=user.ID, returns 201 with trip JSON (AI stub: returns trip with hardcoded 3-day Paris itinerary)
- [ ] T051 [US1] Register POST /api/v1/trips route in backend/cmd/api/main.go: attach trip creation handler with auth middleware (requires valid JWT)

### Frontend Implementation for US1

- [ ] T052 [P] [US1] Create Button primitive in frontend/src/components/primitives/Button.tsx: <Button> with variant (primary/secondary/danger), size (small/medium/large), disabled, loading props; uses design tokens; keyboard accessible
- [ ] T053 [P] [US1] Create Input primitive in frontend/src/components/primitives/Input.tsx: <Input> with type, label, error, disabled props; uses design tokens; associated label for accessibility
- [ ] T054 [P] [US1] Create Label primitive in frontend/src/components/primitives/Label.tsx: <Label> with htmlFor, required indicator; uses design tokens
- [ ] T055 [P] [US1] Create ErrorMessage primitive in frontend/src/components/primitives/ErrorMessage.tsx: <ErrorMessage> displays validation/API errors with icon; uses error color token; aria-live for screen readers
- [ ] T056 [US1] Create Form composite in frontend/src/components/composites/Form.tsx: <Form> with onSubmit, children; prevents default, handles loading state, disables submit during loading
- [ ] T057 [P] [US1] Create LoadingSpinner feature in frontend/src/components/features/LoadingSpinner.tsx: <LoadingSpinner> with message prop; uses primary color token; aria-busy for screen readers
- [ ] T058 [US1] Create useRegister hook in frontend/src/features/auth/hooks/useRegister.ts: TanStack Query mutation for POST /auth/register, updates authStore on success, handles errors via errorHandler
- [ ] T059 [P] [US1] Create authApi service in frontend/src/features/auth/services/authApi.ts: register(email, password, fullName, paymentMethodToken) calls axios POST /api/v1/auth/register with withCredentials
- [ ] T060 [US1] Create RegisterForm component in frontend/src/features/auth/components/RegisterForm.tsx: form with email, password, fullName inputs, "Continue to Payment" checkbox, "Create Free Account" button; uses useRegister hook, validates client-side, shows loading/error states
- [ ] T061 [US1] Create RegisterPage in frontend/src/features/auth/pages/RegisterPage.tsx: renders RegisterForm, heading "Create Your Account", links to /login; redirects to /dashboard on success
- [ ] T062 [P] [US1] Create useTrips hook in frontend/src/features/trips/hooks/useTrips.ts: TanStack Query query for GET /trips, returns user's owned trips and collaborations
- [ ] T063 [P] [US1] Create useCreateTrip hook in frontend/src/features/trips/hooks/useCreateTrip.ts: TanStack Query mutation for POST /trips, invalidates trips query on success
- [ ] T064 [P] [US1] Create tripsApi service in frontend/src/features/trips/services/tripsApi.ts: getTrips(), createTrip(destination, startDate, endDate) calls axios
- [ ] T065 [US1] Create TripCard component in frontend/src/features/trips/components/TripCard.tsx: displays trip destination, dates, "View Details" link; uses Card composite; keyboard accessible
- [ ] T066 [P] [US1] Create Card composite in frontend/src/components/composites/Card.tsx: <Card> with heading, children; uses design tokens for border, padding, shadow
- [ ] T067 [US1] Create TripDashboard component in frontend/src/features/trips/components/TripDashboard.tsx: renders "My Trips" section (if hasSubscription), "Shared with Me" section, "Create Trip" button (enabled if hasSubscription, disabled with tooltip if Free User), uses useTrips hook, displays LoadingSpinner/ErrorMessage/EmptyState
- [ ] T068 [P] [US1] Create EmptyState feature in frontend/src/components/features/EmptyState.tsx: <EmptyState> with message, illustration icon, action button; uses design tokens
- [ ] T069 [US1] Create DashboardPage in frontend/src/features/trips/pages/DashboardPage.tsx: renders TripDashboard, heading "My Trips", protected route (requires auth)
- [ ] T070 [US1] Create TripDetailPage in frontend/src/features/trips/pages/TripDetailPage.tsx: displays trip destination, dates, day-by-day itinerary (stub: hardcoded Paris 3-day plan), "Edit Trip" button (if owner), "Delete Trip" button (if owner), uses useTripDetail hook (fetch GET /trips/:id)

**Checkpoint**: Paid User can now register, create subscription, and generate first trip (US1 complete and independently testable)

---

## Phase 4: User Story 2 - Free User Registration & Accepting Collaboration Invite (Priority: P1)

**Goal**: Enable collaboration invites, Free User registration without payment, single-collaboration limit enforcement

**Independent Test**: Paid User invites collaborator via email, Free User registers without payment, accepts invitation, verifies single-trip limit (attempt second invite shows "Leave current trip first")

### Backend Implementation for US2

- [ ] T071 [P] [US2] Create Collaborator model in backend/internal/collaboration/models.go: Collaborator struct with ID, TripID, UserID, Email, Status (pending/accepted/rejected), InvitedAt, AcceptedAt, CreatedAt
- [ ] T072 [P] [US2] Create Suggestion model in backend/internal/collaboration/models.go: Suggestion struct with ID, TripID, CollaboratorID, SuggestionType, Content, Details (jsonb), Status (pending/approved/rejected), ApprovedAt, RejectedAt, CreatedAt
- [ ] T073 [US2] Create Collaborator repository in backend/internal/collaboration/repository.go: CreateCollaborator(collab), GetCollaboratorsByTripID(tripID), GetCollaboratorsByUserID(userID), UpdateCollaboratorStatus(id, status), DeleteCollaborator(id), CountActiveCollaborations(userID)
- [ ] T074 [US2] Implement POST /trips/:trip_id/collaborators handler in backend/internal/collaboration/handler.go: validates auth, checks user is trip creator OR existing collaborator with invite permission, validates email format, checks basic plan limit (1 collaborator max), creates Collaborator with status='pending', logs invitation to console (email stub), returns 201 with collaborator JSON
- [ ] T075 [US2] Register POST /api/v1/trips/:trip_id/collaborators route in backend/cmd/api/main.go: attach invite handler with auth middleware
- [ ] T076 [US2] Implement PATCH /collaborators/:id/accept handler in backend/internal/collaboration/handler.go: validates auth, checks user owns collaborator record, if user.HasSubscription=false checks CountActiveCollaborations(userID) <= 0 (returns 400 "Leave current trip first" if > 0), updates status='accepted', sets AcceptedAt=now, returns 200 with collaborator JSON
- [ ] T077 [US2] Register PATCH /api/v1/collaborators/:id/accept route in backend/cmd/api/main.go: attach accept handler with auth middleware
- [ ] T078 [US2] Implement DELETE /trips/:trip_id/collaborators/leave handler in backend/internal/collaboration/handler.go: validates auth, checks user is collaborator (not creator), deletes Collaborator record, returns 204 No Content
- [ ] T079 [US2] Register DELETE /api/v1/trips/:trip_id/collaborators/leave route in backend/cmd/api/main.go: attach leave handler with auth middleware
- [ ] T080 [US2] Modify POST /auth/register handler in backend/internal/auth/handler.go: if no paymentMethodToken provided, skip subscription creation, set User.HasSubscription=false, log auth_registration with has_subscription=false detail
- [ ] T081 [US2] Implement GET /invitations handler in backend/internal/collaboration/handler.go: validates auth, returns Collaborator records where UserID=user.ID AND status='pending', includes trip details (destination, creator name)
- [ ] T082 [US2] Register GET /api/v1/invitations route in backend/cmd/api/main.go: attach invitations list handler with auth middleware

### Frontend Implementation for US2

- [ ] T083 [P] [US2] Create Badge primitive in frontend/src/components/primitives/Badge.tsx: <Badge> with variant (free/paid/pending); uses design tokens; displays "Free User" or "Paid User" text
- [ ] T084 [P] [US2] Create Modal composite in frontend/src/components/composites/Modal.tsx: <Modal> with isOpen, onClose, title, children; uses design tokens; keyboard trap, ESC to close, focus management
- [ ] T085 [P] [US2] Create Banner composite in frontend/src/components/composites/Banner.tsx: <Banner> with type (info/warning/error/success), message, action button; uses design tokens; dismissible
- [ ] T086 [US2] Create useInvitations hook in frontend/src/features/collaboration/hooks/useInvitations.ts: TanStack Query query for GET /invitations, returns pending invitations
- [ ] T087 [P] [US2] Create useAcceptInvitation hook in frontend/src/features/collaboration/hooks/useAcceptInvitation.ts: TanStack Query mutation for PATCH /collaborators/:id/accept, invalidates invitations and trips queries on success
- [ ] T088 [P] [US2] Create useLeaveTrip hook in frontend/src/features/collaboration/hooks/useLeaveTrip.ts: TanStack Query mutation for DELETE /trips/:trip_id/collaborators/leave, invalidates trips query, shows success confirmation
- [ ] T089 [P] [US2] Create collaborationApi service in frontend/src/features/collaboration/services/collaborationApi.ts: getInvitations(), acceptInvitation(id), leaveTrip(tripId), inviteCollaborator(tripId, email)
- [ ] T090 [US2] Create InvitationCard component in frontend/src/features/collaboration/components/InvitationCard.tsx: displays trip destination, creator, invited date, "Accept" button, "Reject" button; uses Card, Badge; shows error if collaboration limit reached with "Leave Current Trip" button
- [ ] T091 [US2] Create InvitationsPage in frontend/src/features/collaboration/pages/InvitationsPage.tsx: renders list of InvitationCard, uses useInvitations hook, shows EmptyState if no invitations, LoadingSpinner during fetch
- [ ] T092 [US2] Modify RegisterPage in frontend/src/features/auth/pages/RegisterPage.tsx: add "Sign up as Free User" option alongside "Continue to Payment", conditionally omit paymentMethodToken from request body
- [ ] T093 [US2] Modify TripDashboard in frontend/src/features/trips/components/TripDashboard.tsx: add "Shared with Me" section, display collaborator badge (Free User/Paid User), disable "Create Trip" button for Free Users with tooltip "Upgrade to create your own trips"
- [ ] T094 [US2] Create InviteCollaborator component in frontend/src/features/collaboration/components/InviteCollaborator.tsx: form with email input, "Send Invite" button, displays "Basic plan: 1 collaborator maximum" note, uses inviteCollaborator mutation
- [ ] T095 [US2] Modify TripDetailPage in frontend/src/features/trips/pages/TripDetailPage.tsx: if user is creator, show "Invite Collaborator" button → opens InviteCollaborator modal; if user is collaborator, show badge and "Leave Trip" button
- [ ] T096 [US2] Create LeaveTripButton component in frontend/src/features/collaboration/components/LeaveTripButton.tsx: button with confirmation modal "Are you sure you want to leave this trip?", uses useLeaveTrip hook, redirects to /dashboard on success

**Checkpoint**: Free Users can register, accept invitations, enforce single-collaboration limit, leave trips (US2 complete and independently testable)

---

## Phase 5: User Story 3 - Returning User Login & Trip Management (Priority: P1)

**Goal**: Authenticate returning users, display trip dashboard, enable trip CRUD operations

**Independent Test**: Create test user, logout, login with email/password, verify dashboard loads with trips, test rate limiting (6 failed attempts)

### Backend Implementation for US3

- [ ] T097 [P] [US3] Create LoginRequest/LoginResponse models in backend/internal/auth/models.go: LoginRequest{Email, Password}, LoginResponse{User}
- [ ] T098 [US3] Implement auth service Login method in backend/internal/auth/service.go: validates email format, retrieves user by email (returns 401 "Invalid credentials" if not found, no enumeration), checks rate limit via rate limiter (returns 429 with retry_after if exceeded), compares password hash with bcrypt (returns 401 if mismatch, increments failed attempts), resets failed_login_attempts=0 on success, logs auth_login_success or auth_login_failure security event
- [ ] T099 [US3] Implement POST /auth/login handler in backend/internal/auth/handler.go: validates request body, calls authService.Login, generates JWT tokens (access 24h, refresh 30d), sets HTTP-only cookies, returns 200 with user JSON
- [ ] T100 [US3] Register POST /api/v1/auth/login route in backend/cmd/api/main.go: attach login handler with rate limit middleware (progressive delay after 5 failures)
- [ ] T101 [P] [US3] Implement POST /auth/logout handler in backend/internal/auth/handler.go: validates auth, revokes refresh token (mark as revoked in RefreshToken table), clears access_token and refresh_token cookies, logs auth_logout security event, returns 204 No Content
- [ ] T102 [P] [US3] Register POST /api/v1/auth/logout route in backend/cmd/api/main.go: attach logout handler with auth middleware
- [ ] T103 [P] [US3] Implement GET /trips handler in backend/internal/collaboration/handler.go: validates auth, returns trips where CreatorID=user.ID (owned trips) OR user in Collaborator records (shared trips), excludes archived trips by default (optional ?include_archived=true for creator only), returns 200 with trips array JSON
- [ ] T104 [P] [US3] Register GET /api/v1/trips route in backend/cmd/api/main.go: attach trips list handler with auth middleware
- [ ] T105 [P] [US3] Implement GET /trips/:id handler in backend/internal/collaboration/handler.go: validates auth, retrieves trip, checks user is creator OR collaborator (returns 404 if neither, prevents ID enumeration), returns 200 with trip JSON including full itinerary
- [ ] T106 [P] [US3] Register GET /api/v1/trips/:id route in backend/cmd/api/main.go: attach trip detail handler with auth middleware
- [ ] T107 [P] [US3] Implement PUT /trips/:id handler in backend/internal/collaboration/handler.go: validates auth, checks user is creator (returns 403 if not), checks subscription active OR in grace period (returns 403 "Renew subscription" if grace period expired), validates request body, updates trip with optimistic locking check (returns 409 if version mismatch), returns 200 with updated trip JSON
- [ ] T108 [P] [US3] Register PUT /api/v1/trips/:id route in backend/cmd/api/main.go: attach trip update handler with auth middleware
- [ ] T109 [P] [US3] Implement DELETE /trips/:id handler in backend/internal/collaboration/handler.go: validates auth, checks user is creator (returns 403 if not), checks subscription active (returns 403 if grace period or cancelled), deletes trip (CASCADE deletes collaborators and suggestions), returns 204 No Content
- [ ] T110 [P] [US3] Register DELETE /api/v1/trips/:id route in backend/cmd/api/main.go: attach trip delete handler with auth middleware

### Frontend Implementation for US3

- [ ] T111 [P] [US3] Create useLogin hook in frontend/src/features/auth/hooks/useLogin.ts: TanStack Query mutation for POST /auth/login, updates authStore on success, handles 401 "Invalid credentials" and 429 "Rate limit" errors, displays retry_after countdown
- [ ] T112 [P] [US3] Create useLogout hook in frontend/src/features/auth/hooks/useLogout.ts: TanStack Query mutation for POST /auth/logout, clears authStore on success, redirects to /login
- [ ] T113 [US3] Create LoginForm component in frontend/src/features/auth/components/LoginForm.tsx: form with email, password inputs, "Log In" button, "Forgot Password?" link; uses useLogin hook, validates client-side, shows loading/error states, displays rate limit countdown if 429 error
- [ ] T114 [US3] Create LoginPage in frontend/src/features/auth/pages/LoginPage.tsx: renders LoginForm, heading "Welcome Back", link to /register; redirects to /dashboard on success, preserves ?redirect query param
- [ ] T115 [P] [US3] Create useTripDetail hook in frontend/src/features/trips/hooks/useTripDetail.ts: TanStack Query query for GET /trips/:id, returns trip with full itinerary
- [ ] T116 [P] [US3] Create useUpdateTrip hook in frontend/src/features/trips/hooks/useUpdateTrip.ts: TanStack Query mutation for PUT /trips/:id, invalidates trip and trips queries on success
- [ ] T117 [P] [US3] Create useDeleteTrip hook in frontend/src/features/trips/hooks/useDeleteTrip.ts: TanStack Query mutation for DELETE /trips/:id, invalidates trips query, redirects to /dashboard on success
- [ ] T118 [US3] Modify TripDetailPage in frontend/src/features/trips/pages/TripDetailPage.tsx: if user is creator and hasSubscription=true, show "Edit Trip" button → opens edit modal; show "Delete Trip" button → opens confirmation modal; if user is creator and in grace period, show "Renew Subscription" banner with disabled edit button
- [ ] T119 [P] [US3] Create DeleteTripModal component in frontend/src/features/trips/components/DeleteTripModal.tsx: confirmation modal "Are you sure you want to delete this trip?", uses useDeleteTrip hook
- [ ] T120 [US3] Modify authStore in frontend/src/stores/authStore.ts: add logout action that clears user state, add setUser action that updates user from API response
- [ ] T121 [US3] Modify protected route wrapper in frontend/src/routes/router.tsx: check authStore.isAuthenticated, redirect to /login?redirect=<current-path> if not authenticated, restore redirect after login
- [ ] T122 [US3] Create Navigation component in frontend/src/components/features/Navigation.tsx: displays user name, has_subscription badge, "Log Out" button, uses useLogout hook, responsive mobile menu

**Checkpoint**: Users can login, view dashboard, manage trips, logout (US3 complete and independently testable)

---

## Phase 6: User Story 4 - Collaboration: Invite & Manage Suggestions (Priority: P2)

**Goal**: Enable collaborators to submit suggestions, creators to approve/reject suggestions, update trip itinerary

**Independent Test**: Paid User invites Free User, Free User submits suggestion, Paid User reviews and approves, verify trip updated with suggestion

### Backend Implementation for US4

- [ ] T123 [P] [US4] Create Suggestion repository in backend/internal/collaboration/repository.go: CreateSuggestion(suggestion), GetSuggestionsByTripID(tripID), GetSuggestionByID(id), UpdateSuggestionStatus(id, status), ApplySuggestionToTrip(suggestionID) (stub: logs to console "Applied suggestion [id] to trip")
- [ ] T124 [US4] Implement POST /trips/:trip_id/suggestions handler in backend/internal/collaboration/handler.go: validates auth, checks user is collaborator on trip (not creator, returns 403 if creator), validates request body (suggestion_type, content, details), creates Suggestion with status='pending', returns 201 with suggestion JSON
- [ ] T125 [US4] Register POST /api/v1/trips/:trip_id/suggestions route in backend/cmd/api/main.go: attach create suggestion handler with auth middleware
- [ ] T126 [P] [US4] Implement GET /trips/:trip_id/suggestions handler in backend/internal/collaboration/handler.go: validates auth, checks user is creator OR collaborator, returns suggestions for trip filtered by status (optional ?status=pending query param), returns 200 with suggestions array JSON
- [ ] T127 [P] [US4] Register GET /api/v1/trips/:trip_id/suggestions route in backend/cmd/api/main.go: attach list suggestions handler with auth middleware
- [ ] T128 [US4] Implement PATCH /suggestions/:id/approve handler in backend/internal/collaboration/handler.go: validates auth, retrieves suggestion, checks user is trip creator (returns 403 if not), checks suggestion status='pending' (returns 409 if already processed), updates status='approved', sets ApprovedAt=now, calls ApplySuggestionToTrip (stub implementation), returns 200 with suggestion JSON and message "Suggestion approved"
- [ ] T129 [US4] Register PATCH /api/v1/suggestions/:id/approve route in backend/cmd/api/main.go: attach approve suggestion handler with auth middleware
- [ ] T130 [P] [US4] Implement PATCH /suggestions/:id/reject handler in backend/internal/collaboration/handler.go: validates auth, checks user is trip creator, checks suggestion status='pending' (returns 409 if already processed), updates status='rejected', sets RejectedAt=now, returns 200 with suggestion JSON and message "Suggestion rejected"
- [ ] T131 [P] [US4] Register PATCH /api/v1/suggestions/:id/reject route in backend/cmd/api/main.go: attach reject suggestion handler with auth middleware

### Frontend Implementation for US4

- [ ] T132 [P] [US4] Create useSuggestions hook in frontend/src/features/collaboration/hooks/useSuggestions.ts: TanStack Query query for GET /trips/:trip_id/suggestions, returns trip suggestions grouped by status
- [ ] T133 [P] [US4] Create useCreateSuggestion hook in frontend/src/features/collaboration/hooks/useCreateSuggestion.ts: TanStack Query mutation for POST /trips/:trip_id/suggestions, invalidates suggestions query on success
- [ ] T134 [P] [US4] Create useApproveSuggestion hook in frontend/src/features/collaboration/hooks/useApproveSuggestion.ts: TanStack Query mutation for PATCH /suggestions/:id/approve, invalidates suggestions and trip queries on success
- [ ] T135 [P] [US4] Create useRejectSuggestion hook in frontend/src/features/collaboration/hooks/useRejectSuggestion.ts: TanStack Query mutation for PATCH /suggestions/:id/reject, invalidates suggestions query on success
- [ ] T136 [US4] Create SuggestionCard component in frontend/src/features/collaboration/components/SuggestionCard.tsx: displays suggestion type, content, details, collaborator name, status badge (pending/approved/rejected), timestamp; if user is creator and status='pending', show "Approve" and "Reject" buttons; uses Card, Badge
- [ ] T137 [US4] Create SuggestionsList component in frontend/src/features/collaboration/components/SuggestionsList.tsx: renders list of SuggestionCard, grouped by status, uses useSuggestions hook, shows EmptyState if no suggestions
- [ ] T138 [US4] Modify TripDetailPage in frontend/src/features/trips/pages/TripDetailPage.tsx: if user is collaborator (not creator), replace "Edit" buttons with "Suggest Change" button next to itinerary items; if user is creator, add "Review Suggestions" section displaying pending suggestions count badge, click opens SuggestionsList modal
- [ ] T139 [US4] Create SuggestChangeModal component in frontend/src/features/collaboration/components/SuggestChangeModal.tsx: form with suggestion_type select, content textarea, details JSONB editor (simple key-value inputs), "Submit Suggestion" button, uses useCreateSuggestion hook, shows success message "Suggestion submitted. Waiting for creator approval."

**Checkpoint**: Collaborators can suggest changes, creators can approve/reject, suggestions workflow complete (US4 complete and independently testable)

---

## Phase 7: User Story 5 - Password Management & Account Security (Priority: P2)

**Goal**: Implement password reset flow, password change in settings, session management

**Independent Test**: Request password reset, receive link (check console logs), complete reset, login with new password; change password in settings, verify sessions revoked on other devices

### Backend Implementation for US5

- [ ] T140 [P] [US5] Create PasswordResetToken repository in backend/internal/auth/repository.go: CreatePasswordResetToken(token), GetPasswordResetTokenByHash(hash), MarkTokenAsUsed(id), InvalidatePreviousTokens(userID), CleanupExpiredTokens()
- [ ] T141 [P] [US5] Create RefreshToken repository in backend/internal/auth/repository.go: CreateRefreshToken(token), GetRefreshTokenByHash(hash), RevokeRefreshToken(id), RevokeAllRefreshTokens(userID), RevokeAllRefreshTokensExceptCurrent(userID, currentTokenID)
- [ ] T142 [US5] Implement POST /auth/password-reset-request handler in backend/internal/auth/handler.go: validates email format, retrieves user by email (if not found, still returns 200 "Link sent" to prevent enumeration, but logs security event with user_exists=false), invalidates previous reset tokens for user, generates 32-byte random token, hashes with SHA-256, creates PasswordResetToken with expires_at=now+1h, logs to console "[DEMO] Password reset link: http://localhost:3000/password-reset/complete?token=[token]", logs auth_password_reset_request security event, returns 200 with generic message "If an account with that email exists, a password reset link has been sent."
- [ ] T143 [US5] Register POST /api/v1/auth/password-reset-request route in backend/cmd/api/main.go: attach password reset request handler with rate limit middleware (10/hour per IP)
- [ ] T144 [US5] Implement POST /auth/password-reset-complete handler in backend/internal/auth/handler.go: validates request body (token, new_password), hashes token with SHA-256, retrieves PasswordResetToken by token_hash, validates expires_at>now AND used_at IS NULL (returns 400 "Invalid or expired token" if validation fails), retrieves user, validates new password format, hashes new password with bcrypt cost 12, updates user.password_hash, marks token as used (set used_at=now), revokes all refresh tokens for user (force re-login on all devices), logs auth_password_reset_complete security event, returns 200 with message "Password reset successfully"
- [ ] T145 [US5] Register POST /api/v1/auth/password-reset-complete route in backend/cmd/api/main.go: attach password reset complete handler with rate limit middleware
- [ ] T146 [P] [US5] Implement POST /auth/password-change handler in backend/internal/auth/handler.go: validates auth, validates request body (current_password, new_password, log_out_all_other_devices boolean), retrieves user, compares current_password hash with bcrypt (returns 401 if mismatch), validates new password format, hashes new password with bcrypt cost 12, updates user.password_hash, if log_out_all_other_devices=true revokes all refresh tokens except current token, if log_out_all_other_devices=false no token revocation, logs auth_password_change security event, returns 200 with message "Password changed successfully"
- [ ] T147 [P] [US5] Register POST /api/v1/auth/password-change route in backend/cmd/api/main.go: attach password change handler with auth middleware
- [ ] T148 [P] [US5] Implement POST /auth/refresh handler in backend/internal/auth/handler.go: extracts refresh_token from cookie, validates refresh token via JWT validator, checks token not revoked in RefreshToken table (returns 401 "Invalid refresh token" if revoked or expired), generates new access token (24h), optionally rotates refresh token (generate new 30d, revoke old one), sets new cookies, logs auth_token_refresh security event, returns 200 with message "Token refreshed"
- [ ] T149 [P] [US5] Register POST /api/v1/auth/refresh route in backend/cmd/api/main.go: attach token refresh handler (no auth middleware, uses refresh token from cookie)
- [ ] T150 [US5] Modify auth middleware in backend/internal/security/middleware.go: on 401 Unauthorized (expired access token), return 401 with error="TOKEN_EXPIRED"; frontend will detect and call /auth/refresh

### Frontend Implementation for US5

- [ ] T151 [P] [US5] Create usePasswordResetRequest hook in frontend/src/features/auth/hooks/usePasswordResetRequest.ts: TanStack Query mutation for POST /auth/password-reset-request, shows success message "Link sent"
- [ ] T152 [P] [US5] Create usePasswordResetComplete hook in frontend/src/features/auth/hooks/usePasswordResetComplete.ts: TanStack Query mutation for POST /auth/password-reset-complete, redirects to /login on success with success banner "Password reset successfully"
- [ ] T153 [P] [US5] Create usePasswordChange hook in frontend/src/features/auth/hooks/usePasswordChange.ts: TanStack Query mutation for POST /auth/password-change, shows success message, optionally logs out other devices
- [ ] T154 [US5] Create PasswordResetRequestForm component in frontend/src/features/auth/components/PasswordResetRequestForm.tsx: form with email input, "Send Reset Link" button, uses usePasswordResetRequest hook, shows generic success message
- [ ] T155 [US5] Create PasswordResetRequestPage in frontend/src/features/auth/pages/PasswordResetRequestPage.tsx: renders PasswordResetRequestForm, heading "Reset Your Password", link back to /login
- [ ] T156 [P] [US5] Create PasswordResetCompleteForm component in frontend/src/features/auth/components/PasswordResetCompleteForm.tsx: form with new_password, confirm_password inputs, "Reset Password" button, extracts token from URL query param, uses usePasswordResetComplete hook, shows error if token invalid/expired
- [ ] T157 [P] [US5] Create PasswordResetCompletePage in frontend/src/features/auth/pages/PasswordResetCompletePage.tsx: renders PasswordResetCompleteForm, heading "Set New Password"
- [ ] T158 [P] [US5] Create PasswordChangeForm component in frontend/src/features/auth/components/PasswordChangeForm.tsx: form with current_password, new_password, confirm_password inputs, "Log out all other devices" checkbox, "Change Password" button, uses usePasswordChange hook, shows success message
- [ ] T159 [P] [US5] Create AccountSettingsPage in frontend/src/features/auth/pages/AccountSettingsPage.tsx: renders PasswordChangeForm, heading "Account Settings", protected route
- [ ] T160 [US5] Modify axios interceptor in frontend/src/services/api.ts: on 401 response with error="TOKEN_EXPIRED", call POST /auth/refresh, retry original request with new access token; if refresh fails (401 "Invalid refresh token"), clear authStore, redirect to /login?redirect=<original-path> with message "Session expired. Please log in again."

**Checkpoint**: Password reset and password change flows complete, session management with automatic refresh (US5 complete and independently testable)

---

## Phase 8: User Story 6 - UI Component Library & Design System Basics (Priority: P3)

**Goal**: Establish consistent visual identity, accessible components, error/loading/empty states

**Independent Test**: Navigate through all features, verify consistent styling (colors, spacing, typography), keyboard navigation works, focus indicators visible, no raw errors displayed

### Frontend Implementation for US6

- [ ] T161 [P] [US6] Create Checkbox primitive in frontend/src/components/primitives/Checkbox.tsx: <Checkbox> with label, checked, onChange, disabled props; uses design tokens; keyboard accessible; associated label
- [ ] T162 [P] [US6] Create Select primitive in frontend/src/components/primitives/Select.tsx: <Select> with options, label, error, disabled props; uses design tokens; keyboard accessible; ARIA attributes
- [ ] T163 [P] [US6] Create Textarea primitive in frontend/src/components/primitives/Textarea.tsx: <Textarea> with label, error, disabled, maxLength props; uses design tokens; character counter
- [ ] T164 [P] [US6] Create Link primitive in frontend/src/components/primitives/Link.tsx: <Link> wraps React Router Link; uses design tokens; keyboard accessible; visible focus indicator
- [ ] T165 [P] [US6] Create Toast composite in frontend/src/components/composites/Toast.tsx: <Toast> notification with type (success/error/info/warning), message, auto-dismiss; uses design tokens; aria-live for screen readers
- [ ] T166 [P] [US6] Create Tooltip composite in frontend/src/components/composites/Tooltip.tsx: <Tooltip> with trigger, content; uses design tokens; keyboard accessible (on focus); ARIA attributes
- [ ] T167 [US6] Create ErrorBoundary feature in frontend/src/components/features/ErrorBoundary.tsx: React error boundary component, catches runtime errors, displays user-friendly "Something went wrong" message with "Reload Page" button, logs error to console (no stack trace to user)
- [ ] T168 [P] [US6] Audit all forms for accessibility: verify all <Input>/<Select>/<Checkbox> have associated <Label> with htmlFor, verify error messages use ErrorMessage component with aria-live, verify focus indicators visible on all interactive elements
- [ ] T169 [P] [US6] Audit all interactive elements for keyboard accessibility: verify all buttons/links/form elements reachable via Tab, verify Enter/Space activate buttons, verify ESC closes modals, verify focus trap in modals
- [ ] T170 [P] [US6] Audit all text for color contrast: verify all text meets WCAG AA 4.5:1 contrast ratio against background, adjust design tokens if needed, use browser DevTools Accessibility panel
- [ ] T171 [US6] Audit all async actions for loading states: verify all mutations (register, login, create trip, submit suggestion, etc.) show <LoadingSpinner> during processing, disable submit button during loading to prevent double submission
- [ ] T172 [US6] Audit all error scenarios for user-friendly messages: verify all API errors mapped via errorHandler.ts to clear messages, verify no raw error messages or stack traces displayed, verify actionable next steps provided (e.g., "Try again" button)
- [ ] T173 [US6] Audit all empty states for EmptyState component: verify "No trips yet" → "Create Your First Trip" button, "No invitations" → "Invitations appear here" message, "No suggestions" → "Suggestions appear here" message
- [ ] T174 [P] [US6] Create documentation for design system in frontend/src/styles/README.md: document all design tokens (colors, spacing, typography, focus), document Atomic Design component layering (Primitives → Composites → Features), document accessibility guidelines (keyboard nav, focus indicators, WCAG AA), provide usage examples for each token

**Checkpoint**: Design system complete, all components accessible, consistent visual identity (US6 complete and independently testable)

---

## Phase 9: Subscription Lifecycle Management (Additional Priority: P2)

**Goal**: Handle subscription cancellation, 30-day grace period with read-only trips, trip archival, resubscription with restoration

**Independent Test**: Paid User cancels subscription, verifies grace period banner displayed, owned trips become read-only (edit disabled), after 30 days trips archived (hidden), resubscribe to restore trips

### Backend Implementation for Subscription Lifecycle

- [ ] T175 [P] Implement DELETE /subscriptions/:id handler in backend/internal/subscription/handler.go: validates auth, checks user owns subscription, updates Subscription set status='cancelled', cancelled_at=now, grace_period_ends_at=now+30days, updates User.has_subscription=false, returns 200 with subscription JSON and message "Subscription cancelled. Read-only access until [grace_period_ends_at]. Renew anytime."
- [ ] T176 [P] Register DELETE /api/v1/subscriptions/:id route in backend/cmd/api/main.go: attach cancel subscription handler with auth middleware
- [ ] T177 [P] Implement POST /subscriptions/:id/renew handler in backend/internal/subscription/handler.go: validates auth, checks user owns subscription, checks subscription status='cancelled', calls payment provider ProcessPayment (stub succeeds), updates Subscription set status='active', cancelled_at=null, grace_period_ends_at=null, updates User.has_subscription=true, updates all owned trips set archived=false, returns 200 with subscription JSON and message "Subscription renewed successfully"
- [ ] T178 [P] Register POST /api/v1/subscriptions/:id/renew route in backend/cmd/api/main.go: attach renew subscription handler with auth middleware
- [ ] T179 Create scheduled job for trip archival in backend/cmd/api/archival_job.go: runs daily (cron), queries subscriptions where grace_period_ends_at < now AND status='cancelled', updates all trips where creator_id IN (affected users) set archived=true, logs to console "Archived [count] trips for [count] users"
- [ ] T180 Modify POST /subscriptions handler in backend/internal/subscription/handler.go: if user already has subscription with status='cancelled', allow resubscription by reusing existing subscription record (set status='active', clear cancellation fields), if user has active subscription return 409 "Subscription already active"

### Frontend Implementation for Subscription Lifecycle

- [ ] T181 [P] Create useCancelSubscription hook in frontend/src/features/subscription/hooks/useCancelSubscription.ts: TanStack Query mutation for DELETE /subscriptions/:id, invalidates subscription and user queries on success
- [ ] T182 [P] Create useRenewSubscription hook in frontend/src/features/subscription/hooks/useRenewSubscription.ts: TanStack Query mutation for POST /subscriptions/:id/renew, invalidates subscription, user, and trips queries on success
- [ ] T183 [P] Create useUpgrade hook in frontend/src/features/subscription/hooks/useUpgrade.ts: TanStack Query mutation for POST /subscriptions (upgrade Free User to Paid User), updates authStore.user.has_subscription=true on success, invalidates trips query to show restored/enabled trip creation
- [ ] T184 Create GracePeriodBanner component in frontend/src/features/subscription/components/GracePeriodBanner.tsx: Banner with type='warning', message "Your subscription has expired. Renew to continue editing your trips. Read-only access until [grace_period_ends_at].", "Renew Subscription" button, uses useRenewSubscription hook
- [ ] T185 Create UpgradePrompt component in frontend/src/features/subscription/components/UpgradePrompt.tsx: Banner with type='info', message "Upgrade to create your own trips", "Subscribe Now" button opens payment flow, uses useUpgrade hook
- [ ] T186 Create SubscriptionCheckout component in frontend/src/features/subscription/components/SubscriptionCheckout.tsx: form with plan selection (Monthly $9.99), payment stub labeled "[DEMO] Subscription Checkout", "Complete Subscription" button (always succeeds), uses useUpgrade or POST /subscriptions mutation depending on context (register vs upgrade)
- [ ] T187 Modify TripDashboard in frontend/src/features/trips/components/TripDashboard.tsx: if user.has_subscription=false AND subscription.status='cancelled' AND grace_period_ends_at NOT NULL, display GracePeriodBanner at top; if user.has_subscription=false AND no subscription, display UpgradePrompt
- [ ] T188 Modify TripDetailPage in frontend/src/features/trips/pages/TripDetailPage.tsx: if user is creator AND in grace period (has_subscription=false, grace_period_ends_at>now), disable "Edit Trip" and "Delete Trip" buttons with tooltip "Renew subscription to edit", show GracePeriodBanner
- [ ] T189 Create UpgradePage in frontend/src/features/subscription/pages/UpgradePage.tsx: renders SubscriptionCheckout, heading "Upgrade to Paid User", displays plan benefits (create unlimited trips, invite collaborators, AI itinerary generation), uses useUpgrade hook, redirects to /dashboard on success
- [ ] T190 Modify Navigation component in frontend/src/components/features/Navigation.tsx: if user.has_subscription=false AND no subscription, show "Upgrade" button in nav; if in grace period, show "Renew" button

**Checkpoint**: Subscription lifecycle complete with grace period, archival, and restoration (additional feature complete and independently testable)

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements affecting multiple user stories

- [ ] T191 [P] Add comprehensive error logging to backend: wrap all errors with context using fmt.Errorf, log to CloudWatch Logs with correlation ID, structured JSON format
- [ ] T192 [P] Add request logging middleware in backend/internal/security/middleware.go: log all requests with method, path, status, duration, correlation ID to CloudWatch
- [ ] T193 [P] Add security headers middleware in backend/internal/security/middleware.go: set Content-Security-Policy, X-Frame-Options, X-Content-Type-Options, Strict-Transport-Security headers
- [ ] T194 [P] Run quickstart.md validation: execute all 6 scenarios (Paid User registration, Free User registration, Login, Password reset, Collaboration, Free User upgrade), verify all acceptance criteria met, document any deviations
- [ ] T195 [P] Backend code cleanup: remove unused imports, remove commented code, run gofmt on all files, run golangci-lint and fix all errors
- [ ] T196 [P] Frontend code cleanup: remove unused imports, remove commented code, run Prettier on all files, run ESLint and fix all errors
- [ ] T197 [P] Update README in backend/: document environment variables, database setup, migration commands, how to run server locally, API endpoint documentation
- [ ] T198 [P] Update README in frontend/: document environment variables, how to run dev server, how to run tests (Vitest, Playwright), component library structure
- [ ] T199 [P] Create E2E test for US1 in e2e/specs/auth/paid-user-registration.spec.ts: Playwright test navigates to /register, fills form with payment, submits, verifies redirect to /dashboard, creates trip, verifies trip in list
- [ ] T200 [P] Create E2E test for US2 in e2e/specs/auth/free-user-registration.spec.ts: Playwright test for Free User registration, accept invitation, verify single-collaboration limit enforced
- [ ] T201 [P] Create E2E test for US3 in e2e/specs/auth/login-and-dashboard.spec.ts: Playwright test for login with valid credentials, verify dashboard loads, verify rate limiting after 6 failed attempts
- [ ] T202 [P] Create E2E test for US4 in e2e/specs/collaboration/invite-and-suggestions.spec.ts: Playwright test for invite collaborator, submit suggestion, approve suggestion, verify trip updated
- [ ] T203 [P] Create E2E test for US5 in e2e/specs/auth/password-management.spec.ts: Playwright test for password reset request, complete reset via link, login with new password, change password in settings
- [ ] T204 [P] Create E2E accessibility test in e2e/specs/accessibility/wcag-compliance.spec.ts: Playwright test with axe-core integration, verify all pages pass WCAG AA checks for color contrast, keyboard navigation, ARIA attributes
- [ ] T205 Security audit: review all authentication endpoints for vulnerabilities (SQL injection, XSS, CSRF), verify secrets in environment variables, verify no sensitive data in logs, verify rate limiting on all public endpoints
- [ ] T206 Performance optimization: add indexes to frequently queried columns (email, user_id, trip_id, status), optimize N+1 queries with joins/batch loading, add caching headers to static assets

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1: Setup**: No dependencies - can start immediately
- **Phase 2: Foundational**: Depends on Setup completion - **BLOCKS all user stories**
- **Phase 3-8: User Stories**: All depend on Foundational phase completion
  - Phase 3 (US1): Can start after Foundational ✅
  - Phase 4 (US2): Can start after Foundational ✅ (independent of US1)
  - Phase 5 (US3): Can start after Foundational ✅ (independent of US1/US2)
  - Phase 6 (US4): Can start after Foundational ✅, integrates with US2 (collaboration model) but independently testable
  - Phase 7 (US5): Can start after Foundational ✅ (independent of other stories)
  - Phase 8 (US6): Can start after Foundational ✅ (independent of other stories)
- **Phase 9: Subscription Lifecycle**: Depends on US1 (subscription model) and US2 (collaboration model)
- **Phase 10: Polish**: Depends on all desired user stories being complete

### User Story Completion Order for MVP

**Recommended Priority**:
1. **Phase 3 (US1)** - Paid User Registration & First Trip (P1) → Delivers core value proposition
2. **Phase 5 (US3)** - Login & Trip Management (P1) → Enables returning users
3. **Phase 4 (US2)** - Free User Registration & Collaboration (P1) → Enables viral growth
4. **Phase 6 (US4)** - Collaboration Suggestions (P2) → Completes collaboration workflow
5. **Phase 7 (US5)** - Password Management (P2) → Account security
6. **Phase 9** - Subscription Lifecycle (P2) → Revenue protection
7. **Phase 8 (US6)** - Design System (P3) → Polish

**MVP Checkpoint**: Stop after Phase 3 + Phase 5 (US1 + US3) for minimum viable product (Paid User can register, create trips, login, view trips). Deploy and validate before continuing.

### Parallel Opportunities

**Phase 1 (Setup)**: T002, T004, T005, T006, T007, T008 can run in parallel (different directories)

**Phase 2 (Foundational)**: 
- T010-T017 migrations can run in parallel (different files)
- T019-T036 can run in parallel (different files/domains)

**User Story Phases (3-8)**: After Foundational complete, all user stories can be implemented in parallel by different developers if team capacity allows. Each story is independently testable.

**Within Each User Story**:
- Backend tasks marked [P] can run in parallel (different files)
- Frontend tasks marked [P] can run in parallel (different components)
- Backend and Frontend work for same story can run in parallel (different developers)

**Phase 10 (Polish)**: T191-T206 can run in parallel (different concerns/files)

### Parallel Example: User Story 1 (Phase 3)

```bash
# Backend team can work on these simultaneously:
T037 [P] User model (auth/models.go)
T038 [P] RegisterRequest model (auth/models.go - different struct)
T039 [P] Subscription model (subscription/models.go - different file)
T042 [P] Payment provider interface (subscription/payment/provider.go)
T043 [P] Stub payment provider (subscription/payment/stub.go)

# Frontend team can work on these simultaneously:
T052 [P] Button primitive
T053 [P] Input primitive
T054 [P] Label primitive
T055 [P] ErrorMessage primitive
T057 [P] LoadingSpinner feature

# After models complete, services can start:
T040 User repository (depends on T037)
T041 [P] Subscription repository (depends on T039)
T044 Auth service Register (depends on T040, T041)
T045 Subscription service CreateSubscription (depends on T041, T043)

# Once services ready, handlers can start:
T046 POST /auth/register handler (depends on T044, T045)
T050 POST /trips handler (depends on T048, T049)

# Frontend integration after API ready:
T058 useRegister hook (depends on T046 backend complete)
T060 RegisterForm (depends on T058, primitives T052-T055)
T061 RegisterPage (depends on T060)
```

---

## Implementation Strategy

### MVP First (User Story 1 + 3 Only)

1. Complete Phase 1: Setup (8 tasks)
2. Complete Phase 2: Foundational (28 tasks) ⚠️ **CRITICAL BLOCKER**
3. Complete Phase 3: User Story 1 (34 tasks) → Paid User registration + first trip
4. Complete Phase 5: User Story 3 (12 tasks) → Login + trip management
5. **STOP and VALIDATE**: Test US1 + US3 independently via quickstart scenarios 1 & 3
6. Deploy MVP, gather feedback

**Total MVP Tasks**: 82 tasks  
**Estimated MVP Time**: 2-3 weeks (single full-time developer)

### Incremental Delivery (After MVP)

1. Add Phase 4: User Story 2 (26 tasks) → Free User + collaboration invites
2. Validate independently, deploy
3. Add Phase 6: User Story 4 (19 tasks) → Suggestion workflow
4. Validate independently, deploy
5. Add Phase 7: User Story 5 (21 tasks) → Password management
6. Validate independently, deploy
7. Add Phase 9: Subscription Lifecycle (16 tasks) → Grace period + archival
8. Validate independently, deploy
9. Add Phase 8: User Story 6 (14 tasks) → Design system polish
10. Complete Phase 10: Polish (16 tasks) → Cross-cutting concerns

**Total All Tasks**: 206 tasks  
**Estimated Full Time**: 4-6 weeks (single full-time developer)

### Parallel Team Strategy

With 3 developers after Foundational phase completes:

- **Developer A**: Phase 3 (US1) → Phase 5 (US3) → Phase 9 (Subscription)
- **Developer B**: Phase 4 (US2) → Phase 6 (US4) → Phase 10 (Polish backend)
- **Developer C**: Phase 7 (US5) → Phase 8 (US6) → Phase 10 (Polish frontend + E2E)

**Estimated Parallel Time**: 2-3 weeks with 3 developers

---

## Summary

**Total Tasks**: 206  
**Total Phases**: 10  
**User Stories**: 6 (US1-US6)  
**MVP Scope**: 82 tasks (Phases 1, 2, 3, 5) → Paid User registration + trip creation + login  
**Priority Distribution**:
- P1 (MVP-critical): 72 tasks across US1, US2, US3
- P2 (Post-MVP): 100 tasks across US4, US5, Subscription Lifecycle
- P3 (Polish): 34 tasks across US6, Phase 10

**Parallel Opportunities Identified**: 
- Phase 1: 6 parallel tasks
- Phase 2: 28 parallel tasks (migrations + infrastructure)
- Per User Story: ~15-20 parallel tasks (backend/frontend split + [P] markers)
- Phase 10: 16 parallel tasks

**Independent Test Criteria**: Each user story has clear validation scenario in quickstart.md, can be tested without other stories complete, delivers standalone value.

**Ready for Implementation**: All tasks include exact file paths, follow checklist format with [Task ID], [P] markers, [Story] labels, and clear descriptions. Start with Phase 1 → Phase 2 → Phase 3 (US1 MVP).
