# Development Guide

## Prerequisites

- Go 1.24.0+
- Python 3.x + pip
- Docker + Docker Compose
- nginx (included in docker-compose)

## Quick Start

```bash
cp env.example .env
make setup        # copy .env + install deps + build
make up           # start all services

# Test
make test-api     # quick POST /solve with example 4×4 matrix
make health       # curl /health
```

## Running Locally (Without Docker)

```bash
# Terminal 1: Python solver
make dev-python   # flask run on :5000

# Terminal 2: Go gateway
make dev-go       # go run api/cmd/main.go on :8080
```

No nginx in local mode — connect directly to `:8080`.

## Testing

```bash
make test              # All tests (Docker)
make test-unit        # Python unit tests only
make test-integration # End-to-end (nginx → gateway → flask)
make test-coverage    # HTML coverage report
make test-local       # Without Docker (local Python + Go)

# Single Go test
cd api && go test ./internal/handlers/... -run TestAssignmentHandler_SolveAssignment -v

# Single Python test
pytest hopfield/tests/test_hopfield_solver.py::test_solve_simple_2x2 -v
```

## Makefile Reference

| Target | Description |
|--------|-------------|
| `make build` | `docker-compose build` |
| `make up` | Start all services (background) |
| `make down` | Stop all services |
| `make restart` | Restart services |
| `make status` | Show container status |
| `make logs` | Follow all logs |
| `make logs-api` | Go gateway logs |
| `make logs-hopfield` | Python solver logs |
| `make health` | curl /health |
| `make monitor` | docker stats watch |
| `make lint` | flake8 + golangci-lint |
| `make format` | black + go fmt |
| `make dev-python` | Flask locally on :5000 |
| `make dev-go` | Go gateway locally on :8080 |
| `make install-deps` | pip install + go mod download |
| `make ssl-generate` | Self-signed certs for dev HTTPS |
| `make backup` | Timestamped .tar.gz |
| `make clean` | Remove containers + images + volumes |

## Go Dependencies

```
github.com/gin-gonic/gin v1.9.1
github.com/sirupsen/logrus v1.9.3
github.com/google/uuid v1.6.0
github.com/stretchr/testify v1.8.3
```

## Python Dependencies

```
Flask==3.1.3
numpy==1.26.4
gunicorn==23.0.0
pydantic==2.6.0
python-dotenv==1.0.1
```

## Validation Constraints

| Constraint | Go side | Python side |
|------------|---------|-------------|
| Minimum size | 1×1 (no explicit min) | 2×2 |
| Maximum size | no explicit max | 50×50 |
| Min value | ≥ 0 | ≥ 0 |
| Max value | no explicit max | ≤ 1e9 |
| NaN/Inf | checked | checked |
| Batch max | 100 | 100 |

Both layers validate independently. The Python layer is the authoritative constraint source.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `API_KEY` | `""` | Enable auth if set (Go gateway) |
| `PORT` | `8080` | Go gateway listen port |
| `GIN_MODE` | `debug` | `release` for production |
| `HOPFIELD_SERVICE_URL` | `http://hopfield-service:5000` | Python solver URL |
| `FLASK_ENV` | `development` | Flask environment |

## Project Structure

```
api/
├── cmd/main.go                    Router setup, graceful shutdown
├── internal/
│   ├── handlers/
│   │   ├── assignment.go          SolveAssignment(), SolveBatch()
│   │   ├── health.go              Health endpoints
│   │   ├── assignment_test.go     Handler tests (MockHTTPClient)
│   │   └── health_test.go
│   └── models/
│       ├── assignment.go          AssignmentRequest, Response, Batch structs
│       └── assignment_test.go
└── pkg/middleware/
    ├── cors.go
    ├── logging.go
    ├── auth.go
    └── request_context.go

hopfield/
├── src/
│   ├── hopfield_solver.py         HopfieldAssignmentSolver (1000 iter, greedy decode)
│   ├── api_server.py              Flask /solve, /solve/batch, /health
│   ├── validation.py              Constraints: 2–50, [0, 1e9], no NaN
│   ├── metrics.py                 Request metrics collector
│   └── logging_config.py
└── tests/
    ├── test_hopfield_solver.py    12 algorithm tests
    ├── test_api_server.py         Flask endpoint tests
    └── test_validation.py         Validation rule tests

nginx/
├── nginx.conf                     Rate limiting, proxy, security headers
└── ssl/                           cert.pem + key.pem (dev certs)

tests/
└── integration_test.py            Full stack tests

docker-compose.yml                 hopfield-service + api-gateway + nginx
```
