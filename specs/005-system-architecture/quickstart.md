# Quickstart: System Architecture and Technology Stack

**Feature**: System Architecture and Technology Stack  
**Created**: 2026-07-03  
**Purpose**: Validation scenarios to prove architecture works end-to-end

## Overview

This quickstart provides runnable validation scenarios for backend, frontend, and infrastructure architectural foundations. Each scenario validates a specific architectural layer without implementing full features. Success means the architecture is correctly configured and ready for feature development.

---

## Scenario 1: Backend Skeleton Validation

**Goal**: Validate backend structure, middleware chain, database client, and health check.

### Prerequisites

- Go 1.24+ installed
- PostgreSQL 15.4 running locally or accessible
- Environment variables configured

### Setup

1. **Create backend structure**:
```bash
cd backend
mkdir -p cmd/api internal/middleware internal/database internal/errors config
```

2. **Create configuration** (`config/config.go`):
```go
package config

import (
    "fmt"
    "os"
)

type Config struct {
    HTTPPort      string
    DatabaseURL   string
    LogLevel      string
}

func Load() (*Config, error) {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        return nil, fmt.Errorf("DATABASE_URL not set")
    }

    return &Config{
        HTTPPort:    getEnv("HTTP_PORT", "8080"),
        DatabaseURL: dbURL,
        LogLevel:    getEnv("LOG_LEVEL", "info"),
    }, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

3. **Create database client** (`internal/database/client.go`):
```go
package database

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

func NewClient(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, err
    }

    config.MinConns = 5
    config.MaxConns = 25

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, err
    }

    if err := pool.Ping(ctx); err != nil {
        return nil, err
    }

    return pool, nil
}
```

4. **Create request ID middleware** (`internal/middleware/request_id.go`):
```go
package middleware

