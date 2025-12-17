# Mundo Tá Lendo 2026 🌍📚

Dashboard de telemetria global do desafio de leitura "Mundo Tá Lendo 2026". Descubra colaborativamente culturas ao redor do mundo através de um mapa interativo que mostra a jornada coletiva de leitura com sistema de progresso visual.

## 🌟 Conceito

Este é um projeto **colaborativo** sobre **descobrir culturas** através da leitura. À medida que participantes leem livros de diferentes países ao longo de 2026, o mapa vai revelando a jornada coletiva de descoberta cultural com **transparência dinâmica** baseada no progresso de leitura.

## 🚀 Ambientes

### Produção
- **Frontend**: https://mundotalendo.com.br *(a configurar)*
- **API**: https://api.mundotalendo.com.br *(a configurar)*

### Desenvolvimento
- **Frontend**: https://dev.mundotalendo.com.br ✅
- **API**: https://api.dev.mundotalendo.com.br ✅

## ✨ Funcionalidades

- 🗺️ **Mapa interativo** com MapLibre GL JS mostrando 193 países
- 🎨 **Cores vibrantes** - 12 meses com cores distintas
- 📊 **Sistema de progresso** - Transparência visual de 0-100%
  - 0% → 30% opaco (descoberta inicial)
  - 100% → 100% opaco (completamente explorado)
- 🔄 **Atualização em tempo real** - Polling a cada 15s
- 🇧🇷 **Labels em português** - Todos os países com nomes PT-BR
- 📱 **Responsivo** - Funciona em desktop e mobile
- 🎯 **Tooltip interativo** - Mostra progresso ao hover
- 🌊 **Oceano clareado** - Design visual agradável

## 🏗️ Arquitetura

### Backend (Serverless)
- **Runtime**: Go 1.23+ (ARM64/Graviton)
- **Platform**: AWS Lambda
- **Database**: DynamoDB (2 tabelas)
  - `Leituras` - Eventos de leitura com progresso
  - `Falhas` - Log de erros para análise
- **API**: API Gateway V2 (HTTP API)
- **Region**: us-east-2 (Ohio)

### Frontend
- **Framework**: Next.js 16 (App Router)
- **Language**: JavaScript + JSDoc
- **Styling**: Tailwind CSS v4
- **Maps**: MapLibre GL JS 5.14.0
- **Data Fetching**: SWR (polling 15s)
- **Deploy**: CloudFront + S3

### Infraestrutura
- **IaC**: SST v3.17.25 (Ion)
- **DNS**: AWS Route 53
- **SSL**: AWS Certificate Manager
- **CDN**: CloudFront

## 📁 Estrutura do Projeto

```
mundotalendo/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── layout.js           # Layout raiz
│   │   ├── page.js             # Página principal
│   │   └── globals.css         # Estilos + MapLibre CSS
│   ├── components/
│   │   └── Map.jsx             # Mapa com transparência dinâmica
│   ├── config/
│   │   ├── countries.js        # 193 países ISO → PT-BR
│   │   ├── countryCentroids.js # 1 ponto exato por país
│   │   └── months.js           # 12 meses → cores → países
│   └── hooks/
│       └── useStats.js         # SWR polling /stats
├── packages/functions/         # Lambdas Go
│   ├── types/
│   │   └── types.go            # Structs compartilhados
│   ├── mapping/
│   │   └── countries.go        # Nome PT-BR → ISO3
│   ├── webhook/                # POST /webhook
│   │   ├── main.go
│   │   └── go.mod
│   ├── stats/                  # GET /stats
│   │   ├── main.go
│   │   └── go.mod
│   ├── seed/                   # POST /test/seed
│   │   ├── main.go
│   │   └── go.mod
│   └── clear/                  # POST /clear
│       ├── main.go
│       └── go.mod
├── sst.config.ts               # Configuração SST
├── next.config.js              # Next.js config
├── postcss.config.js           # Tailwind v4
├── CLAUDE.md                   # Contexto técnico completo
└── project.md                  # Especificação original
```

## 🔌 API Endpoints

### `POST /webhook`
Recebe eventos de leitura do Maratona.app

