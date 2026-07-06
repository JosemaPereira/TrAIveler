# Quickstart: Authentication & Collaboration UX Validation

**Feature**: Authentication & Collaboration UX  
**Date**: 2026-07-06  
**Phase**: 1 (Design)

## Overview

This document provides runnable validation scenarios to prove the authentication, session management, subscription lifecycle, and collaboration workflows work end-to-end. Each scenario maps to a user story from spec.md.

**Prerequisites**:
- Backend server running on `http://localhost:8080`
- Frontend app running on `http://localhost:3000`
- PostgreSQL database initialized with migrations
- Email service configured (stub implementation for dev)
- Payment provider configured (stub implementation for dev)

---

## Scenario 1: Paid User Registration → First Trip Creation

**User Story**: US1 - Paid User Onboarding (Priority: P1)

**Goal**: Register as Paid User with payment method, verify subscription activation, create first trip.

### Steps

1. **Navigate to registration page**:
   ```bash
   open http://localhost:3000/register
   ```

2. **Fill registration form**:
   - Email: `paid-user@example.com`
   - Password: `SecureP@ss123`
   - Full Name: `Jane Doe`
   - Payment Method: Select credit card `4242 4242 4242 4242` (Stripe test card)
   - Agree to terms

3. **Submit form** → Backend calls `POST /api/v1/auth/register` with payment_method_token

4. **Expected Result**:
   - ✅ User account created with `has_subscription=true`
   - ✅ Subscription record created with `status='active'`
   - ✅ Redirected to `/dashboard` with welcome message
   - ✅ Dashboard shows "Create Trip" button (not disabled)
   - ✅ HTTP-only cookies set: `access_token`, `refresh_token`

5. **Create first trip**:
   - Click "Create Trip" button
   - Enter trip details: Destination "Paris", Dates "2026-09-01 to 2026-09-07"
   - Submit → Backend calls `POST /api/v1/trips`

6. **Expected Result**:
   - ✅ Trip created with `creator_id = user.id`
   - ✅ Redirected to trip detail page `/trips/:id`
   - ✅ Trip details displayed (Paris, 7 days)

### Validation Commands

```bash
# Check user registration
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "paid-user@example.com",
    "password": "SecureP@ss123",
    "full_name": "Jane Doe",
    "payment_method_token": "tok_visa_test_123"
  }'

# Expected: 201 Created, response contains user.id and subscription.id

# Check trip creation
curl -X POST http://localhost:8080/api/v1/trips \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<JWT_FROM_REGISTRATION>" \
  -d '{
    "destination": "Paris",
    "start_date": "2026-09-01",
    "end_date": "2026-09-07"
  }'

# Expected: 201 Created, response contains trip.id with creator_id = user.id
```

### Database Validation

```sql
-- Check user and subscription
SELECT u.email, u.has_subscription, s.status, s.current_period_end
FROM users u
LEFT JOIN subscriptions s ON s.user_id = u.id
WHERE u.email = 'paid-user@example.com';

-- Expected: has_subscription=true, status='active', current_period_end ~30 days from now

-- Check trip
SELECT t.id, t.destination, t.creator_id, t.archived
FROM trips t
JOIN users u ON t.creator_id = u.id
WHERE u.email = 'paid-user@example.com';

-- Expected: destination='Paris', archived=false
```

---

## Scenario 2: Free User Registration → Accept Collaboration Invitation

**User Story**: US2 - Free User Onboarding (Priority: P1)

**Goal**: Register as Free User (no payment), receive collaboration invitation, accept invitation, verify single-collaboration limit.

### Steps

1. **Paid User invites Free User** (prerequisite: run Scenario 1 first):
   - Log in as `paid-user@example.com`
   - Navigate to trip detail page
   - Click "Invite Collaborator"
   - Enter email: `free-user@example.com`
   - Submit → Backend calls `POST /api/v1/trips/:id/collaborators`

2. **Free User registers**:
   ```bash
   open http://localhost:3000/register
   ```
   - Email: `free-user@example.com`
   - Password: `FreeUserP@ss456`
   - Full Name: `John Smith`
   - **Do not** provide payment method
   - Submit → Backend calls `POST /api/v1/auth/register` without payment_method_token

