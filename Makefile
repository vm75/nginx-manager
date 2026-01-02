.PHONY: help build run stop logs clean restart shell test

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Docker image
	docker-compose build

up: ## Start containers in detached mode
	docker-compose up -d

run: up ## Alias for up

stop: ## Stop containers
	docker-compose down

logs: ## Show container logs
	docker-compose logs -f

restart: ## Restart containers
	docker-compose restart

shell: ## Open shell in running container
	docker-compose exec nginx-editor sh

status: ## Show status of services
	docker-compose exec nginx-editor supervisorctl status

nginx-test: ## Test nginx configuration
	docker-compose exec nginx-editor nginx -t

nginx-reload: ## Reload nginx
	docker-compose exec nginx-editor nginx -s reload

fail2ban-status: ## Show fail2ban status
	docker-compose exec nginx-editor fail2ban-client status

clean: ## Stop and remove containers, networks, and volumes
	docker-compose down -v

rebuild: clean build up ## Clean rebuild and start

dev: ## Build and run for development
	docker-compose up --build

prune: ## Remove unused Docker resources
	docker system prune -af --volumes
