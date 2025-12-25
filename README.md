# Mundo Tá Lendo 2026 🌍📚

Global map of the 2026 Mundo Tá Lendo marathon. Collaboratively discover cultures around the world through an interactive map that shows the collective reading journey with a visual progress system.

## 🌟 Concept

This is a **collaborative** project about **discovering cultures** through reading. As participants read books from different countries throughout 2026, the map reveals the collective journey of cultural discovery with **dynamic transparency** based on reading progress.

## 🚀 Environments

### Production
- **Frontend**: https://mundotalendo.com.br *(to be configured)*
- **API**: https://api.mundotalendo.com.br *(to be configured)*

### Development
- **Frontend**: https://dev.mundotalendo.com.br ✅
- **API**: https://api.dev.mundotalendo.com.br ✅

## ✨ Features

- 🗺️ **Interactive map** with MapLibre GL JS showing 193 countries
- 🎨 **5-tier color system** - 60 distinct colors (12 months × 5 progress levels)
- 📊 **Progress visualization** - Color intensity based on reading progress (≥1% to show)
  - 0%: Gray (unexplored - not colored)
  - Tier 1 (1-20%): Light shade - "Iniciado"
  - Tier 2 (21-40%): Light - "Em Progresso"
  - Tier 3 (41-60%): Medium - "No Meio"
  - Tier 4 (61-80%): Dark - "Quase Completo"
  - Tier 5 (81-100%): Vibrant full color - "Completo"
- 🎛️ **Collapsible legend** - Toggle to show/hide month colors (starts hidden)
- 🔄 **Real-time updates** - Polling every 60s with retry logic
- 🇧🇷 **Portuguese labels** - All countries with PT-BR names
- 📱 **Responsive** - Works on desktop and mobile
- 🎯 **Interactive tooltip** - Shows country, month, progress % and tier label
- 🌊 **Lightened ocean** - Pleasant visual design
- 🖼️ **Logo header** - Mundo Tá Lendo 2026 logo image
- 🛡️ **Error Boundary** - Graceful error handling with reload option
- 🔁 **Auto-retry** - 3 attempts with exponential backoff on API failures
- 🔐 **Security headers** - X-Frame-Options, X-Content-Type-Options
- ⚡ **Performance** - Lambda concurrency limits, DynamoDB pagination, PITR backups
- 📍 **User markers** - GPS-style circular avatars showing latest user location (DEV only)
- 📖 **Book tracking** - Hover tooltips display current book being read

## 🏗️ Architecture

### Backend (Serverless)
- **Runtime**: Go 1.23+ (ARM64/Graviton)
- **Platform**: AWS Lambda
- **Database**: DynamoDB (Single Table Design with GSI)
  - **DataTable** - Single table with UUID-based partition keys:
    - `EVENT#LEITURA#<uuid>` - Reading events grouped by webhook (v1.0.2+)
    - `WEBHOOK#PAYLOAD#<uuid>` - Original payload stored once per webhook (v1.0.2+)
    - `ERROR#<uuid>` - Failed webhook processing logs with UUID tracking
    - `APIKEY#*` - API keys for authentication
  - **UserIndex GSI** - Global Secondary Index for efficient user queries:
    - hashKey: `user` (participant name)
    - rangeKey: `PK` (partition key)
    - Enables fast deletion of old user readings
  - **Storage Optimization**: 99% reduction (2.9 GB → 35 MB for 100 users)
- **API**: API Gateway V2 (HTTP API with CORS)
- **Authentication**: API Key via `X-API-Key` header (in-memory validation)
- **Region**: us-east-2 (Ohio)

### Frontend
- **Framework**: Next.js 16 (App Router)
- **Language**: JavaScript + JSDoc
- **Styling**: Tailwind CSS v4
- **Maps**: MapLibre GL JS 5.14.0
- **Data Fetching**: SWR (polling 60s, 3 retries, 10s timeout)
- **Error Handling**: React Error Boundary
- **Deploy**: CloudFront + S3

### Infrastructure
- **IaC**: SST v3.17.25 (Ion)
- **DNS**: AWS Route 53
- **SSL**: AWS Certificate Manager
- **CDN**: CloudFront

## 📁 Project Structure

