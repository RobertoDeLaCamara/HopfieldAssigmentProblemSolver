
### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GIN_MODE` | Gin mode ("release" or "debug") | "debug" |
| `HOPFIELD_SERVICE_URL` | URL of the Hopfield service | "http://hopfield-service:5000" |

## API Endpoints

### Health Checks
- `GET /health` - General health check
- `GET /health/ready` - Readiness check
- `GET /health/live` - Liveness check

### Assignment Problems
- `POST /solve` - Solve a single assignment problem
- `POST /solve/batch` - Solve multiple assignment problems in batch

### Example Request

#### Single Assignment Problem
