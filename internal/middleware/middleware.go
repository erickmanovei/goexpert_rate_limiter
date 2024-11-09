package middleware

import (
	"context"
	"net/http"

	"github.com/erickmanovei/goexpert_rate_limiter/internal/ratelimiter"
)

func RateLimiterMiddleware(rl *ratelimiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()
			ip := r.RemoteAddr
			token := r.Header.Get("API_KEY")

			rateLimit := rl.IpRateLimit
			key := "ip:" + ip
			if token != "" {
				rateLimit = rl.TokenRateLimit
				key = "token:" + token
			}

			rateLimited, err := rl.IsRateLimited(ctx, key, rateLimit)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if rateLimited {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
