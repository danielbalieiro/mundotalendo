.PHONY: help build clean dev deploy-dev deploy-prod check-deps test-api test-frontend test-backend test-all test-coverage seed stats users clear logs-webhook logs-stats logs-all alarms metrics alarms-prod metrics-prod logs-all-prod info info-prod unlock

# ⚠️ IMPORTANT: This project uses us-east-2 (Ohio) region
# All AWS commands MUST use --region us-east-2
# Resources in us-east-1 were deleted on 2025-12-23

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
	@(cd packages/functions/types && go build .)
	@(cd packages/functions/webhook && go build .)
	@(cd packages/functions/consumer && go build .)
	@(cd packages/functions/stats && go build .)
	@(cd packages/functions/seed && go build .)
	@(cd packages/functions/clear && go build .)
	@(cd packages/functions/users && go build .)
	@echo "$(GREEN)Build completed!$(NC)"

tidy: ## Update Go dependencies
	@echo "$(GREEN)Updating Go dependencies...$(NC)"
	@(cd packages/functions/webhook && go mod tidy)
	@(cd packages/functions/consumer && go mod tidy)
	@(cd packages/functions/stats && go mod tidy)
	@(cd packages/functions/seed && go mod tidy)
	@(cd packages/functions/clear && go mod tidy)
	@(cd packages/functions/users && go mod tidy)
	@echo "$(GREEN)Dependencies updated!$(NC)"

clean: ## Clean builds and cache
	@echo "$(YELLOW)Cleaning builds...$(NC)"
	@rm -rf .sst .open-next .next
	@find packages/functions -type f -name "webhook" -o -name "stats" -o -name "seed" -o -name "clear" | xargs rm -f
	@echo "$(GREEN)Cleanup completed!$(NC)"

# Deploy
unlock: ## Unlock stuck deployment (use STAGE=prod for production)
	@echo "$(YELLOW)Unlocking deployment...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	npx sst unlock --stage $$STAGE

deploy-dev: ## Deploy to dev environment and fix env vars
	@echo "$(GREEN)Deploying to DEV...$(NC)"
	@echo "\n$(YELLOW)Force rebuilding Go functions...$(NC)"
	@find packages/functions -type f -name "bootstrap" | xargs rm -f
	@cd packages/functions/webhook && go build -o bootstrap .
	@cd packages/functions/consumer && go build -o bootstrap .
	@cd packages/functions/stats && go build -o bootstrap main.go
	@cd packages/functions/users && go build -o bootstrap main.go
	@cd packages/functions/seed && go build -o bootstrap main.go
	@cd packages/functions/clear && go build -o bootstrap main.go
	@echo "\n$(YELLOW)Syncing API key to SST Secret before deploy...$(NC)"
	@$(MAKE) update-secret
	@npx sst deploy --stage dev
	@echo "\n$(YELLOW)Fixing Lambda environment variables (SST bug workaround)...$(NC)"
	@$(MAKE) fix-env
	@echo "\n$(YELLOW)Updating .env.local with current API key...$(NC)"
	@$(MAKE) update-env-local

deploy-prod: ## Deploy to prod environment and fix env vars
	@echo "$(RED)Deploying to PRODUCTION...$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "\n$(YELLOW)Force rebuilding Go functions...$(NC)"; \
		find packages/functions -type f -name "bootstrap" | xargs rm -f; \
		(cd packages/functions/webhook && go build -o bootstrap .); \
		(cd packages/functions/consumer && go build -o bootstrap .); \
		(cd packages/functions/stats && go build -o bootstrap main.go); \
		(cd packages/functions/users && go build -o bootstrap main.go); \
		(cd packages/functions/seed && go build -o bootstrap main.go); \
		(cd packages/functions/clear && go build -o bootstrap main.go); \
		echo "\n$(YELLOW)Syncing API key to SST Secret before deploy...$(NC)"; \
		STAGE=prod $(MAKE) update-secret; \
		npx sst deploy --stage prod; \
		echo "\n$(YELLOW)Fixing Lambda environment variables (SST bug workaround)...$(NC)"; \
		STAGE=prod $(MAKE) fix-env; \
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

update-secret: ## Update SST Secret with current API key from DynamoDB (use STAGE=prod for production)
	@echo "$(YELLOW)Syncing SST Secret with DynamoDB API key...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	API_KEY=$$(STAGE=$$STAGE $(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found for stage $$STAGE. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	npx sst secret set FrontendApiKey "$$API_KEY" --stage $$STAGE; \
	echo "$(GREEN)✅ SST Secret updated with: $$API_KEY (stage: $$STAGE)$(NC)"

