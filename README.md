# Hopfield Assignment Problem Solver

A service that solves the optimal resource-task assignment problem using Hopfield neural networks.

## Architecture

- **Backend API**: Implemented in Go for high performance
- **Hopfield Algorithm**: Implemented in Python for scientific flexibility
- **Deployment**: Docker containers for easy distribution

## Project Structure

```
├── api/                    # REST API in Go
│   ├── cmd/
│   ├── internal/
│   ├── pkg/
│   └── Dockerfile
├── hopfield/              # Hopfield algorithm in Python
│   ├── src/
│   ├── requirements.txt
│   └── Dockerfile
├── docker-compose.yml     # Service orchestration
├── .gitignore
└── README.md
```

## Usage

```bash
# Start all services
docker-compose up -d

# Solve an assignment problem
curl -X POST http://localhost:8080/solve \
  -H "Content-Type: application/json" \
  -d '{"cost_matrix": [[1,2,3],[4,5,6],[7,8,9]]}'
```

## Requirements

- Docker
- Docker Compose