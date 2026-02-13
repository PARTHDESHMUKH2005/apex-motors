package main

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)


func generateTokenPair(username string) (accessToken, refreshToken string, err error) {
	
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