update-env-local: ## Update .env.local with current API key
	@API_KEY=$$($(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found in DynamoDB$(NC)"; \
		exit 1; \
	fi; \
	echo "# Development API URL - configured in sst.config.ts for stage 'dev'" > .env.local; \
	echo "NEXT_PUBLIC_API_URL=$(API_DEV)" >> .env.local; \
	echo "" >> .env.local; \
	echo "# API Key - auto-updated by deploy script" >> .env.local; \
	echo "NEXT_PUBLIC_API_KEY=$$API_KEY" >> .env.local; \
	echo "" >> .env.local; \
	echo "# Feature flag: Show user markers on map (DEV only)" >> .env.local; \
	echo "NEXT_PUBLIC_SHOW_USER_MARKERS=true" >> .env.local; \
	echo "" >> .env.local; \
	echo "# NOTE: This file is auto-generated by 'make deploy-dev' or 'make deploy-local'" >> .env.local; \
	echo "$(GREEN)✅ .env.local updated with API key: $$API_KEY$(NC)"

deploy-local: ## Deploy to DEV, fix env vars, validate, and start local server
	@echo "$(GREEN)🚀 Deploy Local - Complete Setup$(NC)"
	@echo "\n$(YELLOW)Step 1/5: Deploying to DEV...$(NC)"
	@npx sst deploy --stage dev
	@echo "\n$(YELLOW)Step 2/5: Fixing Lambda environment variables...$(NC)"
	@$(MAKE) -s fix-env
	@echo "\n$(YELLOW)Step 3/5: Validating endpoints...$(NC)"
	@API_KEY=$$($(MAKE) -s get-api-key); \
	echo "Testing /stats endpoint..."; \
	STATS_RESPONSE=$$(curl -s -w "\n%{http_code}" $(API_DEV)/stats -H "X-API-Key: $$API_KEY"); \
	HTTP_CODE=$$(echo "$$STATS_RESPONSE" | tail -n1); \
	if [ "$$HTTP_CODE" = "200" ]; then \
		echo "  ✅ /stats working"; \
	else \
		echo "  ❌ /stats failed (HTTP $$HTTP_CODE)"; \
	fi; \
	echo "Testing /users/locations endpoint..."; \
	USERS_RESPONSE=$$(curl -s -w "\n%{http_code}" $(API_DEV)/users/locations -H "X-API-Key: $$API_KEY"); \
	HTTP_CODE=$$(echo "$$USERS_RESPONSE" | tail -n1); \
	if [ "$$HTTP_CODE" = "200" ]; then \
		echo "  ✅ /users/locations working"; \
	else \
		echo "  ❌ /users/locations failed (HTTP $$HTTP_CODE)"; \
		echo "  Response: $$USERS_RESPONSE"; \
	fi
	@echo "\n$(YELLOW)Step 4/5: Updating .env.local...$(NC)"
	@$(MAKE) -s update-env-local
	@echo "\n$(YELLOW)Step 5/5: Starting local server...$(NC)"
	@echo "$(GREEN)✨ Setup complete! Opening http://localhost:3000$(NC)"
	@npm run dev:local

# API Testing (requires API key)
get-api-key: ## Get first active API key for testing
	@STAGE=$${STAGE:-dev}; \
	DATA_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query "TableNames[?contains(@, 'mundotalendo-$$STAGE-DataTable')]" --output text); \
	if [ -z "$$DATA_TABLE" ]; then \
		echo "None"; \
		exit 0; \
	fi; \
	aws dynamodb scan --region $(REGION) --table-name $$DATA_TABLE \
		--filter-expression "begins_with(PK, :pk) AND #active = :active" \
		--expression-attribute-names '{"#active":"active"}' \
		--expression-attribute-values '{":pk":{"S":"APIKEY#"},":active":{"BOOL":true}}' \
		--query 'Items[0].key.S' --output text 2>/dev/null | head -1 || echo "None"

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

seed: ## Populate database with random data (count=20) - DEV ONLY (not supported in prod for safety)
	@echo "$(GREEN)Populating database...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	if [ "$$STAGE" != "dev" ]; then \
		echo "$(RED)Error: seed command is DEV-only for safety. Use webhook-full for testing in dev.$(NC)"; \
		exit 1; \
	fi; \
	API_KEY=$$(STAGE=$$STAGE $(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	curl -s -X POST $(API_DEV)/test/seed \
		-H "Content-Type: application/json" \
		-H "X-API-Key: $$API_KEY" \
		-d '{"count": 20}' | jq .

stats: ## Get reading statistics from API (use STAGE=prod for production)
	@echo "$(GREEN)Fetching stats...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	API_KEY=$$(STAGE=$$STAGE $(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	if [ "$$STAGE" = "prod" ]; then \
		API_URL=$(API_PROD); \
	else \
		API_URL=$(API_DEV); \
	fi; \
	echo "$(YELLOW)Stage: $$STAGE | URL: $$API_URL$(NC)"; \
	curl -s $$API_URL/stats \
		-H "X-API-Key: $$API_KEY" | jq .

users: ## Get user locations from API (use STAGE=prod for production)
	@echo "$(GREEN)Fetching user locations...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	API_KEY=$$(STAGE=$$STAGE $(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	if [ "$$STAGE" = "prod" ]; then \
		API_URL=$(API_PROD); \
	else \
		API_URL=$(API_DEV); \
	fi; \
	echo "$(YELLOW)Stage: $$STAGE | URL: $$API_URL$(NC)"; \
	curl -s $$API_URL/users/locations \
		-H "X-API-Key: $$API_KEY" | jq .

readings: ## Get readings for a country (use STAGE=prod for production, iso3=BRA required)
	@if [ -z "$(iso3)" ]; then \
		echo "$(RED)Error: iso3 parameter required. Usage: make readings iso3=BRA$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Fetching readings for $(iso3)...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	API_KEY=$$(STAGE=$$STAGE $(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	if [ "$$STAGE" = "prod" ]; then \
		API_URL=$(API_PROD); \
	else \
		API_URL=$(API_DEV); \
	fi; \
	echo "$(YELLOW)Stage: $$STAGE | URL: $$API_URL$(NC)"; \
	curl -s $$API_URL/readings/$(iso3) \
		-H "X-API-Key: $$API_KEY" | jq .

clear: ## Clear all database tables - DEV ONLY (not supported in prod for safety)
	@echo "$(RED)Clearing database...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	if [ "$$STAGE" != "dev" ]; then \
		echo "$(RED)Error: clear command is DEV-only for safety. NEVER clear production data.$(NC)"; \
		exit 1; \
	fi; \
	API_KEY=$$(STAGE=$$STAGE $(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	curl -s -X POST $(API_DEV)/clear \
		-H "X-API-Key: $$API_KEY" | jq .

migrate: ## Migrate existing data to populate book covers (capaURL) - supports STAGE=prod
	@echo "$(YELLOW)Running migration to populate book covers...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	API_URL=$$(if [ "$$STAGE" = "prod" ]; then echo "$(API_PROD)"; else echo "$(API_DEV)"; fi); \
	API_KEY=$$(STAGE=$$STAGE $(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	echo "$(YELLOW)Stage: $$STAGE$(NC)"; \
	echo "$(YELLOW)API URL: $$API_URL$(NC)"; \
	echo "$(YELLOW)This may take a few minutes for large datasets...$(NC)"; \
	curl -s -X POST $$API_URL/migrate \
		-H "X-API-Key: $$API_KEY" | jq .

webhook-test: ## Test webhook with sample payload - DEV ONLY (not supported in prod for safety)
	@echo "$(GREEN)Testing webhook...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	if [ "$$STAGE" != "dev" ]; then \
		echo "$(RED)Error: webhook-test is DEV-only for safety. Real webhooks come from Maratona.app in prod.$(NC)"; \
		exit 1; \
	fi; \
	API_KEY=$$(STAGE=$$STAGE $(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	curl -s -X POST $(API_DEV)/webhook \
		-H "Content-Type: application/json" \
		-H "X-API-Key: $$API_KEY" \
		-d '{ \
			"perfil": {"nome": "Test User", "link": "https://test.com", "imagem": "https://assets.maratona.app/uploads/users/21cce56bf08058ce473c58096676adf9a67e940a14958389b50309d93d090b15/ff9f8b72-aca1-4ae7-ba66-66fc75842d06.png"}, \
			"maratona": {"nome": "Test", "identificador": "maratona-lendo-paises"}, \
			"desafios": [{ \
				"descricao": "Brasil", \
				"categoria": "Janeiro", \
				"concluido": true, \
				"tipo": "leitura", \
				"vinculados": [{"progresso": 85, "updatedAt": "2024-12-16T10:00:00Z", "edicao": {"titulo": "Dom Casmurro"}}] \
			}] \
		}' | jq .

