package main

import (
	"math/rand"
	"sync"
	"time"
)

// ─── Car Store ────────────────────────────────────────────────────────────────

var (
	carStore = make(map[int]CarListing)
	nextID   = 1
	storeMu  sync.RWMutex // RWMutex: many concurrent readers, one writer
)

// ─── Refresh Token Store ──────────────────────────────────────────────────────
// Maps token string → username.
// Kept server-side so we can revoke tokens immediately (logout, rotation).

var (
	refreshTokens   = make(map[string]string)
	refreshTokensMu sync.RWMutex
)

// ─── Rate Limit Store ─────────────────────────────────────────────────────────
// Maps IP address → slice of request timestamps (sliding window).

var (
	rateLimiter   = make(map[string][]time.Time)
	rateLimiterMu sync.Mutex
)

// ─── Seed Demo Data ───────────────────────────────────────────────────────────

// seedDemoInventory populates the car store with realistic demo listings.
// Called once at startup so the API has data to return immediately.
func seedDemoInventory() {
	demo := []CarListing{
		{
			Make: "McLaren", Model: "765LT", Year: 2021, Price: 358000,
			Mileage: 8200, FuelType: "petrol", Transmission: "automatic", Condition: "used",
			Description: "Track-focused supercar. Titanium exhaust, MSO carbon pack.",
			ImageURL:    "https://images.unsplash.com/photo-1621135802920-133df287f89c?w=800",
		},
		{
			Make: "Porsche", Model: "911 GT3", Year: 2023, Price: 185000,
			Mileage: 3100, FuelType: "petrol", Transmission: "manual", Condition: "certified",
			Description: "Weissach package, carbon ceramics, clubsport seats.",
			ImageURL:    "https://images.unsplash.com/photo-1503376780353-7e6692767b70?w=800",
		},
		{
			Make: "BMW", Model: "M5 CS", Year: 2022, Price: 142000,
			Mileage: 14000, FuelType: "petrol", Transmission: "automatic", Condition: "used",
			Description: "630hp, carbon roof, M bucket seats.",
			ImageURL:    "https://images.unsplash.com/photo-1555215695-3004980ad54e?w=800",
		},
		{
			Make: "Tesla", Model: "Model S Plaid", Year: 2023, Price: 118000,
			Mileage: 5600, FuelType: "electric", Transmission: "automatic", Condition: "certified",
			Description: "1020hp tri-motor. 0-100 in 2.1s. FSD included.",
			ImageURL:    "https://images.unsplash.com/photo-1560958089-b8a1929cea89?w=800",
		},
		{
			Make: "Mercedes", Model: "AMG GT R", Year: 2021, Price: 167000,
			Mileage: 11300, FuelType: "petrol", Transmission: "automatic", Condition: "used",
			Description: "585hp V8, aero package, Green Hell Magno paint.",
			ImageURL:    "https://images.unsplash.com/photo-1618843479313-40f8afb4b4d8?w=800",
		},
		{
			Make: "Audi", Model: "R8 V10 Plus", Year: 2022, Price: 195000,
			Mileage: 6700, FuelType: "petrol", Transmission: "automatic", Condition: "new",
			Description: "620hp naturally aspirated V10. Laser headlights.",
			ImageURL:    "https://images.unsplash.com/photo-1606016159991-dfe4f2746ad5?w=800",
		},
	}

	storeMu.Lock()
	defer storeMu.Unlock()
	for i, car := range demo {
		car.ID = nextID
		car.Seller = "demo"
		car.ListedAt = time.Now().Add(-time.Duration(i*5) * 24 * time.Hour).Format(time.RFC3339)
		car.Views = rand.Intn(200) + 10
		carStore[car.ID] = car
		nextID++
	}
}
