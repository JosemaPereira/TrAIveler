# Feature Specification: Authentication & Collaboration User Experience

**Feature Branch**: `008-auth-collaboration-ux`

**Created**: 2026-07-06

**Status**: Draft

**Input**: User description: "Define user types (system administrators, public users), capabilities, and workflows from login to account creation, including how to create a trip and invite collaborators. It also verifies the basic user interface (we currently have no graphical or manual UX/UI identity). Finally, it checks for any outstanding edge cases related to security, UI, and backend."

## User Types & Capabilities

The system supports two end-user types with distinct capabilities:

**Paid User** (requires active subscription):
- Full trip ownership: Create, read, update, delete own trips
- AI-powered itinerary generation with conversational multi-turn input
- Invite up to 1 collaborator per trip (basic plan limit)
- Approve or reject collaborator suggestions
- Full CRUD on own trip itinerary items
- Can collaborate on other Paid Users' trips (as Free User behavior when invited)

**Free User** (no subscription required):
- Registration and account creation without payment
- Can be invited to collaborate on ONE trip at a time
- Must leave current collaboration before accepting a new invitation
- View-only access to invited trip itineraries
- Submit suggestions for modifications (cannot directly edit)
- Can upgrade to Paid User by purchasing a subscription
- Cannot create own trips until subscription purchased

## User Scenarios & Testing *(mandatory)*

### User Story 1 - New Paid User Registration & First Trip Creation (Priority: P1)

A first-time visitor discovers TrAIveler, creates an account through a subscription flow, and creates their first trip with AI assistance.

**Why this priority**: This is the primary onboarding funnel for Paid Users - every trip creator must complete this flow to use the product. Without a working registration and first-trip experience, no other features matter.

**Independent Test**: Can be fully tested by navigating to the landing page, completing registration with stub payment, creating a trip via AI generation, and verifying the trip appears in the dashboard. Delivers a complete end-to-end experience from anonymous visitor to active user with content.

**Acceptance Scenarios**:

1. **Given** a visitor lands on the homepage, **When** they click "Get Started" or "Sign Up", **Then** they see a registration form with email, password, and two options: "Continue to Payment" (become Paid User) or "Create Free Account" (become Free User)
2. **Given** a user selects "Continue to Payment" and enters valid email/password, **When** they proceed to payment, **Then** they see a stub payment page clearly labeled "[DEMO] Subscription Checkout" with plan details and "Complete Subscription" button
3. **Given** a user on the stub payment page, **When** they click "Complete Subscription", **Then** payment always succeeds, they are logged in automatically as a Paid User, and redirected to the trip creation page
4. **Given** a newly registered user on the trip creation page, **When** they describe their desired trip in natural language (destination, duration, interests), **Then** the AI asks clarifying questions before generating an itinerary
5. **Given** the AI has generated an itinerary, **When** the user views the results, **Then** they see a day-by-day plan with destinations, activities, and food recommendations organized by day
6. **Given** a user has a generated trip, **When** they navigate to "My Trips", **Then** they see their trip listed with title, destination summary, and "View Details" action

---

### User Story 2 - Free User Registration & Accepting Collaboration Invite (Priority: P1)

A user receives a collaboration invitation email, creates a free account (no payment required), and accepts the invitation to view and collaborate on the trip.

**Why this priority**: Free User onboarding enables viral growth and reduces collaboration friction. This is the primary path for non-paying users to experience the product and potentially convert to Paid Users.

**Independent Test**: Can be tested by creating a Paid User account, sending a collaboration invite to an email, registering as a Free User with that email, accepting the invitation, and verifying view + suggest capabilities. Delivers complete free collaboration onboarding independently.

**Acceptance Scenarios**:

