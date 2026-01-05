.PHONY: help build run down up logs clean restart shell test

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Docker image
	docker compose build

up: ## Start containers in detached mode
	docker compose up -d

down: ## Stop containers
	docker compose down

run: up ## Alias for up

logs: ## Show container logs
	docker compose logs -f

restart: ## Restart containers
	docker compose restart

shell: ## Open shell in running container
	docker compose exec server-manager sh

status: ## Show status of services
	docker compose exec server-manager supervisorctl status

nginx-test: ## Test nginx configuration
	docker compose exec server-manager nginx -t

nginx-reload: ## Reload nginx
	docker compose exec server-manager nginx -s reload

fail2ban-status: ## Show fail2ban status
	docker compose exec server-manager fail2ban-client status

clean: ## Stop and remove containers, networks, and volumes
	docker compose down -v

rebuild: clean build up ## Clean rebuild and start

build-restart: build down up

dev: ## Build and run for development
	docker compose up --build

prune: ## Remove unused Docker resources
	docker system prune -af --volumes
