# Data Model: System Architecture and Technology Stack

**Feature**: System Architecture and Technology Stack  
**Created**: 2026-07-03  
**Status**: Complete

## Overview

This document defines the architectural components (not database entities) for backend, frontend, and infrastructure layers. Each component has clear responsibilities, dependencies, and integration contracts. Components are organized by layer to reflect the three-tier architecture.

---

## Backend Core Components

### HTTPServer

**Description**: Chi router-based HTTP server managing the request/response lifecycle.

**Responsibilities**:
- Start HTTP listener on configured port (default: 8080)
- Register domain handlers with Chi router
- Apply global middleware chain to all requests
- Graceful shutdown on SIGTERM/SIGINT (drain in-flight requests)
- Health check endpoint (`GET /healthz`)

**Dependencies**:
- Chi router (`github.com/go-chi/chi/v5`)
- Middleware components (RequestIDGenerator, Logger, PanicRecovery, CORS, BodySizeLimiter)
- Domain handlers (`internal/<domain>/handler.go`)

**Configuration**:
- `HTTP_PORT`: Server listen port (default: 8080)
- `READ_TIMEOUT`: Request read timeout (default: 30s)
- `WRITE_TIMEOUT`: Response write timeout (default: 60s for AI streaming)
- `SHUTDOWN_TIMEOUT`: Graceful shutdown drain period (default: 30s)

**Lifecycle**:
- `NewServer(config, middleware, handlers)`: Constructor with dependency injection
- `Start()`: Begin listening (blocking)
- `Shutdown(ctx)`: Graceful shutdown with context deadline

---

### Middleware Chain

**Description**: Global HTTP middleware applied to all requests in order.

**Components**:
1. **RequestIDGenerator**: Generates unique correlation ID (UUID v4), injects into context, adds `X-Request-ID` response header
2. **Logger**: Structured JSON logging (`log/slog`) with request method, path, status, duration, correlation ID
3. **PanicRecovery**: Catches panics, logs stack trace, returns `500 Internal Server Error` with correlation ID
4. **CORS**: Configures CORS headers for frontend origin (staging/production domains)
5. **BodySizeLimiter**: Rejects requests > 10 MB with `413 Payload Too Large`

**Application Order**: RequestIDGenerator → Logger → PanicRecovery → CORS → BodySizeLimiter → Handler

**Dependencies**:
- `log/slog` for structured logging
- `context.Context` for correlation ID propagation

**Configuration**:
- `ALLOWED_ORIGINS`: CORS origins (comma-separated, default: `https://staging.traveler.example.com`)
- `MAX_BODY_SIZE`: Request body size limit in bytes (default: 10485760 = 10 MB)

---

### DatabaseClient

**Description**: pgx connection pool managing PostgreSQL connections.

**Responsibilities**:
- Establish connection pool on startup with min/max connection limits
- Provide `Exec`, `Query`, `QueryRow` methods to repositories
- Health check (`Ping()` for `/healthz`)
- Graceful connection closure on shutdown

**Dependencies**:
- `github.com/jackc/pgx/v5/pgxpool` for connection pooling
- PostgreSQL 15.4 on Amazon RDS

**Configuration**:
- `DATABASE_URL`: PostgreSQL connection string (retrieved from AWS Secrets Manager)
- `DB_POOL_MIN_CONNS`: Minimum persistent connections (default: 5)
- `DB_POOL_MAX_CONNS`: Maximum connections (default: 25)
- `DB_POOL_MAX_CONN_LIFETIME`: Connection max age (default: 1h)
- `DB_POOL_MAX_CONN_IDLE_TIME`: Idle connection timeout (default: 30m)

**Error Handling**:
- Connection errors wrapped with context: `fmt.Errorf("database connection failed: %w", err)`
- Query errors returned to repository layer (not logged directly)

---

### AIClient

**Description**: Anthropic SDK wrapper for itinerary generation API calls.

**Responsibilities**:
- Send itinerary generation requests to Claude API
- Apply prompt validation (PromptValidator) before sending
- Apply output sanitization (OutputSanitizer) on responses
- Handle streaming responses (multi-turn conversation)
- Retry logic with exponential backoff on transient errors (429, 503)