```
mundotalendo/
├── public/
│   └── mundotalendo.png        # Logo image
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── layout.js           # Root layout with Error Boundary
│   │   ├── page.js             # Main page with collapsible legend
│   │   ├── globals.css         # Styles + MapLibre CSS
│   │   ├── api/
│   │   │   └── proxy-image/    # CORS proxy for user avatars
│   │   │       └── route.js
│   │   └── test-colors/        # Color testing page
│   │       └── page.js         # Visual validation of 60 color combinations
│   ├── components/
│   │   ├── Map.jsx             # Interactive map with 5-tier color system
│   │   ├── ErrorBoundary.jsx   # Error boundary for graceful failures
│   │   └── MapLegend.jsx       # Legacy legend component (not used)
│   ├── config/
│   │   ├── countries.js        # 193 countries ISO3 → PT-BR names
│   │   ├── countryCentroids.js # 1 exact point per country (no duplicates)
│   │   └── months.js           # 12 months → 5-tier color gradients (60 colors)
│   ├── hooks/
│   │   ├── useStats.js         # SWR with retry logic, 60s polling, 10s timeout
│   │   └── useUserLocations.js # SWR hook for user marker locations (60s polling)
│   └── utils/
│       ├── colorTiers.js       # Tier calculation utilities
│       └── logger.js           # Conditional logging (dev only)
├── packages/functions/         # Go Lambda Functions
│   ├── types/
│   │   └── types.go            # Shared structs (WebhookPayload, LeituraItem, etc.)
│   ├── mapping/
│   │   └── countries.go        # PT-BR country name → ISO3 code (208 countries)
│   ├── auth/
│   │   └── auth.go             # API key validation (in-memory match)
│   ├── webhook/                # POST /webhook - Process reading events
│   │   ├── main.go
│   │   └── go.mod
│   ├── stats/                  # GET /stats - Return country progress
│   │   ├── main.go
│   │   └── go.mod
│   ├── users/                  # GET /users/locations - Return user locations with avatars
│   │   ├── main.go
│   │   └── go.mod
│   ├── seed/                   # POST /test/seed - Generate test data
│   │   ├── main.go
│   │   └── go.mod
│   └── clear/                  # POST /clear - Clear all data
│       ├── main.go
│       └── go.mod
├── sst.config.ts               # SST Ion configuration (IaC)
├── next.config.js              # Next.js + Turbopack/Webpack config
├── postcss.config.js           # Tailwind CSS v4 config
├── Makefile                    # Dev commands (deploy, test, logs, etc.)
├── .env.local                  # Environment variables (API_URL, API_KEY)
├── CLAUDE.md                   # Technical context and decision history
└── project.md                  # Original project specification
```

## 🔌 API Endpoints

**⚠️ All endpoints require authentication via `X-API-Key` header.**

### `POST /webhook`
Receives reading events from Maratona.app

**Validations:**
- ✅ Filters by `identificador = "maratona-lendo-paises"` OR `"mundotalendo-2026"`
- ✅ Accepts `tipo = "leitura"` OR `"atividade"`
- ✅ If `concluido = true`, forces progress = 100%
- ✅ Calculates maximum progress among vinculados
- ✅ Extracts book title from `vinculados[].edicao.titulo`
- ✅ Saves user avatar URL from `perfil.imagem`
- ✅ Saves complete payload in JSON metadata
- ✅ Logs failures in separate table

**Response Structure:**
```json
{
  "success": true,
  "processed": 2,
  "failed": 1,
  "total": 3,
  "status": "PARTIAL",
  "errors": [
    {
      "code": "COUNTRY_NOT_FOUND",
      "message": "Country not mapped in ISO table",
      "details": "XYZ"
    }
  ]
}
```

**Status Codes:**
- `COMPLETED` - All items processed successfully
- `PARTIAL` - Some items processed, some failed
- `FAILED` - No items processed
- `NO_DATA` - No valid data to process

**Error Codes:**
| Code | Description |
|------|-------------|
| `UNAUTHORIZED` | Invalid or missing API key |
| `UNMARSHAL_ERROR` | Failed to parse JSON payload |
| `COUNTRY_NOT_FOUND` | Country name not found in ISO mapping table |
| `METADATA_MARSHAL_ERROR` | Failed to serialize metadata |
| `DYNAMODB_MARSHAL_ERROR` | Failed to marshal item for DynamoDB |
| `DYNAMODB_PUT_ERROR` | Failed to save item to DynamoDB |

