.PHONY: help build clean dev deploy-dev deploy-prod check-deps test-api test-frontend test-backend test-all test-coverage seed clear logs-webhook logs-stats unlock

# Variables
API_DEV := https://api.dev.mundotalendo.com.br
API_PROD := https://api.mundotalendo.com.br
REGION := us-east-2

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help
	@echo "$(GREEN)Mundo Tá Lendo 2026 - Available commands:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

# Build
build: ## Build all Go functions
	@echo "$(GREEN)Building Go functions...$(NC)"
	@cd packages/functions/types && go build .
	@cd packages/functions/webhook && go build .
	@cd packages/functions/stats && go build .
	@cd packages/functions/seed && go build .
	@cd packages/functions/clear && go build .
	@echo "$(GREEN)Build completed!$(NC)"

tidy: ## Update Go dependencies
	@echo "$(GREEN)Updating Go dependencies...$(NC)"
	@cd packages/functions/webhook && go mod tidy
	@cd packages/functions/stats && go mod tidy
	@cd packages/functions/seed && go mod tidy
	@cd packages/functions/clear && go mod tidy
	@echo "$(GREEN)Dependencies updated!$(NC)"

clean: ## Clean builds and cache
	@echo "$(YELLOW)Cleaning builds...$(NC)"
	@rm -rf .sst .open-next .next
	@find packages/functions -type f -name "webhook" -o -name "stats" -o -name "seed" -o -name "clear" | xargs rm -f
	@echo "$(GREEN)Cleanup completed!$(NC)"

# Deploy
unlock: ## Unlock stuck deployment
	@echo "$(YELLOW)Unlocking deployment...$(NC)"
	@npx sst unlock --stage dev

deploy-dev: ## Deploy to dev environment and fix env vars
	@echo "$(GREEN)Deploying to DEV...$(NC)"
	@npx sst deploy --stage dev
	@echo "\n$(YELLOW)Fixing Lambda environment variables (SST bug workaround)...$(NC)"
	@$(MAKE) fix-env

deploy-prod: ## Deploy to prod environment and fix env vars
	@echo "$(RED)Deploying to PRODUCTION...$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		npx sst deploy --stage prod; \
		echo "\n$(YELLOW)Fixing Lambda environment variables (SST bug workaround)...$(NC)"; \
		$(MAKE) fix-env; \
	else \
		echo "Deploy cancelled."; \
	fi

remove-dev: ## Remove dev stack
	@echo "$(RED)Removing DEV stack...$(NC)"
	@npx sst remove --stage dev

# Development
dev: ## Start local Next.js server
	@echo "$(GREEN)Starting local server...$(NC)"
	@npm run dev:local

