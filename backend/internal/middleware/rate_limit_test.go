package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// okHandler is a minimal next handler that always writes 200.
func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// doRequest sends one request from a fixed client IP through handler.
func doRequest(handler http.Handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/things", nil)
	req.RemoteAddr = "203.0.113.7:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestRateLimit_WithinBudget_AllowsAndSetsHeaders(t *testing.T) {
	handler := RateLimit(3, time.Minute)(okHandler())

	rec := doRequest(handler)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "3", rec.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "2", rec.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimit_ExceedingBudget_Returns429WithRetryAfter(t *testing.T) {
	handler := RateLimit(2, time.Minute)(okHandler())

	assert.Equal(t, http.StatusOK, doRequest(handler).Code)
	assert.Equal(t, http.StatusOK, doRequest(handler).Code)

	rec := doRequest(handler)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "0", rec.Header().Get("X-RateLimit-Remaining"))
	assert.Contains(t, rec.Body.String(), "rate_limit_exceeded")

	retryAfter, err := strconv.Atoi(rec.Header().Get("Retry-After"))
	require.NoError(t, err)
	assert.Positive(t, retryAfter, "a blocked client must be told to retry in >=1 second")
}

func TestRateLimit_SeparateIPs_TrackedIndependently(t *testing.T) {
	handler := RateLimit(1, time.Minute)(okHandler())

	first := httptest.NewRequest(http.MethodGet, "/", nil)
	first.RemoteAddr = "198.51.100.1:1000"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, first)

	second := httptest.NewRequest(http.MethodGet, "/", nil)
	second.RemoteAddr = "198.51.100.2:2000"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, second)

	assert.Equal(t, http.StatusOK, rec1.Code)
	assert.Equal(t, http.StatusOK, rec2.Code,
		"a different IP has its own budget and must not be blocked")
}

func TestRateLimit_WindowElapses_BudgetResets(t *testing.T) {
	clock := time.Now()
	rl := newRateLimiter(1, time.Minute, func() time.Time { return clock })
	handler := rl.middleware(okHandler())

	assert.Equal(t, http.StatusOK, doRequest(handler).Code)
	assert.Equal(t, http.StatusTooManyRequests, doRequest(handler).Code)

	clock = clock.Add(time.Minute + time.Second)

	assert.Equal(t, http.StatusOK, doRequest(handler).Code,
		"once the window elapses the client's budget should reset")
}

func TestRateLimit_ConcurrentRequests_IsRaceFree(t *testing.T) {
	handler := RateLimit(1000, time.Minute)(okHandler())

	const goroutines = 50
	var okCount atomic.Int64
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if doRequest(handler).Code == http.StatusOK {
				okCount.Add(1)
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, int64(goroutines), okCount.Load(),
		"within a generous budget every concurrent request should be allowed")
}
