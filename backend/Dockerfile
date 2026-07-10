# Multi-stage Docker build for TrAIveler backend API
#
# This Dockerfile creates an optimized production image for deployment to AWS ECS Fargate.
# It uses a multi-stage build to separate the build environment from the runtime environment,
# resulting in a minimal final image (~50MB vs ~1.2GB if we kept the Go toolchain).
#
# Stage 1 (builder): Compiles the Go binary with all dependencies
# Stage 2 (runtime):  Runs the binary in a minimal Alpine Linux container
#
# Security features:
# - Static binary (CGO_ENABLED=0) with no external C dependencies
# - Non-root user (appuser:1000) for container execution
# - Minimal attack surface (Alpine base + ca-certificates only)
# - Automated health checks for orchestration reliability

# ============================================================================
# Stage 1: Builder
# ============================================================================
# --platform=$BUILDPLATFORM pins this stage to the CI runner's native
# architecture (amd64), even though the final image targets linux/arm64
# (see the runtime stage below). Without this, buildx runs the *entire*
# builder stage — including `go build` — under QEMU emulation for arm64,
# which measured ~7.5 minutes for this project's `go build` step alone vs.
# well under a minute natively. Go's own toolchain cross-compiles for arm64
# natively (no CGO, no emulation needed) via GOARCH below, so pinning the
# builder to the build machine's real architecture is strictly faster with
# no downside.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

# Populated automatically by buildx from the --platform flag passed to
# `docker buildx build` (e.g. linux/arm64 -> TARGETOS=linux, TARGETARCH=arm64).
ARG TARGETOS
ARG TARGETARCH

# Install build-time dependencies required for go build and git operations
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory for the build
WORKDIR /build

# Copy Go module files first to leverage Docker layer caching
# These files change less frequently than source code, so Docker can reuse
# this layer when only application code changes
COPY go.mod go.sum ./

# Download dependencies - this layer is cached unless go.mod/go.sum change
RUN go mod download && go mod verify

# Copy all source code into the builder
COPY . .

# Build the binary with optimization flags:
# - CGO_ENABLED=0: Build a static binary with no C dependencies (portable across Linux distros)
# - GOOS/GOARCH=$TARGETOS/$TARGETARCH: cross-compile for the requested target
#   platform (e.g. linux/arm64 for ECS Fargate Graviton2) while this stage
#   itself runs natively — see the --platform=$BUILDPLATFORM note above
# - -ldflags="-w -s": Strip debug information and symbol table (reduces binary size ~30%)
# - -o api: Output binary named "api"
# - ./cmd/api: Entry point package to compile
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -ldflags="-w -s" \
    -o api \
    ./cmd/api

# ============================================================================
# Stage 2: Runtime
# ============================================================================
# 3.19 (released Nov 2023) is past its ~2-year support window and no longer
# receives security backports, which is what the "2 high vulnerabilities"
# scan finding was flagging. 3.23 is the current maintained branch.
FROM alpine:3.23

# Install runtime dependencies:
# - ca-certificates: Required for HTTPS connections to external APIs (Anthropic)
# - tzdata: Timezone data for accurate timestamp logging
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user for running the application
# Security best practice: containers should not run as root
# UID/GID 1000 is standard for application users
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory for the application
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /build/api /app/api

# Copy database migrations directory (required for startup migration execution)
# The API server may run migrations automatically on startup
COPY --from=builder /build/migrations /app/migrations

# Change ownership of application files to the non-root user
RUN chown -R appuser:appuser /app

# Switch from root to the non-root user
# All subsequent commands and the CMD will run as this user
USER appuser

# Expose the HTTP server port
# This is documentation only - actual port binding happens at runtime via -p flag
EXPOSE 8080

# Configure container health check for orchestration platforms (ECS, Kubernetes)
# Health checks determine if the container is ready to receive traffic and when to restart
# Parameters:
#   --interval=30s: Check health every 30 seconds after the container starts
#   --timeout=10s: Wait up to 10 seconds for the health check to complete
#   --start-period=30s: Grace period before the first health check (allows app to initialize)
#   --retries=3: Mark container unhealthy after 3 consecutive failed checks
#
# The health check calls the /healthz endpoint which verifies database connectivity
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

# Run the binary
# Note: This is the default command. Can be overridden at runtime with docker run
CMD ["/app/api"]
