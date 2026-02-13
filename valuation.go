package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ─── POST /api/valuate ────────────────────────────────────────────────────────

// valuateHandler estimates a car's market value using a rule-based engine.
// It applies multipliers for depreciation, mileage, condition, fuel type,
// and transmission — each explained in the response `factors` array.
//
// This is intentionally simple and transparent so it can be explained in
// a portfolio/interview context without complex ML dependencies.
func valuateHandler(w http.ResponseWriter, r *http.Request) {
	var req ValuationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, nil, "invalid request body")
		return
	}

	if req.Make == "" || req.Year == 0 {
		respond(w, http.StatusBadRequest, nil, "make and year are required")
		return
	}

	value, factors := calculateValue(req)
	variance := value * 0.07 // ±7% range for min/max estimate

	respond(w, http.StatusOK, ValuationResponse{
		EstimatedMin: roundToHundred(value - variance),
		EstimatedMax: roundToHundred(value + variance),
		Confidence:   "medium",
		Factors:      factors,
	}, "")
}

// calculateValue runs the pricing engine and returns the estimated value
// along with a human-readable list of factors that affected the price.
func calculateValue(req ValuationRequest) (float64, []string) {
	value := basePriceFor(req.Make)
	var factors []string

	// ── Step 1: Depreciation ──────────────────────────────────────────────────
	// Cars depreciate ~12% per year after the first 3 years.
	age := time.Now().Year() - req.Year
	if age > 3 {
		for i := 3; i < age; i++ {
			value *= 0.88
		}
		factors = append(factors, "Annual depreciation applied (12%/yr after year 3)")
	}

	// ── Step 2: Mileage ───────────────────────────────────────────────────────
	switch {
	case req.Mileage > 150000:
		value *= 0.72
		factors = append(factors, "Very high mileage (>150k km): -28%")
	case req.Mileage > 100000:
		value *= 0.82
		factors = append(factors, "High mileage (>100k km): -18%")
	case req.Mileage < 10000:
		value *= 1.12
		factors = append(factors, "Very low mileage (<10k km): +12%")
	case req.Mileage < 30000:
		value *= 1.06
		factors = append(factors, "Low mileage (<30k km): +6%")
	}

	// ── Step 3: Condition ─────────────────────────────────────────────────────
	switch strings.ToLower(req.Condition) {
	case "new":
		value *= 1.15
		factors = append(factors, "New condition: +15%")
	case "certified":
		value *= 1.06
		factors = append(factors, "Certified pre-owned: +6%")
	default:
		factors = append(factors, "Standard used vehicle pricing")
	}

	// ── Step 4: Fuel Type ─────────────────────────────────────────────────────
	switch strings.ToLower(req.FuelType) {
	case "electric":
		value *= 1.18
		factors = append(factors, "Electric: strong demand premium (+18%)")
	case "hybrid":
		value *= 1.08
		factors = append(factors, "Hybrid: efficiency premium (+8%)")
	case "diesel":
		value *= 0.94
		factors = append(factors, "Diesel: regulatory risk discount (-6%)")
	}

	// ── Step 5: Transmission ──────────────────────────────────────────────────
	if strings.ToLower(req.Transmission) == "automatic" {
		value *= 1.03
		factors = append(factors, "Automatic gearbox: +3%")
	}

	return value, factors
}

// basePriceFor returns a tier-based starting price for a given car make.
// Unrecognised makes fall back to a sensible mid-market default.
func basePriceFor(make string) float64 {
	tiers := map[string]float64{
		"rolls royce":  350000,
		"ferrari":      260000,
		"lamborghini":  230000,
		"mclaren":      200000,
		"bentley":      200000,
		"aston martin": 160000,
		"tesla":        70000,
		"porsche":      90000,
		"mercedes":     58000,
		"bmw":          52000,
		"audi":         50000,
		"jaguar":       55000,
		"lexus":        48000,
		"ford":         30000,
		"toyota":       26000,
		"honda":        23000,
		"hyundai":      21000,
	}

	lower := strings.ToLower(make)
	for brand, price := range tiers {
		if strings.Contains(lower, brand) {
			return price
		}
	}
	return 30000 // default mid-market fallback
}

// roundToHundred rounds a value to the nearest 100 for cleaner display.
func roundToHundred(v float64) float64 {
	return float64(int(v/100+0.5)) * 100
}
