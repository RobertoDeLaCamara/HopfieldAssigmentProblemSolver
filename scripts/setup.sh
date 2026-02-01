#!/bin/bash

# Initial setup script for Hopfield Assignment Problem Solver

set -e

echo "🚀 Setting up Hopfield Assignment Problem Solver..."

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored messages
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

print_status "Docker and Docker Compose are installed ✓"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    print_status "Creating .env file from env.example..."
    cp env.example .env
    print_warning "Please review and modify the .env file as needed"
else
    print_status ".env file already exists ✓"
fi

# Create necessary directories
print_status "Creating necessary directories..."
mkdir -p logs
mkdir -p nginx/ssl

# Generate self-signed SSL certificates for development
if [ ! -f nginx/ssl/cert.pem ] || [ ! -f nginx/ssl/key.pem ]; then
    print_status "Generating self-signed SSL certificates for development..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout nginx/ssl/key.pem \
        -out nginx/ssl/cert.pem \
        -subj "/C=ES/ST=Madrid/L=Madrid/O=HopfieldSolver/CN=localhost" 2>/dev/null || {
        print_warning "Could not generate SSL certificates. This is normal if OpenSSL is not installed."
        print_warning "Services will work without HTTPS in development mode."
    }
else
    print_status "SSL certificates already exist ✓"
fi

# Build Docker images
print_status "Building Docker images..."
docker-compose build

# Verify that images were built successfully
if [ $? -eq 0 ]; then
    print_status "Docker images built successfully ✓"
else
    print_error "Error building Docker images"
    exit 1
fi

# Show usage information
echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "To start the services:"
echo "  make up"
echo "  # or"
echo "  docker-compose up -d"
echo ""
echo "To verify everything works:"
echo "  make health"
echo "  # or"
echo "  curl http://localhost:8080/health"
echo ""
echo "To test the API:"
echo "  make test-api"
echo "  # or"
echo "  curl -X POST http://localhost:8080/solve \\"
echo "    -H \"Content-Type: application/json\" \\"
echo "    -d '{\"cost_matrix\": [[1,2],[3,4]]}'"
echo ""
echo "To view logs:"
echo "  make logs"
echo "  # or"
echo "  docker-compose logs -f"
echo ""
echo "To stop the services:"
echo "  make down"
echo "  # or"
echo "  docker-compose down"
echo ""
echo "For more available commands:"
echo "  make help"
echo ""
print_status "Ready to use! 🚀"
