package main

import "net/http"

// ─── GET /api/stats ───────────────────────────────────────────────────────────

// statsHandler returns a live overview of the car marketplace.
// All calculations are done in a single pass over the store for efficiency.
func statsHandler(w http.ResponseWriter, r *http.Request) {
	storeMu.RLock()
	defer storeMu.RUnlock()

	total := len(carStore)
	totalValue := 0.0
	fuelBreakdown := map[string]int{}
	condBreakdown := map[string]int{}

	// Track extremes for the summary cards
	topViewed := CarListing{}
	cheapest := CarListing{Price: 1e12} // start high so first real car wins
	mostExpensive := CarListing{}

	for _, car := range carStore {
		totalValue += car.Price
		fuelBreakdown[car.FuelType]++
		condBreakdown[car.Condition]++

		if car.Views > topViewed.Views {
			topViewed = car
		}
		if car.Price < cheapest.Price {
			cheapest = car
		}
		if car.Price > mostExpensive.Price {
			mostExpensive = car
		}
	}

	avgPrice := 0.0
	if total > 0 {
		avgPrice = totalValue / float64(total)
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"total_listings":      total,
		"total_value":         roundToHundred(totalValue),
		"average_price":       roundToHundred(avgPrice),
		"fuel_breakdown":      fuelBreakdown,
		"condition_breakdown": condBreakdown,
		"most_viewed":         topViewed,
		"cheapest":            cheapest,
		"most_expensive":      mostExpensive,
	}, "")
}