**Payload:**
```json
{
  "perfil": {
    "nome": "Nathy",
    "link": "https://maratona.app/u/nathytalendo",
    "imagem": "https://assets.maratona.app/uploads/users/nathy/avatar.png"
  },
  "maratona": {
    "nome": "Maratona lendo países",
    "identificador": "mundotalendo-2026"
  },
  "desafios": [
    {
      "descricao": "Brasil",
      "categoria": "Janeiro",
      "tipo": "leitura",
      "vinculados": [
        {
          "progresso": 85,
          "updatedAt": "2024-12-16T10:00:00Z",
          "edicao": {
            "titulo": "The Silmarillion"
          }
        }
      ]
    }
  ]
}
```

### `GET /stats`
Returns explored countries with progress

**Response:**
```json
{
  "countries": [
    {"iso3": "BRA", "progress": 85},
    {"iso3": "USA", "progress": 100},
    {"iso3": "JPN", "progress": 42}
  ],
  "total": 3
}
```

### `GET /users/locations`
Returns latest location per user with avatar and book info (for map markers)

**How it works:**
- Queries all reading events from DynamoDB
- Finds most recent reading per user (using SK timestamp)
- Returns user location, avatar URL, and current book title

**Response:**
```json
{
  "users": [
    {
      "user": "DanZaekald",
      "avatarURL": "https://assets.maratona.app/uploads/users/danzaekald/avatar.png",
      "iso3": "MAR",
      "pais": "Marrocos",
      "livro": "The Silmarillion",
      "timestamp": "TIMESTAMP#2025-12-23T14:00:00Z#0"
    },
    {
      "user": "Nathy",
      "avatarURL": "https://assets.maratona.app/uploads/users/nathy/avatar.png",
      "iso3": "BRA",
      "pais": "Brasil",
      "livro": "Dom Casmurro",
      "timestamp": "TIMESTAMP#2025-12-22T10:00:00Z#0"
    }
  ],
  "total": 2
}
```

**Frontend Integration:**
- Hook: `useUserLocations()` polls this endpoint every 60s
- Map renders GPS-style circular avatars at country centroids
- Tooltip shows: "📍 {user} - Lendo: {livro}"
- Feature flag: `NEXT_PUBLIC_SHOW_USER_MARKERS` (ON in dev, OFF in prod initially)

### `POST /test/seed`
Populates database with random data (development)

**Payload:**
```json
{
  "count": 20
}
```

### `POST /clear`
Clears all data from DataTable (development only)

**Response:**
```json
{
  "success": true,
  "leiturasDeleted": 15,
  "falhasDeleted": 3
}
```

**Note:** This endpoint clears all reading events (`EVENT#LEITURA`) and error logs (`ERROR#*`) from the Single Table, but preserves API keys.

## 🔐 API Key Authentication

All API endpoints require authentication using an API key passed via the `X-API-Key` header.

**Validation Method:** API keys are validated by scanning all active keys from DynamoDB and matching in-memory (Go code), instead of using DynamoDB filter expressions. This ensures reliable authentication for all keys.

### Creating API Keys

```bash
# Create a new API key
make create-api-key name=frontend

# Output example:
# frontend-7665ec5b-c42e-4baa-93ef-c7247199b11f-2025-12-17
```

### Managing API Keys

```bash
# List all API keys
make list-api-keys

# Delete an API key
make delete-api-key name=frontend
```

### Using API Keys

**In requests:**
```bash
curl https://api.dev.mundotalendo.com.br/stats \
  -H "X-API-Key: your-api-key-here"
```

**In frontend (.env.local):**
```bash
NEXT_PUBLIC_API_KEY=frontend-uuid-date
```

The frontend automatically includes the API key in all requests when configured.

## 🚀 Local Setup

### Prerequisites
- Node.js 18+ (recommended 24.6.0)
- Go 1.23+
- AWS CLI configured
- AWS Account
- Make (pre-installed on macOS/Linux)

### Installation

```bash
# 1. Clone the repository
git clone git@github.com:danielbalieiro/mundotalendo.git
cd mundotalendo

# 2. Install all dependencies
make install

# Or manually:
npm install
cd packages/functions/webhook && go mod tidy && cd ../../..
cd packages/functions/stats && go mod tidy && cd ../../..
cd packages/functions/seed && go mod tidy && cd ../../..
cd packages/functions/clear && go mod tidy && cd ../../..
```

