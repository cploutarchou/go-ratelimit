### Package ratelimiter provides a rate limiter middleware for HTTP servers using Redis as the backend for storing request rate information.
_________
#### Usage
##### Importing
```go
import (
	"github.com/redis/go-redis/v9"
	"github.com/cploutarchou/go-ratelimit"
)
```
##### Creating a Limiter
```go
client := redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

limiter := ratelimiter.NewRateLimiter(client, 10, time.Minute)
```

In the example above, a new Redis client is created using the `go-redis/redis `library and passed to the `NewRateLimiter` function, along with a maximum request rate of 10 requests per minute.

##### Limiting Requests
```go
testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "test response")
})

limitedHandler := limiter.Limit(testHandler)

http.ListenAndServe(":8080", limitedHandler)
```
The `Limit` method of the `RateLimiter` struct is used to create a new HTTP handler that limits the rate of incoming requests. In the example above, the testHandler is wrapped in the `limitedHandler`, which limits the rate of incoming requests to 10 requests per minute. The resulting handler can be passed to `http.ListenAndServe` to start serving HTTP traffic.

##### Customizing Rate Limiting
The rate at which requests are limited can be customized by passing a different interval to the `NewRateLimiter` function. For example, to limit requests to `10` per second, use:

```go
limiter := ratelimiter.NewRateLimiter(client, 10, time.Second)
```
##### Customizing Response Headers
The rate limiter middleware adds three response headers to each response: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset`.

The `X-RateLimit-Limit` header specifies the maximum number of requests that can be made in the given interval.

The `X-RateLimit-Remaining` header specifies the number of requests remaining in the current interval.

The `X-RateLimit-Reset` header specifies the time (in seconds) until the current interval resets.

These headers can be customized by modifying the `http.ResponseWriter` directly in the handler.

##### Example
```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/cploutarchou/go-ratelimit"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	limiter := ratelimiter.NewRateLimiter(client, 10, time.Minute)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "test response")
	})

	limitedHandler := limiter.Limit(testHandler)

	http.ListenAndServe(":8080", limitedHandler)
}
```
##### Testing
The package includes a set of unit tests that can be run using the go test command:

```sh
$ go test .
```
The tests require a running instance of Redis to be available on the default port (6379). If Redis is not available, the tests will fail. A Docker container is used to start a Redis instance during testing. If Docker is not installed, the tests will also fail.