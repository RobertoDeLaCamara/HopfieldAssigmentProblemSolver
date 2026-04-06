# HopfieldAssigmentProblemSolver — Wiki

Go API Gateway (Gin, port 8080) proxies to a Python Flask Hopfield neural network solver (port 5000). Nginx reverse proxy in front (port 80). Solves the assignment problem (Hungarian method alternative) using energy minimization with flow conservation constraints.

## Quick Start

```bash
cp env.example .env
make install-deps && make build && make up

# Test
curl -X POST http://localhost/api/solve \
  -H "Content-Type: application/json" \
  -d '{"cost_matrix": [[9,2,7,8],[6,4,3,7],[5,8,1,8],[7,6,9,4]]}'

# Expected: assignments: [1, 2, 0, 3], total_cost: 13.0
```

## Stack

| Component | Technology | Port |
|-----------|-----------|------|
| Reverse proxy | nginx:alpine | 80, 443 |
| API Gateway | Go 1.24.0 + Gin + Logrus | 8080 |
| Solver | Python Flask 3.1.3 + NumPy | 5000 |

## Wiki Pages

- [Architecture and Data Flow](Architecture-and-Data-Flow.md)
- [Hopfield Algorithm](Hopfield-Algorithm.md)
- [API Reference](API-Reference.md)
- [Development Guide](Development-Guide.md)

## Key Layout

```
api/
├── cmd/main.go                   Go entry point (Gin router setup)
├── internal/handlers/
│   ├── assignment.go             /solve and /solve/batch handlers
│   └── health.go                 /health/* handlers
├── internal/models/assignment.go Request/response structs + validation
└── pkg/middleware/
    ├── cors.go                   CORS (all origins)
    ├── logging.go                Request logging (logrus JSON)
    ├── auth.go                   Optional API key (constant-time comparison)
    └── request_context.go        UUID request IDs + metrics

hopfield/src/
├── hopfield_solver.py            HopfieldAssignmentSolver class
├── api_server.py                 Flask /solve and /solve/batch
└── validation.py                 Input constraints (max 50×50)

nginx/nginx.conf                  Rate limiting (10 req/s API, 5 req/s Hopfield)
```

## Non-Obvious Facts

- **Hopfield is heuristic, not exact** — the algorithm may not find the optimal assignment every run. Results can vary between calls due to random neuron initialization (`u ~ N(0, 0.1)`).
- **Greedy decoder guarantees a valid permutation** — even if the network doesn't fully converge to binary values, the greedy step extracts a complete assignment.
- **Matrix size limit is 50×50** — enforced in `validation.py`. Larger matrices cause exponentially longer convergence.
- **Auth uses constant-time comparison** — `subtle.ConstantTimeCompare()` prevents timing attacks on the API key.
- **Batch errors are per-problem** — one invalid problem in a batch returns 200 with `success: false` on that item, not a 400 for the whole batch.
- **Request IDs flow end-to-end** — Go middleware generates UUID, sets it in `X-Request-ID` response header, and includes it in all log entries.
