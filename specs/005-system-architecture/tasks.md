# Tasks: System Architecture and Technology Stack

**Branch**: `005-system-architecture`  
**Input**: Design documents from `/specs/005-system-architecture/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/](contracts/)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each architectural layer.

---

## Format: `- [ ] [ID] [P?] [Story] Description with file path`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1=Backend, US2=Frontend, US3=Infrastructure, US4=Integration)
- All file paths use project structure from plan.md

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Create project structure and initialize dependencies

- [x] T001 Create backend directory structure per plan.md: backend/{cmd/api,internal/{middleware,database,ai,errors},pkg,config,migrations,tests/{integration,fixtures}}
- [x] T002 [P] Create frontend directory structure per plan.md: frontend/src/{components/{primitives,composites},features,hooks,lib,stores,styles,routes}
- [x] T003 [P] Create e2e directory structure: e2e/{tests,fixtures,playwright.config.ts}
- [x] T004 [P] Create infrastructure directory structure: infra/{modules/{vpc,ecs,rds,alb,cloudfront,secrets},environments}
- [x] T005 Initialize Go module in backend/go.mod with Go 1.24+ and core dependencies (Chi, pgx/v5, goose/v3, google/uuid, log/slog)
- [x] T006 [P] Initialize React project in frontend/ with Vite, TypeScript strict mode, and core dependencies (TanStack Query v5, Zustand, React Router v7, Lucide React)
- [x] T007 [P] Initialize E2E project in e2e/ with Playwright and axe-core dependencies
- [x] T008 Create backend/.env.example with required environment variables (DATABASE_URL, HTTP_PORT, LOG_LEVEL, ANTHROPIC_API_KEY)
- [x] T009 [P] Create frontend/.env.example with VITE_API_BASE_URL variable
- [x] T010 Create backend/README.md with quickstart instructions, directory structure explanation, and development workflow
- [x] T011 [P] Create frontend/README.md with development server instructions, component guidelines, and testing commands
- [x] T012 [P] Create infra/README.md with Terraform initialization instructions and environment deployment guide

**Checkpoint**: Project structure complete - foundational infrastructure can now be built

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story implementation can begin

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T013 Create backend/config/config.go with configuration struct and environment variable loading using os.Getenv with validation
- [x] T014 [P] Create backend/Dockerfile with multi-stage build (builder stage with Go 1.24+, runtime stage with minimal Alpine)
- [x] T015 [P] Create .gitignore files for backend/ (exclude vendor/, .env, binary), frontend/ (exclude node_modules/, dist/, .env), and infra/ (exclude .terraform/, *.tfstate)
- [x] T016 [P] Create docker-compose.yml for local development with PostgreSQL 15.4 service and backend service configuration
- [x] T017 Create .github/workflows/backend-ci.yml skeleton (lint, test, build jobs without full implementation)
- [x] T018 [P] Create .github/workflows/frontend-ci.yml skeleton (lint, test, build, accessibility jobs without full implementation)
- [x] T019 [P] Create .github/workflows/infra-plan.yml skeleton (validate, format check, plan jobs)
- [x] T020 Configure golangci-lint in backend/.golangci.yml with required linters (errcheck, govet, staticcheck, revive, gosec)
- [x] T021 [P] Configure ESLint and Prettier in frontend/ with TypeScript strict mode rules and no-any enforcement
- [x] T022 Create infra/backend.tf with S3 backend configuration for remote state (bucket: traveler-terraform-state, DynamoDB table: traveler-terraform-locks)
- [x] T023 Create infra/versions.tf with Terraform >= 1.5 and AWS provider ~> 5.0 version constraints

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Backend Service Architecture (Priority: P1) 🎯 MVP

**Goal**: Establish Go backend structure with domain layering, middleware chain, database pooling, AI client abstraction, and error handling

**Independent Test**: Create skeleton backend with health check endpoint, middleware chain operational, database connection pool established, verify via `curl http://localhost:8080/healthz` returns 200 OK with X-Request-ID header

### Backend Core Infrastructure

