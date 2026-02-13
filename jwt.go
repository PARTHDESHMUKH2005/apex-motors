package main

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ─── Token Generation ─────────────────────────────────────────────────────────

// generateTokenPair creates a short-lived access token and a long-lived
// refresh token for the given username.
//
// The access token is used on every protected API request (Authorization header).
// The refresh token is used only once — to obtain a new token pair.
func generateTokenPair(username string) (accessToken, refreshToken string, err error) {
	// --- Access Token ---
	// Short-lived (15 min). Sent in Authorization: Bearer <token> header.
	atClaims := &Claims{
		Username:  username,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims).SignedString(jwtSecret)
	if err != nil {
		return
	}

	// --- Refresh Token ---
	// Long-lived (7 days). Stored server-side to allow revocation.
	rtClaims := &Claims{
		Username:  username,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims).SignedString(jwtSecret)
	if err != nil {
		return
	}

	// Persist in server-side store so we can revoke it on logout
	refreshTokensMu.Lock()
	refreshTokens[refreshToken] = username
	refreshTokensMu.Unlock()

	return
}

// ─── Token Validation ─────────────────────────────────────────────────────────

// validateJWT parses and validates a token string.
// expectedType must match the token's TokenType claim ("access" or "refresh").
// This prevents a refresh token from being accepted on a protected route.
func validateJWT(tokenString, expectedType string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Reject tokens signed with anything other than HMAC
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	// Enforce the token type to prevent misuse
	if claims.TokenType != expectedType {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