3. **Expected Result**:
   - ✅ User account created with `has_subscription=false`
   - ✅ Redirected to `/dashboard`
   - ✅ Dashboard shows banner: "You have 1 pending invitation"
   - ✅ "Create Trip" button disabled with tooltip "Subscribe to create your own trips"

4. **Accept invitation**:
   - Click "View Invitations" → Navigate to `/invitations`
   - See invitation from `paid-user@example.com` for "Paris" trip
   - Click "Accept" → Backend calls `PATCH /api/v1/collaborators/:id/accept`

5. **Expected Result**:
   - ✅ Collaborator status changed to `accepted`
   - ✅ Redirected to trip detail page `/trips/:id`
   - ✅ Trip displayed as read-only for Free User (can view, suggest, cannot edit)
   - ✅ Badge shows "Free User" next to name

6. **Verify single-collaboration limit**:
   - Paid User invites Free User to **second trip** (different destination)
   - Free User receives second invitation
   - Free User attempts to accept second invitation

7. **Expected Result**:
   - ✅ Error message: "You are already collaborating on Paris. Leave that trip first to accept this invitation."
   - ✅ "Leave Current Trip" button displayed
   - ✅ Second invitation remains in "Pending" status

### Validation Commands

```bash
# Check Free User registration
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "free-user@example.com",
    "password": "FreeUserP@ss456",
    "full_name": "John Smith"
  }'

# Expected: 201 Created, user.has_subscription=false, no subscription object in response

# Check collaboration invitation
curl -X POST http://localhost:8080/api/v1/trips/<TRIP_ID>/collaborators \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<PAID_USER_JWT>" \
  -d '{"email": "free-user@example.com"}'

# Expected: 201 Created, collaborator.status='pending'

# Accept invitation
curl -X PATCH http://localhost:8080/api/v1/collaborators/<COLLABORATOR_ID>/accept \
  -H "Cookie: access_token=<FREE_USER_JWT>"

# Expected: 200 OK, collaborator.status='accepted'

# Attempt second collaboration (should fail)
curl -X PATCH http://localhost:8080/api/v1/collaborators/<SECOND_COLLABORATOR_ID>/accept \
  -H "Cookie: access_token=<FREE_USER_JWT>"

# Expected: 400 Bad Request, error="COLLABORATION_LIMIT_EXCEEDED"
```

### Database Validation

```sql
-- Check Free User
SELECT email, has_subscription FROM users WHERE email = 'free-user@example.com';
-- Expected: has_subscription=false

-- Check collaboration count
SELECT COUNT(*) FROM collaborators
WHERE user_id = (SELECT id FROM users WHERE email = 'free-user@example.com')
AND status = 'accepted';
-- Expected: 1 (single collaboration)
```

---

## Scenario 3: Login → Trip Dashboard Access

**User Story**: US3 - Login and Trip Management (Priority: P1)

**Goal**: Existing user logs in, accesses trip dashboard, verifies session persistence.

### Steps

1. **Navigate to login page**:
   ```bash
   open http://localhost:3000/login
   ```

2. **Fill login form**:
   - Email: `paid-user@example.com`
   - Password: `SecureP@ss123`
   - Submit → Backend calls `POST /api/v1/auth/login`

3. **Expected Result**:
   - ✅ Redirected to `/dashboard`
   - ✅ HTTP-only cookies set: `access_token` (24h), `refresh_token` (30d)
   - ✅ Dashboard shows user trips (Paris trip from Scenario 1)
   - ✅ Navigation shows user name "Jane Doe" and "Paid User" badge

4. **Verify session persistence**:
   - Refresh page (F5) → No redirect to login
   - Close browser tab, reopen `http://localhost:3000/dashboard` → Still authenticated (cookies persist)

5. **Verify failed login rate limiting**:
   - Log out
   - Attempt login with incorrect password 6 times within 1 minute
   - On 6th attempt, observe progressive delay message

6. **Expected Result**:
   - ✅ First 5 failures: Immediate "Invalid credentials" error
   - ✅ 6th failure: 429 Too Many Requests, "Too many failed login attempts. Try again in 1 second."
   - ✅ 7th failure (after delay): "Try again in 2 seconds."
   - ✅ 8th failure: "Try again in 4 seconds." (exponential backoff)

### Validation Commands

