# API Documentation

## Overview

The Hopfield Assignment Problem Solver service exposes a REST API to solve optimal assignment problems using Hopfield neural networks.

## Endpoints

### Service Health

#### GET /health
Checks the general status of the service.

**Response:**
```json
{
  "status": "healthy",
  "service": "hopfield-assignment-api",
  "version": "1.0.0"
}
```

#### GET /health/ready
Checks if the service is ready to receive traffic.

#### GET /health/live
Checks if the service is alive.

### Assignment Problem Solving

#### POST /api/v1/solve
Solves an individual assignment problem.

**Request Body:**
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

**Response (Success):**
```json
{
  "success": true,
  "result": {
    "assignments": [1, 2, 0, 3],
    "total_cost": 13.0,
    "iterations": 156,
    "cost_matrix": [
      [9, 2, 7, 8],
      [6, 4, 3, 7],
      [5, 8, 1, 8],
      [7, 6, 9, 4]
    ]
  }
}
```

**Response (Error):**
```json
{
  "success": false,
  "error": "The cost matrix must be square"
}
```

#### POST /api/v1/solve/batch
Solves multiple assignment problems in batch.

**Request Body:**
```json
{
  "problems": [
    {
      "id": "problem_1",
      "cost_matrix": [
        [1, 2],
        [3, 4]
      ]
    },
    {
      "id": "problem_2",
      "cost_matrix": [
        [5, 6, 7],
        [8, 9, 10],
        [11, 12, 13]
      ]
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "results": [
    {
      "id": "problem_1",
      "success": true,
      "result": {
        "assignments": [0, 1],
        "total_cost": 5.0,
        "iterations": 45,
        "cost_matrix": [[1, 2], [3, 4]]
      }
    },
    {
      "id": "problem_2",
      "success": true,
      "result": {
        "assignments": [0, 1, 2],
        "total_cost": 30.0,
        "iterations": 78,
        "cost_matrix": [[5, 6, 7], [8, 9, 10], [11, 12, 13]]
      }
    }
  ]
}
```

## HTTP Status Codes

- `200 OK`: Successful request
- `400 Bad Request`: Request error (invalid format, validation failed)
- `500 Internal Server Error`: Internal server error

## Validations

### Cost Matrix
- Must be a square matrix (n×n)
- All elements must be non-negative numbers
- Cannot be empty
- Recommended maximum: 50×50 (for optimal performance)

### Rate Limits
- API Gateway: 10 requests per second
- Hopfield Service: 5 requests per second

## Usage Examples

### cURL

```bash
# Solve a simple problem
curl -X POST http://localhost:8080/api/v1/solve \
  -H "Content-Type: application/json" \
  -d '{
    "cost_matrix": [
      [1, 2, 3],
      [4, 5, 6],
      [7, 8, 9]
    ]
  }'

# Check service health
curl http://localhost:8080/health
```

### Python

```python
import requests
import json

# Configure base URL
base_url = "http://localhost:8080"

# Solve a problem
cost_matrix = [
    [9, 2, 7, 8],
    [6, 4, 3, 7],
    [5, 8, 1, 8],
    [7, 6, 9, 4]
]

response = requests.post(
    f"{base_url}/api/v1/solve",
    json={"cost_matrix": cost_matrix}
)

if response.status_code == 200:
    result = response.json()
    print(f"Assignments: {result['result']['assignments']}")
    print(f"Total cost: {result['result']['total_cost']}")
else:
    print(f"Error: {response.json()['error']}")
```

### JavaScript

```javascript
const solveAssignment = async (costMatrix) => {
  try {
    const response = await fetch('http://localhost:8080/api/v1/solve', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ cost_matrix: costMatrix })
    });
    
    const result = await response.json();
    
    if (result.success) {
      console.log('Assignments:', result.result.assignments);
      console.log('Total cost:', result.result.total_cost);
    } else {
      console.error('Error:', result.error);
    }
  } catch (error) {
    console.error('Network error:', error);
  }
};

// Usage example
const costMatrix = [
  [1, 2, 3],
  [4, 5, 6],
  [7, 8, 9]
];

solveAssignment(costMatrix);
```

## Monitoring and Logs

The service generates structured logs in JSON format that include:
- Request timestamp
- HTTP method and route
- Response status code
- Latency time
- Client IP
- User-Agent

Logs are available in Docker containers and can be collected by monitoring systems like ELK Stack, Prometheus, etc.
