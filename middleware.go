package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"
)


// Middleware wraps an http.HandlerFunc and returns a new one.
// This lets us compose behaviors cleanly without nesting callbacks.
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Chain applies a list of middlewares to a handler.
// Declaration order = execution order (first listed = outermost wrapper).
//
// Example:
//
//	Chain(myHandler, LoggingMiddleware, AuthMiddleware, MethodMiddleware("POST"))
//	// executes: Logging → Auth → Method check → myHandler
func Chain(h http.HandlerFunc, mws ...Middleware) http.HandlerFunc {
	// Apply in reverse so the first middleware ends up on the outside
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// ─── Logging Middleware ───────────────────────────────────────────────────────

// LoggingMiddleware logs the method, path, client IP, and response duration.
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("→ %s %s [%s]", r.Method, r.URL.Path, getIP(r))
		next(w, r)
		log.Printf("← %s %s (%v)", r.Method, r.URL.Path, time.Since(start))
	}
}


// MethodMiddleware rejects requests that don't match the allowed HTTP method.
// OPTIONS is always allowed so CORS preflight passes through.
func MethodMiddleware(method string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method && r.Method != http.MethodOptions {
				respond(w, http.StatusMethodNotAllowed, nil, "method not allowed")
				return
			}
			next(w, r)
		}
	}
}


// RateLimitMiddleware uses a sliding window to cap requests per IP.
// Applied to auth endpoints to prevent brute-force attacks.
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isRateLimited(getIP(r)) {
			w.Header().Set("Retry-After", "60")
			respond(w, http.StatusTooManyRequests, nil, "rate limit exceeded — retry in 60s")
			return
		}
		next(w, r)
	}
}


// AuthMiddleware validates the JWT access token in the Authorization header.
// On success it injects the parsed Claims into the request context so
// downstream handlers can read the username without re-parsing the token.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respond(w, http.StatusUnauthorized, nil, "authorization header required")
			return
		}

		// Strip the "Bearer " prefix before validating
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := validateJWT(tokenStr, "access")
		if err != nil {
			respond(w, http.StatusUnauthorized, nil, "invalid or expired access token")
			return
		}

		// Store claims in context so handlers can access them without re-parsing
		ctx := context.WithValue(r.Context(), ctxKey("claims"), claims)
		next(w, r.WithContext(ctx))
	}
}


// isRateLimited uses a sliding window to count requests per IP per minute.
// Returns true if the IP has exceeded rateLimitMax requests.
func isRateLimited(ip string) bool {
	rateLimiterMu.Lock()
	defer rateLimiterMu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rateLimitWindow)

	// Keep only timestamps within the current window
	var fresh []time.Time
	for _, t := range rateLimiter[ip] {
		if t.After(windowStart) {
			fresh = append(fresh, t)
		}
	}
	fresh = append(fresh, now)
	rateLimiter[ip] = fresh

	return len(fresh) > rateLimitMax
}

// getIP extracts the client IP from the request, honouring X-Forwarded-For
// for requests that pass through a reverse proxy.
func getIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}