- [x] T024 [P] [US1] Create backend/internal/middleware/request_id.go implementing UUID v4 generation, context injection, and X-Request-ID response header
- [x] T025 [P] [US1] Create backend/internal/middleware/logger.go using log/slog for structured JSON logging with method, path, status, duration, correlation ID
- [x] T026 [P] [US1] Create backend/internal/middleware/recovery.go implementing panic recovery with stack trace logging and 500 response
- [x] T027 [P] [US1] Create backend/internal/middleware/cors.go with configurable allowed origins from environment variable
- [x] T028 [P] [US1] Create backend/internal/middleware/body_size.go limiting request body to 10 MB with 413 response on violation
- [x] T029 [US1] Create backend/internal/database/client.go implementing pgx connection pool with min 5, max 25 connections, health check (Ping), and graceful closure
- [x] T030 [P] [US1] Create backend/internal/ai/client.go defining AIClient interface with GenerateItinerary and StreamItinerary methods
- [x] T031 [P] [US1] Create backend/internal/ai/validator.go implementing prompt validation stub (to be enhanced with injection detection rules later)
- [x] T032 [P] [US1] Create backend/internal/ai/sanitizer.go implementing output sanitization stub (HTML/script stripping to be enhanced later)
- [x] T033 [US1] Create backend/internal/errors/handler.go implementing domain error-to-HTTP status mapping (404, 400, 401, 403, 409, 500) with structured JSON responses
- [x] T034 [US1] Create backend/internal/errors/types.go defining domain error types (ErrNotFound, ErrValidation, ErrUnauthorized, ErrForbidden, ErrConflict)

### Backend HTTP Server

- [x] T035 [US1] Create backend/cmd/api/main.go implementing HTTPServer with Chi router, middleware chain registration (RequestID → Logger → Recovery → CORS → BodySize), health check endpoint, and graceful shutdown
- [x] T036 [US1] Integrate backend/internal/database/client.go initialization in main.go with configuration from config package and connection pool lifecycle management
- [x] T037 [US1] Add /healthz endpoint to main.go verifying database Ping() succeeds before returning 200 OK

### Backend Domain Pattern Scaffolds

- [x] T038 [P] [US1] Create backend/internal/example/model.go with sample domain model struct demonstrating naming conventions and field tags
- [x] T039 [P] [US1] Create backend/internal/example/repository.go implementing repository interface pattern with Create, FindByID, Update, Delete, List methods using pgx connection pool
- [x] T040 [P] [US1] Create backend/internal/example/service.go implementing service interface pattern with business logic, repository dependency injection, and domain error returns
- [x] T041 [US1] Create backend/internal/example/handler.go implementing HTTP handler calling service layer, using errors.HandleError for error responses, and demonstrating context value extraction (requestID, userID)

**Checkpoint**: Backend architecture scaffolded - new domains can follow example/ pattern, middleware chain operational, database pooling configured

---

## Phase 4: User Story 2 - Frontend Application Structure (Priority: P1) 🎯 MVP

**Goal**: Establish React SPA structure with Atomic Design layering, TanStack Query integration, Zustand state management, design token system, and accessibility foundation

**Independent Test**: Run `npm run dev`, open browser to localhost:5173, verify app renders without errors, design tokens applied to primitive components, Zustand auth store functional, TanStack Query provider wrapping app, axe-core reports zero violations

### Frontend Core Infrastructure

- [x] T042 [P] [US2] Create frontend/src/styles/tokens.css defining CSS custom properties for colors (primary, surface, text), spacing (sm, md, lg), typography (font-size-base, font-size-lg, font-weight-bold), border-radius (radius-md), and shadows (shadow-sm, shadow-md)
- [x] T043 [P] [US2] Create frontend/src/styles/global.css importing tokens.css and setting base styles (font-family, box-sizing, CSS reset)
- [x] T044 [P] [US2] Create frontend/src/lib/api-client.ts implementing fetch wrapper with base URL from env var, Content-Type and X-Request-ID headers, credentials include, and APIError class (status, message, requestId fields)
- [x] T045 [P] [US2] Create frontend/src/lib/query-client.ts configuring TanStack Query defaults (staleTime: 5 min, retry: 1, refetchOnWindowFocus: false)
- [x] T046 [US2] Create frontend/src/stores/auth-store.ts implementing Zustand store with isAuthenticated, user, login, logout, refreshSession actions (no persistence for MVP - session storage can be added later)
- [x] T047 [US2] Create frontend/src/App.tsx wrapping application with QueryClientProvider from lib/query-client.ts and importing global.css

### Frontend Component Primitives