**Dependencies**:
- Anthropic Go SDK (`github.com/anthropics/anthropic-sdk-go`)
- PromptValidator (prompt injection detection)
- OutputSanitizer (XSS prevention)

**Configuration**:
- `ANTHROPIC_API_KEY`: API key (retrieved from AWS Secrets Manager)
- `AI_TIMEOUT_NON_STREAMING`: Request timeout for non-streaming (default: 60s)
- `AI_TIMEOUT_STREAMING`: No timeout for streaming (infinite, relies on HTTP/2 keep-alive)
- `AI_MAX_RETRIES`: Retry attempts on transient failures (default: 3)

**Interface** (`internal/ai/client.go`):
```go
type AIClient interface {
    GenerateItinerary(ctx context.Context, prompt string) (*ItineraryResponse, error)
    StreamItinerary(ctx context.Context, prompt string) (<-chan ItineraryChunk, <-chan error)
}
```

**Error Handling**:
- 400 (invalid prompt): Return validation error to caller (user input issue)
- 429 (rate limit): Respect `Retry-After` header, return `503 Service Unavailable` to client
- 503 (AI provider unavailable): Fail fast, return `503` to client with `Retry-After: 60`

---

### RepositoryPattern

**Description**: Generic interface for data access abstraction across domains.

**Responsibilities**:
- Encapsulate SQL queries (SELECT, INSERT, UPDATE, DELETE)
- Return domain models (not raw SQL rows)
- Wrap database errors with context before returning to service layer

**Interface** (example for `TripRepository`):
```go
type TripRepository interface {
    Create(ctx context.Context, trip *Trip) error
    FindByID(ctx context.Context, id string) (*Trip, error)
    Update(ctx context.Context, trip *Trip) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, userID string, filters ListFilters) ([]*Trip, error)
}
```

**Dependencies**:
- DatabaseClient (pgx connection pool)
- Domain models (`internal/<domain>/model.go`)

**Error Wrapping**:
```go
if err != nil {
    return fmt.Errorf("failed to create trip: %w", err)
}
```

**Implementation**: Each domain has `repository.go` file implementing domain-specific repository interface.

---

### ServicePattern

**Description**: Business logic layer encapsulating domain operations.

**Responsibilities**:
- Validate business rules before database/AI operations
- Orchestrate repository calls and external API calls (AI, future integrations)
- Return domain errors (not HTTP status codes)

**Dependencies**:
- Repository interfaces (injected via constructor)
- AIClient interface (for itinerary-related services)

**Interface** (example for `TripService`):
```go
type TripService interface {
    CreateTrip(ctx context.Context, req CreateTripRequest) (*Trip, error)
    GetTrip(ctx context.Context, tripID, userID string) (*Trip, error)
    UpdateTrip(ctx context.Context, req UpdateTripRequest) error
    DeleteTrip(ctx context.Context, tripID, userID string) error
}
```

**Testability**: Services have no HTTP dependencies; can be unit tested with repository mocks.

---

### ErrorHandler

**Description**: Centralized error-to-HTTP status mapping.

**Responsibilities**:
- Map domain errors to HTTP status codes (404, 400, 500, etc.)
- Return structured JSON error responses with correlation ID
- Log errors with appropriate severity (WARN for 4xx, ERROR for 5xx)