```bash
# Successful login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "email": "paid-user@example.com",
    "password": "SecureP@ss123"
  }'

# Expected: 200 OK, Set-Cookie headers with access_token and refresh_token

# Access protected endpoint
curl -X GET http://localhost:8080/api/v1/trips \
  -b cookies.txt

# Expected: 200 OK, list of user's trips

# Failed login rate limiting
for i in {1..8}; do
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
      "email": "paid-user@example.com",
      "password": "WrongPassword"
    }' \
    -w "\n%{http_code}\n"
  sleep 0.5
done

# Expected: First 5 attempts return 401, 6th+ return 429 with increasing retry_after
```

---

## Scenario 4: Password Reset → Password Change

**User Story**: US4 - Password Management (Priority: P2)

**Goal**: User forgets password, requests reset, completes reset via email link, then changes password while authenticated.

### Steps (Part A: Password Reset)

1. **Request password reset**:
   ```bash
   open http://localhost:3000/password-reset
   ```
   - Enter email: `paid-user@example.com`
   - Submit → Backend calls `POST /api/v1/auth/password-reset-request`

2. **Expected Result**:
   - ✅ Message displayed: "If an account with that email exists, a password reset link has been sent."
   - ✅ Email sent to `paid-user@example.com` with reset link (check logs or email stub)
   - ✅ PasswordResetToken record created in database

3. **Open reset link from email**:
   - Extract token from email: `http://localhost:3000/password-reset/complete?token=<TOKEN>`
   - Open link in browser

4. **Complete password reset**:
   - Enter new password: `NewSecureP@ss789`
   - Confirm new password: `NewSecureP@ss789`
   - Submit → Backend calls `POST /api/v1/auth/password-reset-complete`

5. **Expected Result**:
   - ✅ Success message: "Password reset successfully. You can now log in."
   - ✅ Redirected to `/login`
   - ✅ PasswordResetToken marked as used (`used_at` set)
   - ✅ All refresh tokens revoked (user logged out on all devices)

6. **Verify token single-use**:
   - Attempt to reuse same reset link
   - Expected: Error "This password reset link has already been used."

### Steps (Part B: Password Change)

7. **Log in with new password**:
   - Email: `paid-user@example.com`
   - Password: `NewSecureP@ss789`
   - Submit → Backend calls `POST /api/v1/auth/login`

8. **Navigate to account settings**:
   ```bash
   open http://localhost:3000/settings/account
   ```
   - Click "Change Password"

9. **Change password**:
   - Current Password: `NewSecureP@ss789`
   - New Password: `AnotherP@ss000`
   - Confirm New Password: `AnotherP@ss000`
   - Submit → Backend calls `POST /api/v1/auth/password-change`

10. **Expected Result**:
    - ✅ Success message: "Password changed successfully."
    - ✅ All refresh tokens except current session revoked (logged out on other devices)
    - ✅ Current session remains active (no redirect)

### Validation Commands

```bash
# Request password reset
curl -X POST http://localhost:8080/api/v1/auth/password-reset-request \
  -H "Content-Type: application/json" \
  -d '{"email": "paid-user@example.com"}'

# Expected: 200 OK, message="If an account with that email exists..."

# Complete password reset (extract token from email/logs)
curl -X POST http://localhost:8080/api/v1/auth/password-reset-complete \
  -H "Content-Type: application/json" \
  -d '{
    "token": "<TOKEN_FROM_EMAIL>",
    "new_password": "NewSecureP@ss789"
  }'

# Expected: 200 OK, message="Password reset successfully"

# Reuse token (should fail)
curl -X POST http://localhost:8080/api/v1/auth/password-reset-complete \
  -H "Content-Type: application/json" \
  -d '{
    "token": "<SAME_TOKEN>",
    "new_password": "AnotherPassword"
  }'

# Expected: 400 Bad Request, error="TOKEN_ALREADY_USED"

# Change password (authenticated)
curl -X POST http://localhost:8080/api/v1/auth/password-change \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<JWT>" \
  -d '{
    "current_password": "NewSecureP@ss789",
    "new_password": "AnotherP@ss000"
  }'

# Expected: 200 OK, message="Password changed successfully"
```

### Database Validation

