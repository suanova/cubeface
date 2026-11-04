package common

import (
	"sync"
	"time"
)

// RateLimiter is a small fixed-window in-memory limiter. It is intended for a
// single-process deployment; replace it with Redis if the API is scaled out.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
	limit   int
	window  time.Duration
}

type rateEntry struct {
	count   int
	resetAt time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit < 1 {
		limit = 1
	}
	return &RateLimiter{
		entries: make(map[string]*rateEntry),
		limit:   limit,
		window:  window,
	}
}

// Allow reports whether a request for key may proceed and, when denied, how
// long until the window resets.
func (r *RateLimiter) Allow(key string) (bool, time.Duration) {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	// Opportunistic cleanup keeps the map bounded without a goroutine.
	if len(r.entries) > 10_000 {
		for k, e := range r.entries {
			if now.After(e.resetAt) {
				delete(r.entries, k)
			}
		}
	}

	entry, ok := r.entries[key]
	if !ok || now.After(entry.resetAt) {
		r.entries[key] = &rateEntry{count: 1, resetAt: now.Add(r.window)}
		return true, 0
	}
	if entry.count >= r.limit {
		return false, time.Until(entry.resetAt)
	}
	entry.count++
	return true, 0
}