import (
    "context"
    "net/http"
    "github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
        w.Header().Set("X-Request-ID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

5. **Create health check handler** (`cmd/api/main.go`):
```go
package main

import (
    "context"
    "encoding/json"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "backend/config"
    "backend/internal/database"
    "backend/internal/middleware"
    "github.com/go-chi/chi/v5"
)

func main() {
    ctx := context.Background()

    cfg, err := config.Load()
    if err != nil {
        slog.Error("failed to load config", "error", err)
        os.Exit(1)
    }

    db, err := database.NewClient(ctx, cfg.DatabaseURL)
    if err != nil {
        slog.Error("failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    r := chi.NewRouter()
    r.Use(middleware.RequestID)

    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        if err := db.Ping(r.Context()); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy"})
            return
        }
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })

    server := &http.Server{
        Addr:    ":" + cfg.HTTPPort,
        Handler: r,
    }

    go func() {
        slog.Info("starting server", "port", cfg.HTTPPort)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("server error", "error", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    slog.Info("shutting down server")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        slog.Error("server shutdown error", "error", err)
    }
}
```

### Run

```bash
# Set environment variables
export DATABASE_URL="postgres://user:pass@localhost:5432/traveler?sslmode=disable"
export HTTP_PORT="8080"

# Install dependencies
go mod init backend
go get github.com/go-chi/chi/v5
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/google/uuid

# Run server
go run cmd/api/main.go
```

### Test

```bash
# Health check should return 200 OK with X-Request-ID header
curl -v http://localhost:8080/healthz

# Expected output:
# < HTTP/1.1 200 OK
# < X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
# < Content-Type: application/json
# {"status":"ok"}
```

### Expected Outcome

✅ Server starts without errors  
✅ Database connection pool established (min 5, max 25 connections)  
✅ `/healthz` returns `200 OK` with `{"status":"ok"}`  
✅ Response includes `X-Request-ID` header (UUID format)  
✅ Graceful shutdown on SIGTERM (drains in-flight requests)

---

## Scenario 2: Frontend Scaffold Validation

**Goal**: Validate React app structure, TanStack Query, Zustand store, and design tokens.

### Prerequisites

- Node.js 20+ installed
- npm or yarn

### Setup

1. **Create React app with Vite**:
```bash
npm create vite@latest frontend -- --template react-ts
cd frontend
npm install
```

2. **Install dependencies**:
```bash
npm install @tanstack/react-query zustand react-router-dom
npm install -D vitest @testing-library/react @testing-library/jest-dom
```

3. **Create design tokens** (`src/styles/tokens.css`):
```css
:root {
  /* Colors */
  --color-primary: #1976d2;
  --color-surface: #ffffff;
  --color-text: #212121;

  /* Spacing */
  --space-sm: 8px;
  --space-md: 16px;
  --space-lg: 24px;

  /* Typography */
  --font-size-base: 14px;
  --font-size-lg: 16px;

  /* Border Radius */
  --radius-md: 8px;
}
```

4. **Create API client** (`src/lib/api-client.ts`):
```tsx
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public requestId?: string
  ) {
    super(message);
    this.name = 'APIError';
  }
}

export async function apiRequest<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new APIError(
      error.message || 'Request failed',
      response.status,
      error.request_id
    );
  }

  return response.json();
}
```

5. **Create auth store** (`src/stores/auth-store.ts`):
```tsx
import { create } from 'zustand';

interface User {
  id: string;
  email: string;
}

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  login: (user: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  user: null,
  login: (user) => set({ isAuthenticated: true, user }),
  logout: () => set({ isAuthenticated: false, user: null }),
}));
```

6. **Create primitive component** (`src/components/primitives/Button.tsx`):
```tsx
import styles from './Button.module.css';

interface ButtonProps {
  variant?: 'primary' | 'secondary';
  onClick?: () => void;
  children: React.ReactNode;
}

export function Button({ variant = 'primary', onClick, children }: ButtonProps) {
  return (
    <button
      className={styles[`button-${variant}`]}
      onClick={onClick}
    >
      {children}
    </button>
  );
}
```

7. **Create Button styles** (`src/components/primitives/Button.module.css`):
```css
.button-primary {
  background-color: var(--color-primary);
  color: white;
  padding: var(--space-md);
  border-radius: var(--radius-md);
  border: none;
  cursor: pointer;
  font-size: var(--font-size-base);
}

.button-primary:hover {
  opacity: 0.9;
}
```

8. **Update App.tsx**:
```tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Button } from './components/primitives/Button';
import { useAuthStore } from './stores/auth-store';
import './styles/tokens.css';

const queryClient = new QueryClient();

function App() {
  const { isAuthenticated, user, login, logout } = useAuthStore();

  return (
    <QueryClientProvider client={queryClient}>
      <div style={{ padding: 'var(--space-lg)' }}>
        <h1>TrAIveler Architecture Validation</h1>
        
        {isAuthenticated ? (
          <>
            <p>Welcome, {user?.email}!</p>
            <Button variant="primary" onClick={logout}>Logout</Button>
          </>
        ) : (
          <Button 
            variant="primary" 
            onClick={() => login({ id: '1', email: 'test@example.com' })}
          >
            Login
          </Button>
        )}
      </div>
    </QueryClientProvider>
  );
}

export default App;
```

### Run

```bash
# Set environment variable
export VITE_API_BASE_URL=http://localhost:8080

# Run dev server
npm run dev
```

### Test

1. Open browser to `http://localhost:5173`
2. Click "Login" button
3. Verify text changes to "Welcome, test@example.com!"
4. Click "Logout" button
5. Verify text changes back to "Login"

### Expected Outcome

✅ React app renders without errors  
✅ Design tokens applied (button uses `--color-primary`, `--space-md`, etc.)  
✅ Zustand store updates on login/logout (no prop drilling)  
✅ TanStack Query provider wraps app (ready for API calls)  
✅ Button component uses CSS Modules (no hard-coded styles)  
✅ All styling references tokens from `tokens.css`

---

## Scenario 3: Infrastructure Plan Validation

**Goal**: Validate Terraform module structure and staging environment plan.

### Prerequisites

- Terraform 1.5+ installed
- AWS CLI configured with credentials
- S3 bucket and DynamoDB table for remote state (created manually)

### Setup

1. **Create Terraform structure**:
```bash
mkdir -p infra/{modules/{vpc,ecs},environments}
cd infra
```

2. **Create VPC module** (`modules/vpc/main.tf`):
```hcl
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.environment}-vpc"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

resource "aws_subnet" "public" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index)
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name        = "${var.environment}-public-subnet-${count.index + 1}"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

output "vpc_id" {
  value = aws_vpc.main.id
}

output "public_subnet_ids" {
  value = aws_subnet.public[*].id
}
```

3. **Create VPC variables** (`modules/vpc/variables.tf`):
```hcl
variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
}

variable "availability_zones" {
  description = "Availability zones"
  type        = list(string)
}
```

4. **Create root module** (`main.tf`):
```hcl
terraform {
  required_version = ">= 1.5"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

module "vpc" {
  source = "./modules/vpc"

  environment        = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
}
```

5. **Create root variables** (`variables.tf`):
```hcl
variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
}

variable "availability_zones" {
  description = "Availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}
```

6. **Create staging config** (`environments/staging.tfvars`):
```hcl
environment = "staging"
vpc_cidr    = "10.0.0.0/16"
```

7. **Create backend config** (`backend.tf`):
```hcl
terraform {
  backend "s3" {
    bucket         = "traveler-terraform-state"
    key            = "infra/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "traveler-terraform-locks"
    encrypt        = true
  }
}
```

### Run

```bash
# Initialize Terraform
terraform init

# Validate configuration
terraform validate

# Format check
terraform fmt -check

# Plan staging environment
terraform plan -var-file=environments/staging.tfvars
```

### Test Plan Output

```bash
terraform plan -var-file=environments/staging.tfvars | grep "Plan:"
```

### Expected Outcome

✅ `terraform init` succeeds (remote state configured)  
✅ `terraform validate` returns no errors  
✅ `terraform fmt -check` returns no formatting issues  
✅ `terraform plan` shows resources to create:
  - 1 VPC with CIDR 10.0.0.0/16
  - 2 public subnets across us-east-1a and us-east-1b
✅ All resources tagged with Environment=staging, ManagedBy=terraform, Project=traveler  
✅ Remote state locking works (DynamoDB lock acquired/released)

---

## Scenario 4: End-to-End Integration Validation

**Goal**: Validate frontend → backend → database integration.

### Prerequisites

- Backend server running (Scenario 1)
- Frontend app running (Scenario 2)
- PostgreSQL with test data

### Setup

1. **Add test endpoint to backend** (in `cmd/api/main.go`):
```go
r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message":    "Architecture validation successful",
        "request_id": middleware.GetRequestID(r.Context()),
    })
})
```

2. **Create test component in frontend** (`src/components/ArchitectureTest.tsx`):
```tsx
import { useQuery } from '@tanstack/react-query';
import { apiRequest } from '../lib/api-client';
import { Button } from './primitives/Button';

interface TestResponse {
  message: string;
  request_id: string;
}

export function ArchitectureTest() {
  const { data, error, isLoading, refetch } = useQuery({
    queryKey: ['test'],
    queryFn: () => apiRequest<TestResponse>('/test'),
    enabled: false, // Manual trigger
  });

  return (
    <div>
      <Button variant="primary" onClick={() => refetch()}>
        Test Backend Connection
      </Button>

      {isLoading && <p>Loading...</p>}
      {error && <p style={{ color: 'red' }}>Error: {error.message}</p>}
      {data && (
        <div>
          <p>✅ {data.message}</p>
          <p>Request ID: {data.request_id}</p>
        </div>
      )}
    </div>
  );
}
```

3. **Add to App.tsx**:
```tsx
import { ArchitectureTest } from './components/ArchitectureTest';

// Inside return statement:
<ArchitectureTest />
```

### Run

1. Ensure backend is running on port 8080
2. Ensure frontend dev server is running
3. Open browser to frontend URL
4. Click "Test Backend Connection" button

### Expected Outcome

✅ Frontend makes fetch request to backend `/test` endpoint  
✅ Backend returns JSON with message and request_id  
✅ TanStack Query handles loading/success states  
✅ Correlation ID propagated from frontend through backend  
✅ Response rendered in frontend UI  
✅ No CORS errors (CORS middleware configured correctly)

---

## Scenario 5: Accessibility Validation

**Goal**: Validate WCAG 2.1 AA compliance and design token usage.

### Prerequisites

- Frontend app running (Scenario 2)
- axe-core browser extension installed OR Playwright with axe-core

### Setup (Manual with Browser Extension)

1. Install axe DevTools browser extension
2. Open frontend in browser
3. Open DevTools → axe DevTools tab
4. Click "Scan All of My Page"

### Setup (Automated with Playwright)

1. **Install Playwright and axe-core**:
```bash
cd frontend
npm install -D @playwright/test @axe-core/playwright
```

2. **Create accessibility test** (`tests/accessibility.spec.ts`):
```tsx
import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Accessibility', () => {
  test('should not have WCAG 2.1 AA violations', async ({ page }) => {
    await page.goto('http://localhost:5173');

    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa'])
      .analyze();

    expect(accessibilityScanResults.violations).toEqual([]);
  });
});
```

3. **Run test**:
```bash
npx playwright test tests/accessibility.spec.ts
```

### Expected Outcome

✅ Zero WCAG 2.1 AA violations detected  
✅ All buttons have accessible names  
✅ Color contrast meets 4.5:1 minimum ratio  
✅ Focus indicators visible on all interactive elements  
✅ All colors, spacing, typography use design tokens (verify in DevTools)

---

## Summary

All validation scenarios passed confirm:

1. **Backend Architecture**: Domain-driven structure, middleware chain, database pooling, health check all operational
2. **Frontend Architecture**: Atomic Design layering, TanStack Query integration, Zustand state management, design tokens enforced
3. **Infrastructure**: Terraform modules structured correctly, remote state locking works, staging plan validates
4. **Integration**: Frontend ↔ Backend communication works with correlation IDs, CORS configured correctly
5. **Accessibility**: WCAG 2.1 AA compliance validated, no hard-coded styling values

**Next Steps**: Architecture validated. Ready for feature implementation. Create tasks via `/speckit.tasks`.
