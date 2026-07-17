package ratelimit

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testEmail = "user@example.com"

// newTestLimiter returns a Limiter whose clock is driven by the returned
// pointer, so tests can advance time deterministically instead of sleeping.
func newTestLimiter(t *testing.T) (*Limiter, *time.Time) {
	t.Helper()

	clock := time.Now()
	l := New()
	l.store.now = func() time.Time { return clock }
	return l, &clock
}

func TestLimiter_CheckRateLimit_BelowThreshold_NoDelay(t *testing.T) {
	l, _ := newTestLimiter(t)

	for i := 0; i < freeAttempts-1; i++ {
		l.RecordFailure(testEmail)
		assert.Zero(t, l.CheckRateLimit(testEmail),
			"when fewer than freeAttempts failures accrue it should not delay")
	}
}

func TestLimiter_CheckRateLimit_ProgressiveBackoffCurve(t *testing.T) {
	tests := []struct {
		failures  int
		wantDelay time.Duration
	}{
		{failures: 5, wantDelay: 1 * time.Second},
		{failures: 6, wantDelay: 2 * time.Second},
		{failures: 7, wantDelay: 4 * time.Second},
		{failures: 8, wantDelay: 8 * time.Second},
		{failures: 9, wantDelay: 16 * time.Second},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.failures)+"failures", func(t *testing.T) {
			l, _ := newTestLimiter(t)

			for i := 0; i < tt.failures; i++ {
				l.RecordFailure(testEmail)
			}

			assert.Equal(t, tt.wantDelay, l.CheckRateLimit(testEmail),
				"after N failures it should delay 2^(N-5) seconds")
		})
	}
}

func TestLimiter_CheckRateLimit_DelayClampedToWindow(t *testing.T) {
	l, _ := newTestLimiter(t)

	// Far past the point where 2^(n-5)s would exceed the 15-minute window.
	for i := 0; i < 30; i++ {
		l.RecordFailure(testEmail)
	}

	assert.Equal(t, l.window, l.CheckRateLimit(testEmail),
		"a runaway backoff should be clamped to the window, never overflow")
}

func TestLimiter_CheckRateLimit_WindowExpiry_ResetsCount(t *testing.T) {
	l, clock := newTestLimiter(t)

	for i := 0; i < freeAttempts; i++ {
		l.RecordFailure(testEmail)
	}
	assert.NotZero(t, l.CheckRateLimit(testEmail))

	*clock = clock.Add(defaultWindow + time.Second)

	assert.Zero(t, l.CheckRateLimit(testEmail),
		"failures older than the window should no longer count")
}

func TestLimiter_Reset_ClearsHistory(t *testing.T) {
	l, _ := newTestLimiter(t)

	for i := 0; i < freeAttempts; i++ {
		l.RecordFailure(testEmail)
	}
	assert.NotZero(t, l.CheckRateLimit(testEmail))

	l.Reset(testEmail)

	assert.Zero(t, l.CheckRateLimit(testEmail),
		"a successful-login reset should clear the account's delay")
}

func TestLimiter_ConcurrentAccess_IsRaceFree(t *testing.T) {
	l := New()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			l.RecordFailure(testEmail)
			_ = l.CheckRateLimit(testEmail)
		}()
	}
	wg.Wait()

	// All 50 failures land in the same live window.
	assert.Equal(t, goroutines, l.store.attempts(testEmail))
}
