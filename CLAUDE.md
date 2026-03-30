# HopfieldAssigmentProblemSolver

Hybrid microservices: Go API Gateway (Gin) + Python Hopfield Solver (Flask), Docker Compose + Nginx.

## Key Commands

```bash
make help          # list all 50+ targets
make build         # build images
make up            # start stack (dev)
make dev           # dev mode with hot reload
make test          # run all tests
make test-unit     # unit only
make test-integration  # integration only
make test-coverage # coverage report
make lint          # flake8 + golangci-lint
make format        # black + isort (Python), gofmt (Go)
```

## Project Structure

- `api/cmd/main.go` — Go API gateway entry point
- `api/internal/handlers/assignment.go` — uses `HTTPDoer` interface (testable HTTP client)
- `hopfield/src/hopfield_solver.py` — core Hopfield algorithm
- `docker-compose.ci.yml` — CI-specific (no host port mappings, avoids conflict with registry on :5000)
- `Jenkinsfile` — declarative pipeline, `agent any` (runs on agent-45)

## CI — Jenkins Build #9 (all passing)

Stages: Prepare → Lint (Python+Go parallel) → Test (Python+Go parallel) → Build Images → Integration Tests

**Integration test env vars:** `API_BASE_URL`, `HOPFIELD_BASE_URL` (configurable for CI vs local)

## Gotchas

- Python formatting uses Docker for version consistency (black/isort versions must match CI)
- Jenkinsfile uses `--profile black` for isort compatibility
- Go tests: io.ReadCloser return types required in test helpers; nil-safe mock responses needed
- flake8: W503 in extend-ignore (matches pre-commit config)

## Remotes

- `origin` → Gitea (192.168.1.62:9090)
- `github` → GitHub (RobertoDeLaCamara/HopfieldAssigmentProblemSolver)
- License: Apache 2.0
