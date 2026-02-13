package main

import "time"

// ─── Application Configuration ────────────────────────────────────────────────
// All tuneable constants live here. Easy to swap to env vars later.

var jwtSecret = []byte("apex-motors-secret-change-in-production")

const (
	// Token lifetimes
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour

	// Rate limiting: max requests per IP per minute on auth endpoints
	rateLimitWindow = time.Minute
	rateLimitMax    = 10

	// Server settings
	serverAddr     = ":5001"
	serverReadTTO  = 15 * time.Second
	serverWriteTTO = 15 * time.Second
	serverIdleTTO  = 60 * time.Second

	// Demo credentials (hard-coded for portfolio demo purposes)
	demoUsername = "seller"
	demoPassword = "carmarket123"
)

// allowedOrigins controls which origins the CORS middleware accepts.
// In production, replace with your actual front-end domain.
var allowedOrigins = []string{"http://localhost:5001"}
