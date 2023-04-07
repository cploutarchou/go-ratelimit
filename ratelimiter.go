package ratelimiter

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
	
	"github.com/redis/go-redis/v9"
	
	"github.com/go-redis/redis_rate/v10"
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
		
		fmt.Println("result.Allowed :", result.Allowed)
		fmt.Println("rl.maxRate :", rl.maxRate)
		fmt.Println(result.Allowed >= rl.maxRate)
		if result.Allowed >= rl.maxRate {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		log.Println("X-Rate-Limit-Limit:", r.Header.Get("X-Rate-Limit-Limit"))
		log.Println("X-Rate-Limit-Remaining:", r.Header.Get("X-Rate-Limit-Remaining"))
		log.Println("X-Rate-Limit-Reset:", r.Header.Get("X-Rate-Limit-Reset"))
		w.Header().Set("X-Rate-Limit-Limit", strconv.Itoa(rl.maxRate))
		w.Header().Set("X-Rate-Limit-Remaining", strconv.Itoa(int(rl.maxRate-result.Remaining)))
		w.Header().Set("X-Rate-Limit-Reset", strconv.Itoa(int(result.ResetAfter/time.Second)))
		log.Println("X-Rate-Limit-Limit:", r.Header.Get("X-Rate-Limit-Limit"))
		log.Println("X-Rate-Limit-Remaining:", r.Header.Get("X-Rate-Limit-Remaining"))
		log.Println("X-Rate-Limit-Reset:", r.Header.Get("X-Rate-Limit-Reset"))
		
		rl.next.ServeHTTP(w, r)
	})
}
