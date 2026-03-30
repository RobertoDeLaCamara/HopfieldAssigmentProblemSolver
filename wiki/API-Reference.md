# API Reference

Base URL: `http://localhost/api` (via nginx on port 80)

Direct Go gateway: `http://localhost:8080`

## Authentication

Optional. Set `API_KEY` environment variable to enable.

When enabled:
```
X-API-Key: <value>
# or
Authorization: Bearer <value>
```

Returns `401` if header missing, `403` if invalid. Uses constant-time comparison.

---

## POST /api/solve

Solve a single assignment problem.

**Request body**:
```json
{
  "cost_matrix": [
    [9, 2, 7, 8],
    [6, 4, 3, 7],
    [5, 8, 1, 8],
    [7, 6, 9, 4]
  ]
}
```

**Constraints**:
- Matrix must be square (n×n)
- n must be 2–50
- All values must be ≥ 0 and ≤ 1e9
- No NaN or Inf values

**Response 200**:
```json
{
  "success": true,
  "result": {
    "assignments": [1, 2, 0, 3],
    "total_cost": 14.0,
    "iterations": 156,
    "cost_matrix": [[9,2,7,8],[6,4,3,7],[5,8,1,8],[7,6,9,4]]
  }
}
```

`assignments[i]` is the job index assigned to worker `i`.

**Response 400** — Invalid input:
```json
{
  "success": false,
  "error": "The cost matrix must be square. Row 1 has 3 elements, expected 4."
}
```

---

## POST /api/solve/batch

Solve multiple assignment problems in one request.

**Request body**:
```json
{
  "problems": [
    {"id": "p1", "cost_matrix": [[1,2],[3,4]]},
    {"id": "p2", "cost_matrix": [[5,6],[7,8]]}
  ]
}
```

**Constraints**:
- `problems` array must be non-empty
- Maximum 100 problems per batch
- Each `cost_matrix` follows single-solve constraints

**Response 200** — Batch always returns 200, errors are per-problem:
```json
{
  "success": true,
  "results": [
    {
      "id": "p1",
      "success": true,
      "result": {
        "assignments": [0, 1],
        "total_cost": 5.0,
        "iterations": 45,
        "cost_matrix": [[1,2],[3,4]]
      }
    },
    {
      "id": "p2",
      "success": false,
      "error": "Matrix contains negative values."
    }
  ]
}
```

Individual problem failures do not affect other problems in the batch.

---

## Health Endpoints

```
GET /health          {"status": "healthy", "service": "hopfield-assignment-api", "version": "1.0.0"}
GET /health/ready    {"status": "ready", ...}
GET /health/live     {"status": "alive", ...}
GET /time            {"time": "2026-03-30T12:00:00Z"}
```

---

## Nginx Direct Routes

For debugging, bypass the Go gateway:

```
POST /hopfield/solve         Direct to Flask :5000/solve
GET  /health                 Proxied to Go gateway /health
GET  /status                 {"status": "ok", "service": "hopfield-assignment-solver"}
```

Rate limit for `/hopfield/` is 5 req/s (stricter than the 10 req/s for `/api/`).

---

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success (including batch with individual failures) |
| 400 | Validation error |
| 401 | API key missing (when auth enabled) |
| 403 | API key invalid |
| 500 | Internal error or solver failure |
| 504 | Gateway timeout (30s exceeded) |

---

## Request ID Tracking

Every request receives an `X-Request-ID` response header. Pass `X-Request-ID` in your request to preserve your own ID through the system:

```bash
curl -H "X-Request-ID: my-trace-123" http://localhost/api/solve ...
# Response headers will include: X-Request-ID: my-trace-123
```

This ID appears in all Go middleware log entries for the request.