```sql
-- Check password reset token
SELECT user_id, expires_at, used_at, created_at
FROM password_reset_tokens
WHERE user_id = (SELECT id FROM users WHERE email = 'paid-user@example.com')
ORDER BY created_at DESC LIMIT 1;

-- Expected: used_at IS NOT NULL (token marked as used)

-- Check security events
SELECT event_type, severity, created_at
FROM security_events
WHERE user_id = (SELECT id FROM users WHERE email = 'paid-user@example.com')
ORDER BY created_at DESC LIMIT 5;

-- Expected: auth_password_reset_request, auth_password_reset_complete, auth_password_change events
```

---

## Scenario 5: Collaboration → Suggestion → Approval

**User Story**: US5 - Collaboration Workflow (Priority: P2)

**Goal**: Free User collaborates on trip, creates suggestion, Paid User (creator) approves suggestion, verify trip updated.

### Steps

1. **Free User creates suggestion** (prerequisite: Free User accepted invitation in Scenario 2):
   - Log in as `free-user@example.com`
   - Navigate to trip detail page `/trips/:id` (Paris trip)
   - Click "Suggest Change" button
   - Select type: "Activity"
   - Enter content: "Visit the Louvre Museum"
   - Add details: Day 2, Duration 3 hours
   - Submit → Backend calls `POST /api/v1/trips/:id/suggestions`

2. **Expected Result**:
   - ✅ Suggestion created with `status='pending'`
   - ✅ Success message: "Suggestion submitted. Waiting for creator approval."
   - ✅ Suggestion appears in "Pending Suggestions" section

3. **Paid User reviews suggestion**:
   - Log in as `paid-user@example.com` (trip creator)
   - Navigate to trip detail page `/trips/:id`
   - See notification badge: "1 pending suggestion"
   - Click "Review Suggestions" → Navigate to suggestions list

4. **Paid User approves suggestion**:
   - See suggestion from `free-user@example.com`: "Visit the Louvre Museum"
   - Click "Approve" → Backend calls `PATCH /api/v1/suggestions/:id/approve`

5. **Expected Result**:
   - ✅ Suggestion status changed to `approved`
   - ✅ Success message: "Suggestion approved. The trip has been updated."
   - ✅ New activity "Visit the Louvre Museum" added to Day 2 of trip
   - ✅ Free User sees notification: "Your suggestion was approved!"

6. **Verify rejection flow**:
   - Free User creates another suggestion: "Stay at Hotel XYZ"
   - Paid User rejects suggestion → Backend calls `PATCH /api/v1/suggestions/:id/reject`
   - Expected: Suggestion status `rejected`, no trip changes, Free User notified

### Validation Commands

```bash
# Create suggestion (Free User)
curl -X POST http://localhost:8080/api/v1/trips/<TRIP_ID>/suggestions \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<FREE_USER_JWT>" \
  -d '{
    "suggestion_type": "activity",
    "content": "Visit the Louvre Museum",
    "details": {"day_number": 2, "estimated_duration": "3 hours"}
  }'

# Expected: 201 Created, suggestion.status='pending'

# Approve suggestion (Paid User)
curl -X PATCH http://localhost:8080/api/v1/suggestions/<SUGGESTION_ID>/approve \
  -H "Cookie: access_token=<PAID_USER_JWT>"

# Expected: 200 OK, suggestion.status='approved', message="Suggestion approved..."

# Verify trip updated
curl -X GET http://localhost:8080/api/v1/trips/<TRIP_ID> \
  -H "Cookie: access_token=<PAID_USER_JWT>"

# Expected: Response includes new activity "Visit the Louvre Museum" in Day 2
```

### Database Validation

```sql
-- Check suggestion
SELECT id, collaborator_id, suggestion_type, content, status, approved_at
FROM suggestions
WHERE trip_id = '<TRIP_ID>'
AND content = 'Visit the Louvre Museum';

-- Expected: status='approved', approved_at IS NOT NULL

-- Check activity created from suggestion
SELECT id, trip_id, day_number, name, duration
FROM activities
WHERE trip_id = '<TRIP_ID>'
AND name LIKE '%Louvre%';

-- Expected: day_number=2, duration='3 hours', activity exists
```

---

## Scenario 6: Free User Upgrade → Collaboration Preserved

**User Story**: US2 (extension) - Free User Upgrade (Priority: P2)

**Goal**: Free User upgrades to Paid User mid-collaboration, verify collaboration continues and trip creation enabled.

### Steps