### ⚡ Makefile - Quick Commands

The project includes a Makefile to facilitate common operations:

```bash
# View all available commands
make help

# Build and Deploy
make build          # Compile all Go functions
make tidy           # Update Go dependencies
make deploy-dev     # Deploy to dev
make deploy-prod    # Deploy to prod
make clean          # Clean builds and cache

# Development
make dev            # Start local Next.js server

# Testing and API
make test           # Test all endpoints
make seed           # Populate database with 20 random countries
make clear          # Clear all tables
make webhook-test   # Test webhook with sample payload

# Logs (real-time)
make logs-webhook   # Webhook Lambda logs
make logs-stats     # Stats Lambda logs

# API Key Management
make create-api-key name=myapp  # Create new API key
make list-api-keys              # List all keys
make delete-api-key name=myapp  # Remove a key

# Utilities
make info           # Show AWS resources
make unlock         # Unlock stuck deployment
```

### Configuration

Copy the configuration template:

```bash
cp .env.local.example .env.local
```

Edit `.env.local`:

```bash
# AWS API Gateway URL (get after deploy)
NEXT_PUBLIC_API_URL=https://api.dev.mundotalendo.com.br

# API Key (create with: make create-api-key name=frontend)
NEXT_PUBLIC_API_KEY=frontend-uuid-date

# Feature flag: Show user markers on map (default: true in dev, false in prod)
NEXT_PUBLIC_SHOW_USER_MARKERS=true
```

### Development

```bash
# With Makefile (recommended)
make dev

# Or manually
npm run dev:local
```

Access: http://localhost:3000

## 📦 Deploy

### Deploy to DEV

```bash
# With Makefile (recommended)
make deploy-dev

# Or manually
npx sst deploy --stage dev
```

**What happens:**
1. SST builds and compiles all Go Lambda functions
2. Deploys infrastructure (API Gateway, DynamoDB, CloudFront)
3. Configures Lambda environment variables automatically via `link: [dataTable]`
4. Outputs URLs for API and frontend

**Note:** First deploy requires creating an API key:
```bash
make create-api-key name=frontend
# Copy the generated key to .env.local as NEXT_PUBLIC_API_KEY
```

### Deploy to PROD

```bash
# With Makefile (requires confirmation)
make deploy-prod

# Or manually
npx sst deploy --stage prod
```

### Remove Stack

```bash
# Dev
make remove-dev

# Or manually
npx sst remove --stage dev
npx sst remove --stage prod
```

## 🎨 Color System

The project uses a **5-tier color progression system** where each of the 12 months has 5 distinct color shades based on reading progress, totaling **60 unique colors**.

### Tier Levels

| Tier | Progress Range | Description | Visual Intensity |
|------|----------------|-------------|------------------|
| **Tier 1** | 0-20% | Iniciado | Lightest shade |
| **Tier 2** | 21-40% | Em Progresso | Light |
| **Tier 3** | 41-60% | No Meio | Medium |
| **Tier 4** | 61-80% | Quase Completo | Dark |
| **Tier 5** | 81-100% | Completo | Vibrant full color |

### Month Colors (Tier 5 - Full Intensity)

Each month has a distinct base color shown at maximum progress:

| Month | Color | Tier 5 Hex |
|-------|-------|------------|
| Janeiro | Vibrant Red | `#FF1744` |
| Fevereiro | Bright Cyan | `#00E5FF` |
| Março | Lemon Yellow | `#FFD600` |
| Abril | Vibrant Green | `#00E676` |
| Maio | Intense Orange | `#FF6F00` |
| Junho | Vibrant Purple | `#D500F9` |
| Julho | Royal Blue | `#2979FF` |
| Agosto | Vibrant Pink | `#FF4081` |
| Setembro | Bright Teal | `#1DE9B6` |
| Outubro | Flaming Orange | `#FF9100` |
| Novembro | Deep Violet | `#651FFF` |
| Dezembro | Intense Magenta | `#F50057` |

**Map Elements:**
- **Ocean**: `#6BB6FF` (light blue)
- **Unexplored countries**: `#F5F5F5` (light gray)

### Testing Colors