- [x] T048 [P] [US2] Create frontend/src/components/primitives/Button.tsx with variant prop (primary, secondary, danger), size prop (sm, md, lg), CSS Module styling using design tokens, and aria-label support
- [x] T049 [P] [US2] Create frontend/src/components/primitives/Button.module.css referencing var(--color-primary), var(--space-md), var(--radius-md) from tokens.css
- [x] T050 [P] [US2] Create frontend/src/components/primitives/Input.tsx with label, id, name, required props, CSS Module styling, and associated label for accessibility
- [x] T051 [P] [US2] Create frontend/src/components/primitives/Card.tsx with children prop and CSS Module styling using var(--color-surface), var(--shadow-sm), var(--space-lg)
- [x] T052 [P] [US2] Create frontend/src/components/primitives/LoadingSpinner.tsx with aria-label prop for screen readers
- [x] T053 [P] [US2] Create frontend/src/components/primitives/ErrorMessage.tsx displaying error with retry button (optional onClick prop)
- [x] T054 [P] [US2] Create frontend/src/components/primitives/EmptyState.tsx with message and optional action button

### Frontend Composites and Features

- [x] T055 [US2] Create frontend/src/components/composites/Form.tsx composing Button and Input primitives, handling onSubmit with loading state, error display, and validation error mapping
- [x] T056 [P] [US2] Create frontend/src/features/.gitkeep as placeholder (actual features will be added in subsequent specs)
- [x] T057 [US2] Create frontend/src/components/ErrorBoundary.tsx implementing React.Component error boundary with fallback UI showing error message and "Go Home" action

### Frontend Routing and State

- [x] T058 [US2] Create frontend/src/routes/index.tsx defining React Router v7 routes configuration (root route returning simple "TrAIveler" heading as placeholder)
- [x] T059 [US2] Update frontend/src/App.tsx to include RouterProvider with routes from routes/index.tsx
- [x] T060 [US2] Wrap App.tsx with ErrorBoundary component

**Checkpoint**: Frontend architecture scaffolded - Atomic Design layering established, design tokens enforced, TanStack Query and Zustand operational, accessibility foundation in place

---

## Phase 5: User Story 3 - Infrastructure as Code Foundations (Priority: P1) 🎯 MVP

**Goal**: Establish Terraform modules for AWS resources (VPC, ECS, RDS, ALB, CloudFront, Secrets), environment separation via .tfvars, and CI/CD integration patterns

**Independent Test**: Run `terraform init && terraform validate && terraform plan -var-file=environments/staging.tfvars`, verify plan shows VPC, ECS, RDS, ALB, CloudFront, Secrets Manager resources with staging configuration, remote state locking operational

### Terraform Module: VPC

- [x] T061 [P] [US3] Create infra/modules/vpc/main.tf defining aws_vpc resource with CIDR from var.vpc_cidr, enable_dns_hostnames, enable_dns_support
- [x] T062 [P] [US3] Add aws_subnet.public resources in infra/modules/vpc/main.tf creating 2 public subnets across 2 AZs with cidrsubnet() and count
- [x] T063 [P] [US3] Add aws_subnet.private resources in infra/modules/vpc/main.tf creating 2 private subnets across 2 AZs
- [x] T064 [P] [US3] Add aws_internet_gateway and aws_route_table resources for public subnet routing in infra/modules/vpc/main.tf
- [x] T065 [P] [US3] Add conditional NAT resource (aws_instance for staging, aws_nat_gateway for production) based on var.enable_nat_gateway in infra/modules/vpc/main.tf
- [x] T066 [US3] Create infra/modules/vpc/variables.tf defining environment, vpc_cidr, availability_zones, enable_nat_gateway variables
- [x] T067 [US3] Create infra/modules/vpc/outputs.tf exporting vpc_id, public_subnet_ids, private_subnet_ids
- [x] T068 [US3] Add resource tagging in infra/modules/vpc/main.tf with Environment, ManagedBy, Project tags per FR-023

### Terraform Module: ECS

