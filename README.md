# Hopfield Assignment API

This project implements an API for solving assignment problems using a Hopfield neural network approach. The system consists of a Go-based API gateway that interfaces with a Python-based Hopfield solver service.

## Architecture

[Client] --> [API Gateway (Go)] --> [Hopfield Solver (Python)]

## Endpoints

### Health Endpoints
- `GET /health` - Check service health
- `GET /health/ready` - Check readiness 
- `GET /health/live` - Check liveness
- `GET /time` - Get current time in ISO 8601 format

### Assignment Endpoints
- `POST /solve` - Solve a single assignment problem
- `POST /solve/batch` - Solve multiple assignment problems in batch

## Getting Started

### Prerequisites
- Go 1.21+
- Docker (for containerized deployment)

### Running Locally