1. **Free User upgrades** (prerequisite: Free User collaborating on Paris trip from Scenario 2):
   - Log in as `free-user@example.com`
   - Navigate to `/dashboard`
   - See banner: "Upgrade to create your own trips and collaborate on unlimited trips"
   - Click "Upgrade Now" button → Navigate to `/upgrade`

2. **Complete upgrade**:
   - Select plan: "Monthly Subscription - $9.99/month"
   - Enter payment method: `4242 4242 4242 4242` (Stripe test card)
   - Submit → Backend calls `POST /api/v1/subscriptions`

3. **Expected Result**:
   - ✅ Subscription created with `status='active'`
   - ✅ `User.has_subscription` set to `true`
   - ✅ Redirected to `/dashboard`
   - ✅ Success message: "Subscription activated! You can now create trips."
   - ✅ Dashboard shows "Create Trip" button (now enabled)
   - ✅ Paris trip (collaboration) still visible in "Collaborating" section

4. **Verify collaboration preserved**:
   - Navigate to Paris trip detail page
   - Badge now shows "Paid User" instead of "Free User"
   - Collaboration continues (can still view, suggest)
   - Still cannot edit directly (not trip creator)

5. **Verify multi-collaboration enabled**:
   - Paid User invites Free User (now upgraded) to **second trip** (different destination)
   - Upgraded user accepts second invitation → No "Leave current trip first" error
   - User now collaborating on **two trips** (Paris + new trip)

6. **Create own trip**:
   - Click "Create Trip" button
   - Enter destination: "London", Dates "2026-10-01 to 2026-10-07"
   - Submit → Backend calls `POST /api/v1/trips`
   - Expected: Trip created with `creator_id = user.id` (now owns a trip)

### Validation Commands

```bash
# Upgrade to Paid User
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<FREE_USER_JWT>" \
  -d '{
    "plan_id": "<MONTHLY_PLAN_ID>",
    "payment_method_token": "tok_visa_test_123"
  }'

# Expected: 201 Created, subscription.status='active', user.has_subscription=true

# Check collaboration count (should allow > 1)
curl -X GET http://localhost:8080/api/v1/trips \
  -H "Cookie: access_token=<UPGRADED_USER_JWT>"

# Expected: 200 OK, list includes both owned trips and collaborations (> 1 collaboration)

# Create own trip (now allowed)
curl -X POST http://localhost:8080/api/v1/trips \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<UPGRADED_USER_JWT>" \
  -d '{
    "destination": "London",
    "start_date": "2026-10-01",
    "end_date": "2026-10-07"
  }'

# Expected: 201 Created, trip.creator_id = upgraded user ID
```

### Database Validation

```sql
-- Check user upgrade
SELECT u.email, u.has_subscription, s.status
FROM users u
LEFT JOIN subscriptions s ON s.user_id = u.id
WHERE u.email = 'free-user@example.com';

-- Expected: has_subscription=true, subscription.status='active'

-- Check collaboration count (should be >= 2 after accepting second invitation)
SELECT COUNT(*) FROM collaborators
WHERE user_id = (SELECT id FROM users WHERE email = 'free-user@example.com')
AND status = 'accepted';

-- Expected: >= 2 (no single-collaboration limit after upgrade)

-- Check owned trips
SELECT COUNT(*) FROM trips
WHERE creator_id = (SELECT id FROM users WHERE email = 'free-user@example.com');

-- Expected: >= 1 (London trip created)
```

---

## Summary

**Total Scenarios**: 6 (mapping to 6 user stories in spec.md)
- Scenario 1: Paid User registration + first trip (US1)
- Scenario 2: Free User registration + accept invitation (US2)
- Scenario 3: Login + dashboard access + rate limiting (US3)
- Scenario 4: Password reset + password change (US4)
- Scenario 5: Collaboration + suggestion + approval (US5)
- Scenario 6: Free User upgrade + collaboration preserved (US2 extension)

**Validation Methods**:
- **Browser**: Manual UI testing via frontend app
- **cURL**: Direct API testing with request/response validation
- **Database**: SQL queries to verify data integrity and business rules

**Expected Test Execution Time**: ~30 minutes for all scenarios (manual browser testing)

**Automation**: All scenarios can be converted to E2E Playwright tests in `e2e/specs/auth/` and `e2e/specs/collaboration/` directories (see plan.md Project Structure).