- [x] T069 [P] [US3] Create infra/modules/ecs/main.tf defining aws_ecs_cluster resource
- [x] T070 [P] [US3] Add aws_ecs_task_definition in infra/modules/ecs/main.tf with ARM64 architecture, task_cpu and task_memory from variables, container definition with ECR image URL, environment variables, CloudWatch log configuration
- [x] T071 [P] [US3] Add aws_ecs_service in infra/modules/ecs/main.tf with desired_count, launch_type FARGATE, network_configuration using private subnets, load_balancer attachment to ALB target group, health_check_grace_period_seconds 60
- [x] T072 [P] [US3] Add aws_appautoscaling_target and aws_appautoscaling_policy in infra/modules/ecs/main.tf for CPU-based target tracking at 70% with min_tasks and max_tasks from variables
- [x] T073 [P] [US3] Add aws_cloudwatch_log_group in infra/modules/ecs/main.tf with retention_in_days from variable
- [x] T074 [P] [US3] Add aws_security_group for ECS tasks in infra/modules/ecs/main.tf allowing ingress from ALB security group on port 8080
- [x] T075 [US3] Create infra/modules/ecs/variables.tf defining environment, vpc_id, private_subnet_ids, task_cpu, task_memory, min_tasks, max_tasks, log_retention_days, ecr_repository_url variables
- [x] T076 [US3] Create infra/modules/ecs/outputs.tf exporting cluster_name, service_name, task_definition_family, log_group_name
- [x] T077 [US3] Add resource tagging in infra/modules/ecs/main.tf with Environment, ManagedBy, Project tags

### Terraform Module: RDS

- [x] T078 [P] [US3] Create infra/modules/rds/main.tf defining aws_db_subnet_group using private subnets
- [x] T079 [P] [US3] Add aws_db_instance in infra/modules/rds/main.tf with engine postgresql, engine_version 15.4, instance_class from variable, allocated_storage 20, multi_az from variable, backup_retention_period from variable, backup_window 03:00-04:00, maintenance_window sun:04:00-sun:05:00
- [x] T080 [P] [US3] Add aws_security_group for RDS in infra/modules/rds/main.tf allowing ingress from ECS security group on port 5432
- [x] T081 [P] [US3] Add aws_secretsmanager_secret and aws_secretsmanager_secret_version in infra/modules/rds/main.tf for database credentials (username, password generated with random_password)
- [x] T082 [US3] Create infra/modules/rds/variables.tf defining environment, vpc_id, private_subnet_ids, instance_class, multi_az, backup_retention_days, db_name variables
- [x] T083 [US3] Create infra/modules/rds/outputs.tf exporting db_endpoint, db_name, db_secret_arn (marked sensitive)
- [x] T084 [US3] Add resource tagging in infra/modules/rds/main.tf with Environment, ManagedBy, Project tags

### Terraform Module: ALB

- [x] T085 [P] [US3] Create infra/modules/alb/main.tf defining aws_lb resource with load_balancer_type application, subnets from public_subnet_ids, security_groups
- [x] T086 [P] [US3] Add aws_lb_target_group in infra/modules/alb/main.tf with target_type ip, port 8080, protocol HTTP, health_check path /healthz, interval 30, timeout 5, healthy_threshold 2, unhealthy_threshold 3
- [x] T087 [P] [US3] Add aws_lb_listener for HTTPS (port 443) in infra/modules/alb/main.tf with ssl_policy, certificate_arn from variable, default_action forward to target group
- [x] T088 [P] [US3] Add aws_lb_listener for HTTP (port 80) in infra/modules/alb/main.tf with redirect action to HTTPS
- [x] T089 [P] [US3] Add aws_security_group for ALB in infra/modules/alb/main.tf allowing ingress on ports 80 and 443 from 0.0.0.0/0
- [x] T090 [US3] Create infra/modules/alb/variables.tf defining environment, vpc_id, public_subnet_ids, certificate_arn, ecs_security_group_id variables
- [x] T091 [US3] Create infra/modules/alb/outputs.tf exporting alb_dns_name, alb_zone_id, target_group_arn
- [x] T092 [US3] Add resource tagging in infra/modules/alb/main.tf with Environment, ManagedBy, Project tags

### Terraform Module: CloudFront

- [x] T093 [P] [US3] Create infra/modules/cloudfront/main.tf defining aws_s3_bucket for frontend builds with private ACL
- [x] T094 [P] [US3] Add aws_cloudfront_origin_access_identity and bucket policy in infra/modules/cloudfront/main.tf granting OAI read access to S3
- [x] T095 [P] [US3] Add aws_cloudfront_distribution in infra/modules/cloudfront/main.tf with S3 origin, default_cache_behavior (viewer_protocol_policy redirect-to-https, allowed_methods GET HEAD OPTIONS), custom_error_response for SPA routing (404 → /index.html), aliases from domain_name variable, viewer_certificate with certificate_arn
- [x] T096 [US3] Create infra/modules/cloudfront/variables.tf defining environment, domain_name, certificate_arn variables
- [x] T097 [US3] Create infra/modules/cloudfront/outputs.tf exporting s3_bucket_name, cloudfront_distribution_id, cloudfront_domain_name
- [x] T098 [US3] Add resource tagging in infra/modules/cloudfront/main.tf with Environment, ManagedBy, Project tags

