# Contributing to Hopfield Assignment Problem Solver

Thank you for your interest in contributing to the Hopfield Assignment Problem Solver! We welcome contributions from the community.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)

## 🤝 Code of Conduct

This project follows a code of conduct to ensure a welcoming environment for all contributors. By participating, you agree to:

- Be respectful and inclusive
- Focus on constructive feedback
- Accept responsibility for mistakes
- Show empathy towards other contributors
- Help create a positive community

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or later
- Python 3.8 or later
- Docker 20.10+
- Docker Compose 2.0+
- Git

### Quick Setup

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/your-username/HopfieldAssigmentProblemSolver.git
   cd HopfieldAssigmentProblemSolver
   ```
3. Set up the development environment:
   ```bash
   ./scripts/setup.sh
   ```
4. Start development services:
   ```bash
   make dev
   ```

## 🛠️ Development Setup

### Environment Configuration

1. Copy the example environment file:
   ```bash
   cp env.example .env
   ```

2. Modify `.env` for your development needs:
   ```bash
   ENV=development
   GIN_MODE=debug
   LOG_LEVEL=debug
   ```

### Installing Dependencies

```bash
# Install all dependencies
make install-deps

# Or install individually:
# Python dependencies
cd hopfield && pip install -r requirements.txt

# Go dependencies
cd api && go mod download
```

### Running Services Locally

```bash
# Run Hopfield service
make dev-python

# Run API Gateway (in another terminal)
make dev-go

# Or run both with Docker
make dev
```

## 📁 Project Structure

```
HopfieldAssigmentProblemSolver/
├── api/                    # Go API Gateway
│   ├── cmd/
│   │   └── main.go        # Application entry point
│   ├── internal/
│   │   ├── handlers/      # HTTP handlers
│   │   └── models/        # Data models
│   └── pkg/
│       └── middleware/    # Custom middleware
├── hopfield/              # Python Hopfield Solver
│   ├── src/
│   │   ├── api_server.py # Flask API server
│   │   └── hopfield_solver.py # Core algorithm
│   └── tests/             # Unit tests
├── nginx/                 # Reverse proxy configuration
├── docs/                  # Documentation
├── scripts/               # Setup and utility scripts
├── tests/                 # Integration tests
├── docker-compose.yml     # Docker services
├── Makefile              # Build automation
└── README.md             # Project documentation
```

## 🔄 Development Workflow

### 1. Choose an Issue

- Check the [Issues](https://github.com/your-repo/issues) page
- Look for issues labeled `good first issue` or `help wanted`
- Comment on the issue to indicate you're working on it

### 2. Create a Branch

```bash
# Create and switch to a new branch
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 3. Make Changes

- Write clear, focused commits
- Test your changes thoroughly
- Follow the coding standards

### 4. Test Your Changes

```bash
# Run all tests
make test

# Run tests locally
make test-local

# Test API functionality
make test-api
```

### 5. Update Documentation

- Update README.md if needed
- Update API documentation in `docs/API.md`
- Add code comments for complex logic

## 🧪 Testing

### Test Coverage

The project maintains comprehensive test coverage with:
- **29 unit tests** for individual components and algorithms
- **13 integration tests** for end-to-end functionality
- **100% test pass rate** across all test suites

### Running Tests

```bash
# Run all tests in Docker (recommended)
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run tests with coverage report
make test-coverage

# Run tests locally (without Docker)
make test-local

# Test API functionality manually
make test-api
```

### Test Structure

- **Unit Tests** (`hopfield/tests/`): Test Hopfield algorithm, API validation, error handling
- **Integration Tests** (`tests/integration_test.py`): Test service communication, performance, batch processing
- **API Tests**: Test REST endpoints, health checks, error responses

### Writing Tests

#### Python Tests (Hopfield Service)

- Place test files in `hopfield/tests/`
- Use `pytest` framework
- Follow naming convention: `test_*.py`

```python
import pytest
from src.hopfield_solver import HopfieldAssignmentSolver

def test_solver_initialization():
    solver = HopfieldAssignmentSolver()
    assert solver is not None

def test_solve_simple_problem():
    solver = HopfieldAssignmentSolver()
    cost_matrix = [[1, 2], [3, 4]]
    assignments, cost, iterations = solver.solve(cost_matrix)
    assert len(assignments) == 2
    assert cost >= 0
    assert iterations > 0
```
    assert 'assignments' in result
    assert 'total_cost' in result
```

#### Go Tests (API Gateway)

- Place test files alongside source code with `_test.go` suffix
- Use Go's built-in testing framework

```go
package handlers

import (
    "testing"
    "net/http/httptest"
)

func TestHealthCheck(t *testing.T) {
    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()

    HealthCheck(w, req)

    if w.Code != 200 {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

### Test Coverage

Aim for high test coverage, especially for:
- Core algorithm logic
- API endpoints
- Error handling
- Edge cases

## 💅 Code Style

### Python (Hopfield Service)

- Follow PEP 8 style guide
- Use `black` for code formatting
- Use `flake8` for linting
- Maximum line length: 88 characters

```bash
# Format code
make format

# Lint code
make lint
```

### Go (API Gateway)

- Follow standard Go formatting (`go fmt`)
- Use `golangci-lint` for linting
- Follow Go naming conventions
- Use meaningful variable and function names

```bash
# Format Go code
cd api && go fmt ./...

# Lint Go code
cd api && golangci-lint run
```

### General Guidelines

- Write clear, self-documenting code
- Add comments for complex algorithms
- Use consistent naming conventions
- Keep functions small and focused
- Handle errors appropriately

## 📝 Submitting Changes

### Commit Guidelines

- Write clear, concise commit messages
- Use present tense ("Add feature" not "Added feature")
- Reference issue numbers when applicable

```bash
# Good commit messages
git commit -m "Add rate limiting middleware"
git commit -m "Fix division by zero in cost calculation (#123)"
git commit -m "Update API documentation with new endpoints"
```

### Pull Request Process

1. **Ensure your branch is up to date**:
   ```bash
   git fetch origin
   git rebase origin/main
   ```

2. **Run tests and linting**:
   ```bash
   make test
   make lint
   make format
   ```

3. **Create a Pull Request**:
   - Use a descriptive title
   - Provide detailed description of changes
   - Reference related issues
   - Add screenshots for UI changes

4. **Code Review**:
   - Address review comments
   - Make requested changes
   - Ensure CI checks pass

5. **Merge**:
   - Squash commits if requested
   - Delete branch after merge

## 🐛 Reporting Issues

### Bug Reports

When reporting bugs, please include:

- **Description**: Clear description of the issue
- **Steps to reproduce**: Step-by-step instructions
- **Expected behavior**: What should happen
- **Actual behavior**: What actually happens
- **Environment**: OS, Go/Python versions, Docker versions
- **Logs**: Relevant error messages or logs
- **Screenshots**: If applicable

### Feature Requests

For feature requests, please include:

- **Description**: What feature you'd like to see
- **Use case**: Why this feature would be useful
- **Implementation ideas**: If you have thoughts on how to implement it
- **Alternatives**: Other solutions you've considered

## 📚 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Python Documentation](https://docs.python.org/3/)
- [Gin Web Framework](https://gin-gonic.com/)
- [Flask Documentation](https://flask.palletsprojects.com/)
- [Docker Documentation](https://docs.docker.com/)

## 🙏 Recognition

Contributors will be recognized in the project README and release notes. Thank you for helping make this project better!

---

**Happy contributing! 🎉**