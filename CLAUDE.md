# Claude Context - Mundo Tá Lendo 2026

> **Versão:** v1.1.0 | **Status:** EM PRODUÇÃO
> **URLs:** https://mundotalendo.com.br | https://api.mundotalendo.com.br
> **Dev:** https://dev.mundotalendo.com.br | https://api.dev.mundotalendo.com.br

## O Projeto

Dashboard de **descoberta cultural colaborativa** através da leitura. Mapeia em tempo real países sendo explorados por participantes do desafio "Mundo Tá Lendo 2026".

**Conceito:** Não é sobre "conquista" de países - é sobre **descobrir culturas** colaborativamente.

## SISTEMA EM PRODUÇÃO

**ANTES de qualquer mudança:**
1. Testar localmente com `npm run dev:local`
2. Validar compatibilidade com dados existentes
3. Deploy em DEV antes de produção
4. Ter plano de rollback

**NUNCA:**
- Apagar dados sem backup
- Mudar schema do DynamoDB sem migração
- Deploy direto em produção sem testar em DEV

## Stack

| Componente | Tecnologia |
|------------|-----------|
| IaC | SST (Ion) 3.x |
| Frontend | Next.js 16, Tailwind v4 |
| Maps | MapLibre GL JS 5.x |
| Backend | Go 1.23+, ARM64 |
| Database | DynamoDB (Single Table) |
| Queue | SQS + DLQ |
| Storage | S3 (webhook payloads) |
| API | API Gateway V2 |
| Region | us-east-2 (Ohio) |

## Arquitetura Async (Webhook)

```
POST /webhook → Webhook Lambda → S3 (payload) + SQS (message)
                                        ↓
                              Consumer Lambda → DynamoDB
                                        ↓ (falhas)
                                       DLQ
```

**Benefícios:**
- Resposta rápida (~100ms vs ~2s)
- Retries automáticos via SQS (3x)
- Dead Letter Queue para falhas persistentes
- Alarmes para panics/crashes

## Estrutura DynamoDB

```
LeituraItem:
  PK: "EVENT#LEITURA"
  SK: "<uuid>#<iso3>#<index>"
  Fields: webhookUUID, updatedAt, iso3, pais, categoria, progresso, user, avatarURL, capaURL, livro

WebhookItem:
  PK: "WEBHOOK#PAYLOAD#<uuid>"
  SK: "TIMESTAMP#<RFC3339>"
  Fields: user, payload (JSON)
```

## Estrutura de Arquivos

```
src/
├── app/
│   ├── page.js              # Página principal
│   ├── globals.css          # Tailwind + MapLibre CSS
│   └── api/proxy-image/     # Proxy CORS para imagens
├── components/
│   ├── Map.jsx              # Mapa com cores e markers
│   └── CountryPopup.jsx     # Popup de leituras
├── config/
│   ├── countries.js         # ISO → Nome PT-BR
│   ├── countryCentroids.js  # ISO → [lng, lat]
│   └── months.js            # Mês → Cor
└── hooks/
    ├── useStats.js          # Polling de estatísticas
    ├── useCountryReadings.js # Fetch leituras por país
    └── useAsyncImages.js    # Carregamento de avatares

packages/functions/
├── types/types.go           # Structs compartilhados
├── mapping/iso.go           # Nome país → ISO code
├── webhook/main.go          # POST /webhook (async: S3 + SQS)
├── consumer/main.go         # SQS Consumer (processa webhooks)
├── stats/main.go            # GET /stats
├── users/main.go            # GET /users/locations
├── readings/main.go         # GET /readings/{iso3}
├── seed/main.go             # POST /test/seed
└── clear/main.go            # POST /test/clear
```

## Comandos Make

**Deploy:**
```bash
make deploy-dev          # Deploy para DEV
make deploy-prod         # Deploy para PROD (confirmação)
make unlock              # Desbloquear deploy travado
```

**Consultas (suportam STAGE=prod):**
```bash
make stats               # Ver estatísticas
make users               # Ver localizações
make readings iso3=BRA   # Ver leituras de um país
make get-api-key         # Pegar API key
```

**Apenas DEV (bloqueados em prod):**
```bash
make seed                # Popular BD com dados teste
make clear               # Limpar BD
make webhook-full        # Gerar webhook completo
```

**Monitoring:**
```bash
make logs-all            # Logs em tempo real
make info                # Ver recursos AWS
```

## Webhook Maratona.app

Payload esperado:
```json
{
  "perfil": { "nome": "...", "imagemURL": "..." },
  "desafios": [{
    "descricao": "Brasil",
    "categoria": "Janeiro",
    "tipo": "leitura",
    "progresso": 75,
    "vinculados": [{
      "edicao": { "titulo": "...", "capa": "..." },
      "progresso": 75,
      "updatedAt": "..."
    }]
  }]
}
```

## Decisões Técnicas

- **MapLibre direto** (não react-map-gl) - mais controle
- **Webpack** (não Turbopack) - compatibilidade MapLibre
- **Centroids custom** - 1 label por país (vector tiles duplicam)
- **go.mod por Lambda** - requisito SST
- **Proxy de imagens** - resolve CORS

## Design

- **Oceano:** #0077BE
- **País não explorado:** #F5F5F5
- **Cores dos meses:** definidas em `src/config/months.js`
- **Labels:** PT-BR, Noto Sans Bold 12px, halo branco
- **Zoom:** 1-6 (evita divisões estaduais)

## Avisos

1. Labels sempre em português PT-BR
2. Usar `npm run dev:local` (Webpack, não Turbopack)
3. Região AWS: us-east-2 (Ohio)
4. Comandos destrutivos bloqueados em produção

## Changelog

### v1.1.0 - Node.js 22.x Runtime Upgrade
- **Server Lambda runtime**: Atualizado de `nodejs20.x` para `nodejs22.x`
- **SST atualizado**: v3.17.25 → v3.17.38
- **Nota**: ImageOptimizer e Revalidation ainda em `nodejs20.x` (hardcoded no SST, aguardando upstream fix)

### v1.0.9 - Async Webhook Processing
- **Arquitetura async**: Webhook → S3 + SQS → Consumer → DynamoDB
- **Novo Consumer Lambda**: Processa webhooks de forma assíncrona
- **S3 PayloadBucket**: Armazena payloads (90 dias lifecycle)
- **SQS + DLQ**: Fila com 3 retries e dead letter queue
- **Alarmes de panic/crash**: Detecta crashes em webhook e consumer
- **fix-env melhorado**: Atualiza env vars de todas as Lambdas incluindo Consumer

### v1.0.8 - Race Condition Fix
- Fix race conditions no carregamento do mapa

### v1.0.7 - Country Readings Popup
- Popup mostra todas as leituras do país

### v1.0.6 - Country Popup & Book Covers
- Popup de país com capas de livros