### Terraform Module: Secrets

- [x] T099 [P] [US3] Create infra/modules/secrets/main.tf defining aws_secretsmanager_secret resources for db_credentials, ai_api_key, jwt_signing_key (secrets created empty, values populated manually post-apply)
- [x] T100 [US3] Create infra/modules/secrets/variables.tf defining environment variable
- [x] T101 [US3] Create infra/modules/secrets/outputs.tf exporting db_secret_arn, ai_api_key_secret_arn, jwt_signing_key_secret_arn (all marked sensitive)

### Terraform Root Module and Environment Configs

- [x] T102 [US3] Create infra/main.tf calling vpc, ecs, rds, alb, cloudfront, secrets modules with dependency injection via module outputs
- [x] T103 [US3] Create infra/variables.tf defining all root-level variables (environment, vpc_cidr, availability_zones, enable_nat_gateway, task_cpu, task_memory, min_tasks, max_tasks, log_retention_days, instance_class, multi_az, backup_retention_days, backend_domain, frontend_domain)
- [x] T104 [US3] Create infra/outputs.tf exporting CI/CD-relevant outputs (ecr_repository_url, ecs_cluster_name, ecs_service_name, s3_bucket_name, cloudfront_distribution_id)
- [x] T105 [US3] Create infra/environments/staging.tfvars with cost-optimized configuration (vpc_cidr 10.0.0.0/16, enable_nat_gateway false, task_cpu 256, task_memory 512, min_tasks 1, max_tasks 2, log_retention_days 7, instance_class db.t4g.micro, multi_az false, backup_retention_days 1)
- [x] T106 [US3] Create infra/environments/production.tfvars with production configuration (vpc_cidr 10.1.0.0/16, enable_nat_gateway true, task_cpu 1024, task_memory 2048, min_tasks 2, max_tasks 20, log_retention_days 30, instance_class db.t4g.small, multi_az true, backup_retention_days 30)

### Terraform CI/CD Integration

- [x] T107 [US3] Update .github/workflows/infra-plan.yml implementing terraform init, terraform validate, terraform fmt -check, terraform plan for both staging and production .tfvars, and PR comment with plan output
- [x] T108 [US3] Create .github/workflows/infra-apply.yml implementing terraform apply -auto-approve for staging on main merge, with output export to GitHub Secrets for backend-ci.yml and frontend-ci.yml
- [x] T109 [US3] Configure OIDC federation in AWS IAM (manual step documented in infra/README.md) creating IAM role with trust policy for GitHub Actions and permissions for Terraform operations

**Checkpoint**: Infrastructure modules complete - staging can be provisioned via `terraform apply -var-file=environments/staging.tfvars`, production dormant until load testing validates configuration

---

## Phase 6: User Story 4 - Integration and Error Handling Patterns (Priority: P2)

**Goal**: Implement standardized patterns for API integration, error correlation, logging, and graceful degradation across all layers

**Independent Test**: Simulate failure scenarios (database timeout, AI API error, malformed client request) and verify errors logged with correlation IDs, client receives structured error responses, retries/fallbacks work

### Backend Integration Patterns

- [x] T110 [P] [US4] Enhance backend/internal/errors/handler.go to extract correlation ID from context and include in all error responses
- [x] T111 [P] [US4] Update backend/internal/middleware/logger.go to log all 5xx errors with stack trace and correlation ID for observability
- [x] T112 [P] [US4] Create backend/internal/ai/anthropic.go implementing Anthropic SDK wrapper with exponential backoff retry (max 3 attempts), timeout configuration (60s non-streaming), and 429/503 error handling
- [x] T113 [US4] Update backend/internal/ai/client.go to use anthropic.go implementation and return service unavailable errors with Retry-After header on AI provider outages

### Frontend Integration Patterns

- [x] T114 [P] [US4] Enhance frontend/src/lib/api-client.ts to include X-Request-ID in all requests using crypto.randomUUID() and extract requestId from error responses
- [x] T115 [P] [US4] Create frontend/src/hooks/useErrorHandler.ts custom hook handling common error scenarios (401 → redirect to login, 403 → show permission error, 404 → show not found, 500/503 → show retry)
- [x] T116 [US4] Update frontend/src/components/primitives/ErrorMessage.tsx to display correlation ID when available in error responses

