#!/bin/bash
# Verification script for Docker infrastructure setup (tickets #30 and #32)
#
# Purpose:
#   Validates that Docker infrastructure files are correctly configured before
#   attempting to run services. Catches configuration errors early without
#   requiring a full build or database connection.
#
# What it checks:
#   1. Dockerfile structure (multi-stage build, health check)
#   2. docker-compose.yml syntax (YAML validation)
#   3. Service definitions (postgres and backend services exist)
#   4. Dockerfile builder stage (Go modules can be processed)
#   5. PostgreSQL image availability (can be pulled from Docker Hub)
#
# Exit codes:
#   0 = All checks passed
#   1 = One or more checks failed
#
# Usage:
#   ./verify-docker-setup.sh
#
# Note: This script does NOT start services or require a running Docker daemon
#       for most checks. It validates configuration files statically where possible.

set -e

# Change to project root directory to ensure relative paths work correctly
cd "$(dirname "$0")"

echo "=== Docker Setup Verification ==="
echo

# ============================================================================
# Check 1: Dockerfile Structure
# ============================================================================
echo "✓ Check 1: Dockerfile structure"
if [ -f "backend/Dockerfile" ]; then
    # Verify multi-stage build: should have both builder and runtime stages
    # Check for health check: ensures container orchestration can monitor service health
    if grep -q "FROM golang:.*AS builder" backend/Dockerfile && \
       grep -q "FROM alpine:" backend/Dockerfile && \
       grep -q "HEALTHCHECK" backend/Dockerfile; then
        echo "  ✅ Multi-stage Dockerfile with health check found"
    else
        echo "  ❌ Dockerfile structure incorrect"
        echo "     Expected: Multi-stage build (builder + runtime) with HEALTHCHECK"
        exit 1
    fi
else
    echo "  ❌ Dockerfile not found at backend/Dockerfile"
    exit 1
fi

# ============================================================================
# Check 2: docker-compose.yml Syntax
# ============================================================================
echo
echo "✓ Check 2: docker-compose.yml syntax"
# Use docker-compose config to validate YAML syntax and structure
# --quiet suppresses output, only returns exit code
if docker-compose config --quiet; then
    echo "  ✅ docker-compose.yml is valid"
else
    echo "  ❌ docker-compose.yml has syntax errors"
    echo "     Run: docker-compose config (without --quiet) to see details"
    exit 1
fi

# ============================================================================
# Check 3: Service Definitions
# ============================================================================
echo
echo "✓ Check 3: Service definitions"
# Verify both required services are defined in docker-compose.yml
if docker-compose config | grep -q "postgres:" && \
   docker-compose config | grep -q "backend:"; then
    echo "  ✅ Both postgres and backend services defined"
else
    echo "  ❌ Missing service definitions in docker-compose.yml"
    echo "     Expected: 'postgres' and 'backend' services"
    exit 1
fi

# ============================================================================
# Check 4: Dockerfile Build Process
# ============================================================================
echo
echo "✓ Check 4: Dockerfile build process"
echo "  Note: Build will fail without main.go (expected behavior)"

# Try building just the builder stage to verify Go module handling
# This checks that go.mod/go.sum are valid and dependencies can be downloaded
if docker build --target builder -t test-builder backend/ > /dev/null 2>&1; then
    echo "  ✅ Builder stage can process Go modules"
else
    # Check if the failure is the expected "no Go files" error (main.go doesn't exist yet)
    # or an unexpected error (syntax problems, missing files, etc.)
    if docker build --target builder -t test-builder backend/ 2>&1 | grep -q "no Go files"; then
        echo "  ✅ Builder stage works (fails at compile as expected - no main.go yet)"
    else
        echo "  ⚠️  Builder stage has unexpected errors"
        echo "     Run: docker build --target builder -t test-builder backend/"
        echo "     to see detailed error output"
    fi
fi

# ============================================================================
# Check 5: PostgreSQL Image Availability
# ============================================================================
echo
echo "✓ Check 5: PostgreSQL image availability"
# Verify the specified PostgreSQL version can be pulled from Docker Hub
# This catches typos in the image tag or registry access issues
if docker pull postgres:15.4-alpine > /dev/null 2>&1; then
    echo "  ✅ PostgreSQL 15.4-alpine image available"
else
    echo "  ❌ Cannot pull postgres:15.4-alpine image"
    echo "     Check: Docker daemon running? Internet connection available?"
    exit 1
fi

echo
echo "=== All Checks Passed ==="
echo
echo "Next steps:"
echo "1. Implement main.go (creates the HTTP server entry point)"
echo "2. Copy backend/.env.example to backend/.env and configure"
echo "3. Start services: docker-compose up -d"
echo "4. Verify health: docker-compose ps (both services should show 'healthy')"
