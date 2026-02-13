package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ─── GET /api/cars ────────────────────────────────────────────────────────────

// getCarsHandler returns all listings, with optional filtering and sorting.
//
// Query params:
//
//	make        — filter by make (partial, case-insensitive)
//	fuel        — filter by fuel type (petrol/diesel/electric/hybrid)
//	condition   — filter by condition (new/used/certified)
//	min_price   — lower price bound
//	max_price   — upper price bound
//	sort        — price_asc | price_desc | year_desc
func getCarsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	makeF := strings.ToLower(q.Get("make"))
	fuelF := strings.ToLower(q.Get("fuel"))
	condF := strings.ToLower(q.Get("condition"))
	minP, _ := strconv.ParseFloat(q.Get("min_price"), 64)
	maxP, _ := strconv.ParseFloat(q.Get("max_price"), 64)

	storeMu.RLock()
	var listings []CarListing
	for _, car := range carStore {
		if makeF != "" && !strings.Contains(strings.ToLower(car.Make), makeF) {
			continue
		}
		if fuelF != "" && strings.ToLower(car.FuelType) != fuelF {
			continue
		}
		if condF != "" && strings.ToLower(car.Condition) != condF {
			continue
		}
		if minP > 0 && car.Price < minP {
			continue
		}
		if maxP > 0 && car.Price > maxP {
			continue
		}
		listings = append(listings, car)
	}
	storeMu.RUnlock()

	// Simple bubble sort — fine for small in-memory slices
	switch q.Get("sort") {
	case "price_asc":
		sortBy(listings, func(a, b CarListing) bool { return a.Price < b.Price })
	case "price_desc":
		sortBy(listings, func(a, b CarListing) bool { return a.Price > b.Price })
	case "year_desc":
		sortBy(listings, func(a, b CarListing) bool { return a.Year > b.Year })
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"listings": listings,
		"count":    len(listings),
	}, "")
}

// sortBy is a tiny generic-style helper for sorting CarListing slices.
func sortBy(lst []CarListing, less func(a, b CarListing) bool) {
	for i := 0; i < len(lst)-1; i++ {
		for j := i + 1; j < len(lst); j++ {
			if !less(lst[i], lst[j]) {
				lst[i], lst[j] = lst[j], lst[i]
			}
		}
	}
}

// ─── GET /api/cars/{id} ───────────────────────────────────────────────────────

// getCarHandler returns a single listing by ID and increments its view counter.
func getCarHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseCarID(r.URL.Path)
	if err != nil {
		respond(w, http.StatusBadRequest, nil, "invalid car id")
		return
	}

	storeMu.Lock()
	car, ok := carStore[id]
	if !ok {
		storeMu.Unlock()
		respond(w, http.StatusNotFound, nil, "car not found")
		return
	}
	car.Views++
	carStore[id] = car
	storeMu.Unlock()

	respond(w, http.StatusOK, car, "")
}

// ─── POST /api/cars/add ───────────────────────────────────────────────────────

// addCarHandler creates a new listing. Requires authentication.
// The seller field is set from the JWT claims — clients cannot spoof it.
func addCarHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(ctxKey("claims")).(*Claims)

	var car CarListing
	if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
		respond(w, http.StatusBadRequest, nil, "invalid request body")
		return
	}

	// Basic validation — all required fields must be present
	if car.Make == "" || car.Model == "" || car.Year == 0 || car.Price <= 0 {
		respond(w, http.StatusBadRequest, nil, "make, model, year and price are required")
		return
	}

	storeMu.Lock()
	car.ID = nextID
	car.Seller = claims.Username // always from JWT, never from client body
	car.ListedAt = time.Now().Format(time.RFC3339)
	car.Views = 0
	carStore[car.ID] = car
	nextID++
	storeMu.Unlock()

	respond(w, http.StatusCreated, car, "")
}

// ─── DELETE /api/cars/{id} ────────────────────────────────────────────────────

// deleteCarHandler removes a listing. Only the original seller may delete it.
// Returns 403 Forbidden (not 404) so the client knows the car exists but
// they don't own it — this is intentional information disclosure here.
func deleteCarHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(ctxKey("claims")).(*Claims)

	id, err := parseCarID(r.URL.Path)
	if err != nil {
		respond(w, http.StatusBadRequest, nil, "invalid car id")
		return
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	car, ok := carStore[id]
	if !ok {
		respond(w, http.StatusNotFound, nil, "car not found")
		return
	}

	if car.Seller != claims.Username {
		// 403 Forbidden — authenticated but not authorised
		respond(w, http.StatusForbidden, nil, "you can only delete your own listings")
		return
	}

	delete(carStore, id)
	respond(w, http.StatusOK, map[string]string{"message": "listing deleted"}, "")
}

// ─── Helper ───────────────────────────────────────────────────────────────────

// parseCarID strips the /api/cars/ prefix and parses the remaining ID.
func parseCarID(path string) (int, error) {
	raw := strings.TrimPrefix(path, "/api/cars/")
	return strconv.Atoi(raw)
}