# API Testing (requires API key)
get-api-key: ## Get first active API key for testing
	@DATA_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-dev-DataTable`)]' --output text); \
	aws dynamodb scan --region $(REGION) --table-name $$DATA_TABLE \
		--filter-expression "begins_with(PK, :pk) AND #active = :active" \
		--expression-attribute-names '{"#active":"active"}' \
		--expression-attribute-values '{":pk":{"S":"APIKEY#"},":active":{"BOOL":true}}' \
		--query 'Items[0].key.S' --output text

# Unit Tests
check-deps: ## Check if test dependencies are installed
	@if ! npm list jest > /dev/null 2>&1; then \
		echo "$(YELLOW)Installing test dependencies...$(NC)"; \
		npm install --include=dev --legacy-peer-deps || npm install --include=dev --force; \
	fi

test-frontend: check-deps ## Run frontend unit tests (Jest)
	@echo "$(GREEN)Running frontend tests...$(NC)"
	@npm test

test-frontend-watch: check-deps ## Run frontend tests in watch mode
	@echo "$(GREEN)Running frontend tests in watch mode...$(NC)"
	@npm run test:watch

test-backend: ## Run backend unit tests (Go)
	@echo "$(GREEN)Running backend tests...$(NC)"
	@cd packages/functions && go test ./... -v

test-backend-coverage: ## Run backend tests with coverage
	@echo "$(GREEN)Running backend tests with coverage...$(NC)"
	@cd packages/functions && go test ./... -cover

test-all: check-deps ## Run all unit tests (frontend + backend)
	@echo "$(GREEN)Running all unit tests...$(NC)"
	@echo "\n$(YELLOW)=== Frontend Tests ===$(NC)"
	@npm test
	@echo "\n$(YELLOW)=== Backend Tests ===$(NC)"
	@cd packages/functions && go test ./... -v

test-coverage: check-deps ## Generate coverage reports for all tests
	@echo "$(GREEN)Generating coverage reports...$(NC)"
	@echo "\n$(YELLOW)=== Frontend Coverage ===$(NC)"
	@npm run test:coverage
	@echo "\n$(YELLOW)=== Backend Coverage ===$(NC)"
	@cd packages/functions && go test ./... -coverprofile=coverage.out
	@cd packages/functions && go tool cover -html=coverage.out -o coverage.html
	@echo "\n$(GREEN)Coverage reports generated!$(NC)"
	@echo "Backend HTML report: packages/functions/coverage.html"

test-bench: ## Run Go benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@cd packages/functions && go test ./... -bench=. -benchmem

# API Integration Tests
test-api: ## Test dev API endpoints
	@echo "$(GREEN)Testing DEV API...$(NC)"
	@API_KEY=$$($(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	echo "$(YELLOW)Using API key: $$API_KEY$(NC)"; \
	echo "\n$(YELLOW)GET /stats:$(NC)"; \
	curl -s $(API_DEV)/stats -H "X-API-Key: $$API_KEY" | jq .; \
	echo "\n$(YELLOW)Available endpoints:$(NC)"; \
	echo "  - GET  $(API_DEV)/stats"; \
	echo "  - POST $(API_DEV)/webhook"; \
	echo "  - POST $(API_DEV)/test/seed"; \
	echo "  - POST $(API_DEV)/clear"

seed: ## Populate database with random data (count=20)
	@echo "$(GREEN)Populating database...$(NC)"
	@API_KEY=$$($(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	curl -s -X POST $(API_DEV)/test/seed \
		-H "Content-Type: application/json" \
		-H "X-API-Key: $$API_KEY" \
		-d '{"count": 20}' | jq .

clear: ## Clear all database tables
	@echo "$(RED)Clearing database...$(NC)"
	@API_KEY=$$($(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	curl -s -X POST $(API_DEV)/clear \
		-H "X-API-Key: $$API_KEY" | jq .

webhook-test: ## Test webhook with sample payload
	@echo "$(GREEN)Testing webhook...$(NC)"
	@API_KEY=$$($(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	curl -s -X POST $(API_DEV)/webhook \
		-H "Content-Type: application/json" \
		-H "X-API-Key: $$API_KEY" \
		-d '{ \
			"perfil": {"nome": "Test User", "link": "https://test.com"}, \
			"maratona": {"nome": "Test", "identificador": "maratona-lendo-paises"}, \
			"desafios": [{ \
				"descricao": "Brasil", \
				"categoria": "Janeiro", \
				"concluido": true, \
				"tipo": "leitura", \
				"vinculados": [{"progresso": 85, "updatedAt": "2024-12-16T10:00:00Z"}] \
			}] \
		}' | jq .

# Logs
logs-webhook: ## Show webhook Lambda logs
	@echo "$(GREEN)Webhook logs (last 5min):$(NC)"
	@WEBHOOK_FN=$$(aws lambda list-functions --region $(REGION) --query 'Functions[?contains(FunctionName, `mundotalendo-dev-ApiRouteBahoda`)].FunctionName' --output text); \
	if [ -n "$$WEBHOOK_FN" ]; then \
		aws logs tail /aws/lambda/$$WEBHOOK_FN --follow --region $(REGION) --since 5m 2>&1 || echo "$(YELLOW)Log group not found or no logs yet$(NC)"; \
	else \
		echo "$(RED)Webhook Lambda not found$(NC)"; \
	fi

logs-stats: ## Show stats Lambda logs
	@echo "$(GREEN)Stats logs (last 5min):$(NC)"
	@STATS_FN=$$(aws lambda list-functions --region $(REGION) --query 'Functions[?contains(FunctionName, `mundotalendo-dev-ApiRouteNodhex`)].FunctionName' --output text); \
	if [ -n "$$STATS_FN" ]; then \
		aws logs tail /aws/lambda/$$STATS_FN --follow --region $(REGION) --since 5m 2>&1 || echo "$(YELLOW)Log group not found or no logs yet$(NC)"; \
	else \
		echo "$(RED)Stats Lambda not found$(NC)"; \
	fi

# AWS Info
info: ## Show AWS resources information
	@echo "$(GREEN)AWS Resources - Stage: dev$(NC)"
	@echo "\n$(YELLOW)DynamoDB Tables:$(NC)"
	@aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-dev`)]' --output table
	@echo "\n$(YELLOW)Lambda Functions:$(NC)"
	@aws lambda list-functions --region $(REGION) --query 'Functions[?contains(FunctionName, `mundotalendo-dev`)].FunctionName' --output table
	@echo "\n$(YELLOW)API Gateway:$(NC)"
	@aws apigatewayv2 get-apis --region $(REGION) --query 'Items[?contains(Name, `mundotalendo`)].{Name:Name,Endpoint:ApiEndpoint}' --output table

