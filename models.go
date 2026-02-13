package main

import "github.com/golang-jwt/jwt/v5"

// ─── Auth Models ──────────────────────────────────────────────────────────────

// User is the login request payload.
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims is embedded inside every JWT.
// TokenType distinguishes "access" tokens from "refresh" tokens so that
// a refresh token cannot be used directly on protected API routes.
type Claims struct {
	Username  string `json:"username"`
	TokenType string `json:"token_type"` // "access" | "refresh"
	jwt.RegisteredClaims
}

// LoginResponse is what the client receives after a successful login or refresh.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds until access token expires
	Message      string `json:"message"`
}

// ─── Car Models ───────────────────────────────────────────────────────────────

// CarListing represents a single car in the marketplace.
type CarListing struct {
	ID           int     `json:"id"`
	Make         string  `json:"make"`
	Model        string  `json:"model"`
	Year         int     `json:"year"`
	Mileage      int     `json:"mileage"`
	FuelType     string  `json:"fuel_type"`    // petrol | diesel | electric | hybrid
	Transmission string  `json:"transmission"` // manual | automatic
	Condition    string  `json:"condition"`    // new | used | certified
	Price        float64 `json:"price"`
	Description  string  `json:"description"`
	ImageURL     string  `json:"image_url"`
	Seller       string  `json:"seller"`
	ListedAt     string  `json:"listed_at"`
	Views        int     `json:"views"`
}

// ─── Valuation Models ─────────────────────────────────────────────────────────

// ValuationRequest is the input to the rule-based pricing engine.
type ValuationRequest struct {
	Make         string `json:"make"`
	Year         int    `json:"year"`
	Mileage      int    `json:"mileage"`
	Condition    string `json:"condition"`
	FuelType     string `json:"fuel_type"`
	Transmission string `json:"transmission"`
}

// ValuationResponse is the output of the pricing engine.
type ValuationResponse struct {
	EstimatedMin float64  `json:"estimated_min"`
	EstimatedMax float64  `json:"estimated_max"`
	Confidence   string   `json:"confidence"`
	Factors      []string `json:"factors"` // human-readable explanation of adjustments
}

// ─── API Envelope ─────────────────────────────────────────────────────────────

// APIResponse is a consistent JSON wrapper for every response.
// Every endpoint returns { success, data?, error? } so the client
// can always check `success` first and branch accordingly.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ─── Context Key ──────────────────────────────────────────────────────────────

// ctxKey is the typed key used to store/retrieve values from request context.
// Using a custom type prevents collisions with other packages.
type ctxKey string
