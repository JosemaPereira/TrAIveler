package ratelimit

import "time"

const (
	// defaultWindow is the rolling window over which failed attempts accrue,
	// per specs/008 data-model.md ("within 15 minutes").
	defaultWindow = 15 * time.Minute
	// freeAttempts is the number of failures allowed before a progressive
	// delay applies; the freeAttempts-th failure is the first to incur one.
	freeAttempts = 5
	// maxExponent caps the backoff exponent so the 1<<n shift stays well in
	// range; the resulting delay is clamped to the window regardless.
	maxExponent = 20
)

// Limiter applies progressive-delay rate limiting keyed by account identifier
// (email). Once freeAttempts failures accrue within the window, each recorded
// failure doubles the enforced delay: 2^(attempts-freeAttempts) seconds
// (1s, 2s, 4s, 8s, 16s...), clamped to the window. It is concurrency-safe.
type Limiter struct {
	store  *store
	window time.Duration
}

// New builds a Limiter with the default 15-minute window.
func New() *Limiter {
	return &Limiter{store: newStore(defaultWindow), window: defaultWindow}
}

// CheckRateLimit returns the delay the caller must enforce before honoring
// another attempt for email, or 0 when the account is under the threshold. It
// does not record the attempt — call RecordFailure for that.
func (l *Limiter) CheckRateLimit(email string) time.Duration {
	attempts := l.store.attempts(email)
	if attempts < freeAttempts {
		return 0
	}
	return l.delayFor(attempts)
}

// RecordFailure registers one failed attempt for email.
func (l *Limiter) RecordFailure(email string) {
	l.store.recordFailure(email)
}

// Reset clears email's failed-attempt history, e.g. after a successful login.
func (l *Limiter) Reset(email string) {
	l.store.reset(email)
}

// delayFor computes 2^(attempts-freeAttempts) seconds, clamped to the window.
// CheckRateLimit only calls this once attempts >= freeAttempts, so exp is
// always >= 0; maxExponent bounds the shift so it can never overflow int64.
func (l *Limiter) delayFor(attempts int) time.Duration {
	exp := attempts - freeAttempts
	if exp > maxExponent {
		exp = maxExponent
	}
	delay := time.Duration(int64(1)<<exp) * time.Second
	if delay > l.window {
		return l.window
	}
	return delay
}
