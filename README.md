# Apex Motors — Go Backend

A clean, production-ready REST API built with Go's standard library, JWT authentication, and CORS middleware — no frameworks, just Go.

---

## Quick Start

```bash
# 1. Install dependencies
go mod tidy

# 2. Run the server
go run .

# 3. Open the app
open http://localhost:5001

# Demo login:  seller / carmarket123
```

---

## Project Structure

```
apex-motors/
├── main.go          # Server setup, routing, CORS configuration
├── config.go        # All constants (JWT secret, timeouts, rate limits)
├── models.go        # Data types: User, Claims, CarListing, etc.
├── store.go         # In-memory data stores + demo seed data
├── jwt.go           # Token generation and validation
├── middleware.go    # Auth, rate limit, method, logging middleware
├── auth.go          # Login, refresh, logout handlers
├── cars.go          # Car listing CRUD handlers
├── valuation.go     # Rule-based car pricing engine
├── stats.go         # Market statistics handler
├── response.go      # Shared JSON response helper
├── go.mod           # Module definition
├── go.sum           # Dependency checksums
└── static/          # Frontend HTML/CSS/JS (served as-is)
```

---

## API Reference

All protected routes require: `Authorization: Bearer <access_token>`

### Auth

| Method | Route           | Auth? | Description                              |
|--------|-----------------|-------|------------------------------------------|
| POST   | `/api/login`    | No    | Returns access + refresh token pair      |
| POST   | `/api/refresh`  | No    | Rotates tokens (old refresh → new pair)  |
| POST   | `/api/logout`   | Yes   | Revokes refresh token server-side        |

**Login request:**
```json
{ "username": "seller", "password": "carmarket123" }
```

**Login response:**
```json
{
  "success": true,
  "data": {
    "access_token":  "eyJ...",
    "refresh_token": "eyJ...",
    "expires_in":    900,
    "message":       "login successful"
  }
}
```

### Cars

| Method | Route             | Auth? | Description                                |
|--------|-------------------|-------|--------------------------------------------|
| GET    | `/api/cars`       | Yes   | List all cars (filterable + sortable)      |
| GET    | `/api/cars/{id}`  | Yes   | Get single car (increments view count)     |
| POST   | `/api/cars/add`   | Yes   | Create a new listing                       |
| DELETE | `/api/cars/{id}`  | Yes   | Delete a listing (owner only)              |

**GET /api/cars query params:**

| Param       | Example          | Description              |
|-------------|------------------|--------------------------|
| `make`      | `?make=porsche`  | Partial match, case-insensitive |
| `fuel`      | `?fuel=electric` | petrol / diesel / electric / hybrid |
| `condition` | `?condition=new` | new / used / certified   |
| `min_price` | `?min_price=100000` | Lower price bound     |
| `max_price` | `?max_price=300000` | Upper price bound     |
| `sort`      | `?sort=price_asc`   | price_asc / price_desc / year_desc |

**POST /api/cars/add body:**
```json
{
  "make": "Ferrari", "model": "F8 Tributo", "year": 2022,
  "price": 280000, "mileage": 5000,
  "fuel_type": "petrol", "transmission": "automatic", "condition": "used",
  "description": "Full carbon pack.", "image_url": "https://..."
}
```

### Valuation & Stats

| Method | Route          | Auth? | Description                          |
|--------|----------------|-------|--------------------------------------|
| POST   | `/api/valuate` | Yes   | Rule-based price estimate            |
| GET    | `/api/stats`   | Yes   | Live marketplace overview            |

**POST /api/valuate body:**
```json
{
  "make": "BMW", "year": 2020, "mileage": 45000,
  "condition": "used", "fuel_type": "petrol", "transmission": "automatic"
}
```

---

## Architecture Decisions

### JWT — Access + Refresh Token Pattern

- **Access token** (15 min): short-lived, sent with every API request in the `Authorization` header.
- **Refresh token** (7 days): long-lived, stored server-side. Used only to get a new token pair.
- **Token rotation**: each `/api/refresh` call deletes the old refresh token and issues a new one — a stolen token can only be replayed once.
- **Server-side revocation**: logout deletes the refresh token from the in-memory store so it can never be reused.

### Middleware Chain

```go
// Middlewares compose cleanly in declaration order:
Chain(handler, LoggingMiddleware, AuthMiddleware, MethodMiddleware("POST"))
// Execution: Logging → Auth check → Method check → handler
```

Each middleware is a simple `func(http.HandlerFunc) http.HandlerFunc` — easy to test and extend.

### CORS

Locked to `http://localhost:5001` (same origin as the server). In production, replace `allowedOrigins` in `config.go` with your actual domain.

### Rate Limiting

Sliding window per IP on auth endpoints — maximum 10 requests per minute. Returns `429 Too Many Requests` with a `Retry-After: 60` header.

### Consistent API Response Envelope

Every response follows the same shape:
```json
{ "success": true,  "data": { ... } }
{ "success": false, "error": "human-readable message" }
```
The frontend can always check `json.success` before accessing `json.data`.

---

## Running Tests (manual)

```bash
# Login
curl -X POST http://localhost:5001/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"seller","password":"carmarket123"}'

# Use returned access_token for protected routes
TOKEN="<access_token_from_above>"

# Get all cars
curl http://localhost:5001/api/cars \
  -H "Authorization: Bearer $TOKEN"

# Add a car
curl -X POST http://localhost:5001/api/cars/add \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"make":"Ferrari","model":"F8","year":2022,"price":280000,"mileage":5000,"fuel_type":"petrol","transmission":"automatic","condition":"used"}'

# Get market stats
curl http://localhost:5001/api/stats \
  -H "Authorization: Bearer $TOKEN"

# Valuate a car
curl -X POST http://localhost:5001/api/valuate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"make":"BMW","year":2020,"mileage":45000,"condition":"used","fuel_type":"petrol","transmission":"automatic"}'
```
