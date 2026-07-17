package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// rateLimiter is a fixed-window, per-client-IP request counter shared by all
// requests flowing through one RateLimit middleware instance. State is
// in-memory and per process — it does not survive a restart or coordinate
// across ECS tasks (an accepted MVP limitation). Concurrency-safe; now is
// injectable so tests can drive the window without sleeping.
type rateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	now      func() time.Time
	counters map[string]*ipWindow
}

// ipWindow is one client IP's request count within its current window.
type ipWindow struct {
	count int
	start time.Time
}

// RateLimit returns middleware that allows at most limit requests per client IP
// within each fixed window. Every response carries X-RateLimit-Limit,
// X-RateLimit-Remaining, and X-RateLimit-Reset; once the budget is exhausted it
// returns 429 with Retry-After and a rate_limit_exceeded envelope instead of
// calling next.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return newRateLimiter(limit, window, time.Now).middleware
}

func newRateLimiter(limit int, window time.Duration, now func() time.Time) *rateLimiter {
	return &rateLimiter{
		limit:    limit,
		window:   window,
		now:      now,
		counters: make(map[string]*ipWindow),
	}
}

func (rl *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed, remaining, reset := rl.take(clientIP(r))

		h := w.Header()
		h.Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		h.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		h.Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))

		if !allowed {
			h.Set("Retry-After", strconv.Itoa(rl.retryAfterSeconds(reset)))
			writeErrorEnvelope(w, r, http.StatusTooManyRequests, "rate_limit_exceeded", "Too many requests")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// take registers one request from key and reports whether it is within budget,
// the remaining allowance for the window, and when the window resets.
func (rl *rateLimiter) take(key string) (allowed bool, remaining int, reset time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	win, ok := rl.counters[key]
	if !ok || now.Sub(win.start) >= rl.window {
		win = &ipWindow{start: now}
		rl.counters[key] = win
	}
	win.count++

	remaining = rl.limit - win.count
	if remaining < 0 {
		remaining = 0
	}
	return win.count <= rl.limit, remaining, win.start.Add(rl.window)
}

// retryAfterSeconds is the ceiling of the time until reset, floored at 1 so a
// blocked client is never told to retry in 0 seconds.
func (rl *rateLimiter) retryAfterSeconds(reset time.Time) int {
	seconds := int(math.Ceil(reset.Sub(rl.now()).Seconds()))
	if seconds < 1 {
		return 1
	}
	return seconds
}

// clientIP is the rate-limiting key: the host portion of RemoteAddr. Behind the
// ALB this is the load balancer's address, so per-IP isolation of real clients
// depends on a future trusted X-Forwarded-For hop — deliberately not trusted
// here, since an unvalidated XFF header is client-spoofable and would let a
// caller evade its own limit.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