**Error Response Format**:
```json
{
  "error": "resource_not_found",
  "message": "Trip not found",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Mapping**:
- `ErrNotFound` → 404
- `ErrValidation` → 400
- `ErrUnauthorized` → 401
- `ErrForbidden` → 403
- `ErrConflict` → 409
- All other errors → 500

**Implementation**: `internal/errors/handler.go` exports `HandleError(w, r, err)`

---

## Frontend Core Components

### AppRouter

**Description**: React Router v7 configuration defining application routes.

**Responsibilities**:
- Define route paths and associated components
- Implement protected routes (redirect to `/login` if unauthenticated)
- Lazy load route components for code splitting

**Dependencies**:
- React Router v7 (`react-router-dom`)
- AuthStore (authentication state)

**Configuration**:
```tsx
const routes = [
  { path: '/', element: <LandingPage /> },
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
  { path: '/dashboard', element: <ProtectedRoute><Dashboard /></ProtectedRoute> },
  { path: '/trips/:id', element: <ProtectedRoute><TripDetailPage /></ProtectedRoute> },
  { path: '/generate', element: <ProtectedRoute><GeneratePage /></ProtectedRoute> },
];
```

**Protected Route Logic**:
```tsx
const ProtectedRoute = ({ children }) => {
  const isAuthenticated = useAuthStore(state => state.isAuthenticated);
  return isAuthenticated ? children : <Navigate to="/login" />;
};
```

---

### QueryProvider

**Description**: TanStack Query provider with global configuration.

**Responsibilities**:
- Configure default query behavior (staleTime, cacheTime, retry policy)
- Provide query client to all child components via context

**Dependencies**:
- TanStack Query v5 (`@tanstack/react-query`)
- APIClient (fetch wrapper)

**Configuration**:
```tsx
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,  // 5 minutes
      cacheTime: 1000 * 60 * 10, // 10 minutes
      retry: 1,                   // Retry once on failure
      refetchOnWindowFocus: false,
    },
  },
});
```

**Usage**: Wrap `<App />` with `<QueryClientProvider client={queryClient}>`

---

### AuthStore

**Description**: Zustand store managing global authentication state.

**Responsibilities**:
- Store authentication status (`isAuthenticated`, `user`)
- Provide typed actions: `login(user)`, `logout()`, `refreshSession()`
- Persist state to sessionStorage (not localStorage, for security)

**Dependencies**:
- Zustand (`zustand`)

**State Shape**:
```tsx
interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  login: (user: User) => void;
  logout: () => void;
  refreshSession: () => Promise<void>;
}
```

**Implementation**: `src/stores/auth-store.ts`

**Usage**:
```tsx
const { isAuthenticated, user, logout } = useAuthStore();
```

---

### APIClient

**Description**: Thin wrapper around native fetch API for backend communication.

**Responsibilities**:
- Inject base URL (staging/production backend API URL)
- Inject default headers (`Content-Type: application/json`, `X-Request-ID`)
- Inject auth tokens from HTTP-only cookies (sent automatically)
- Normalize responses: 2xx → return `data`, 4xx/5xx → throw structured error

**Dependencies**:
- Native fetch API (no external library)

**Configuration**:
- `VITE_API_BASE_URL`: Backend API base URL (e.g., `https://api.staging.traveler.example.com`)

**Implementation** (`src/lib/api-client.ts`):
```tsx
export async function apiRequest<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'X-Request-ID': crypto.randomUUID(),
      ...options?.headers,
    },
    credentials: 'include', // Send HTTP-only cookies
  });

  if (!response.ok) {
    const error = await response.json();
    throw new APIError(error.message, response.status, error.request_id);
  }

  return response.json();
}
```

---

### DesignTokens

**Description**: CSS custom properties defining the design system.

**Responsibilities**:
- Define color palette, spacing scale, typography, border-radius, shadows
- Provide light/dark mode variants (future)
- Ensure WCAG 2.1 AA contrast ratios

**Location**: `src/styles/tokens.css`

**Structure**:
```css
:root {
  /* Colors */
  --color-primary: #1976d2;
  --color-secondary: #dc004e;
  --color-text: #212121;
  --color-background: #fafafa;

  /* Spacing (8px base unit) */
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 16px;
  --space-lg: 24px;
  --space-xl: 32px;

  /* Typography */
  --font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --font-size-sm: 12px;
  --font-size-base: 14px;
  --font-size-lg: 16px;
  --font-size-xl: 20px;

  /* Border Radius */
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;

  /* Shadows */
  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.12);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.16);
  --shadow-lg: 0 10px 20px rgba(0, 0, 0, 0.20);
}
```

**Usage**: Reference tokens in CSS Modules:
```css
.button {
  background-color: var(--color-primary);
  padding: var(--space-md);
  border-radius: var(--radius-md);
}
```

---

### ComponentLibrary

**Description**: Reusable UI components following Atomic Design layering.