**Validações:**
- ✅ Filtra por `identificador = "maratona-lendo-paises"`
- ✅ Aceita `tipo = "leitura"` OU `"atividade"`
- ✅ Se `concluido = true`, força progresso = 100%
- ✅ Calcula progresso máximo entre vinculados
- ✅ Salva payload completo em metadata JSON
- ✅ Loga falhas em tabela separada

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
    "link": "https://maratona.app/u/nathytalendo"
  },
  "maratona": {
    "nome": "Maratona lendo países",
    "identificador": "maratona-lendo-paises"
  },
  "desafios": [
    {
      "descricao": "Brasil",
      "categoria": "Janeiro",
      "tipo": "leitura",
      "vinculados": [
        {
          "progresso": 85,
          "updatedAt": "2024-12-16T10:00:00Z"
        }
      ]
    }
  ]
}
```

### `GET /stats`
Retorna países explorados com progresso

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

### `POST /test/seed`
Popula banco com dados aleatórios (desenvolvimento)

**Payload:**
```json
{
  "count": 20
}
```

### `POST /clear`
Limpa todas as tabelas (desenvolvimento)

**Response:**
```json
{
  "success": true,
  "leiturasDeleted": 15,
  "falhasDeleted": 3
}
```

## 🚀 Setup Local

### Pré-requisitos
- Node.js 18+ (recomendado 24.6.0)
- Go 1.23+
- AWS CLI configurado
- Conta AWS
- Make (já vem no macOS/Linux)

### Instalação

```bash
# 1. Clone o repositório
git clone git@github.com:danielbalieiro/mundotalendo.git
cd mundotalendo

# 2. Instale todas as dependências
make install

# Ou manualmente:
# npm install
# cd packages/functions/webhook && go mod tidy && cd ../..
cd packages/functions/stats && go mod tidy && cd ../..
cd packages/functions/seed && go mod tidy && cd ../..
cd packages/functions/clear && go mod tidy && cd ../..
```

### ⚡ Makefile - Comandos Rápidos

O projeto inclui um Makefile para facilitar operações comuns:

```bash
# Ver todos os comandos disponíveis
make help

# Build e Deploy
make build          # Compila todas as funções Go
make tidy           # Atualiza dependências Go
make deploy-dev     # Deploy para dev
make deploy-prod    # Deploy para prod
make clean          # Limpa builds e cache

# Desenvolvimento
make dev            # Inicia servidor Next.js local

# Testes e API
make test           # Testa todos os endpoints
make seed           # Popula banco com 20 países aleatórios
make clear          # Limpa todas as tabelas
make webhook-test   # Testa webhook com payload de exemplo

# Logs (tempo real)
make logs-webhook   # Logs do webhook Lambda
make logs-stats     # Logs do stats Lambda

# Utilidades
make info           # Mostra recursos AWS
make unlock         # Desbloqueia deploy travado
```

### Configuração

Crie `.env.local`:

```bash
# Desenvolvimento com API AWS (recomendado)
NEXT_PUBLIC_API_URL=https://api.dev.mundotalendo.com.br

# OU desenvolvimento com mock local
# NEXT_PUBLIC_API_URL=/api
```

### Desenvolvimento

```bash
# Com Makefile (recomendado)
make dev

# Ou manualmente
npm run dev:local
```

Acesse: http://localhost:3000

## 📦 Deploy

### Deploy para DEV

```bash
# Com Makefile (recomendado - já configura env vars automaticamente)
make deploy-dev

# Ou manualmente
npx sst deploy --stage dev
make fix-env  # Necessário após deploy (bug do SST)
```

### Deploy para PROD

```bash
# Com Makefile (confirmação + auto-fix env vars)
make deploy-prod

# Ou manualmente
npx sst deploy --stage prod
make fix-env  # Necessário após deploy (bug do SST)
```

### Remover Stack

```bash
# Dev
make remove-dev

