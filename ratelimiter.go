package ratelimiter

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"
	
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	limiter  *redis_rate.Limiter
	next     http.Handler
	maxRate  int
	interval time.Duration
	mu       sync.Mutex
}

func NewRateLimiter(client *redis.Client, maxRate int, interval time.Duration) *RateLimiter {
	limiter := redis_rate.NewLimiter(client)
	return &RateLimiter{
		limiter:  limiter,
		maxRate:  maxRate,
		interval: interval,
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	rl.next = next
	var err error
	
	var result *redis_rate.Result
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl.mu.Lock()
		defer rl.mu.Unlock()
		clientIP := r.RemoteAddr // change this to your actual client IP extraction method
		switch rl.interval {
		case time.Second:
			result, err = rl.limiter.Allow(context.Background(), clientIP, redis_rate.PerSecond(rl.maxRate))
		case time.Minute:
			result, err = rl.limiter.Allow(context.Background(), clientIP, redis_rate.PerMinute(rl.maxRate))
		case time.Hour:
			result, err = rl.limiter.Allow(context.Background(), clientIP, redis_rate.PerHour(rl.maxRate))
		default:
			result, err = rl.limiter.Allow(context.Background(), clientIP, redis_rate.PerMinute(rl.maxRate))
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		if result.Allowed == 0 && result.Remaining == 0 {
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.maxRate))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(int(result.Remaining)))
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(result.ResetAfter/time.Second)))
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.maxRate))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(int(result.Remaining)))
		w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(result.ResetAfter/time.Second)))
		rl.next.ServeHTTP(w, r)
	})
}
