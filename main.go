package main

import (
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"

	"github.com/rs/cors"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Populate the in-memory store with demo cars
	seedDemoInventory()

	// ── Router ────────────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	// ── Static page routes ────────────────────────────────────────────────────
	// Each route serves the corresponding HTML file from the static/ directory.
	// The frontend uses shared.css and shared.js for styling and common behavior.
	mux.HandleFunc("/", LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "index.html"))
	}))
	mux.HandleFunc("/login", LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "login.html"))
	}))
	mux.HandleFunc("/collection", LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "collection.html"))
	}))
	mux.HandleFunc("/experience", LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "experience.html"))
	}))
	mux.HandleFunc("/contact", LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "contact.html"))
	}))
	mux.HandleFunc("/services", LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "services.html"))
	}))
	mux.HandleFunc("/new-arrivals", LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "new-arrivals.html"))
	}))
	mux.HandleFunc("/sold-archive", LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "sold-archive.html"))
	}))

	// Serve static assets (CSS, JS, images) from the static/ folder
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// ── Auth API ──────────────────────────────────────────────────────────────
	// Rate limited to prevent brute-force attacks.
	mux.HandleFunc("/api/login",
		LoggingMiddleware(Chain(loginHandler,
			RateLimitMiddleware,
			MethodMiddleware("POST"),
		)))

	// Token rotation — client sends old refresh token, gets a new pair back
	mux.HandleFunc("/api/refresh",
		LoggingMiddleware(Chain(refreshHandler,
			RateLimitMiddleware,
			MethodMiddleware("POST"),
		)))

	// Logout revokes the refresh token server-side
	mux.HandleFunc("/api/logout",
		LoggingMiddleware(Chain(logoutHandler,
			AuthMiddleware,
			MethodMiddleware("POST"),
		)))

	// ── Car Listing API ───────────────────────────────────────────────────────
	// All car routes require a valid JWT access token.

	// GET  /api/cars         — list all (with optional filters)
	mux.HandleFunc("/api/cars",
		LoggingMiddleware(Chain(getCarsHandler,
			AuthMiddleware,
			MethodMiddleware("GET"),
		)))

	// POST /api/cars/add     — create a new listing
	mux.HandleFunc("/api/cars/add",
		LoggingMiddleware(Chain(addCarHandler,
			AuthMiddleware,
			MethodMiddleware("POST"),
		)))

	// GET|DELETE /api/cars/{id} — view or remove a single listing
	mux.HandleFunc("/api/cars/",
		LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				Chain(getCarHandler, AuthMiddleware)(w, r)
			case http.MethodDelete:
				Chain(deleteCarHandler, AuthMiddleware)(w, r)
			default:
				respond(w, http.StatusMethodNotAllowed, nil, "method not allowed")
			}
		}))

	// ── Utility API ───────────────────────────────────────────────────────────

	// POST /api/valuate — rule-based car valuation engine
	mux.HandleFunc("/api/valuate",
		LoggingMiddleware(Chain(valuateHandler,
			AuthMiddleware,
			MethodMiddleware("POST"),
		)))

	// GET /api/stats — live marketplace overview
	mux.HandleFunc("/api/stats",
		LoggingMiddleware(Chain(statsHandler,
			AuthMiddleware,
			MethodMiddleware("GET"),
		)))

	// ── CORS ──────────────────────────────────────────────────────────────────
	// Configured to only accept requests from our own origin.
	// In production, set allowedOrigins to your actual domain.
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300, // cache preflight for 5 minutes
	})

	// ── Server ────────────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      c.Handler(mux),
		ReadTimeout:  serverReadTTO,
		WriteTimeout: serverWriteTTO,
		IdleTimeout:  serverIdleTTO,
	}

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("  APEX MOTORS  →  http://localhost:5001      ")
	log.Println("  Login:  seller / carmarket123              ")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
