# Validation contract

The Go gateway (`api/`) and the Python solver (`hopfield/`) both validate cost matrices. If they disagree, a request the gateway accepts comes back from the solver as an opaque 500 instead of a clear 400.

| File | Role |
|---|---|
| `validation_cases.json` | Single source of truth: the `limits` and a list of valid/invalid matrices and batch sizes. |
| `openapi.yaml` | The same contract as an OpenAPI 3 description, for readers and API tooling. |

Both suites run every case in `validation_cases.json`:

- Go: `api/internal/models/contract_test.go` (also checks the `Min/MaxMatrixSize`, `Min/MaxCostValue` and `MaxBatchSize` constants against `limits`).
- Python: `hopfield/tests/test_contract.py` (also checks `get_validation_summary()` and that `openapi.yaml` states the same numbers).

**To change a limit:** edit `limits` here, `openapi.yaml`, the Go constants in `api/internal/models/assignment.go` and the Python constants in `hopfield/src/validation.py`. Whichever you forget, a test fails. To add a rule, add a case to `validation_cases.json` and both sides must satisfy it.

NaN and infinity cannot appear in JSON, so they are covered only by the solver's own tests.

The tests read this directory, so CI and `make test` mount the repository root, not just `api/` or `hopfield/`.