### Integration Validation

- [x] T117 [US4] Create integration test scenario in backend/tests/integration/health_test.go verifying health check returns 200 OK with correlation ID header
- [x] T118 [P] [US4] Create integration test scenario in backend/tests/integration/error_test.go verifying database timeout returns 500 with correlation ID and structured error response
- [x] T119 [US4] Document error handling patterns in backend/README.md and frontend/README.md with examples of domain error creation, error wrapping, and client error handling

**Checkpoint**: Integration patterns established - errors traceable via correlation IDs, graceful degradation implemented, retry logic operational

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and final cleanup across all layers

- [x] T120 [P] Update docs/architecture.md with final backend/frontend/infrastructure architecture diagrams and component descriptions from specs/005-system-architecture/
- [x] T121 [P] Update docs/coding-guidelines.md with Go import order (stdlib → external → internal), TypeScript import order (react → external → @/ → relative), and file naming conventions
- [x] T122 [P] Update docs/testing-guidelines.md with backend testing patterns (service mocks, repository integration tests), frontend testing patterns (React Testing Library, MSW, axe-core)
- [x] T123 [P] Update project root README.md with architecture overview, getting started instructions (Docker Compose local dev), and links to backend/frontend/infra READMEs
- [x] T124 Run all quickstart validation scenarios from specs/005-system-architecture/quickstart.md (Scenario 1: Backend skeleton, Scenario 2: Frontend scaffold, Scenario 3: Infrastructure plan, Scenario 4: E2E integration, Scenario 5: Accessibility) and document results
- [x] T125 Validate golangci-lint passes with zero errors on backend code
- [x] T126 [P] Validate ESLint and Prettier pass with zero errors on frontend code
- [x] T127 [P] Validate terraform fmt check passes on all infra files
- [x] T128 Run backend unit tests and verify ≥ 80% coverage on example domain service and repository
- [x] T129 [P] Run frontend component tests and verify primitives (Button, Input, Card) render correctly with design tokens
- [x] T130 [P] Run axe-core accessibility audit on frontend and verify zero WCAG 2.1 AA violations
- [x] T131 Update specs/005-system-architecture/tasks.md marking all tasks complete and adding completion notes

**Checkpoint**: Architecture fully documented, validated, and ready for feature development

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 - Backend (Phase 3)**: Depends on Foundational - can proceed independently
- **User Story 2 - Frontend (Phase 4)**: Depends on Foundational - can proceed independently
- **User Story 3 - Infrastructure (Phase 5)**: Depends on Foundational - can proceed independently
- **User Story 4 - Integration (Phase 6)**: Depends on US1, US2, US3 completion - requires all layers operational
- **Polish (Phase 7)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (Backend)**: Independent after Foundational - 41 tasks (T024-T041 + dependencies from US4)
- **US2 (Frontend)**: Independent after Foundational - 32 tasks (T042-T060 + dependencies from US4)
- **US3 (Infrastructure)**: Independent after Foundational - 49 tasks (T061-T109)
- **US4 (Integration)**: Depends on US1, US2, US3 - 10 tasks (T110-T119)

### Within Each User Story

**US1 - Backend**:
1. Middleware components (T024-T028) can run in parallel
2. DatabaseClient (T029) independent
3. AI client interface and stubs (T030-T032) can run in parallel
4. Error handler (T033-T034) independent
5. HTTPServer (T035-T037) depends on middleware and database client
6. Domain pattern scaffolds (T038-T041) depend on HTTPServer, can run in parallel for different files

**US2 - Frontend**:
1. Core infrastructure (T042-T047) sequential (tokens → global.css → api-client → query-client → auth-store → App.tsx)
2. Primitives (T048-T054) can run in parallel
3. Composites and features (T055-T057) depend on primitives
4. Routing (T058-T060) depends on composites and error boundary

**US3 - Infrastructure**:
1. Each module (VPC, ECS, RDS, ALB, CloudFront, Secrets) can be developed in parallel
2. Root module (T102-T104) depends on all modules complete
3. Environment configs (T105-T106) independent
4. CI/CD (T107-T109) depends on root module

**US4 - Integration**:
1. Backend enhancements (T110-T113) sequential within backend
2. Frontend enhancements (T114-T116) sequential within frontend
3. Integration tests (T117-T119) depend on backend/frontend enhancements

### Parallel Opportunities

