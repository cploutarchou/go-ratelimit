package ratelimiter

import (
	"context"
	"fmt"
	"github.com/ory/dockertest"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Limit(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to Docker: %s", err)
	}
	
	resource, err := pool.Run("redis", "6", nil)
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}
	defer pool.Purge(resource)
	
	var client *redis.Client
	
	if err := pool.Retry(func() error {
		client = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp")),
		})
		return client.Ping(context.Background()).Err()
	}); err != nil {
		t.Fatalf("Could not connect to Docker: %s", err)
	}
	defer client.Close()
	
	rl := NewRateLimiter(client, 2, time.Minute)
	
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "test response")
	})
	
	limitedHandler := rl.Limit(testHandler)
	
	rr := httptest.NewRecorder()
	
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// First two requests should be allowed
	limitedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "test response\n", rr.Body.String())
	assert.Equal(t, "2", rr.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "1", rr.Header().Get("X-RateLimit-Remaining"))
	
	rr = httptest.NewRecorder()
	limitedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "test response\n", rr.Body.String())
	assert.Equal(t, "2", rr.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", rr.Header().Get("X-RateLimit-Remaining"))
	
	// Third request should be rejected
	rr = httptest.NewRecorder()
	limitedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Equal(t, "Too many requests\n", rr.Body.String())
	assert.Equal(t, "2", rr.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", rr.Header().Get("X-RateLimit-Remaining"))
	
}
