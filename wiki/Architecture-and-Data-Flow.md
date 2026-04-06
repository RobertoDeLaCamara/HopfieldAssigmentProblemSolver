# Architecture & Data Flow

## System Overview

```
┌─────────────────────────────────────────────────────────────┐
│  Client                                                     │
│  POST /api/solve  {"cost_matrix": [[...]]}                  │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  nginx :80/:443                                             │
│  Rate limit: 10 req/s (API), 5 req/s (Hopfield direct)      │
│  Security headers: X-Frame-Options, X-Content-Type-Options   │
│  proxy_pass → api-gateway:8080                              │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Go API Gateway :8080   (api/cmd/main.go)                   │
│                                                             │
│  Middleware stack:                                          │
│  1. LoggingMiddleware (logrus JSON)                         │
│  2. CORS (allow all origins)                                │
│  3. APIKeyAuth (optional, constant-time compare)            │
│  4. RequestContext (UUID + metrics)                         │
│                                                             │
│  Handler (api/internal/handlers/assignment.go):             │
│  - Validate cost matrix (square, non-negative, 2–50 dim)    │
│  - Context timeout: 30 seconds                              │
│  - POST → http://hopfield-service:5000/solve                │
└─────────────────────────┬───────────────────────────────────┘
                          │ POST /solve {cost_matrix}
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Flask Solver :5000   (hopfield/src/api_server.py)          │
│                                                             │
│  validate_matrix(cost_matrix)                               │
│  → 2×2 minimum, 50×50 maximum                               │
│  → all values in [0, 1e9], no NaN/Inf                       │
│                                                             │
│  HopfieldAssignmentSolver.solve(cost_matrix)                │
│  → Normalize to [0,1]                                       │
│  → Initialize neurons: u ~ N(0, 0.1)                        │
│  → Energy minimization (1000 iterations max)                │
│  → Greedy decoding → assignments[]                          │
│  → total_cost = sum(cost[i][assignments[i]])                │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
                  {"assignments": [1,2,0,3],
                   "total_cost": 13.0,
                   "iterations": 156,
                   "cost_matrix": [...]}
```

## Request Lifecycle

```
Client: POST /api/solve
  │
  Nginx: rate limit check → proxy to :8080
  │
  Go middleware:
  ├─ Generate request UUID (or accept X-Request-ID header)
  ├─ Log: method, path, client_ip, user_agent
  ├─ CORS preflight (OPTIONS) → handled
  ├─ APIKeyAuth: if API_KEY set, check X-API-Key header
  │   └─ subtle.ConstantTimeCompare() prevents timing attacks
  │   └─ 401 if missing, 403 if invalid
  │
  Go handler:
  ├─ Decode JSON body → AssignmentRequest
  ├─ Validate CostMatrix:
  │   ├─ Non-empty
  │   ├─ Square (rows == cols)
  │   ├─ All values ≥ 0
  │   └─ No NaN/Inf
  ├─ Create context with 30s timeout
  ├─ HTTP POST to hopfield-service:5000/solve
  │   └─ If timeout → 500 with error
  │   └─ If service error → propagate error
  │
  Flask:
  ├─ Validate (Python-side, max size 50×50)
  ├─ HopfieldAssignmentSolver.solve(matrix)
  └─ Return JSON result
  │
  Go: wrap in APIResponse{Success: true, Result: ...}
  Nginx: forward response to client
  Go middleware: log final status + latency
```

## Batch Processing

```
POST /api/solve/batch {"problems": [{id, cost_matrix}, ...]}
  │
  Go handler:
  ├─ Validate batch (non-empty, max 100 problems)
  ├─ For each problem:
  │   ├─ Validate individually
  │   ├─ If invalid: add BatchResult{id, success: false, error: msg}
  │   └─ If valid: POST /solve with 30s timeout per problem
  │
  └─ Return BatchResponse{success: true, results: [...]}
      Note: batch-level success = true even if individual problems fail
```

## Container Network

```
hopfield-network (bridge):
  nginx → api-gateway:8080     (internal DNS)
  api-gateway → hopfield-service:5000

Health check chain:
  hopfield-service: curl http://localhost:5000/health (30s interval)
  api-gateway: depends_on hopfield-service (service_healthy)
  nginx: depends_on api-gateway
```

## Go Middleware Stack

**1. LoggingMiddleware** — structured JSON via logrus:
Fields: `timestamp`, `status`, `latency`, `client_ip`, `method`, `path`, `user_agent`, `error`, `request_id`

**2. CORS** — allows all origins:
Methods: GET, POST, PUT, DELETE, OPTIONS
Headers: Origin, Content-Type, Accept-Encoding, X-CSRF-Token, Authorization
Credentials: true

**3. APIKeyAuth** — optional:
- Reads `API_KEY` env var
- Checks `X-API-Key` header or `Authorization: Bearer <token>`
- Uses `crypto/subtle.ConstantTimeCompare()` to prevent timing attacks
- Skipped entirely if `API_KEY` not set

**4. RequestContext** — per-request metadata:
- UUID from header or generated (`github.com/google/uuid`)
- Set in response as `X-Request-ID`
- Counters: total_requests, error_count, avg/max latency