**Primitives** (`src/components/primitives/`):
- `Button.tsx`: Primary, secondary, danger variants
- `Input.tsx`: Text input with label, error message, validation state
- `Card.tsx`: Container with optional header, body, footer
- `Modal.tsx`: Accessible dialog with backdrop, close button

**Composites** (`src/components/composites/`):
- `Form.tsx`: Form wrapper with validation, submit handling
- `DataTable.tsx`: Sortable, paginated table with row actions
- `NavigationBar.tsx`: Top navigation with logo, user menu, logout

**Responsibilities**:
- Enforce WCAG 2.1 AA compliance (keyboard operability, focus indicators, ARIA labels)
- Use design tokens for all styling (no hard-coded values)
- Handle Loading, Error, Empty states where applicable

---

### ErrorBoundary

**Description**: React error boundary catching unhandled rendering errors.

**Responsibilities**:
- Catch errors in child component tree
- Display fallback UI with recovery action (retry, navigate home)
- Log errors to monitoring service (future: Sentry integration)

**Implementation** (`src/components/ErrorBoundary.tsx`):
```tsx
class ErrorBoundary extends React.Component {
  state = { hasError: false, error: null };

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error('ErrorBoundary caught:', error, errorInfo);
    // TODO: Send to monitoring service
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback error={this.state.error} reset={() => this.setState({ hasError: false })} />;
    }
    return this.props.children;
  }
}
```

---

## Infrastructure Core Modules

### VPCModule

**Description**: Terraform module creating VPC with isolated subnets.

**Resources**:
- VPC with configurable CIDR block (staging: 10.0.0.0/16, production: 10.1.0.0/16)
- Public subnets across 2 AZs (for ALB)
- Private subnets across 2 AZs (for ECS, RDS)
- Internet Gateway for public subnet routing
- NAT instance (staging, cost-optimized) or NAT Gateway (production, production-grade)
- Route tables associating subnets with gateways

**Inputs**:
- `environment`: "staging" or "production"
- `vpc_cidr`: CIDR block (e.g., "10.0.0.0/16")
- `availability_zones`: List of AZs (e.g., ["us-east-1a", "us-east-1b"])

**Outputs**:
- `vpc_id`: VPC identifier
- `public_subnet_ids`: List of public subnet IDs
- `private_subnet_ids`: List of private subnet IDs

---

### ECSModule

**Description**: Terraform module creating ECS Fargate cluster and service.

**Resources**:
- ECS cluster
- Task definition with environment-specific CPU/memory (staging: 0.25 vCPU/0.5 GB RAM, production: 1 vCPU/2 GB RAM)
- ECS service with auto-scaling policies (CPU target tracking at 70%)
- CloudWatch log group with environment-specific retention (staging: 7 days, production: 30 days)

**Inputs**:
- `environment`: "staging" or "production"
- `task_cpu`: CPU units (e.g., "256" for 0.25 vCPU)
- `task_memory`: Memory in MB (e.g., "512" for 0.5 GB)
- `min_tasks`: Minimum task count (staging: 1, production: 2)
- `max_tasks`: Maximum task count (staging: 2, production: 20)
- `log_retention_days`: CloudWatch log retention (staging: 7, production: 30)

**Outputs**:
- `cluster_name`: ECS cluster name
- `service_name`: ECS service name
- `task_definition_arn`: Task definition ARN for CI/CD updates

---

### RDSModule

**Description**: Terraform module creating PostgreSQL RDS instance.

**Resources**:
- RDS instance (db.t4g.micro for staging, db.t4g.small for production)
- Subnet group (private subnets)
- Security group (allow 5432 from ECS security group only)
- Parameter group (PostgreSQL 15.4 optimized settings)
- Automated backups with configurable retention (staging: 1 day, production: 30 days)

**Inputs**:
- `environment`: "staging" or "production"
- `instance_class`: RDS instance type (e.g., "db.t4g.micro")
- `multi_az`: Boolean (staging: false, production: true)
- `backup_retention_days`: Backup retention period (staging: 1, production: 30)

**Outputs**:
- `db_endpoint`: RDS instance endpoint (host:port)
- `db_name`: Database name
- `db_secret_arn`: ARN of AWS Secrets Manager secret containing credentials