# Ou manualmente
npx sst remove --stage dev
npx sst remove --stage prod
```

## 🎨 Sistema de Cores

Cada mês tem uma cor vibrante específica:

| Mês | Cor | Hex |
|-----|-----|-----|
| Janeiro | Vermelho vibrante | `#FF1744` |
| Fevereiro | Ciano brilhante | `#00E5FF` |
| Março | Amarelo limão | `#FFD600` |
| Abril | Verde vibrante | `#00E676` |
| Maio | Laranja intenso | `#FF6F00` |
| Junho | Roxo vibrante | `#D500F9` |
| Julho | Azul royal | `#2979FF` |
| Agosto | Rosa vibrante | `#FF4081` |
| Setembro | Teal brilhante | `#1DE9B6` |
| Outubro | Laranja flamejante | `#FF9100` |
| Novembro | Violeta profundo | `#651FFF` |
| Dezembro | Magenta intenso | `#F50057` |

**Oceano**: `#6BB6FF` (azul claro)
**Países não explorados**: `#F5F5F5` (cinza claro)

## 🧪 Testes

### Testar API DEV

```bash
# Testar todos os endpoints
make test

# Popular com dados aleatórios (20 países)
make seed

# Limpar banco
make clear

# Testar webhook com payload de exemplo
make webhook-test

# Ou manualmente:
# Limpar banco
curl -X POST https://api.dev.mundotalendo.com.br/clear

# Popular com dados aleatórios
curl -X POST https://api.dev.mundotalendo.com.br/test/seed \
  -H "Content-Type: application/json" \
  -d '{"count": 20}'

# Ver estatísticas
curl https://api.dev.mundotalendo.com.br/stats | jq .
```

## 📊 Monitoramento

### CloudWatch Logs

```bash
# Com Makefile (recomendado)
make logs-webhook   # Logs do webhook em tempo real
make logs-stats     # Logs do stats em tempo real

# Ver informações dos recursos AWS
make info

# Ou manualmente:
# Stats Lambda
aws logs tail /aws/lambda/mundotalendo-dev-ApiRouteNodhexHandlerFunction --follow --region us-east-2

# Webhook Lambda
aws logs tail /aws/lambda/mundotalendo-dev-ApiRouteBahodaHandlerFunction --follow --region us-east-2
```

### DynamoDB Tables

```bash
# Ver todas as tabelas do projeto
make info

# Ou manualmente:
# Ver tabelas
aws dynamodb list-tables --region us-east-2 | grep mundotalendo

# Scan Leituras
aws dynamodb scan --table-name <nome-tabela-leituras> --region us-east-2

# Scan Falhas
aws dynamodb scan --table-name <nome-tabela-falhas> --region us-east-2
```

## 🔧 Troubleshooting

### Mapa não carrega
1. Verificar `NEXT_PUBLIC_API_URL` em `.env.local`
2. Verificar se API está respondendo: `curl https://api.dev.mundotalendo.com.br/stats`
3. Verificar console do browser (F12)

### Stats retorna erro
1. Verificar se Lambdas têm variável `SST_Resource_Leituras_name`
2. Ver logs no CloudWatch
3. Verificar se tabela DynamoDB existe

### Webhook não processa
1. Validar JSON payload
2. Verificar `identificador = "maratona-lendo-paises"`
3. Verificar se país existe no mapeamento (208 países)
4. Consultar tabela Falhas para ver erros

### Deploy SST falha
1. Verificar credenciais AWS: `aws sts get-caller-identity`
2. Verificar `go.mod` em cada função Lambda
3. Consultar CLAUDE.md para workarounds conhecidos

## 📚 Documentação Adicional

- **CLAUDE.md** - Contexto técnico completo, histórico de decisões, bugs conhecidos
- **project.md** - Especificação original do projeto
- **SST Docs** - https://sst.dev/docs

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch feature (`git checkout -b feature/nova-feature`)
3. Commit com mensagens descritivas (`git commit -m 'Add: nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

### Convenção de Commits

- `Add:` - Nova funcionalidade
- `Update:` - Atualização de funcionalidade existente
- `Fix:` - Correção de bug
- `Refactor:` - Refatoração de código
- `Docs:` - Atualização de documentação
- `Style:` - Formatação, lint
- `Test:` - Adicionar/atualizar testes

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para mais detalhes.

## 👥 Autor

**Daniel Balieiro**
- GitHub: [@danielbalieiro](https://github.com/danielbalieiro)

---

**🌍 Descubra o mundo através da leitura!**
