package main

import (
	"encoding/json"
	"net/http"
)

// respond writes a consistent JSON envelope to the response writer.
// Every endpoint uses this so the client always gets the same shape:
//
//	{ "success": true,  "data": { ... } }
//	{ "success": false, "error": "some message" }
//
// Parameters:
//
//	w      — the response writer
//	code   — HTTP status code (200, 201, 400, 401, 403, 404, 500 …)
//	data   — payload for success responses (nil for errors)
//	errMsg — error message; empty string means success
func respond(w http.ResponseWriter, code int, data interface{}, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(APIResponse{
		Success: errMsg == "",
		Data:    data,
		Error:   errMsg,
	})
}
