package ratelimit

import (
	"sync"
	"time"
)

type bucket struct {
	capacity int
	tokens   float64
	last     time.Time
}

type Limiter struct {
	rate     float64
	capacity int
	buckets  map[string]*bucket
	mu       sync.Mutex
}

func NewLimiter(rate float64, capacity int) *Limiter {
	return &Limiter{
		rate:     rate,
		capacity: capacity,
		buckets:  make(map[string]*bucket),
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	b, exists := l.buckets[key]
	if !exists {
		b = &bucket{
			capacity: l.capacity,
			tokens:   float64(l.capacity),
			last:     now,
		}
		l.buckets[key] = b
	}

	// Add leaked tokens
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * l.rate

	if b.tokens > float64(b.capacity) {
		b.tokens = float64(b.capacity)
	}

	b.last = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}
