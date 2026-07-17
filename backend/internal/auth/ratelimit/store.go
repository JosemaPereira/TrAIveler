// Package ratelimit implements progressive-delay, account-level rate limiting
// for login brute-force protection. Failed-attempt state is tracked in memory
// per process: it does not survive a restart or scale across multiple ECS
// tasks, an accepted MVP limitation (single-task assumption). The auth login
// service is its intended consumer — it records failures, checks the enforced
// delay, and resets the counter on a successful login.
package ratelimit

import (
	"sync"
	"time"
)

// attemptRecord is one account's failed-attempt tally for the current window.
type attemptRecord struct {
	count          int
	firstAttemptAt time.Time
}

// store is a concurrency-safe, in-memory map of account key -> failed-attempt
// record with a fixed time window. A record whose window has elapsed is treated
// as absent and reset on the next failure, giving each account a rolling window
// without a background sweeper. Distinct keys that never return are the only
// unbounded-growth vector, acceptable at MVP scale (see package doc). now is
// injectable so tests can advance the clock deterministically.
type store struct {
	mu      sync.Mutex
	window  time.Duration
	now     func() time.Time
	records map[string]attemptRecord
}

// newStore builds an empty store with the given window and the real clock.
func newStore(window time.Duration) *store {
	return &store{
		window:  window,
		now:     time.Now,
		records: make(map[string]attemptRecord),
	}
}

// attempts returns the failed-attempt count live for key within the current
// window, or 0 when no live record exists.
func (s *store) attempts(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.records[key]
	if !ok || s.now().Sub(rec.firstAttemptAt) > s.window {
		return 0
	}
	return rec.count
}

// recordFailure increments key's failed-attempt count, opening a fresh window
// when none is live.
func (s *store) recordFailure(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	rec, ok := s.records[key]
	if !ok || now.Sub(rec.firstAttemptAt) > s.window {
		s.records[key] = attemptRecord{count: 1, firstAttemptAt: now}
		return
	}
	rec.count++
	s.records[key] = rec
}

// reset clears key's record, e.g. after a successful login.
func (s *store) reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.records, key)
}