1. **Given** a visitor receives a collaboration invitation email, **When** they click the invitation link and don't have an account, **Then** they see a registration form with email pre-filled, password field, and "Create Free Account" button (no payment required)
2. **Given** a new Free User completes registration, **When** they log in, **Then** they are redirected to the invitation acceptance page showing trip details and "Accept Invitation" button
3. **Given** a Free User views the invitation, **When** they click "Accept Invitation", **Then** the system creates a Collaborator record linking them to the trip, they see the trip under "Shared with Me", and are badged as "Free User" with "View Only + Suggest" permissions
4. **Given** a Free User is already collaborating on one trip, **When** they receive a second invitation, **Then** the system displays "You are already collaborating on [Trip Name]. Leave that trip first to accept this invitation." with a "Leave Current Trip" button
5. **Given** a Free User clicks "Leave Current Trip", **When** they confirm the action, **Then** their Collaborator record is deleted, they can now accept the new invitation, and see a confirmation "You have left [Trip Name]. You can now accept new invitations."
6. **Given** a Free User on their dashboard, **When** they click "Create Trip" or "Start Planning", **Then** they see an upgrade prompt "Upgrade to create your own trips" with a "Subscribe Now" button leading to the payment flow
7. **Given** a Free User completes subscription purchase, **When** payment succeeds, **Then** User.has_subscription is set to true, user badge updates from "Free User" to "Paid User", existing collaboration on shared trip remains active and uninterrupted, one-collaboration limit is removed, and user can now create their own trips while continuing to collaborate on the original trip

---

### User Story 3 - Returning User Login & Trip Management (Priority: P1)

An existing user logs in with email/password, views their trip dashboard, and performs basic trip operations (view, edit, delete).

**Why this priority**: Returning user authentication is critical for session persistence and trip management. This is the second most common user flow after registration.

**Independent Test**: Can be tested by creating a test user account, logging out, logging back in with email/password, and verifying access to saved trips. Delivers authentication persistence without requiring any collaboration or advanced features.

**Acceptance Scenarios**:

1. **Given** a returning user visits the homepage, **When** they click "Log In", **Then** they see a login form with email, password, "Log In" button, and "Forgot Password?" link
2. **Given** a user enters valid credentials, **When** they click "Log In", **Then** they are authenticated with a secure session token and redirected to their trip dashboard
3. **Given** a user enters invalid credentials, **When** they click "Log In", **Then** they see "Invalid email or password" error message without revealing which field is incorrect (security)
4. **Given** an authenticated user on their dashboard, **When** they view "My Trips", **Then** they see trips organized into two sections: "My Trips" (trips they own as Paid User) and "Shared with Me" (trips where they are collaborating), with Free Users seeing only "Shared with Me" section and Paid Users seeing both
5. **Given** a user viewing their trip dashboard, **When** they click a trip card, **Then** they navigate to the trip detail page showing the full itinerary organized by day
6. **Given** a Paid User on a trip detail page they own, **When** they click "Edit Trip" or "Delete Trip", **Then** they can modify the trip title/description or permanently delete the trip with a confirmation modal

---

### User Story 3 - Collaboration: Invite Free User & Manage Suggestions (Priority: P2)

A Paid User invites a Free User to collaborate on their trip, the Free User submits modification suggestions, and the Paid User approves or rejects them.

**Why this priority**: Collaboration is a core value proposition but depends on users first having trips (P1). This is the first true multi-user feature and enables shared trip planning.

**Independent Test**: Can be tested by creating two user accounts (one Paid User, one Free User), Paid User invites Free User via email, Free User accepts and views trip, Free User submits suggestion, Paid User reviews and approves/rejects. Delivers complete suggest-then-approve workflow independently of AI generation or advanced trip features.

**Acceptance Scenarios**:

