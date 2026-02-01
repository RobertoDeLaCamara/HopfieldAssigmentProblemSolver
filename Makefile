# Makefile for Hopfield Assignment Problem Solver

.PHONY: help build up down logs clean test lint format

# Variables
DOCKER_COMPOSE = docker-compose
PYTHON = python3
GO = go

# Colors for output
GREEN = \033[0;32m
YELLOW = \033[1;33m
RED = \033[0;31m
NC = \033[0m # No Color

help: ## Show this help
	@echo "$(GREEN)Hopfield Assignment Problem Solver$(NC)"
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

build: ## Build all Docker images
	@echo "$(GREEN)Building Docker images...$(NC)"
	$(DOCKER_COMPOSE) build

up: ## Start all services
	@echo "$(GREEN)Starting services...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Services started. Check with: make status$(NC)"

down: ## Stop all services
	@echo "$(YELLOW)Stopping services...$(NC)"
	$(DOCKER_COMPOSE) down

restart: ## Restart all services
	@echo "$(YELLOW)Restarting services...$(NC)"
	$(DOCKER_COMPOSE) restart

status: ## Show service status
	@echo "$(GREEN)Service status:$(NC)"
	$(DOCKER_COMPOSE) ps

logs: ## Show logs from all services
	@echo "$(GREEN)Showing logs:$(NC)"
	$(DOCKER_COMPOSE) logs -f

logs-api: ## Show API Gateway logs
	@echo "$(GREEN)API Gateway logs:$(NC)"
	$(DOCKER_COMPOSE) logs -f api-gateway

logs-hopfield: ## Show Hopfield service logs
	@echo "$(GREEN)Hopfield service logs:$(NC)"
	$(DOCKER_COMPOSE) logs -f hopfield-service

clean: ## Clean containers, images and volumes
	@echo "$(RED)Cleaning Docker resources...$(NC)"
	$(DOCKER_COMPOSE) down -v --rmi all
	docker system prune -f

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	$(DOCKER_COMPOSE) exec hopfield-service python -m pytest tests/ -v
	$(DOCKER_COMPOSE) exec api-gateway go test ./...

test-unit: ## Run unit tests only
	@echo "$(GREEN)Running unit tests...$(NC)"
	$(DOCKER_COMPOSE) exec hopfield-service python -m pytest tests/ -v --tb=short

test-integration: ## Run integration tests only
	@echo "$(GREEN)Running integration tests...$(NC)"
	python3 -m pytest tests/integration_test.py -v

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(DOCKER_COMPOSE) exec hopfield-service python -m pytest tests/ --cov=src --cov-report=html --cov-report=term

test-local: ## Run tests locally (without Docker)
	@echo "$(GREEN)Running tests locally...$(NC)"
	cd hopfield && $(PYTHON) -m pytest tests/ -v
	cd api && $(GO) test ./...

lint: ## Run linters
	@echo "$(GREEN)Running linters...$(NC)"
	cd hopfield && flake8 src/
	cd api && golangci-lint run

format: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	cd hopfield && black src/
	cd api && go fmt ./...

dev: ## Development mode (start services with hot reload)
	@echo "$(GREEN)Starting development mode...$(NC)"
	$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml up

prod: ## Production mode
	@echo "$(GREEN)Starting production mode...$(NC)"
	$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml up -d

health: ## Check service health
	@echo "$(GREEN)Checking service health...$(NC)"
	@echo "API Gateway:"
	@curl -s http://localhost:8080/health | jq . || echo "Error connecting to API Gateway"
	@echo "Hopfield Service:"
	@curl -s http://localhost:5000/health | jq . || echo "Error connecting to Hopfield Service"

test-api: ## Test the API with an example
	@echo "$(GREEN)Testing the API...$(NC)"
	@curl -X POST http://localhost:8080/solve \
		-H "Content-Type: application/json" \
		-d '{"cost_matrix": [[1,2],[3,4]]}' | jq .

install-deps: ## Install dependencies locally
	@echo "$(GREEN)Installing dependencies...$(NC)"
	cd hopfield && pip install -r requirements.txt
	cd api && go mod download

setup: ## Initial project setup
	@echo "$(GREEN)Setting up project...$(NC)"
	cp env.example .env
	$(MAKE) install-deps
	$(MAKE) build
	@echo "$(GREEN)Setup completed. Run 'make up' to start services.$(NC)"

backup: ## Create configuration backup
	@echo "$(GREEN)Creating backup...$(NC)"
	tar -czf backup-$(shell date +%Y%m%d-%H%M%S).tar.gz \
		--exclude='.git' \
		--exclude='node_modules' \
		--exclude='__pycache__' \
		--exclude='*.pyc' \
		--exclude='.env' \
		.

restore: ## Restore from backup (specify BACKUP=file.tar.gz)
	@if [ -z "$(BACKUP)" ]; then \
		echo "$(RED)Error: Specify BACKUP=file.tar.gz$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Restoring from $(BACKUP)...$(NC)"
	tar -xzf $(BACKUP)

# Specific development commands
dev-python: ## Run Hopfield service locally
	@echo "$(GREEN)Running Hopfield service locally...$(NC)"
	cd hopfield && $(PYTHON) src/api_server.py

dev-go: ## Run API Gateway locally
	@echo "$(GREEN)Running API Gateway locally...$(NC)"
	cd api && $(GO) run cmd/main.go

# Monitoring commands
monitor: ## Monitor system resources
	@echo "$(GREEN)Monitoring resources...$(NC)"
	watch -n 1 'docker stats --no-stream'

# Database commands (if added in the future)
# db-migrate: ## Run database migrations
# 	$(DOCKER_COMPOSE) exec api-gateway ./migrate up

# db-rollback: ## Rollback database migrations
# 	$(DOCKER_COMPOSE) exec api-gateway ./migrate down

# SSL commands (for production)
ssl-generate: ## Generate self-signed SSL certificates for development
	@echo "$(GREEN)Generating SSL certificates...$(NC)"
	mkdir -p nginx/ssl
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
		-keyout nginx/ssl/key.pem \
		-out nginx/ssl/cert.pem \
		-subj "/C=ES/ST=Madrid/L=Madrid/O=HopfieldSolver/CN=localhost"

# Default command
.DEFAULT_GOAL := help