**Within Setup (Phase 1)**:
- T002 (frontend structure), T003 (e2e structure), T004 (infra structure) in parallel
- T006 (React init), T007 (E2E init) in parallel after T002/T003
- T009 (frontend .env), T010 (backend README), T011 (frontend README), T012 (infra README) in parallel

**Within Foundational (Phase 2)**:
- T014 (Dockerfile), T015 (gitignore), T016 (docker-compose), T020 (golangci), T021 (ESLint) in parallel
- T017 (backend CI), T018 (frontend CI), T019 (infra CI) in parallel

**Within US1 - Backend**:
- T024-T028 (all middleware) in parallel
- T030-T032 (AI client components) in parallel
- T038-T040 (model, repository, service) in parallel

**Within US2 - Frontend**:
- T048-T054 (all primitives) in parallel
- T056 (features placeholder), T057 (error boundary) in parallel

**Within US3 - Infrastructure**:
- All module main.tf files (T061, T069, T078, T085, T093, T099) in parallel
- All module variables.tf (T066, T075, T082, T090, T096, T100) in parallel
- All module outputs.tf (T067, T076, T083, T091, T097, T101) in parallel

**Within Phase 7 - Polish**:
- T120-T123 (documentation updates) in parallel
- T125-T127 (linting validation) in parallel
- T128-T130 (testing validation) in parallel

---

## Parallel Example: User Story 1 (Backend)

**Day 1 - Parallel Track**:
- Engineer A: T024-T028 (all middleware files) - 5 tasks
- Engineer B: T029 (database client) - 1 task
- Engineer C: T030-T032 (AI client components) - 3 tasks
- Engineer D: T033-T034 (error handler) - 2 tasks

**Day 2 - Sequential**:
- Any engineer: T035-T037 (HTTPServer integration) - 3 tasks

**Day 3 - Parallel Track**:
- Engineer A: T038 (model), T039 (repository) - 2 tasks
- Engineer B: T040 (service) - 1 task (depends on T039 for interface)
- Engineer C: T041 (handler) - 1 task (depends on T040 for service interface)

**Total**: 11 tasks, 3 days with 4 engineers vs 11 days with 1 engineer

---

## Parallel Example: User Story 2 (Frontend)

**Day 1 - Sequential Setup**:
- Engineer: T042-T047 (core infrastructure chain) - 6 tasks

**Day 2 - Parallel Track**:
- Engineer A: T048-T049 (Button component) - 2 tasks
- Engineer B: T050 (Input), T051 (Card) - 2 tasks
- Engineer C: T052 (LoadingSpinner), T053 (ErrorMessage), T054 (EmptyState) - 3 tasks

**Day 3 - Sequential**:
- Any engineer: T055-T060 (composites, routing, error boundary) - 6 tasks

**Total**: 13 tasks, 3 days with 3 engineers vs 13 days with 1 engineer

---

## Parallel Example: User Story 3 (Infrastructure)

**Week 1 - Parallel Module Development**:
- Engineer A: T061-T068 (VPC module complete) - 8 tasks
- Engineer B: T069-T077 (ECS module complete) - 9 tasks
- Engineer C: T078-T084 (RDS module complete) - 7 tasks
- Engineer D: T085-T092 (ALB module complete) - 8 tasks
- Engineer E: T093-T098 (CloudFront module complete) - 6 tasks
- Engineer F: T099-T101 (Secrets module complete) - 3 tasks

**Week 2 - Integration**:
- Any engineer: T102-T109 (root module, configs, CI/CD) - 8 tasks

**Total**: 49 tasks, 2 weeks with 6 engineers vs 10 weeks with 1 engineer

---

## Implementation Strategy

**Recommended MVP Approach**:

1. **Week 1**: Complete Setup (Phase 1) and Foundational (Phase 2) - all engineers - 23 tasks
2. **Week 2-3**: Parallel user story development:
   - Team A (2 engineers): US1 - Backend - 30 tasks
   - Team B (2 engineers): US2 - Frontend - 19 tasks  
   - Team C (2 engineers): US3 - Infrastructure - 49 tasks
3. **Week 4**: US4 - Integration (all engineers) - 10 tasks
4. **Week 5**: Polish (all engineers) - 12 tasks

**Total Effort**: ~5 weeks with 6 engineers vs ~20 weeks with 1 engineer

**MVP Milestone**: After Phase 6 (US1-US4 complete), the architecture is operational and validated. Phase 7 polish can be incremental.

