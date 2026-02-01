# Hopfield Assignment Problem Solver

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/)
[![Python Version](https://img.shields.io/badge/python-3.11+-blue.svg)](https://python.org/)

A high-performance API service for solving assignment problems using Hopfield neural networks. The system consists of a Go-based API gateway that interfaces with a Python-based Hopfield solver service, all containerized with Docker for easy deployment.

## 🏗️ Architecture

```
[Client] --> [API Gateway (Go)] --> [Hopfield Solver (Python)]
```

- **API Gateway**: RESTful API built with Gin framework in Go
- **Hopfield Solver**: Neural network implementation in Python using Flask
- **Nginx**: Reverse proxy with load balancing and SSL termination
- **Docker**: Containerized deployment with Docker Compose

## 🚀 Quick Start

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 2GB RAM minimum
- 1GB disk space
- Python 3.11+ (for local development)
- Go 1.21+ (for local development)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd HopfieldAssigmentProblemSolver
   ```

2. **Setup environment**
   ```bash
   # Using the setup script
   ./scripts/setup.sh

   # Or manually
   cp env.example .env
   make install-deps
   make build
   ```

3. **Start services**
   ```bash
   make up
   # Or: docker-compose up -d
   ```

4. **Verify installation**
   ```bash
   make health
   # Or: curl http://localhost:8080/health
   ```

## 📖 API Usage

The API is accessible through multiple endpoints:

- **Direct API Gateway**: `http://localhost:8080` (for development)
- **Through Nginx**: `http://localhost/api/` (for production with load balancing)

### Solve Single Assignment Problem

```bash
# Direct access
curl -X POST http://localhost:8080/solve \
  -H "Content-Type: application/json" \
  -d '{
    "cost_matrix": [
      [9, 2, 7, 8],
      [6, 4, 3, 7],
      [5, 8, 1, 8],
      [7, 6, 9, 4]
    ]
  }'

# Through Nginx
curl -X POST http://localhost/api/solve \
  -H "Content-Type: application/json" \
  -d '{
    "cost_matrix": [
      [9, 2, 7, 8],
      [6, 4, 3, 7],
      [5, 8, 1, 8],
      [7, 6, 9, 4]
    ]
  }'
```

### Solve Multiple Problems (Batch)

```bash
# Direct access
curl -X POST http://localhost:8080/solve/batch \
  -H "Content-Type: application/json" \
  -d '{
    "problems": [
      {
        "id": "problem_1",
        "cost_matrix": [[1, 2], [3, 4]]
      },
      {
        "id": "problem_2",
        "cost_matrix": [[5, 6, 7], [8, 9, 10], [11, 12, 13]]
      }
    ]
  }'

# Through Nginx
curl -X POST http://localhost/api/solve/batch \
  -H "Content-Type: application/json" \
  -d '{...}'
```

### Health Check

```bash
# Direct access
curl http://localhost:8080/health

# Through Nginx
curl http://localhost/api/health
```

## 🧪 Testing

The project includes comprehensive test suites to ensure reliability and correctness.

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run tests with coverage
make test-coverage
```

### Test Structure

- **Unit Tests** (29 tests): Test individual components like the Hopfield algorithm, API endpoints, and validation logic
- **Integration Tests** (13 tests): Test end-to-end functionality, service communication, and performance
- **Test Coverage**: All critical paths are covered with automated tests

### Manual Testing

```bash
# Test API endpoints manually
curl -X POST http://localhost:8080/solve \
  -H "Content-Type: application/json" \
  -d '{"cost_matrix": [[1,2],[3,4]]}'

# Test batch processing
curl -X POST http://localhost:8080/solve/batch \
  -H "Content-Type: application/json" \
  -d '{"problems": [{"id": "test", "cost_matrix": [[1,2],[3,4]]}]}'
```

## 🛠️ Development

### Local Development Setup

1. **Install dependencies**
   ```bash
   # Python dependencies
   cd hopfield && pip install -r requirements.txt

   # Go dependencies
   cd api && go mod download
   ```

2. **Run services locally**
   ```bash
   # Terminal 1: Hopfield service
   make dev-python

   # Terminal 2: API Gateway
   make dev-go
   ```

### Testing

```bash
# Run all tests
make test

# Run tests locally
make test-local

# Run specific service tests
make test-api
```

### Code Quality

```bash
# Lint code
make lint

# Format code
make format
```

### Available Commands

```bash
make help          # Show all available commands
make build         # Build Docker images
make up            # Start services
make down          # Stop services
make logs          # View logs
make clean         # Clean Docker resources
make restart       # Restart services
make status        # Show service status
```

## 📚 Documentation

- [API Documentation](docs/API.md) - Complete API reference with examples
- [Deployment Guide](docs/DEPLOYMENT.md) - Production deployment instructions
- [Contributing Guidelines](CONTRIBUTING.md) - Development and contribution guidelines

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ENV` | Execution environment | `development` |
| `PORT` | API Gateway port | `8080` |
| `GIN_MODE` | Gin framework mode | `debug` |
| `HOPFIELD_SERVICE_URL` | Hopfield service URL | `http://hopfield-service:5000` |
| `LOG_LEVEL` | Logging level | `info` |
| `API_RATE_LIMIT` | API rate limit | `10r/s` |
| `HOPFIELD_RATE_LIMIT` | Hopfield service rate limit | `5r/s` |

### Docker Compose Files

- `docker-compose.yml` - Base configuration
- `docker-compose.dev.yml` - Development overrides
- `docker-compose.prod.yml` - Production configuration

## 📊 Monitoring

### Health Endpoints

- `GET /health` - General health check
- `GET /health/ready` - Readiness check
- `GET /health/live` - Liveness check
- `GET /time` - Current server time

### Logs

```bash
# View all logs
make logs

# View specific service logs
make logs-api
make logs-hopfield

# Follow logs in real-time
docker-compose logs -f
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on:

- Setting up development environment
- Code style and standards
- Testing requirements
- Pull request process

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Hopfield neural network algorithm implementation
- Gin web framework for Go
- Flask microframework for Python
- Docker for containerization

## 📞 Support

For questions, issues, or contributions:

1. Check the [documentation](docs/)
2. Open an issue on GitHub
3. Review existing issues and pull requests

## 📋 Changelog

See [CHANGELOG.md](CHANGELOG.md) for a complete list of changes and version history.

---

**Happy optimizing! 🎯**