1. **Given** a Paid User on their trip detail page, **When** they click "Invite Collaborator", **Then** they see a form to enter an email address with a "Send Invite" button and a note "Basic plan: 1 collaborator maximum"
2. **Given** a Paid User enters a valid email address, **When** they click "Send Invite", **Then** the system creates a Collaborator record with status 'pending', sends an email notification to the invitee (stub email logs link to console in MVP), and shows a success message "Invitation sent to [email]"
3. **Given** a Free User receives an invite and is not already collaborating, **When** they log in and view their dashboard, **Then** they see the invited trip under "Shared with Me" section with a badge indicating "Free User" role and "View Only + Suggest" permissions
4. **Given** a Free User viewing a shared trip, **When** they want to suggest a change (add/edit/remove activity), **Then** they see a "Suggest Change" button next to itinerary items instead of direct "Edit" actions
5. **Given** a Free User clicks "Suggest Change", **When** they submit a suggestion with a description of the change, **Then** the system creates a Suggestion record with status 'pending', the Paid User sees a notification badge on the trip, and the Free User sees their suggestion listed as "Pending Review"
6. **Given** a Paid User views pending suggestions on their trip, **When** they click "Approve" or "Reject" on a suggestion, **Then** the suggestion status updates to 'approved' or 'rejected', approved suggestions apply changes to the itinerary, and the Free User receives a notification of the decision

---

### User Story 4 - Password Management & Account Security (Priority: P2)

A user resets a forgotten password or changes their password from account settings, with options for session management.

**Why this priority**: Security and account recovery are important but not blocking for initial trip creation. Users need this functionality but it's secondary to core trip planning workflows.

**Independent Test**: Can be tested by initiating password reset flow from login page, receiving reset email (or viewing reset link in logs for stub email), completing password change, and verifying login with new password. Delivers complete self-service account security independently.

**Acceptance Scenarios**:

1. **Given** a user on the login page, **When** they click "Forgot Password?", **Then** they navigate to a password reset request page with an email input field and "Send Reset Link" button
2. **Given** a user enters their registered email, **When** they click "Send Reset Link", **Then** the system generates a password reset token valid for 1 hour, sends an email with reset link (or shows link in console for stub email in MVP), and displays "Password reset link sent to [email]"
3. **Given** a user clicks a valid password reset link, **When** they arrive at the reset password page, **Then** they see a form with "New Password", "Confirm Password" fields, and a "Reset Password" button
4. **Given** a user enters a new valid password, **When** they click "Reset Password", **Then** the password is securely updated, all active sessions on other devices are revoked by default, and the user is redirected to login with a success message "Password reset successfully. Please log in."
5. **Given** an authenticated user in account settings, **When** they initiate a password change, **Then** they see a form with "Current Password", "New Password", "Confirm Password" fields, and a checkbox "Log out all other devices"
6. **Given** a user submits a password change with "Log out all other devices" checked, **When** the change succeeds, **Then** all refresh tokens for that user are revoked, the user remains logged in on the current device, and sees a confirmation "Password changed successfully. All other sessions logged out."

---

### User Story 5 - UI Component Library & Design System Basics (Priority: P3)

Establish a minimal, consistent visual identity and component library to ensure accessible, on-brand user interfaces across all features.

**Why this priority**: Design consistency is important for polish and professionalism, but the product can function with basic unstyled forms in early iterations. This work enables better UX but is not blocking for MVP functionality.

**Independent Test**: Can be tested by navigating through all implemented features and verifying visual consistency (buttons, forms, colors, typography, spacing) and accessibility compliance (keyboard navigation, focus indicators, screen reader labels). Delivers polish layer independently.

**Acceptance Scenarios**:

1. **Given** any page in the application, **When** a developer inspects styles, **Then** all visual design values (spacing, colors, typography) reference the design system tokens (no hard-coded values)
2. **Given** a user navigates any form, **When** they tab through form fields, **Then** they see visible focus indicators meeting accessibility standards for color contrast and visibility
3. **Given** a user encounters an error (validation, server error, network failure), **When** the error is displayed, **Then** they see a consistent error message format with an icon, clear message text, and actionable next steps (never raw API error messages or stack traces)
4. **Given** a user initiates an async action (login, trip generation, suggestion submission), **When** the action is processing, **Then** they see a loading state (spinner + message like "Generating your itinerary...") and the action button becomes disabled to prevent duplicate submissions
5. **Given** a user views a list with no data (no trips, no suggestions, no collaborators), **When** the list renders, **Then** they see an empty state illustration/icon with a message like "No trips yet" and a prominent call-to-action button like "Create Your First Trip"
6. **Given** any interactive element (button, link, touch target), **When** viewed on mobile devices, **Then** touch targets are at least 44×44 pixels with adequate spacing to prevent accidental taps