**First Deliverable**: After US1 + US2 complete, local development environment is functional (Docker Compose backend + frontend dev server). Infrastructure (US3) can proceed in parallel for staging deployment.

---

## Completion Notes (all 131 tasks, Sprints 1–4)

All 131 tasks in this file are complete as of Sprint 4 (2026-07-13). This section summarizes what
shipped, in the same spirit as `docs/roadmap.md`'s per-sprint closure entries (see that file's
"Sprint Plan" section for full task-count/work-item detail) — a narrative summary, not a
task-by-task relisting.

**Phase 1-2 (Setup + Foundational, Sprint 1, 23 tasks)**: Repository directory structure for
`backend/`, `frontend/`, `e2e/`, and `infra/`; Go module and React/Vite project initialization;
Docker/`docker-compose.yml` foundation; CI workflow skeletons (`backend-ci.yml`,
`frontend-ci.yml`, `infra-plan.yml`); `golangci-lint`/ESLint+Prettier configuration; Terraform
backend/version constraints. Shipped across PRs #43–#50.

**Phase 3-4 (Backend + Frontend Architecture, Sprint 2, 37 tasks)**: Backend — Chi middleware
chain (`RequestID → Logger → Recovery → CORS → BodySize`), `pgxpool` database client, the
`AIClient` interface with a working `OllamaClient` for local dev/MVP testing (scope extended
beyond the original stub-only plan — Anthropic-backed client deferred to Phase 6), domain error
handling (`internal/errors/`), the HTTP server entry point, and the `internal/example/`
model→repository→service→handler reference pattern. Frontend — design tokens, the API/query
client, the Zustand auth store, `Button`/`Input`/`Card`/`LoadingSpinner`/`ErrorMessage`/
`EmptyState` primitives, the `Form` composite, `ErrorBoundary`, and React Router v7 wiring.
Shipped across PRs #68, #71, #72, #74, #77–#82.

**Phase 5 (Infrastructure as Code, Sprint 3, 49 tasks)**: All six Terraform modules (VPC, ECS,
RDS, ALB, CloudFront, Secrets), root module wiring with staging/production `.tfvars`, real
`terraform init`/`validate`/`fmt -check` CI (AWS-gated `plan`/`apply` steps), and documented
(not-yet-provisioned) OIDC federation. The accessibility CI gate (`.github/workflows/accessibility.yml`)
and the Terraform remote-state bootstrap script were pulled forward from later specs into this
sprint. Shipped across PRs #95, #98–#100, #102–#107. Per this repo's standing AWS-cost-avoidance
policy, no `terraform apply` has run against real AWS — every module validates and formats
cleanly, but staging/production remain unprovisioned.

**Phase 6 (Integration and Error Handling, Sprint 4, 10 tasks)**: Correlation ID propagation
through `internal/errors/handler.go`, 5xx stack-trace logging, the `AnthropicClient`
implementation (`internal/ai/anthropic.go`, SDK-native retry/backoff) completing the AI client
foundation for staging/production, frontend `X-Request-ID` propagation and the `useErrorHandler`
hook, correlation ID display in `ErrorMessage`, and backend integration tests for `/healthz` and a
DB-timeout scenario. Shipped across PRs #128, #129, #131, #132 (issues #117–#120).

**Phase 7 (Polish & Cross-Cutting Concerns, Sprint 4, 12 tasks)**: A full validation sweep —
quickstart scenarios re-run against the current codebase, `golangci-lint`/ESLint+Prettier/
`terraform fmt` all clean, backend coverage (89.1% full-suite for `internal/example`), frontend
primitive tests (136/136 passing), and an ad-hoc axe-core accessibility scan (0 violations) —
recorded in `specs/005-system-architecture/validation-results.md` (issue #122, PR #133). This same
session brought `docs/architecture.md`, `docs/coding-guidelines.md`, `docs/testing-guidelines.md`,
and the root `README.md` up to date with what Sprints 1-4 actually shipped (issue #121, PR #134),
and marked
every task in this file complete (issue #123, this section).

**Net result**: the system architecture defined by this spec — backend service layering and
middleware, frontend application structure and state management, infrastructure-as-code modules,
and cross-layer integration/error-correlation patterns — is fully scaffolded, documented, and
validated. Real domain functionality (auth, trips, itinerary generation, collaboration) and actual
AWS provisioning are deliberately out of this spec's scope and land in later sprints per
`docs/roadmap.md`.