env-webhook: ## Show webhook Lambda environment variables
	@echo "$(GREEN)Webhook Lambda Environment Variables:$(NC)"
	@WEBHOOK_FN=$$(aws lambda list-functions --region $(REGION) --query 'Functions[?contains(FunctionName, `mundotalendo-dev-ApiRouteBahoda`)].FunctionName' --output text); \
	echo "Function: $$WEBHOOK_FN"; \
	aws lambda get-function-configuration \
		--function-name $$WEBHOOK_FN \
		--region $(REGION) \
		--query 'Environment.Variables' --output json | jq .

list-lambdas: ## List all Lambda functions with their env vars
	@echo "$(GREEN)All Lambda Functions:$(NC)"
	@for fn in $$(aws lambda list-functions --region $(REGION) --query 'Functions[?contains(FunctionName, `mundotalendo-dev-ApiRoute`)].FunctionName' --output text); do \
		echo "\n$(YELLOW)$$fn:$(NC)"; \
		aws lambda get-function-configuration --function-name $$fn --region $(REGION) --query 'Environment.Variables' --output json | jq -c '. | {DataTable: .SST_Resource_DataTable_name}'; \
	done

check-failures: ## Show recent failures from Falhas table
	@echo "$(GREEN)Recent failures (last 10):$(NC)"
	@FALHAS_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-dev-DataTable`)]' --output text); \
	if [ -n "$$FALHAS_TABLE" ]; then \
		aws dynamodb scan --table-name $$FALHAS_TABLE --region $(REGION) --max-items 10 | jq -r '.Items[] | "\(.SK.S): \(.ErrorType.S) - \(.ErrorMessage.S)"'; \
	else \
		echo "$(RED)Falhas table not found$(NC)"; \
	fi

fix-env: ## Fix Lambda environment variables (SST bug workaround)
	@echo "$(YELLOW)Fixing Lambda environment variables...$(NC)"
	@DATA_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-dev-DataTable`)]' --output text); \
	echo "$(GREEN)Found table:$(NC)"; \
	echo "  DataTable: $$DATA_TABLE"; \
	echo "\n$(YELLOW)Updating Lambda functions...$(NC)"; \
	for fn in $$(aws lambda list-functions --region $(REGION) --query 'Functions[?contains(FunctionName, `mundotalendo-dev-ApiRoute`)].FunctionName' --output text); do \
		echo "  Updating $$fn..."; \
		aws lambda update-function-configuration \
			--function-name $$fn \
			--region $(REGION) \
			--environment "Variables={SST_Resource_DataTable_name=$$DATA_TABLE}" \
			--output text --query 'FunctionName' 2>&1 | grep -v "An error occurred" || true; \
	done; \
	echo "\n$(GREEN)Environment variables updated!$(NC)"

# API Key Management
create-api-key: ## Create new API key (make create-api-key name=myapp)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: Use 'make create-api-key name=yourname'$(NC)"; \
		exit 1; \
	fi
	@DATA_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-dev-DataTable`)]' --output text); \
	UUID=$$(uuidgen | tr '[:upper:]' '[:lower:]'); \
	DATE=$$(date +%Y-%m-%d); \
	API_KEY="$(name)-$$UUID-$$DATE"; \
	TIMESTAMP=$$(date -u +"%Y-%m-%dT%H:%M:%SZ"); \
	echo "$(GREEN)Creating API key...$(NC)"; \
	aws dynamodb put-item --region $(REGION) --table-name $$DATA_TABLE \
		--item '{"PK":{"S":"APIKEY#$(name)"},"SK":{"S":"KEY#'$$UUID'"},"name":{"S":"$(name)"},"key":{"S":"'$$API_KEY'"},"createdAt":{"S":"'$$TIMESTAMP'"},"active":{"BOOL":true}}' \
		--output text > /dev/null 2>&1; \
	echo "$(GREEN)API Key created:$(NC)"; \
	echo "$(YELLOW)$$API_KEY$(NC)"; \
	echo "\nAdd to your .env.local:"; \
	echo "NEXT_PUBLIC_API_KEY=$$API_KEY"

list-api-keys: ## List all API keys
	@echo "$(GREEN)Active API Keys:$(NC)"
	@DATA_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-dev-DataTable`)]' --output text); \
	aws dynamodb scan --region $(REGION) --table-name $$DATA_TABLE \
		--filter-expression "begins_with(PK, :pk)" \
		--expression-attribute-values '{":pk":{"S":"APIKEY#"}}' \
		--query 'Items[].{Name:name.S,Key:key.S,Created:createdAt.S,Active:active.BOOL}' \
		--output table

