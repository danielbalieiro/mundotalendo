.PHONY: help build clean dev deploy-dev deploy-prod test seed clear logs-webhook logs-stats unlock

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

# API Testing
test: ## Test dev API endpoints
	@echo "$(GREEN)Testing DEV API...$(NC)"
	@echo "\n$(YELLOW)GET /stats:$(NC)"
	@curl -s $(API_DEV)/stats | jq .
	@echo "\n$(YELLOW)Available endpoints:$(NC)"
	@echo "  - GET  $(API_DEV)/stats"
	@echo "  - POST $(API_DEV)/webhook"
	@echo "  - POST $(API_DEV)/test/seed"
	@echo "  - POST $(API_DEV)/clear"

seed: ## Populate database with random data (count=20)
	@echo "$(GREEN)Populating database...$(NC)"
	@curl -s -X POST $(API_DEV)/test/seed \
		-H "Content-Type: application/json" \
		-d '{"count": 20}' | jq .

clear: ## Clear all database tables
	@echo "$(RED)Clearing database...$(NC)"
	@curl -s -X POST $(API_DEV)/clear | jq .

webhook-test: ## Test webhook with sample payload
	@echo "$(GREEN)Testing webhook...$(NC)"
	@curl -s -X POST $(API_DEV)/webhook \
		-H "Content-Type: application/json" \
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
		aws lambda get-function-configuration --function-name $$fn --region $(REGION) --query 'Environment.Variables' --output json | jq -c '. | {Leituras: .SST_Resource_Leituras_name, Falhas: .SST_Resource_Falhas_name}'; \
	done

check-failures: ## Show recent failures from Falhas table
	@echo "$(GREEN)Recent failures (last 10):$(NC)"
	@FALHAS_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-dev-Falhas`)]' --output text); \
	if [ -n "$$FALHAS_TABLE" ]; then \
		aws dynamodb scan --table-name $$FALHAS_TABLE --region $(REGION) --max-items 10 | jq -r '.Items[] | "\(.SK.S): \(.ErrorType.S) - \(.ErrorMessage.S)"'; \
	else \
		echo "$(RED)Falhas table not found$(NC)"; \
	fi

fix-env: ## Fix Lambda environment variables (SST bug workaround)
	@echo "$(YELLOW)Fixing Lambda environment variables...$(NC)"
	@LEITURAS_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-dev-Leituras`)]' --output text); \
	FALHAS_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-dev-Falhas`)]' --output text); \
	echo "$(GREEN)Found tables:$(NC)"; \
	echo "  Leituras: $$LEITURAS_TABLE"; \
	echo "  Falhas: $$FALHAS_TABLE"; \
	echo "\n$(YELLOW)Updating Lambda functions...$(NC)"; \
	for fn in $$(aws lambda list-functions --region $(REGION) --query 'Functions[?contains(FunctionName, `mundotalendo-dev-ApiRoute`)].FunctionName' --output text); do \
		echo "  Updating $$fn..."; \
		aws lambda update-function-configuration \
			--function-name $$fn \
			--region $(REGION) \
			--environment "Variables={SST_Resource_Leituras_name=$$LEITURAS_TABLE,SST_Resource_Falhas_name=$$FALHAS_TABLE}" \
			--output text --query 'FunctionName' 2>&1 | grep -v "An error occurred" || true; \
	done; \
	echo "\n$(GREEN)Environment variables updated!$(NC)"

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