webhook-full: ## Send webhook with ALL countries (2-5 books each) in ONE request - DEV ONLY
	@echo "$(GREEN)Generating full webhook payload for all countries...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	if [ "$$STAGE" != "dev" ]; then \
		echo "$(RED)Error: webhook-full is DEV-only for safety. Real webhooks come from Maratona.app in prod.$(NC)"; \
		exit 1; \
	fi; \
	API_KEY=$$(STAGE=$$STAGE $(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	COUNTRIES="Brasil|Portugal|Espanha|França|Itália|Alemanha|Reino Unido|Estados Unidos|Canadá|México|Argentina|Chile|Colômbia|Peru|Uruguai|Japão|China|Coreia do Sul|Índia|Austrália|Nova Zelândia|África do Sul|Egito|Marrocos|Nigéria|Quênia|Gana|Senegal|Rússia|Polônia|Ucrânia|Tchéquia|Hungria|Romênia|Grécia|Turquia|Israel|Irã|Iraque|Arábia Saudita|Emirados Árabes Unidos|Tailândia|Vietnã|Filipinas|Indonésia|Malásia|Singapura|Paquistão|Bangladesh|Sri Lanka|Islândia|Noruega|Suécia|Finlândia|Dinamarca|Irlanda|Países Baixos|Bélgica|Suíça|Áustria|Croácia|Sérvia|Bulgária|Eslovênia|Eslováquia|Estônia|Letônia|Lituânia|Geórgia|Armênia|Azerbaijão|Costa Rica|Panamá|Cuba|Jamaica|República Dominicana|Haiti|Bolívia|Paraguai|Equador|Venezuela|Nicarágua|Honduras|Guatemala|El Salvador|Moçambique|Angola|Etiópia|Tanzânia|Uganda|Ruanda|Camarões|Costa do Marfim|Zâmbia|Zimbábue|Botsuana|Namíbia|Líbano|Jordânia|Síria|Iêmen|Omã|Kuwait|Bahrein|Catar|Afeganistão|Cazaquistão|Uzbequistão|Turcomenistão|Quirguistão|Tajiquistão|Mongólia|Nepal|Butão|Mianmar|Camboja|Laos|Brunei|Timor Leste|Papua-Nova Guiné|Fiji|Samoa|Tonga|Vanuatu|Ilhas Salomão|Malta|Chipre|Luxemburgo|Mônaco|Liechtenstein|Andorra|Vaticano|San Marino|Albânia|Macedônia do Norte|Bósnia-Herzegóvina|Montenegro|Moldávia|Bielorrússia|Argélia|Tunísia|Líbia|Sudão|Sudão do Sul|Somália|Eritreia|Djibouti|Mauritânia|Mali|Níger|Chade|Burkina Faso|Benin|Togo|Guiné|Guiné-Bissau|Serra Leoa|Libéria|Guiné Equatorial|Gabão|Congo|República Democrática do Congo|República Centro-Africana|Burundi|Malawi|Madagascar|Maurício|Seychelles|Comores|Cabo Verde|São Tomé e Príncipe|Barbados|Trindade e Tobago|Bahamas|Belize|Guiana|Suriname|Antígua e Barbuda|Santa Lúcia|São Vicente e Grandinas|Granada|Dominica|São Cristóvão e Névis|Maldivas|Coreia do Norte"; \
	MESES="Janeiro|Fevereiro|Março|Abril|Maio|Junho|Julho|Agosto|Setembro|Outubro|Novembro|Dezembro"; \
	AVATAR="https://assets.maratona.app/uploads/users/21cce56bf08058ce473c58096676adf9a67e940a14958389b50309d93d090b15/ff9f8b72-aca1-4ae7-ba66-66fc75842d06.png"; \
	IFS='|' read -ra COUNTRY_ARRAY <<< "$$COUNTRIES"; \
	IFS='|' read -ra MESES_ARRAY <<< "$$MESES"; \
	TOTAL=$${#COUNTRY_ARRAY[@]}; \
	echo "$(YELLOW)Generating payload for $$TOTAL countries (2-5 books each)...$(NC)"; \
	DESAFIOS=""; \
	for ((i=0; i<TOTAL; i++)); do \
		COUNTRY="$${COUNTRY_ARRAY[$$i]}"; \
		MES="$${MESES_ARRAY[$$((RANDOM % 12))]}"; \
		NUM_BOOKS=$$((RANDOM % 4 + 2)); \
		for ((j=0; j<NUM_BOOKS; j++)); do \
			PROGRESSO=$$((RANDOM % 101)); \
			DAYS_AGO=$$((RANDOM % 365)); \
			DATE=$$(date -u -v-$${DAYS_AGO}d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "$$DAYS_AGO days ago" +"%Y-%m-%dT%H:%M:%SZ"); \
			TITULO="Livro $$((j + 1)) de $$COUNTRY"; \
			CAPA="https://picsum.photos/seed/$$COUNTRY-$$j/300/450"; \
			if [ -n "$$DESAFIOS" ]; then DESAFIOS="$$DESAFIOS,"; fi; \
			DESAFIOS="$$DESAFIOS{\"descricao\":\"$$COUNTRY\",\"categoria\":\"$$MES\",\"concluido\":true,\"tipo\":\"leitura\",\"vinculados\":[{\"progresso\":$$PROGRESSO,\"updatedAt\":\"$$DATE\",\"completo\":true,\"edicao\":{\"titulo\":\"$$TITULO\",\"capa\":\"$$CAPA\"}}]}"; \
		done; \
	done; \
	PAYLOAD="{\"perfil\":{\"nome\":\"Test User\",\"link\":\"https://maratona.app\",\"imagem\":\"$$AVATAR\"},\"maratona\":{\"nome\":\"Mundo Tá Lendo 2026\",\"identificador\":\"maratona-lendo-paises\"},\"desafios\":[$$DESAFIOS]}"; \
	echo "$(YELLOW)Sending single request with all countries...$(NC)"; \
	RESPONSE=$$(curl -s -X POST $(API_DEV)/webhook \
		-H "Content-Type: application/json" \
		-H "X-API-Key: $$API_KEY" \
		-d "$$PAYLOAD"); \
	echo "$$RESPONSE" | jq .; \
	echo "\n$(GREEN)✅ Webhook sent! Check stats with: make stats$(NC)"

webhook-multi-users: ## Send webhooks from MULTIPLE users - DEV ONLY (count=1000 default, for testing concentric rings)
	@echo "$(GREEN)Generating webhooks from multiple users...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	if [ "$$STAGE" != "dev" ]; then \
		echo "$(RED)Error: webhook-multi-users is DEV-only for testing. Real webhooks come from Maratona.app in prod.$(NC)"; \
		exit 1; \
	fi; \
	API_KEY=$$(STAGE=$$STAGE $(MAKE) -s get-api-key); \
	if [ -z "$$API_KEY" ] || [ "$$API_KEY" = "None" ]; then \
		echo "$(RED)Error: No API key found. Create one with: make create-api-key name=test$(NC)"; \
		exit 1; \
	fi; \
	COUNT=$${count:-1000}; \
	COUNTRIES="Brasil|Portugal|Espanha|França|Itália|Alemanha|Reino Unido|Estados Unidos|Canadá|Japão"; \
	MESES="Janeiro|Fevereiro|Março|Abril|Maio|Junho|Julho|Agosto|Setembro|Outubro|Novembro|Dezembro"; \
	IFS='|' read -ra COUNTRY_ARRAY <<< "$$COUNTRIES"; \
	IFS='|' read -ra MESES_ARRAY <<< "$$MESES"; \
	echo "$(YELLOW)Sending webhooks for $$COUNT users...$(NC)"; \
	echo "$(YELLOW)Progress: $(NC)"; \
	for ((i=1; i<=COUNT; i++)); do \
		USER="User$$(printf '%04d' $$i)"; \
		AVATAR_ID=$$((i % 70 + 1)); \
		AVATAR="https://i.pravatar.cc/150?img=$$AVATAR_ID"; \
		NUM_COUNTRIES=$$((RANDOM % 2 + 1)); \
		DESAFIOS=""; \
		for ((j=0; j<NUM_COUNTRIES; j++)); do \
			COUNTRY="$${COUNTRY_ARRAY[$$((RANDOM % 10))]}"; \
			MES="$${MESES_ARRAY[$$((RANDOM % 12))]}"; \
			PROGRESSO=$$((RANDOM % 100 + 1)); \
			DAYS_AGO=$$((RANDOM % 30)); \
			DATE=$$(date -u -v-$${DAYS_AGO}d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "$$DAYS_AGO days ago" +"%Y-%m-%dT%H:%M:%SZ"); \
			TITULO="Livro de $$COUNTRY por $$USER"; \
			CAPA="https://picsum.photos/seed/$$USER-$$COUNTRY-$$j/300/450"; \
			if [ -n "$$DESAFIOS" ]; then DESAFIOS="$$DESAFIOS,"; fi; \
			DESAFIOS="$$DESAFIOS{\"descricao\":\"$$COUNTRY\",\"categoria\":\"$$MES\",\"concluido\":true,\"tipo\":\"leitura\",\"vinculados\":[{\"progresso\":$$PROGRESSO,\"updatedAt\":\"$$DATE\",\"completo\":true,\"edicao\":{\"titulo\":\"$$TITULO\",\"capa\":\"$$CAPA\"}}]}"; \
		done; \
		PAYLOAD="{\"perfil\":{\"nome\":\"$$USER\",\"link\":\"https://maratona.app/$$USER\",\"imagem\":\"$$AVATAR\"},\"maratona\":{\"nome\":\"Mundo Tá Lendo 2026\",\"identificador\":\"maratona-lendo-paises\"},\"desafios\":[$$DESAFIOS]}"; \
		curl -s -X POST $(API_DEV)/webhook \
			-H "Content-Type: application/json" \
			-H "X-API-Key: $$API_KEY" \
			-d "$$PAYLOAD" > /dev/null; \
		if [ $$((i % 100)) -eq 0 ]; then \
			echo "  $(CYAN)$$i/$$COUNT users sent...$(NC)"; \
		fi; \
	done; \
	echo "\n$(GREEN)✅ Sent $$COUNT webhooks! Check stats with: make stats$(NC)"; \
	echo "$(GREEN)✅ Check user locations with: make users$(NC)"

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

# Monitoring - DEV
alarms: ## Show CloudWatch alarm status (DEV)
	@echo "$(GREEN)CloudWatch Alarms Status - DEV:$(NC)"
	@aws cloudwatch describe-alarms --region $(REGION) \
		--query 'MetricAlarms[?contains(AlarmName, `mundotalendo-dev`)].{Name:AlarmName,State:StateValue,Reason:StateReason}' \
		--output table

metrics: ## Show custom metrics (DEV)
	@echo "$(GREEN)Custom Metrics - DEV:$(NC)"
	@aws cloudwatch list-metrics --region $(REGION) \
		--namespace MundoTaLendo --output table

logs-all: ## Tail all Lambda logs in real-time (DEV)
	@echo "$(GREEN)Tailing DEV Lambda logs...$(NC)"
	@aws logs tail --follow --region $(REGION) \
		--log-group-name-prefix /aws/lambda/mundotalendo-dev

# Monitoring - PROD
alarms-prod: ## Show CloudWatch alarm status (PROD)
	@echo "$(GREEN)CloudWatch Alarms Status - PROD:$(NC)"
	@aws cloudwatch describe-alarms --region $(REGION) \
		--query 'MetricAlarms[?contains(AlarmName, `mundotalendo-prod`)].{Name:AlarmName,State:StateValue,Reason:StateReason}' \
		--output table

metrics-prod: ## Show custom metrics (PROD)
	@echo "$(GREEN)Custom Metrics - PROD:$(NC)"
	@aws cloudwatch list-metrics --region $(REGION) \
		--namespace MundoTaLendo --output table

logs-all-prod: ## Tail all Lambda logs in real-time (PROD)
	@echo "$(GREEN)Tailing PROD Lambda logs...$(NC)"
	@aws logs tail --follow --region $(REGION) \
		--log-group-name-prefix /aws/lambda/mundotalendo-prod

# AWS Info
info: ## Show AWS resources information (DEV)
	@echo "$(GREEN)AWS Resources - Stage: DEV$(NC)"
	@echo "\n$(YELLOW)DynamoDB Tables:$(NC)"
	@aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-dev`)]' --output table
	@echo "\n$(YELLOW)Lambda Functions:$(NC)"
	@aws lambda list-functions --region $(REGION) --query 'Functions[?contains(FunctionName, `mundotalendo-dev`)].FunctionName' --output table
	@echo "\n$(YELLOW)API Gateway:$(NC)"
	@aws apigatewayv2 get-apis --region $(REGION) --query 'Items[?contains(Name, `mundotalendo`)].{Name:Name,Endpoint:ApiEndpoint}' --output table

info-prod: ## Show AWS resources information (PROD)
	@echo "$(GREEN)AWS Resources - Stage: PROD$(NC)"
	@echo "\n$(YELLOW)DynamoDB Tables:$(NC)"
	@aws dynamodb list-tables --region $(REGION) --query 'TableNames[?contains(@, `mundotalendo-prod`)]' --output table
	@echo "\n$(YELLOW)Lambda Functions:$(NC)"
	@aws lambda list-functions --region $(REGION) --query 'Functions[?contains(FunctionName, `mundotalendo-prod`)].FunctionName' --output table
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
	@STAGE=$${STAGE:-dev}; \
	DATA_TABLE=$$(aws dynamodb list-tables --region $(REGION) --query "TableNames[?contains(@, 'mundotalendo-$$STAGE-DataTable')]" --output text); \
	PAYLOAD_BUCKET=$$(aws s3api list-buckets --query "Buckets[?contains(Name, 'mundotalendo-$$STAGE-payloadbucket')].Name" --output text); \
	WEBHOOK_QUEUE=$$(aws sqs list-queues --region $(REGION) --query "QueueUrls[?contains(@, 'mundotalendo-$$STAGE-WebhookQueueQueue')]" --output text); \
	if [ -z "$$DATA_TABLE" ]; then \
		echo "$(RED)Error: DataTable not found for stage $$STAGE$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)Found resources:$(NC)"; \
	echo "  DataTable: $$DATA_TABLE"; \
	echo "  PayloadBucket: $$PAYLOAD_BUCKET"; \
	echo "  WebhookQueue: $$WEBHOOK_QUEUE"; \
	echo "\n$(YELLOW)Updating API Lambda functions...$(NC)"; \
	for fn in $$(aws lambda list-functions --region $(REGION) --query "Functions[?contains(FunctionName, 'mundotalendo-$$STAGE-ApiRoute')].FunctionName" --output text); do \
		echo "  Updating $$fn..."; \
		if echo "$$fn" | grep -q "Bahoda"; then \
			aws lambda update-function-configuration \
				--function-name $$fn \
				--region $(REGION) \
				--environment "Variables={SST_Resource_DataTable_name=$$DATA_TABLE,SST_Resource_PayloadBucket_name=$$PAYLOAD_BUCKET,SST_Resource_WebhookQueue_url=$$WEBHOOK_QUEUE}" \
				--output text --query 'FunctionName' 2>&1 | grep -v "An error occurred" || true; \
		else \
			aws lambda update-function-configuration \
				--function-name $$fn \
				--region $(REGION) \
				--environment "Variables={SST_Resource_DataTable_name=$$DATA_TABLE}" \
				--output text --query 'FunctionName' 2>&1 | grep -v "An error occurred" || true; \
		fi; \
	done; \
	echo "\n$(YELLOW)Updating Consumer Lambda...$(NC)"; \
	CONSUMER_FN=$$(aws lambda list-functions --region $(REGION) --query "Functions[?contains(FunctionName, 'mundo') && contains(FunctionName, '$$STAGE') && contains(FunctionName, 'WebhookQueueSubscriber')].FunctionName" --output text); \
	if [ -n "$$CONSUMER_FN" ]; then \
		echo "  Updating $$CONSUMER_FN..."; \
		aws lambda update-function-configuration \
			--function-name $$CONSUMER_FN \
			--region $(REGION) \
			--environment "Variables={SST_Resource_DataTable_name=$$DATA_TABLE,SST_Resource_PayloadBucket_name=$$PAYLOAD_BUCKET}" \
			--output text --query 'FunctionName' 2>&1 | grep -v "An error occurred" || true; \
	else \
		echo "  $(YELLOW)Consumer Lambda not found (may not be deployed yet)$(NC)"; \
	fi; \
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

# SQS/DLQ Management
dlq-view: ## View messages in Dead Letter Queue (DEV)
	@echo "$(GREEN)Viewing DLQ messages...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	DLQ_URL=$$(aws sqs list-queues --region $(REGION) --queue-name-prefix "mundotalendo-$$STAGE-WebhookDLQ" --query 'QueueUrls[0]' --output text 2>/dev/null); \
	if [ -z "$$DLQ_URL" ] || [ "$$DLQ_URL" = "None" ]; then \
		echo "$(RED)DLQ not found for stage $$STAGE$(NC)"; \
		exit 1; \
	fi; \
	echo "$(YELLOW)DLQ URL: $$DLQ_URL$(NC)"; \
	ATTRS=$$(aws sqs get-queue-attributes --region $(REGION) --queue-url "$$DLQ_URL" --attribute-names ApproximateNumberOfMessages --query 'Attributes.ApproximateNumberOfMessages' --output text); \
	echo "$(YELLOW)Messages in DLQ: $$ATTRS$(NC)"; \
	if [ "$$ATTRS" != "0" ]; then \
		echo "\n$(YELLOW)Sample message (not deleted):$(NC)"; \
		aws sqs receive-message --region $(REGION) --queue-url "$$DLQ_URL" --max-number-of-messages 1 --visibility-timeout 0 | jq '.Messages[0].Body | fromjson'; \
	fi

dlq-count: ## Count messages in DLQ (use STAGE=prod for production)
	@STAGE=$${STAGE:-dev}; \
	DLQ_URL=$$(aws sqs list-queues --region $(REGION) --queue-name-prefix "mundotalendo-$$STAGE-WebhookDLQ" --query 'QueueUrls[0]' --output text 2>/dev/null); \
	if [ -z "$$DLQ_URL" ] || [ "$$DLQ_URL" = "None" ]; then \
		echo "0"; \
		exit 0; \
	fi; \
	aws sqs get-queue-attributes --region $(REGION) --queue-url "$$DLQ_URL" --attribute-names ApproximateNumberOfMessages --query 'Attributes.ApproximateNumberOfMessages' --output text

dlq-purge: ## Purge all messages from DLQ - DEV ONLY
	@echo "$(RED)Purging DLQ messages...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	if [ "$$STAGE" != "dev" ]; then \
		echo "$(RED)Error: dlq-purge is DEV-only for safety.$(NC)"; \
		exit 1; \
	fi; \
	DLQ_URL=$$(aws sqs list-queues --region $(REGION) --queue-name-prefix "mundotalendo-$$STAGE-WebhookDLQ" --query 'QueueUrls[0]' --output text 2>/dev/null); \
	if [ -z "$$DLQ_URL" ] || [ "$$DLQ_URL" = "None" ]; then \
		echo "$(RED)DLQ not found$(NC)"; \
		exit 1; \
	fi; \
	read -p "Are you sure you want to purge all DLQ messages? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		aws sqs purge-queue --region $(REGION) --queue-url "$$DLQ_URL"; \
		echo "$(GREEN)DLQ purged!$(NC)"; \
	fi

queue-stats: ## Show SQS queue statistics (use STAGE=prod for production)
	@echo "$(GREEN)SQS Queue Statistics...$(NC)"
	@STAGE=$${STAGE:-dev}; \
	echo "$(YELLOW)Stage: $$STAGE$(NC)"; \
	QUEUE_URL=$$(aws sqs list-queues --region $(REGION) --queue-name-prefix "mundotalendo-$$STAGE-WebhookQueue" --query 'QueueUrls[0]' --output text 2>/dev/null); \
	DLQ_URL=$$(aws sqs list-queues --region $(REGION) --queue-name-prefix "mundotalendo-$$STAGE-WebhookDLQ" --query 'QueueUrls[0]' --output text 2>/dev/null); \
	if [ -n "$$QUEUE_URL" ] && [ "$$QUEUE_URL" != "None" ]; then \
		echo "\n$(YELLOW)Main Queue:$(NC)"; \
		aws sqs get-queue-attributes --region $(REGION) --queue-url "$$QUEUE_URL" \
			--attribute-names ApproximateNumberOfMessages ApproximateNumberOfMessagesNotVisible ApproximateNumberOfMessagesDelayed \
			--query 'Attributes' --output table; \
	fi; \
	if [ -n "$$DLQ_URL" ] && [ "$$DLQ_URL" != "None" ]; then \
		echo "\n$(YELLOW)Dead Letter Queue:$(NC)"; \
		aws sqs get-queue-attributes --region $(REGION) --queue-url "$$DLQ_URL" \
			--attribute-names ApproximateNumberOfMessages \
			--query 'Attributes' --output table; \
	fi

test-consumer: ## Run consumer unit tests
	@echo "$(GREEN)Running consumer tests...$(NC)"
	@cd packages/functions/consumer && go test -v

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