delete-api-key: ## Delete API key (make delete-api-key name=myapp)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: Use 'make delete-api-key name=yourname'$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Deleting API key for: $(name)$(NC)"
	@DATA_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-dev-DataTable`)]' --output text); \
	ITEMS=$$(aws dynamodb scan --region $(REGION) --table-name $$DATA_TABLE \
		--filter-expression "PK = :pk" \
		--expression-attribute-values '{":pk":{"S":"APIKEY#$(name)"}}' \
		--query 'Items[].SK.S' --output text); \
	for SK in $$ITEMS; do \
		aws dynamodb delete-item --region $(REGION) --table-name $$DATA_TABLE \
			--key '{"PK":{"S":"APIKEY#$(name)"},"SK":{"S":"'$$SK'"}}' \
			--output text > /dev/null 2>&1; \
		echo "$(GREEN)Deleted: $(name)$(NC)"; \
	done

# PROD API Key Management
create-api-key-prod: ## Create new API key in PROD (make create-api-key-prod name=myapp)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: Use 'make create-api-key-prod name=yourname'$(NC)"; \
		exit 1; \
	fi
	@DATA_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-prod-DataTable`)]' --output text); \
	UUID=$$(uuidgen | tr '[:upper:]' '[:lower:]'); \
	DATE=$$(date +%Y-%m-%d); \
	API_KEY="$(name)-$$UUID-$$DATE"; \
	TIMESTAMP=$$(date -u +"%Y-%m-%dT%H:%M:%SZ"); \
	echo "$(RED)Creating PROD API key...$(NC)"; \
	aws dynamodb put-item --region $(REGION) --table-name $$DATA_TABLE \
		--item '{"PK":{"S":"APIKEY#$(name)"},"SK":{"S":"KEY#'$$UUID'"},"name":{"S":"$(name)"},"key":{"S":"'$$API_KEY'"},"createdAt":{"S":"'$$TIMESTAMP'"},"active":{"BOOL":true}}' \
		--output text > /dev/null 2>&1; \
	echo "$(GREEN)PROD API Key created:$(NC)"; \
	echo "$(YELLOW)$$API_KEY$(NC)"

list-api-keys-prod: ## List all PROD API keys
	@echo "$(RED)PROD Active API Keys:$(NC)"
	@DATA_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-prod-DataTable`)]' --output text); \
	aws dynamodb scan --region $(REGION) --table-name $$DATA_TABLE \
		--filter-expression "begins_with(PK, :pk)" \
		--expression-attribute-values '{":pk":{"S":"APIKEY#"}}' \
		--query 'Items[].{Name:name.S,Key:key.S,Created:createdAt.S,Active:active.BOOL}' \
		--output table

delete-api-key-prod: ## Delete PROD API key (make delete-api-key-prod name=myapp)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: Use 'make delete-api-key-prod name=yourname'$(NC)"; \
		exit 1; \
	fi
	@echo "$(RED)Deleting PROD API key for: $(name)$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		DATA_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-prod-DataTable`)]' --output text); \
		ITEMS=$$(aws dynamodb scan --region $(REGION) --table-name $$DATA_TABLE \
			--filter-expression "PK = :pk" \
			--expression-attribute-values '{":pk":{"S":"APIKEY#$(name)"}}' \
			--query 'Items[].SK.S' --output text); \
		for SK in $$ITEMS; do \
			aws dynamodb delete-item --region $(REGION) --table-name $$DATA_TABLE \
				--key '{"PK":{"S":"APIKEY#$(name)"},"SK":{"S":"'$$SK'"}}' \
				--output text > /dev/null 2>&1; \
			echo "$(GREEN)Deleted: $(name)$(NC)"; \
		done; \
	else \
		echo "Deletion cancelled."; \
	fi

# Git
commit: ## Quick commit (make commit m="message")
	@if [ -z "$(m)" ]; then \
		echo "$(RED)Error: Use 'make commit m=\"your message\"'$(NC)"; \
		exit 1; \
	fi
	@git add .
	@git commit -m "$(m)"
	@git push

# Installation
install: ## Install project dependencies
	@echo "$(GREEN)Installing dependencies...$(NC)"
	@npm install
	@cd packages/functions/webhook && go mod download
	@cd packages/functions/stats && go mod download
	@cd packages/functions/seed && go mod download
	@cd packages/functions/clear && go mod download
	@echo "$(GREEN)Dependencies installed!$(NC)"

# Default
.DEFAULT_GOAL := help
