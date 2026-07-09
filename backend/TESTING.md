# Testing Strategy

## Quick Reference

```bash
make test              # ✅ DEFAULT - Unit tests (no Docker needed)
make test-all          # ⚠️  Requires Colima running
make test-coverage     # CI configuration with coverage
```

## Test Commands Explained

### `make test` (Recommended for Local Development)

**What it does:** Runs unit tests only, skips testcontainers  
**Requirements:** None (no Docker/Colima needed)  
**Used by:** Local development (default), CI pipeline  
**Command:** `go test -tags=test -v ./... -short`

**Why `-short`?**  
Tests with `testing.Short()` check are skipped. This includes:
- Testcontainer tests (need Docker API)
- Long-running integration tests

**Always works locally** ✅

```bash
make test
# Output: 3 core unit tests pass, 7 testcontainer tests skipped
```

---

### `make test-all` (Full Test Suite)

**What it does:** Runs ALL tests including testcontainers  
**Requirements:** ⚠️ **Colima must be running**  
**Used by:** Comprehensive local testing before PR  
**Command:** `go test -tags=test -v ./...`

**Setup required:**
```bash
# Check if Colima is running
colima status

# Start Colima if needed
colima start --cpu 2 --memory 4

# Now run full test suite
make test-all
```

**Crashes locally if Colima not running** ❌

```bash
# Without Colima:
make test-all
# Error: Cannot connect to Docker daemon
```

---

### `make test-coverage` (CI Configuration)

**What it does:** Unit tests + race detector + coverage report  
**Requirements:** None (same as `make test`)  
**Used by:** CI pipeline, pre-commit coverage checks  
**Command:** `go test -tags=test -v ./... -short -race -coverprofile=coverage.out -covermode=atomic`

**Generates:** `coverage.html` with visual coverage report

```bash
make test-coverage
# Opens coverage.html in browser
open coverage.html
```

---

## CI Pipeline Configuration

**File:** `.github/workflows/backend-ci.yml`

**Test strategy:**
1. **Unit tests** with `-short` flag (skips testcontainers)
2. **Race detector** enabled (`-race`)
3. **Coverage reporting** to artifacts
4. **No Colima needed** (testcontainers skipped)

**Why this approach?**
- ✅ Fast CI runs (no container overhead)
- ✅ Consistent between local and CI
- ✅ GitHub Actions has PostgreSQL service for integration tests (future)
- ✅ No Docker-in-Docker complexity

---

## Test Types in This Project

### 1. Unit Tests (Always Run)

**Location:** `*_test.go` files  
**No special tags:** Included in all test runs  
**Examples:**
- `TestNewClient_InvalidURL` - Error handling
- `TestNewClient_ContextCancellation` - Context behavior
- `TestNewClient_RetryLogic` - Retry logic

**These tests:**
- ✅ Run without database
- ✅ Run without Docker/Colima
- ✅ Fast (< 10 seconds)
- ✅ Always run in CI

### 2. Testcontainer Tests (Skipped by Default)

**Location:** `client_test.go`  
**Skip condition:** `if testing.Short() { t.Skip() }`  
**Examples:**
- `TestNewClient_Success` - Real PostgreSQL container
- `TestPing_Success` - Real connection testing
- `TestPool_ConcurrentOperations` - Pool behavior

**These tests:**
- ❌ Need Colima/Docker running
- ❌ Slower (container startup overhead)
- ⚠️  Skipped in CI (by design)
- ✅ Run with `make test-all` only

---

## Why Different Approaches?

### Problem: Docker/Colima Dependency

**Testcontainers** require Docker API:
- Docker Desktop = paid license (not allowed)
- Colima = free, but must be running manually
- CI doesn't have Colima by default

**Solution:** Skip testcontainers locally and in CI

### Trade-offs

**Pros:**
- ✅ Fast feedback loop (unit tests < 10s)
- ✅ No manual Colima management
- ✅ CI works out of the box
- ✅ Consistent local/CI behavior

**Cons:**
- ⚠️  Testcontainer tests skipped by default
- ⚠️  Need manual `make test-all` for full coverage

---

## When to Use Each Command

| Scenario | Command | Reason |
|----------|---------|--------|
| **Quick feedback during TDD** | `make test` | Fast, no Docker needed |
| **Before committing code** | `make test` | Same as CI will run |
| **Before creating PR** | `make test-all` | Full coverage (if Colima available) |
| **Testing database code** | `make test-integration` | Real PostgreSQL behavior |
| **Checking coverage** | `make test-coverage` | Generate coverage.html |
| **CI pipeline** | `make test-coverage` | Unit tests + race + coverage |

---

## Troubleshooting

### "Cannot connect to Docker daemon"

**Cause:** Trying to run testcontainer tests without Colima  
**Sose `make test` instead (skips testcontainers)
2. Start Colima: `colima start --cpu 2 --memory 4`
3. Use `make test-all` after Colima is running

### "DATABASE_URL not set"

**Cause:** Running `make test-integration` without DATABASE_URL  
**Solution:**
```bash
docker-compose up -d postgres
export DATABASE_URL="postgresql://traveler_user:traveler_pass@localhost:5432/traveler_db?sslmode=disable"
make test-integration
```

### Tests pass locally but fail in CI

**Caummary

✅ **For daily development:** `make test` (fast, reliable, no Docker)  
✅ **For comprehensive testing:** `make test-all` (requires Colima)  
✅ **For CI:** `make test-coverage` (same as local default)  
✅ **All commands documented** in Makefile help: `make help`


---

## Historical Note: Why No Separate Integration Tests?

**Previously:** Had duplicate `client_integration_test.go` with `//go:build integration` tag  
**Problem:** 100% duplication with testcontainer tests, compilation bugs, never used  
**Solution:** Removed in favor of testcontainer tests (better isolation, automatic setup)  

**Testcontainers provide everything integration tests did:**
- Real PostgreSQL database testing
- Connection pool behavior validation
- Concurrent operations testing
- Query execution verification

**When we WOULD need separate integration tests:**
- Testing specific PostgreSQL extensions not in containers
- Performance benchmarks against real hardware
- Testing production-like configurations
- Currently: Not applicable