---

### Edge Cases

- What happens when a user tries to register with an email that already exists? (Return 400 Bad Request with "Email already registered" message without revealing whether email exists for security)
- What happens when a Free User already collaborating on one trip receives a second invitation? (Display "You are already collaborating on [Trip Name]. Leave that trip first to accept this invitation." with "Leave Current Trip" button; new invitation cannot be accepted until current collaboration is terminated)
- What happens when a Free User tries to create a trip without upgrading to Paid User? (Show upgrade prompt "Upgrade to create your own trips" with "Subscribe Now" button; redirect to payment flow; trip creation is blocked until subscription is active)
- What happens when a user exceeds the failed login attempt threshold? (After 5 failed login attempts within 15 minutes, subsequent attempts are delayed with exponentially increasing wait times: 1s, 2s, 4s, 8s... Counter resets on successful login)
- What happens when an admin tries to invite more than 1 collaborator on a basic plan? (Return 400 Bad Request with "Basic plan allows 1 collaborator maximum. Upgrade plan to invite more.")
- What happens when a Free User tries to directly edit a trip item without using the suggestion flow? (Return 403 Forbidden with "You don't have permission to edit this trip. Submit a suggestion instead.")
- What happens when a Paid User's subscription is cancelled or expires? (Enter 30-day grace period: set User.has_subscription=false, owned trips remain visible but become read-only, user can view trips and collaborate as Free User on other users' trips, but cannot create new trips or edit owned trips; display banner "Your subscription has expired. Renew to continue editing your trips.")
- What happens when the 30-day grace period expires after subscription cancellation? (System archives (hides) all owned trips - trips are not deleted but invisible in dashboard; user retains Free User capabilities for collaboration; can resubscribe anytime to restore full access and un-archive owned trips)
- What happens when a user in grace period tries to edit their own trip? (Return 403 Forbidden with "Your subscription has expired. Renew your subscription to edit this trip." and display "Renew Subscription" button)
- What happens when a user in grace period or with archived trips resubscribes? (Set User.has_subscription=true, immediately un-archive all previously owned trips, restore full Paid User capabilities including create/edit/delete trips)
- What happens when a Free User who is currently collaborating upgrades to Paid User? (Preserve existing collaboration - set User.has_subscription=true, Collaborator record remains unchanged, user badge updates to "Paid User", user can now create own trips while continuing collaboration on original trip, one-collaboration limit no longer applies)
- What happens when a Paid User (formerly Free User who upgraded) is invited to collaborate on a second trip? (Invitation accepted normally - Paid Users have no collaboration limit, can be invited to multiple trips simultaneously, collaborate as "Paid User" role with view + suggest permissions on trips they don't own)
- What happens when a user clicks a password reset link that has expired (>1 hour old)? (Display "Password reset link has expired. Request a new one." with a link back to password reset request page)
- What happens when a user tries to reuse a password reset link that has already been used? (Display "This password reset link has already been used. Request a new one if needed." with a link back to password reset request page)
- What happens when two admins try to approve the same suggestion simultaneously? (Optimistic locking detects conflict, second approver sees 409 Conflict with "This suggestion has already been processed" message)
- What happens when a user's session expires during active use? (System detects expiration, automatically renews the session seamlessly without forcing re-login)
- What happens when automatic session renewal fails? (Clear expired session, redirect to login page with "Your session has expired. Please log in again." message, preserve intended destination via ?redirect parameter for post-login restoration)
- What happens when a user tries to access a trip they don't own or weren't invited to? (Return 404 Not Found instead of 403 Forbidden to prevent trip ID enumeration attacks)
- What happens when the AI service is unavailable during trip generation? (Display user-friendly error "We're having trouble generating your itinerary. Please try again in a moment." with retry button, log 503 Service Unavailable in backend)
- What happens when the database is unavailable during login or registration? (Return 503 Service Unavailable with Retry-After header (30 seconds) and message "Service temporarily unavailable. Please try again in a moment."; frontend displays error with retry button respecting Retry-After timing)
- What happens when a user navigates to a protected page while not authenticated? (Redirect to login page with `?redirect=/original-path` query parameter, after successful login redirect back to intended destination)
- What happens when a user submits a form with client-side validation disabled (via browser dev tools)? (Backend validation catches invalid data, returns 400 Bad Request with detailed field errors, frontend displays server-side validation errors)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a public landing page accessible to unauthenticated users with "Sign Up" and "Log In" actions clearly visible
- **FR-002**: System MUST provide a registration flow with two paths: (1) Paid User path - collects email, password, validates format, redirects to stub payment checkout, creates User with has_subscription=true after payment; (2) Free User path - collects email, password, validates format, creates User with has_subscription=false, no payment required
- **FR-002a**: System MUST enforce Free User collaboration limit - Free Users (has_subscription=false) can only be invited to collaborate on ONE trip at a time; system checks active Collaborator count before allowing invitation acceptance; displays "Leave current trip first" message if limit reached
- **FR-003**: System MUST integrate a visible stub payment checkout page labeled "[DEMO] Subscription Checkout" that always succeeds, creates a Subscription record with status='stub_pending'→'active', sets User.has_subscription=true, and transitions user to authenticated session as Paid User
- **FR-004**: System MUST provide a login form that accepts email/password, securely validates credentials, establishes an authenticated session valid for 24 hours with automatic renewal capability for 30 days, and redirects to trip dashboard on success
- **FR-005**: System MUST return generic error messages for failed login attempts ("Invalid email or password") without revealing which field is incorrect to prevent account enumeration
- **FR-005a**: System MUST implement progressive delay rate limiting on login attempts - track failed attempts per email address; after 5 failed attempts within 15 minutes, impose exponentially increasing delays (1 second, 2 seconds, 4 seconds, 8 seconds, etc.) before accepting the next login attempt; reset counter on successful login; no permanent account lockout
- **FR-006**: System MUST provide a trip dashboard page showing sections based on user type: Paid Users see "My Trips" (trips they own) and "Shared with Me" (trips they're collaborating on); Free Users see only "Shared with Me" section
- **FR-006a**: System MUST display "Upgrade to create your own trips" prompt with "Subscribe Now" button when Free Users attempt to create a trip; block trip creation until subscription is purchased and User.has_subscription=true
- **FR-007**: System MUST allow Paid Users to invite a Free User or another Paid User to collaborate by email, enforce basic plan limit (1 collaborator maximum), create a Collaborator record with status='pending', and send an email notification (stub email logs invitation link to console in MVP)
- **FR-008**: System MUST display invited trips under "Shared with Me" section for collaborating users with visual badges indicating user type ("Paid User" or "Free User") and permissions ("View Only + Suggest")
- **FR-008a**: System MUST provide "Leave Trip" action for Free Users on shared trips; when confirmed, delete the Collaborator record, remove trip from "Shared with Me", and enable accepting new invitations
- **FR-009**: System MUST replace direct "Edit" actions with "Suggest Change" actions for collaborating users (both Paid and Free) viewing shared trips they don't own, ensuring collaborators cannot bypass suggestion workflow
- **FR-010**: System MUST allow collaborating users to submit suggestions that create Suggestion records with status='pending', description text, and reference to the specific itinerary item being modified
- **FR-011**: System MUST allow Paid Users (trip owners) to approve or reject pending suggestions, apply approved changes to the itinerary atomically (with optimistic locking), update suggestion status to 'approved' or 'rejected', and notify the collaborating user of the decision
- **FR-012**: System MUST provide a "Forgot Password?" flow that generates a single-use password reset token valid for 1 hour, sends an email with reset link (stub email logs link in MVP), validates token has not expired and has not been used (used_at IS NULL), marks token as used on first successful redemption (set used_at timestamp), and allows user to set a new password without knowing the old one
- **FR-013**: System MUST provide an authenticated password change flow in account settings that requires current password verification, validates new password format, securely stores the new password, and provides an optional "Log out all other devices" checkbox
- **FR-014**: System MUST revoke all refresh tokens for a user when "Log out all other devices" is checked during password change, except the current session token
- **FR-015**: System MUST detect expired session tokens and automatically attempt to renew them to maintain user authentication; if renewal fails (expired refresh token, network error, server unavailable), clear the expired session, redirect user to login page with message "Your session has expired. Please log in again.", and preserve the intended destination URL via ?redirect parameter for post-login restoration
- **FR-016**: System MUST redirect unauthenticated users attempting to access protected pages to the login page with a `?redirect` query parameter, and restore intended destination after successful login
- **FR-017**: System MUST implement server-side validation for all user inputs (email format, password complexity, trip title length, suggestion text length) and return 400 Bad Request with structured field-level errors on validation failure
- **FR-018**: System MUST return 404 Not Found (instead of 403 Forbidden) when a user attempts to access a trip they don't own or weren't invited to, preventing trip ID enumeration attacks
- **FR-019**: System MUST display user-friendly error messages for all failure scenarios (network errors, server errors, AI service unavailable) with actionable next steps, never exposing raw API errors or stack traces to end users
- **FR-019a**: System MUST return 503 Service Unavailable with Retry-After header (30 seconds) when database is unavailable during authentication operations (login, registration, password reset), displaying message "Service temporarily unavailable. Please try again in a moment."; frontend should respect Retry-After header and provide retry button
- **FR-020**: System MUST log all authentication and authorization events as structured security events including: event type (login success/failure, registration, password reset request/completion, password change, token refresh, logout), correlation ID for request tracing, user identifier (user ID for successes, email for failures), timestamp, IP address, user agent, and event outcome; logs must exclude passwords, tokens, and PII beyond user identifier
- **FR-021**: System MUST establish a minimal design system with reusable design tokens defining: primary/secondary/error/warning colors (meeting accessibility standards for contrast), spacing scale, typography scale (font sizes, weights, line heights), and focus indicator styles
- **FR-022**: System MUST implement subscription lapse handling with 30-day grace period - when Subscription.status changes to 'cancelled' or expires, set User.has_subscription=false, record grace_period_ends_at = now + 30 days, owned trips remain visible but read-only (cannot edit/delete), user can continue collaborating as Free User on other trips, display banner "Your subscription has expired. Renew to continue editing your trips."
- **FR-023**: System MUST archive owned trips after grace period expires - when current timestamp > Subscription.grace_period_ends_at, set Trip.archived=true for all trips owned by user, trips become invisible in dashboard but are not deleted, user retains Free User collaboration capabilities
- **FR-024**: System MUST restore full access on resubscription - when user resubscribes (completes payment, Subscription.status='active'), set User.has_subscription=true, clear grace_period_ends_at, set Trip.archived=false for all owned trips, immediately restore full Paid User capabilities
- **FR-025**: System MUST preserve collaborations on Free User upgrade - when Free User completes subscription purchase (Subscription.status='stub_pending'→'active'), set User.has_subscription=true, leave existing Collaborator records unchanged, update user badge from "Free User" to "Paid User", remove one-collaboration limit, enable trip creation capability while maintaining collaboration access to original trip

### Key Entities *(include if feature involves data)*

**Note**: Core entities (User, Session, Subscription, Plan, Trip, Collaborator, Suggestion) are already defined in docs/data-model.md. This spec adds UX-specific entities and modifies Subscription and Trip for grace period handling.

- **PasswordResetToken**: Temporary token for password reset flow; includes secure token identifier, user reference, expiration timestamp (1 hour from creation), usage timestamp (null until first redemption; set to current timestamp on successful password reset to prevent reuse), and creation timestamp; validation requires both expires_at > now AND used_at IS NULL
- **UINotification**: Transient notification for user feedback (success/error/info/warning); includes user reference, notification type, message text, read status, and creation timestamp
- **Subscription.grace_period_ends_at** (new field): Timestamp marking end of 30-day grace period after subscription cancellation; null when subscription is active; set to now + 30 days when status changes to 'cancelled'; used to determine when to archive owned trips
- **Trip.archived** (new field): Boolean flag indicating trip is archived and hidden from dashboard; false by default; set to true when grace period expires (now > Subscription.grace_period_ends_at); set to false on resubscription; archived trips are not deleted and can be restored

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new user can complete registration, stub payment, and first trip creation in under 5 minutes from landing page to viewing generated itinerary
- **SC-002**: Returning users can log in and reach their trip dashboard in under 10 seconds (measured from click "Log In" to dashboard fully rendered)
- **SC-003**: 95% of authentication attempts (login, registration) complete successfully on first try with valid credentials, indicating clear form validation and error messaging
- **SC-004**: 100% of form fields and interactive elements are keyboard-accessible with visible focus indicators meeting accessibility standards (sufficient contrast for users with visual impairments)
- **SC-005**: Zero raw API error messages, stack traces, or technical jargon exposed to end users in any error scenario
- **SC-005a**: 100% of authentication and authorization events are logged with structured security event records containing correlation ID, user identifier, timestamp, IP address, and outcome for security monitoring and incident investigation
- **SC-006**: Paid Users can invite a Free User, Free User can submit a suggestion, and Paid User can approve/reject within 3 minutes, demonstrating efficient collaboration workflow
- **SC-006a**: Free Users attempting to accept a second invitation while already collaborating see "Leave current trip first" message and can complete the leave action within 30 seconds
- **SC-007**: Password reset flow completes successfully in under 2 minutes from "Forgot Password?" to setting new password and logging in
- **SC-008**: All critical user actions (login, create trip, submit suggestion, approve suggestion) provide immediate feedback (loading state) within 100ms of user interaction, preventing duplicate submissions
- **SC-009**: Mobile users can complete all primary workflows (register, login, create trip, invite collaborator) with appropriately sized touch targets (minimum 44×44 pixels) and no horizontal scrolling required
- **SC-010**: Design system tokens (colors, spacing, typography) are used consistently across 100% of implemented UI components, ensuring visual consistency throughout the application
- **SC-011**: Users whose subscriptions expire can view their owned trips in read-only mode during the 30-day grace period and can resubscribe to restore full edit access within 2 minutes from "Renew Subscription" button to payment completion
- **SC-012**: Free Users who upgrade to Paid User while collaborating on a trip experience zero disruption - existing collaboration continues uninterrupted, trip remains visible under "Shared with Me", and trip creation capability is enabled within 5 seconds of payment completion

## Assumptions

- Users have stable internet connectivity sufficient for web application use (no offline mode requirements in MVP)
- Email delivery for password reset and collaboration invites is mocked in MVP (emails logged to console; real SMTP/SendGrid integration deferred to post-MVP)
- The application targets modern browsers (Chrome, Firefox, Safari, Edge - last 2 versions); no IE11 support required
- Mobile support is responsive web design (no native mobile apps); touch targets and viewport sizing follow mobile-first design principles
- User accounts are created through self-service registration only; no admin-created accounts or bulk import in MVP
- All UI text and error messages are in English; internationalization (i18n) infrastructure is out of scope per docs/product-vision.md
- The design system uses centralized design tokens for consistent styling across the application
- Accessibility compliance targets industry-standard Level AA conformance; enhanced Level AAA features (audio descriptions, sign language) are not required in MVP
- All forms include client-side validation for immediate feedback, but server-side validation is the authoritative source (no client-side bypass trust)
- The stub payment provider always succeeds and never fails; payment failure testing is deferred until real payment provider integration
- Collaboration invites are sent by email only; no in-app notification system (real-time WebSocket notifications) in MVP
- Users can only belong to one subscription at a time (one-to-one User-Subscription relationship)
- Free Users (no subscription) must have an existing TrAIveler account to receive collaboration invites; no invite-before-registration flow in MVP
- Free Users can only collaborate on ONE trip at a time; this limit encourages upgrading to Paid User for multi-trip collaboration or trip ownership
- When a Free User upgrades to Paid User, the one-collaboration limit is removed; Paid Users can collaborate on unlimited trips simultaneously
- Paid Users who are invited to collaborate on other users' trips behave as Free Users on those trips (view + suggest only; no direct edit access)
- Subscription cancellations in MVP are processed immediately (no prorated refunds, no mid-cycle cancellations); grace period begins immediately upon cancellation
- Grace period duration is fixed at 30 days in MVP; no custom grace periods or extensions
- Archived trips are preserved indefinitely; no automatic deletion after archival (data retention for potential reactivation)
- System-level Administrator role (control panel for viewing all users/trips, troubleshooting) is explicitly out of scope for MVP; internal staff troubleshooting will use database runbook with direct SQL queries for demo data corrections
- Administrator role UI, authentication, and RBAC implementation deferred to post-MVP; MVP authentication covers only end-user flows (Paid User and Free User)

## Clarifications

### Session 2026-07-06

- Q: How should the system handle repeated failed authentication attempts? → A: Progressive delay with account-level tracking - track failed attempts per email address; after 5 failed login attempts within 15 minutes, impose exponentially increasing delays (1s, 2s, 4s, 8s...) before accepting next attempt; reset counter on successful login; no permanent lockout
- Q: What authentication and authorization events should be logged, and what information should each log entry contain? → A: Structured security events with correlation IDs - log all authentication events (login success/failure, registration, password reset request/completion, password change, token refresh, logout) with correlation ID, user ID (or email for failures), timestamp, IP address, user agent, and event outcome; no passwords, tokens, or PII beyond user identifier
- Q: How should the system enforce that a password reset token can only be used once? → A: Mark token as used with timestamp - add used_at timestamp field to PasswordResetToken; set to current timestamp on first redemption; validation checks both expiration (1 hour) and used_at (must be null); preserves audit trail while preventing reuse
- Q: What should happen when automatic session renewal fails (expired refresh token, network error, server unavailable)? → A: Silent logout with redirect to login page - clear the expired session, redirect user to login page with message "Your session has expired. Please log in again." and preserve the intended destination URL via ?redirect parameter
- Q: What should happen when the database is unavailable during authentication attempts? → A: Return 503 Service Unavailable with retry guidance - return 503 status with message "Service temporarily unavailable. Please try again in a moment." and Retry-After header (30 seconds); clear status indication, encourages retry, aligns with HTTP semantics for transient failures
- Q: Is the system-level Administrator role (with control panel access to view users/trips and troubleshoot) in scope for MVP? → A: Out of MVP scope - Administrator is a system-level role for internal staff with platform-wide access; deferred to post-MVP; MVP focuses only on Paid User (trip owner) and Free User (collaborator) authentication and workflows; database runbook will be provided for demo data corrections and troubleshooting via direct database queries
- Q: Can Free Users collaborate without a paid subscription, and what are their limits? → A: Free tier for collaborators only - Free Users can register without payment, receive invitations, and collaborate on trips (view + suggest) BUT can only be invited to ONE trip at a time; must leave existing collaboration before accepting a new invitation; cannot create their own trips until they purchase a subscription and become Paid Users; Paid Users = trip creators with active subscription
- Q: When a Paid User's subscription expires or is cancelled, what happens to their account, owned trips, and system role? → A: Grace period with read-only access - after subscription cancellation, provide 30-day grace period where user can view owned trips (read-only), collaborate as Free User on other trips, but cannot create new trips or edit existing ones; after grace period, owned trips archived (hidden) but not deleted; user can resubscribe anytime to restore full access; during grace period User.has_subscription=false but trips remain visible
- Q: When a Free User who is currently collaborating on a trip upgrades to Paid User, what happens to their existing collaboration? → A: Preserve existing collaboration - Free User upgrades to Paid User (has_subscription=true), existing Collaborator record and trip access remain unchanged, user badge updates to "Paid User", collaboration continues uninterrupted, user can now create their own trips AND remains collaborator on the original trip; no one-collaboration limit after upgrade
