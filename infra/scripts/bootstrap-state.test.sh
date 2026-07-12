#!/bin/bash
# Test suite for infra/scripts/bootstrap-state.sh (issue #96, task 003-T009)
#
# Purpose:
#   Verifies that bootstrap-state.sh idempotently provisions the two AWS resources
#   infra/backend.tf's S3 backend requires (state bucket + lock table), without ever
#   making a real AWS call. Mocks the `aws` CLI as a shell function that logs every
#   invocation to a temp file and returns configurable exit codes for the
#   "does it exist" checks (`s3api head-bucket` / `dynamodb describe-table`), then
#   sources bootstrap-state.sh (so its `main` guard does not auto-execute) and calls
#   `main` directly under two scenarios.
#
# Usage:
#   bash infra/scripts/bootstrap-state.test.sh
#
# Exit codes:
#   0 = all assertions passed
#   1 = one or more assertions failed (or the script under test is missing)
#
# Note: this script never invokes bootstrap-state.sh directly (only sources it) and
# never calls real AWS — the `aws` function below is a full mock for the duration of
# this test run.

set -uo pipefail # no -e: keep running so every assertion gets a chance to report

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_UNDER_TEST="$SCRIPT_DIR/bootstrap-state.sh"

FAILURES=0

pass() { echo "  PASS: $1"; }
fail() {
  echo "  FAIL: $1"
  FAILURES=$((FAILURES + 1))
}

# ---------------------------------------------------------------------------
# Mock aws CLI
# ---------------------------------------------------------------------------
# AWS_CALL_LOG collects one line per invocation (full argv), so assertions can
# inspect exactly what bootstrap-state.sh called without touching real AWS.
AWS_CALL_LOG="$(mktemp)"
trap 'rm -f "$AWS_CALL_LOG"' EXIT

# Configurable via env vars, read fresh on every call so each scenario can flip them:
#   MOCK_BUCKET_EXISTS=1 -> `s3api head-bucket` exits 0 (bucket exists)
#   MOCK_BUCKET_EXISTS=0 -> `s3api head-bucket` exits 1 (404, bucket missing)
#   MOCK_TABLE_EXISTS   -> same idea for `dynamodb describe-table`
aws() {
  echo "$*" >>"$AWS_CALL_LOG"

  if [[ "$1" == "s3api" && "$2" == "head-bucket" ]]; then
    [[ "${MOCK_BUCKET_EXISTS:-0}" == "1" ]] && return 0 || return 1
  fi

  if [[ "$1" == "dynamodb" && "$2" == "describe-table" ]]; then
    [[ "${MOCK_TABLE_EXISTS:-0}" == "1" ]] && return 0 || return 1
  fi

  # Any other aws subcommand (create-bucket, put-bucket-versioning, ...) is a no-op success.
  return 0
}
export -f aws

reset_log() { : >"$AWS_CALL_LOG"; }

# Number of logged lines matching a fixed-string pattern.
count_matching() {
  grep -c -F -- "$1" "$AWS_CALL_LOG" || true
}

assert_count() {
  local pattern="$1" expected="$2" desc="$3"
  local actual
  actual="$(count_matching "$pattern")"
  if [[ "$actual" == "$expected" ]]; then
    pass "$desc"
  else
    fail "$desc (expected $expected call(s), got $actual)"
  fi
}

# ---------------------------------------------------------------------------
# Load script under test (source, so `main` does not auto-execute)
# ---------------------------------------------------------------------------
echo "=== bootstrap-state.sh test suite ==="
echo

if [[ ! -f "$SCRIPT_UNDER_TEST" ]]; then
  fail "$SCRIPT_UNDER_TEST does not exist yet"
  echo
  echo "=== $FAILURES failure(s) ==="
  exit 1
fi

# shellcheck disable=SC1090
source "$SCRIPT_UNDER_TEST"

if ! declare -F main >/dev/null; then
  fail "bootstrap-state.sh does not define a main function"
  echo
  echo "=== $FAILURES failure(s) ==="
  exit 1
fi

# ---------------------------------------------------------------------------
# Assertion group 1: constants must match infra/backend.tf exactly
# ---------------------------------------------------------------------------
echo "-- Constants (must match infra/backend.tf) --"

if [[ "${STATE_BUCKET_NAME:-}" == "traveler-terraform-state" ]]; then
  pass "STATE_BUCKET_NAME is traveler-terraform-state"
else
  fail "STATE_BUCKET_NAME expected 'traveler-terraform-state', got '${STATE_BUCKET_NAME:-<unset>}'"
fi

if [[ "${LOCK_TABLE_NAME:-}" == "traveler-terraform-locks" ]]; then
  pass "LOCK_TABLE_NAME is traveler-terraform-locks"
else
  fail "LOCK_TABLE_NAME expected 'traveler-terraform-locks', got '${LOCK_TABLE_NAME:-<unset>}'"
fi

if [[ "${AWS_REGION_NAME:-}" == "us-east-1" ]]; then
  pass "AWS_REGION_NAME is us-east-1"
else
  fail "AWS_REGION_NAME expected 'us-east-1', got '${AWS_REGION_NAME:-<unset>}'"
fi

# ---------------------------------------------------------------------------
# Scenario A: both resources already exist -> main must be a no-op (idempotency)
# ---------------------------------------------------------------------------
echo
echo "-- Scenario A: both resources already exist (idempotency) --"

export MOCK_BUCKET_EXISTS=1
export MOCK_TABLE_EXISTS=1
reset_log

main

assert_count "create-bucket" "0" "main does not call create-bucket when the bucket already exists"
assert_count "create-table" "0" "main does not call create-table when the table already exists"

# ---------------------------------------------------------------------------
# Scenario B: neither resource exists -> main must create both, exactly once each
# ---------------------------------------------------------------------------
echo
echo "-- Scenario B: neither resource exists (fresh provisioning) --"

export MOCK_BUCKET_EXISTS=0
export MOCK_TABLE_EXISTS=0
reset_log

main

assert_count "create-bucket" "1" "main calls create-bucket exactly once"
assert_count "put-bucket-versioning" "1" "main calls put-bucket-versioning exactly once"

aes256_calls="$(grep -c -F -- "AES256" "$AWS_CALL_LOG" || true)"
if [[ "$aes256_calls" == "1" ]]; then
  pass "main makes exactly one encryption call containing AES256"
else
  fail "main's AES256 encryption call count expected 1, got $aes256_calls"
fi

assert_count "put-public-access-block" "1" "main calls put-public-access-block exactly once"
assert_count "create-table" "1" "main calls create-table exactly once"

create_table_line="$(grep -F -- "create-table" "$AWS_CALL_LOG" || true)"
if [[ "$create_table_line" == *"LockID,AttributeType=S"* ]]; then
  pass "create-table call includes LockID,AttributeType=S"
else
  fail "create-table call missing LockID,AttributeType=S (got: $create_table_line)"
fi

if [[ "$create_table_line" == *"PAY_PER_REQUEST"* ]]; then
  pass "create-table call includes PAY_PER_REQUEST"
else
  fail "create-table call missing PAY_PER_REQUEST (got: $create_table_line)"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo
if [[ "$FAILURES" -eq 0 ]]; then
  echo "=== All checks passed ==="
  exit 0
else
  echo "=== $FAILURES failure(s) ==="
  exit 1
fi
