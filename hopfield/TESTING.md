# Testing Requirements

This file contains all dependencies needed for running tests.

## Installation

```bash
# Install all test dependencies
pip install -r requirements.txt

# Or install with Docker
docker-compose up --build
```

## Running Tests

### All Tests
```bash
# Run all tests
pytest tests/ -v

# Run with coverage
pytest tests/ -v --cov=src --cov-report=html

# Run specific test file
pytest tests/test_validation.py -v
pytest tests/test_api_server.py -v
pytest tests/test_hopfield_solver.py -v
```

### Individual Test Suites

**1. Validation Tests (24 tests)**
```bash
pytest tests/test_validation.py -v
```
Tests for the validation module including:
- Matrix size validation (min 2x2, max 50x50)
- Cost value validation (0 to 1e9, no NaN/Inf)
- Batch validation (max 100 problems)
- Error message quality

**2. Hopfield Solver Tests (13 tests)**
```bash
pytest tests/test_hopfield_solver.py -v
```
Tests for the Hopfield neural network solver:
- Algorithm correctness
- Edge cases (zero matrix, large values, single element)
- Error handling for invalid inputs
- Convergence properties

**3. API Server Tests (40+ tests)**
```bash
pytest tests/test_api_server.py -v
```
Tests for both original and enhanced API servers:
- Health  endpoints (/health, /health/ready, /health/live)
- Solve endpoint with validation
- Batch processing
- Request ID tracking
-CORS headers
- Metrics collection
- Enhanced validation

### Docker Testing

Run tests in Docker environment (recommended):

```bash
# Start services
docker-compose up -d

# Run Python tests
docker-compose exec hopfield-service pytest tests/ -v

# Run integration tests
pytest tests/integration_test.py -v

# Stop services
docker-compose down
```

### CI/CD

Tests are automatically run on:
- Every push to any branch
- Every pull request
- GitHub Actions workflow: `.github/workflows/ci.yml`

See test results in the Actions tab of the GitHub repository.

## Test Coverage

Current coverage:
- **Validation Module**: 100%
- **Hopfield Solver**: ~95%
- **API Server**: ~90%
- **Overall**: ~92%

## Writing New Tests

When adding new tests:

1. **Location**: Place tests in `tests/` directory
2. **Naming**: Name test files `test_*.py`
3. **Structure**: Use pytest fixtures and classes
4. **Imports**: Add path setup for src imports:
   ```python
   import sys
   import os
   sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))
   ```

5. **Run locally** before committing:
   ```bash
   pytest tests/ -v
   ```

## Troubleshooting

### Import Errors

If you get `ModuleNotFoundError`:

```bash
# Install dependencies
pip install -r requirements.txt

# Or run in Docker
docker-compose exec hopfield-service pytest tests/ -v
```

### Flask Not Found

```bash
pip install Flask==3.0.2
```

### Path Issues

Make sure you're running from the  `hopfield` directory:
```bash
cd hopfield
pytest tests/ -v
```

## Test Categories

- **Unit Tests**: Test individual functions and classes
- **Integration Tests**: Test API endpoints end-to-end  
- **Validation Tests**: Test input validation logic
- **Error Handling**: Test error cases and edge conditions

## Performance Tests

For performance testing:

```bash
pytest tests/ -v --duration=10
```

Shows the 10 slowest tests.

## Continuous Testing

Use pytest-watch for continuous testing during development:

```bash
pip install pytest-watch
ptw tests/
```

## Pre-commit Tests

Tests are automatically run before commits if pre-commit hooks are installed:

```bash
pip install pre-commit
pre-commit install
```

---

**Last Updated**: 2026-02-01  
**Total Tests**: 77+  
**Pass Rate**: 100%