---

### ALBModule

**Description**: Terraform module creating Application Load Balancer.

**Resources**:
- ALB in public subnets
- Target group for ECS service (HTTP 8080, health check `/healthz`)
- HTTPS listener (443) with SSL certificate
- HTTP listener (80) redirecting to HTTPS

**Inputs**:
- `environment`: "staging" or "production"
- `vpc_id`: VPC identifier
- `public_subnet_ids`: List of public subnet IDs
- `certificate_arn`: ACM certificate ARN for HTTPS

**Outputs**:
- `alb_dns_name`: ALB DNS name
- `target_group_arn`: Target group ARN for ECS service registration

---

### CloudFrontModule

**Description**: Terraform module creating CloudFront distribution for frontend.

**Resources**:
- S3 bucket for React build artifacts
- CloudFront distribution with S3 origin
- Origin Access Identity (OAI) restricting S3 to CloudFront-only access
- SSL certificate for custom domain
- Cache invalidation on deployment

**Inputs**:
- `environment`: "staging" or "production"
- `domain_name`: Custom domain (e.g., "staging.traveler.example.com")
- `certificate_arn`: ACM certificate ARN

**Outputs**:
- `s3_bucket_name`: S3 bucket name for frontend builds
- `cloudfront_distribution_id`: Distribution ID for cache invalidation in CI/CD

---

### SecretsModule

**Description**: Terraform module creating AWS Secrets Manager secrets.

**Resources**:
- Database credentials secret (username, password)
- Anthropic API key secret
- JWT signing key secret (RS256 private key)

**Inputs**:
- `environment`: "staging" or "production"

**Outputs**:
- `db_secret_arn`: ARN for database credentials
- `ai_api_key_secret_arn`: ARN for Anthropic API key
- `jwt_signing_key_secret_arn`: ARN for JWT signing key

**Secret Rotation**: Manual rotation for MVP; automatic rotation post-MVP.

---

## Component Relationships

### Backend Request Flow
```
HTTP Request → HTTPServer → Middleware Chain → Handler (presentation) → Service (business logic) → Repository (data access) → DatabaseClient → PostgreSQL
                                                                         ↳ AIClient → Anthropic API
```

### Frontend Data Flow
```
Component → TanStack Query Hook → APIClient (fetch wrapper) → Backend API
            ↓
          QueryProvider (cache)
```

### Infrastructure Dependency Graph
```
VPCModule → ECSModule (requires VPC, subnets)
         → RDSModule (requires VPC, subnets)
         → ALBModule (requires VPC, subnets)
         → SecretsModule (independent)
         → CloudFrontModule (independent)
```

---

## Invariants

### Backend
1. All repository errors MUST be wrapped with context before returning to service layer
2. Service layer MUST NOT depend on HTTP-specific types (http.Request, http.ResponseWriter)
3. External API clients (AI) MUST be abstracted behind Go interfaces for testability
4. Configuration MUST be loaded from environment variables; application MUST NOT start if required config missing

### Frontend
5. All styling MUST reference design tokens; hard-coded color/spacing values forbidden
6. All API communication MUST use TanStack Query; direct fetch calls outside APIClient forbidden
7. Global state (auth) MUST use Zustand; prop-drilling authentication forbidden
8. All components MUST handle Loading, Error, Empty states

### Infrastructure
9. All AWS resources MUST be tagged with: Environment (staging/production), ManagedBy (terraform), Project (traveler)
10. All sensitive Terraform outputs (passwords, API keys) MUST be marked `sensitive = true`
11. Terraform state MUST be stored remotely in S3 with DynamoDB locking
12. Environment-specific values MUST be in .tfvars files (not hard-coded in modules)

---

## Summary

Architecture defines 26 core components across three layers: backend (7 components), frontend (7 components), infrastructure (6 modules), plus 6 shared architectural patterns. All components have clear responsibilities, dependencies, and integration contracts. Components align with constitution principles: layered separation (simplicity), interface abstractions (testability), token-driven UI (accessibility), secrets management (security).

Ready for contract definitions and quickstart validation scenarios.
