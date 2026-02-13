package main

import (
	"encoding/json"
	"net/http"
)

// ─── POST /api/login ──────────────────────────────────────────────────────────

// loginHandler authenticates a user and returns an access + refresh token pair.
//
// Request body:  { "username": "seller", "password": "carmarket123" }
// Response:      { "access_token": "...", "refresh_token": "...", "expires_in": 900 }
//
// Protected by: RateLimitMiddleware (brute-force prevention)
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		respond(w, http.StatusBadRequest, nil, "invalid request body")
		return
	}

	// Simple credential check — swap with a database lookup in production
	if creds.Username != demoUsername || creds.Password != demoPassword {
		respond(w, http.StatusUnauthorized, nil, "invalid credentials")
		return
	}

	access, refresh, err := generateTokenPair(creds.Username)
	if err != nil {
		respond(w, http.StatusInternalServerError, nil, "token generation failed")
		return
	}

	respond(w, http.StatusOK, LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(accessTokenTTL.Seconds()),
		Message:      "login successful",
	}, "")
}

// ─── POST /api/refresh ────────────────────────────────────────────────────────

// refreshHandler performs token rotation.
// The client sends its current refresh token; we:
//  1. Validate the JWT signature and type
//  2. Confirm it exists in the server-side store (not revoked)
//  3. Delete the old refresh token (single-use rotation)
//  4. Issue a brand new access + refresh token pair
//
// This means a stolen refresh token can only be used once before it's rotated.
func refreshHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond(w, http.StatusBadRequest, nil, "invalid request body")
		return
	}

	// Validate the JWT itself (signature + expiry + type)
	claims, err := validateJWT(body.RefreshToken, "refresh")
	if err != nil {
		respond(w, http.StatusUnauthorized, nil, "invalid refresh token")
		return
	}

	// Check server-side store — token may have been revoked via logout
	refreshTokensMu.RLock()
	_, exists := refreshTokens[body.RefreshToken]
	refreshTokensMu.RUnlock()

	if !exists {
		respond(w, http.StatusUnauthorized, nil, "refresh token revoked")
		return
	}

	// Rotate: delete the used token before issuing new ones
	refreshTokensMu.Lock()
	delete(refreshTokens, body.RefreshToken)
	refreshTokensMu.Unlock()

	access, refresh, err := generateTokenPair(claims.Username)
	if err != nil {
		respond(w, http.StatusInternalServerError, nil, "token generation failed")
		return
	}

	respond(w, http.StatusOK, LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(accessTokenTTL.Seconds()),
		Message:      "tokens refreshed",
	}, "")
}

// ─── POST /api/logout ─────────────────────────────────────────────────────────

// logoutHandler revokes the client's refresh token server-side.
// Even if the access token is still valid (up to 15 min), the refresh
// token can no longer be used to obtain a new pair.
//
// The client should also delete both tokens from storage on their end.
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	// Decode is best-effort — logout succeeds even if body is missing
	json.NewDecoder(r.Body).Decode(&body)

	if body.RefreshToken != "" {
		refreshTokensMu.Lock()
		delete(refreshTokens, body.RefreshToken)
		refreshTokensMu.Unlock()
	}

	respond(w, http.StatusOK, map[string]string{"message": "logged out"}, "")
}