Visit `/test-colors` to see all 60 color combinations with visual validation:
- 12 months × 5 tiers = 60 distinct colors
- Visual swatches for each tier level
- Boundary value testing (0%, 20%, 21%, 40%, etc.)
- Example: http://localhost:3000/test-colors

## 🧪 Testing

### Unit Tests

The project has comprehensive unit tests for both frontend and backend.

```bash
# Run all tests (frontend + backend)
make test-all

# Frontend tests only (Jest)
make test-frontend
make test-frontend-watch    # Watch mode for development

# Backend tests only (Go)
make test-backend
make test-backend-coverage  # With coverage report

# Generate full coverage reports
make test-coverage          # HTML coverage for both frontend and backend

# Run benchmarks (Go)
make test-bench
```

**Test Coverage:**
- Frontend: ~80 test cases covering utilities, hooks, and components
- Backend: ~75 test cases covering Lambda handlers, auth, and mappings
- Total: ~155 unit tests with ~85% code coverage

📚 **Full testing documentation**: See [TESTING.md](./TESTING.md)

### API Integration Tests

Test the deployed API endpoints:

```bash
# Test all endpoints
make test-api

# Populate with random data (20 countries)
make seed

# Clear database
make clear

# Test webhook with sample payload
make webhook-test

# Or manually:
# Clear database
curl -X POST https://api.dev.mundotalendo.com.br/clear \
  -H "X-API-Key: your-key-here"

# Populate with random data
curl -X POST https://api.dev.mundotalendo.com.br/test/seed \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-key-here" \
  -d '{"count": 20}'

# View statistics
curl https://api.dev.mundotalendo.com.br/stats \
  -H "X-API-Key: your-key-here" | jq .
```

## 📊 Monitoring

### CloudWatch Logs

```bash
# With Makefile (recommended)
make logs-webhook   # Real-time webhook logs
make logs-stats     # Real-time stats logs

# View AWS resource information
make info

# Or manually:
# Stats Lambda
aws logs tail /aws/lambda/mundotalendo-dev-ApiRouteNodhexHandlerFunction --follow --region us-east-2

# Webhook Lambda
aws logs tail /aws/lambda/mundotalendo-dev-ApiRouteBahodaHandlerFunction --follow --region us-east-2
```

### DynamoDB Tables

```bash
# View all project tables
make info

# Or manually:
# View tables
aws dynamodb list-tables --region us-east-2 | grep mundotalendo

# Scan DataTable
aws dynamodb scan --table-name <datatable-name> --region us-east-2
```

## 🔧 Troubleshooting

### Map doesn't load
1. Check `NEXT_PUBLIC_API_URL` in `.env.local`
2. Check if API is responding: `curl https://api.dev.mundotalendo.com.br/stats`
3. Check browser console (F12)

### Stats returns error
1. Check if Lambdas have `SST_Resource_DataTable_name` variable
2. View CloudWatch logs
3. Check if DynamoDB table exists

### Webhook doesn't process
1. Validate JSON payload
2. Check `identificador = "maratona-lendo-paises"`
3. Check if country exists in mapping (208 countries)
4. Query Failures table to see errors

### SST deploy fails
1. Check AWS credentials: `aws sts get-caller-identity`
2. Check `go.mod` in each Lambda function
3. Consult CLAUDE.md for known workarounds

## 📚 Additional Documentation

- **CLAUDE.md** - Complete technical context, decision history, known bugs
- **project.md** - Original project specification
- **SST Docs** - https://sst.dev/docs

## 🤝 Contributing

1. Fork the project
2. Create a feature branch (`git checkout -b feature/new-feature`)
3. Commit with descriptive messages (`git commit -m 'Add: new feature'`)
4. Push to the branch (`git push origin feature/new-feature`)
5. Open a Pull Request

### Commit Convention

- `Add:` - New feature
- `Update:` - Update to existing feature
- `Fix:` - Bug fix
- `Refactor:` - Code refactoring
- `Docs:` - Documentation update
- `Style:` - Formatting, lint
- `Test:` - Add/update tests

## 📄 License

This project is under the MIT license. See the LICENSE file for more details.

## 👥 Author

**Daniel Balieiro**
- GitHub: [@danielbalieiro](https://github.com/danielbalieiro)

---

**🌍 Discover the world through reading!**
