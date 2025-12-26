package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

func RateLimitMiddleware(
	reqPerSec float64,
	burst int,
) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(reqPerSec), burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
