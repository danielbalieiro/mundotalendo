# Claude Context - Mundo Tá Lendo 2026

> **Última atualização:** 2025-12-25 (v1.0.3)
> **Status:** 🔴 EM PRODUÇÃO COM DADOS REAIS - Sistema ativo recebendo leituras reais dos participantes
> **Deploy DEV:** https://dev.mundotalendo.com.br | https://api.dev.mundotalendo.com.br
> **Versão Atual:** v1.0.3 - Critical Bugfixes & Stability

## 📋 Resumo Executivo

Projeto de **descoberta cultural colaborativa** através da leitura. Dashboard que mapeia em tempo real países sendo explorados por participantes do desafio de leitura "Mundo Tá Lendo 2026".

**Conceito importante:** Não é sobre "conquista" de países, é sobre **descobrir culturas** colaborativamente.

## 🚨 ATENÇÃO: SISTEMA EM PRODUÇÃO

**⚠️ MUDANÇAS DEVEM SER FEITAS COM EXTREMO CUIDADO**

O projeto foi **promovido a produção** e está **recebendo dados reais** de participantes do desafio "Mundo Tá Lendo 2026".

### Diretrizes Obrigatórias para Mudanças

**ANTES de implementar QUALQUER mudança, considere:**

1. **Impacto nos dados existentes**
   - Como a mudança afetará os dados já armazenados no DynamoDB?
   - Haverá necessidade de migração de dados?
   - Os dados antigos continuarão funcionando com o novo código?

2. **Compatibilidade retroativa**
   - A mudança quebrará leituras já registradas?
   - O webhook continuará processando dados corretamente?
   - Os endpoints API manterão compatibilidade?

3. **Validação e testes**
   - Teste SEMPRE em ambiente local primeiro
   - Valide com dados reais (não apenas seed)
   - Verifique se não há efeitos colaterais

4. **Deploy gradual**
   - Considere feature flags para mudanças significativas
   - Deploy em DEV antes de produção
   - Monitore métricas após deploy

5. **Rollback plan**
   - Tenha sempre um plano de reversão
   - Mantenha backups antes de mudanças estruturais
   - Documente o processo de rollback

**🔴 NUNCA:**
- Apagar dados de produção sem backup confirmado
- Mudar schema do DynamoDB sem migração planejada
- Fazer deploy direto em produção sem testar em DEV
- Remover campos de API que podem estar em uso
- Alterar lógica de processamento do webhook sem validação completa

**✅ SEMPRE:**
- Testar localmente com `npm run dev:local`
- Validar com `make stats` antes e depois de mudanças
- Verificar logs do CloudWatch após deploys
- Documentar mudanças no CLAUDE.md
- Comunicar breaking changes antecipadamente

## 🎯 Estado Atual do Projeto

### ✅ v1.0.5: Fix CORS Proxy in Deployed Environments (03 Jan 2026)

**🐛 BUGFIX CRÍTICO: Avatares não carregavam em DEV deployado**

**Problema:**
- Imagens de avatares bloqueadas por CORS em https://dev.mundotalendo.com.br
- Erro: "Access to image at 'https://i.pravatar.cc/150?img=44' has been blocked by CORS policy"

**Causa:**
- Código usava `process.env.NODE_ENV === 'development'` para decidir usar proxy
- Em DEV deployado, `NODE_ENV` é sempre "production" (não "development")
- Proxy não era usado → CORS bloqueava imagens externas

**Solução:**
- ✅ Sempre usar proxy `/api/proxy-image` para URLs externas
- ✅ Remove verificação de `NODE_ENV`
- ✅ Proxy já tem cache de 24 horas configurado

**Arquivos modificados:**
- `src/components/Map.jsx` (linha 469): Remove condicional NODE_ENV
- `package.json`: Version bump 1.0.5

**Impacto:**
- Fix apenas no frontend
- Nenhuma mudança no backend
- Compatível com v1.0.4

---

### ✅ v1.0.4: User Markers - Círculos Concêntricos (03 Jan 2026)

**🎨 MELHORIA VISUAL: Círculos concêntricos para marcadores GPS**

**Problema resolvido:**
- Marcadores em linha horizontal criavam "linha bizarra cortando continentes"
- `horizontalSpacing = 2.5` graus → com 10+ usuários = linha de 25 graus
- Visual quebrado e confuso com múltiplos usuários no mesmo país

**Solução implementada:**
- ✅ **Círculos concêntricos** ao redor do centroid do país (360° completo)
- ✅ **Múltiplos anéis** para escalabilidade massiva (1-1000+ usuários)
- ✅ **Raio dinâmico limitado** - cresce até capacidade, depois adiciona anel
- ✅ **Distribuição uniforme** em 360° (conversão polar → cartesiano)

**Constantes configuráveis:**
```javascript
RING_BASE_RADIUS = 1.2       // graus - raio do primeiro anel
RING_INCREMENT = 0.9          // graus - incremento entre anéis
MIN_SPACING_DEGREES = 0.35    // graus - espaçamento mínimo entre avatares
```

**Capacidade por anel:**
- Anel 1 (r=1.2): ~21 usuários
- Anel 2 (r=2.1): ~38 usuários
- Anel 3 (r=3.0): ~54 usuários
- Anel 52 (r=47.1): suporta 1000+ usuários totais

**Arquivos modificados:**
- `src/components/Map.jsx`:
  - Adicionadas constantes (linhas 17-19)
  - Nova função `distributeUsersInRings()` (linhas 26-50)
  - Lógica circular em `buildUserMarkersGeoJSON()` (linhas 97-133)
- `CLAUDE.md` - Documentação atualizada

**Impacto:**
- **Nenhum impacto** em dados existentes (mudança apenas visual/frontend)
- Backend `/users/locations` não muda
- DynamoDB não é afetado
- Compatível com v1.0.3

**Testes:**
- ✅ Compilação local bem-sucedida (`npm run dev:local`)
- ✅ Sem erros de JavaScript
- ✅ Página carregando corretamente
- ⏳ Validação visual pendente (aguardando usuários reais)

---

### ✅ v1.0.3: Critical Bugfixes & Stability (25 Dez 2025)

**🔴 CORREÇÃO EMERGENCIAL: v1.0.2 quebrou site inteiro!**

**Três bugs críticos corrigidos:**

1. **Bug #1: PK Mismatch - Stats/Users retornando vazio**
   - **Problema:** Webhook escrevia `EVENT#LEITURA#<uuid>` mas stats/users consultavam `EVENT#LEITURA`
   - **Impacto:** DynamoDB queries retornavam 0 resultados → mapa sem cores, sem GPS markers
   - **Fix:** PK revertido para `"EVENT#LEITURA"`, UUID movido para campo separado `webhookUUID`
   - **Validação:** `/stats` retornou 174 países, mapa funcional

2. **Bug #2: SK Sobrescrevia Livros Duplicados**
   - **Problema:** SK usando apenas `COUNTRY#<iso3>` não era único por livro
   - **Impacto:** Múltiplos livros no mesmo país → apenas último era salvo, perda de dados
   - **Fix:** SK alterado para `<uuid>#<iso3>#<index>` garantindo unicidade
   - **Validação:** Países com 3-5 livros verificados no DynamoDB

3. **Bug #3: deleteOldUserReadings Apagava Payloads**
   - **Problema:** Função deletava TODOS os items do usuário, incluindo `WEBHOOK#PAYLOAD#*`
   - **Impacto:** Payloads apareciam salvos em logs mas eram imediatamente deletados
   - **Fix:** Adicionado filtro `if !strings.HasPrefix(pkAttr.Value, "EVENT#LEITURA")`
   - **Validação:** Payload `WEBHOOK#PAYLOAD#73d590f5...` confirmado no DynamoDB

**Melhorias de estabilidade:**

4. **GPS Markers apenas para progresso ≥ 1%**
   - Filtro adicionado: `if reading.Progresso < 1 { continue }`
   - Evita marcadores em livros não iniciados

5. **Force Rebuild em Todos os Deploys**
   - **Problema:** SST usava builds em cache, código antigo sendo deployado
   - **Fix:** Makefile agora deleta e recompila todos os binários Go antes do deploy
   - Afeta: `deploy-dev` e `deploy-prod`

6. **get-api-key Retornando Múltiplos Valores**
   - **Problema:** Comando retornava várias API keys, quebrando curl
   - **Impacto:** `make webhook-full` e `make stats` falhavam silenciosamente
   - **Fix:** Adicionado `| head -1` para retornar apenas primeira key ativa

7. **12 Correções de Mapeamento de Países**
   - República Tcheca → Tchéquia
   - Cingapura → Singapura
   - Holanda → Países Baixos
   - Tajiquistão → Tadjiquistão
   - Timor Leste → Timor-Leste
   - San Marino → São Marino
   - Djibuti → Djibouti
   - Congo-Brazzaville → Congo
   - Seicheles → Seychelles
   - Trinidad e Tobago → Trindade e Tobago
   - São Vicente e Granadinas → São Vicente e Grandinas
   - Palestina → Estado da Palestina

**Novas features:**

8. **Comando `make webhook-full`**
   - Gera webhook com TODOS os 185 países (2-5 livros cada)
   - Dados randomizados: progresso 1-100%, datas variadas
   - Útil para popular ambiente DEV com dados realistas
   - **Protegido:** DEV-only, bloqueado em produção

9. **Comando `make stats`** e **`make users`**
   - Fetch rápido de estatísticas e localizações da API
   - Usa API key automaticamente
   - Output formatado com jq
   - **Suporta STAGE=prod:** `make stats STAGE=prod` funciona em produção

10. **Suporte STAGE em comandos Makefile**
    - 6 comandos suportam `STAGE=prod`: stats, users, get-api-key, fix-env, update-secret, unlock
    - 4 comandos protegidos DEV-only: seed, clear, webhook-test, webhook-full
    - Proteção de segurança impede comandos destrutivos em produção

11. **Campo `updatedAt` Adicionado**
    - Salva timestamp RFC3339 do último update do livro
    - Usado para determinar livro mais recente quando usuário lê múltiplos
    - GPS marker aparece no país do livro com maior `updatedAt`

**Arquitetura Final v1.0.3:**

**Estrutura de dados DynamoDB:**
```
LeituraItem (eventos de leitura):
- PK: "EVENT#LEITURA"                    ← Simples, queries funcionam
- SK: "<uuid>#<iso3>#<index>"            ← Único por livro
- webhookUUID: "<uuid>"                  ← Rastreamento de execução
- updatedAt: "2025-12-25T14:30:00Z"      ← Ordenação temporal
- iso3, pais, categoria, progresso, user, imagemURL, livro

WebhookItem (payload original):
- PK: "WEBHOOK#PAYLOAD#<uuid>"           ← Salvo UMA VEZ
- SK: "TIMESTAMP#<RFC3339>"
- user, payload (JSON completo)

FalhaItem (erros):
- PK: "ERROR#<uuid>"
- SK: "TIMESTAMP#<RFC3339>"
- errorType, errorMessage, originalPayload
```

**Rastreamento completo mantido:**
- Webhook UUID: Agrupa todos eventos de uma execução
- Payload salvo separadamente (não deletado)
- Múltiplos livros por país suportados
- Queries eficientes (`PK = "EVENT#LEITURA"`)
- Auto-cleanup protege payloads

**Breaking changes:**
- SK mudou de `COUNTRY#<iso3>` para `<uuid>#<iso3>#<index>`
- Campo `webhookUUID` agora obrigatório
- Campo `updatedAt` agora obrigatório
- Users endpoint usa `updatedAt` para ordenação (não SK)

**Arquivos modificados:**
- `types/types.go` - Campos `WebhookUUID` e `UpdatedAt` em LeituraItem
- `webhook/main.go` - PK simples, SK único, proteção de payloads, import strings
- `users/main.go` - Comparação por `UpdatedAt`, filtro `progresso >= 1`
- `Makefile` - 10 fixes críticos:
  - Suporte STAGE em 6 comandos (stats, users, get-api-key, fix-env, update-secret, unlock)
  - Proteção DEV-only em 4 comandos (seed, clear, webhook-test, webhook-full)
  - Novo comando `make users` para GET /users/locations
  - Force rebuild usando subshells em build/tidy
  - 12 nomes de países corrigidos em webhook-full
- `.gitignore` - Regras para ignorar binários Go compilados
- `CLAUDE.md` - Seção completa "🔧 Comandos Make e STAGE", changelog v1.0.3 atualizado
- `package.json` - Version bump 1.0.3

**Testes:**
- ✅ 26 Go unit tests passando
- ✅ Stats retornando 174 países
- ✅ Múltiplos livros por país confirmados (3-5 livros)
- ✅ Payload salvo e preservado
- ✅ GPS markers filtrados (progresso >= 1%)
- ✅ Deploy force rebuild funcionando

**Migração de dados:**
- Dados v1.0.2 ficam órfãos mas inofensivos
- Próximo webhook do usuário limpa dados antigos automaticamente
- Sem necessidade de migração manual

### ✅ v1.0.2: UUID Architecture & Storage Optimization (25 Dez 2025)

**🚀 OTIMIZAÇÃO MASSIVA: 99% de redução em storage!**

**Arquitetura UUID implementada:**
- ✅ **Payload salvo UMA VEZ** por webhook (`WEBHOOK#PAYLOAD#<uuid>`)
- ✅ **Eventos agrupados** por UUID (`EVENT#LEITURA#<uuid>`)
- ✅ **Erros rastreáveis** com UUID (`ERROR#<uuid>`)
- ✅ **Auto-limpeza** de dados antigos do usuário (mantém apenas última interação)
- ✅ **GSI UserIndex** para queries eficientes por usuário
- ✅ **Validação 100%** dos 195 países do Maratona.app (203 variações)
- ✅ **Threshold ajustado** para ≥1% (países com 0% aparecem cinza)

**Impacto esperado:**
- Storage: 2.9 GB → 35 MB (99% de redução)
- Writes: 195 payloads → 1 payload por webhook (99% menos writes)
- Queries: Items menores (<1KB vs ~50KB) = mais rápidas
- Custo: ~99% de economia em storage + write operations

**Breaking changes:**
- Novos webhooks usam `EVENT#LEITURA#<uuid>` (vs `EVENT#LEITURA` antigo)
- Primeiro webhook após deploy limpa dados antigos automaticamente
- `/stats` não retorna mais países com 0% de progresso

**Arquivos modificados:**
- `types/types.go` - WebhookItem struct, metadata removido
- `webhook/main.go` - UUID functions, deleteOldUserReadings()
- `stats/main.go` - Filtro progress >= 1
- `mapping/iso.go` - 4 novos aliases (Azerbajão, Cabo verde, Irlanda do norte, Suíça)
- `mapping/iso_validation_test.go` - Validação completa de países
- `sst.config.ts` - GSI UserIndex

### ✅ NOVA FEATURE: User Markers GPS-Style (23 Dez 2025 - Atualizado 03 Jan 2026)

**Marcadores de usuários no mapa:**
- ✅ **Avatar circular** dos usuários exibido no mapa (estilo GPS)
- ✅ **Círculos concêntricos** ao redor do nome do país (360° completo) - **Atualizado v1.0.4**
- ✅ **Distribuição em múltiplos anéis** para acomodar 1-1000+ usuários por país
- ✅ **Tooltip ao hover** mostrando nome do usuário e livro sendo lido
- ✅ **Proxy de imagens** para resolver CORS em desenvolvimento
- ✅ **Recorte circular** das imagens usando canvas (fill completo do círculo)
- ✅ **Feature flag** `NEXT_PUBLIC_SHOW_USER_MARKERS` (ON em dev, OFF em prod até validação)
- ✅ **Novo endpoint** `/users/locations` retornando última localização de cada usuário

**Implementação técnica:**
- Backend extrai título do livro de `vinculados[].edicao.titulo`
- Frontend usa MapLibre sprites com ImageData (canvas → circular clip)
- Proxy Next.js API route (`/api/proxy-image`) para bypass CORS
- Dados salvos: `user`, `avatarURL`, `livro`, `iso3`, `pais`, `timestamp`

**Algoritmo de círculos concêntricos (v1.0.4):**
- **Anel 1:** Raio 1.2° - capacidade ~21 usuários
- **Anel 2:** Raio 2.1° - capacidade ~38 usuários
- **Anéis subsequentes:** Incremento de 0.9° entre anéis
- **Capacidade dinâmica:** Calculada como `(2π * raio) / MIN_SPACING_DEGREES`
- **Escalabilidade:** Suporta 1000+ usuários (~52 anéis concêntricos)
- **Distribuição:** Usuários posicionados uniformemente em 360° ao redor do centroid
- **Coordenadas:** Conversão polar → cartesiano: `offsetLng = r * cos(θ)`, `offsetLat = r * sin(θ)`

### ✅ PRODUÇÃO-READY (21 Dez 2025)

**Todas as melhorias críticas implementadas:**
- ✅ **3 bugs críticos** corrigidos (Go 1.25, JSON errors, HTTP status codes)
- ✅ **3 otimizações de performance** (paginação, polling 60s, CORS)
- ✅ **1 essencial de produção** (DynamoDB PITR backups)
- ✅ **4 melhorias de UX** (Error Boundary, retry logic, validation)
- ✅ **3 features opcionais** (concurrency limits, security headers, cleanup)
- ✅ **Todos os testes passando** (26 Go tests + frontend tests)
- ✅ **Deploy DEV funcionando** perfeitamente

**Sistema completo e operacional:**
1. **Backend:** 5 Lambdas Go otimizadas (webhook, stats, users, seed, clear)
2. **Frontend:** Next.js 16 com Error Boundary, retry logic, security headers, user markers
3. **Infraestrutura:** DynamoDB com backups, Lambda com concurrency limits
4. **Performance:** Polling 60s (stats + users), paginação DynamoDB, validações robustas
5. **Segurança:** CORS restrito, input validation, API key authentication, proxy de imagens

### 🧹 Cleanup us-east-1 (23 Dez 2025)

**Problema:** Recursos antigos foram criados acidentalmente em **us-east-1** (Virginia) com stage "danielbalieiro"

**Causa:** SST usa o nome do usuário git como stage padrão quando `--stage` não é especificado. Recursos foram criados na região errada durante testes iniciais.

**Solução:** Todos os recursos em us-east-1 foram deletados com sucesso:
- ✅ 1 CloudFormation stack (`danielbalieiro-mundotalendo-Stack`)
- ✅ 5 Lambda functions
- ✅ 1 DynamoDB table (vazia)
- ✅ 1 API Gateway
- ✅ 4 CloudWatch Log Groups

**Região oficial do projeto:** **us-east-2** (Ohio)
- Prod: `mundotalendo-prod-*`
- Dev: `mundotalendo-dev-*`

**IMPORTANTE:** Sempre usar `--region us-east-2` em comandos AWS CLI

### 🔧 SST Transform Fix (Crítico)

**Problema resolvido:** Variáveis de ambiente não eram configuradas automaticamente

**Causa:** `transform.function: { ... }` com objeto estático substituía completamente a config, removendo as env vars do `link: [dataTable]`

**Solução:**
```typescript
// ❌ ANTES (errado)
transform: { function: { reservedConcurrentExecutions: 10 } }

// ✅ DEPOIS (correto)
transform: { function: (args) => { args.reservedConcurrentExecutions = 10 } }
```

**Resultado:** Deploy agora configura automaticamente `SST_Resource_DataTable_name` em todas as Lambdas

### 🌍 Limitação Conhecida (Aceita)

**Vector Tile - Territórios ultramarinos:**
- Vector tile (`demotiles.maplibre.org`) não separa territórios ultramarinos
- GUF aparece como FRA, GRL como DNK, PRI como USA
- **Decisão:** Aceitar limitação - visual permanece consistente e funcional
- Alternativa (não implementada): GeoJSON completo (2.7MB comprimido)

## 🏗️ Arquitetura Técnica

### Stack

| Componente | Tecnologia | Versão | Notas |
|------------|-----------|---------|-------|
| IaC | SST (Ion) | 3.17.25 | Bug conhecido |
| Frontend | Next.js | 16.0.10 | App Router, JavaScript + JSDoc |
| Bundler | Webpack | - | Turbopack tem conflitos com MapLibre |
| CSS | Tailwind CSS | v4 | Novo plugin @tailwindcss/postcss |
| Maps | MapLibre GL JS | 5.14.0 | Implementação direta (não react-map-gl) |
| Data Fetching | SWR | 2.3.8 | Polling 15s |
| Backend | Go | 1.23+ | ARM64/Graviton |
| Database | DynamoDB | - | Single Table Design |
| API | API Gateway V2 | - | HTTP API (não REST) |
| Region | AWS | us-east-2 | Ohio |

### Recursos AWS Ativos

**API Gateway:**
- URL Dev: `https://api.dev.mundotalendo.com.br` ✅
- URL Raw: `https://q9f0i3fp0d.execute-api.us-east-2.amazonaws.com`
- Domínio custom configurado via SST

**Lambda Functions:**
```
mundotalendo-danielbalieiro-WebhookFunctionFunction-snobkmoh
mundotalendo-danielbalieiro-StatsFunctionFunction-zdvhcmhx
mundotalendo-danielbalieiro-SeedFunctionFunction-kzdzkknw
```

**DynamoDB:**
- Table: `mundotalendo-danielbalieiro-LeiturasTable-hdkkstmu`
- PK: "EVENT#LEITURA"
- SK: "TIMESTAMP#<RFC3339>"
- Attributes: iso3, pais, categoria, status, user

**CloudFormation Stack:**
- Name: `danielbalieiro-mundotalendo-Stack`
- Status: CREATE_COMPLETE (mas incompleto devido ao bug)

### Integrações API Gateway (Manual)

```
POST /webhook → Integration: 2k281yc → Route: 97x0ce6
GET /stats    → Integration: bjwg0eo → Route: qervgwm
POST /test/seed → Integration: r37e3qb → Route: r4b7jx7
```

Permissões Lambda adicionadas manualmente via `aws lambda add-permission`.

## 📝 Histórico de Decisões Técnicas

### Por que não TypeScript?

Decisão do projeto inicial: JavaScript + JSDoc para simplicidade.

### Por que não react-map-gl?

**Erro encontrado:** "Package path . is not exported from package react-map-gl"

**Solução:** Implementação direta do MapLibre GL JS - mais controle, menos abstrações.

### Por que Webpack em vez de Turbopack?

**Erro encontrado:** "This build is using Turbopack, with a `webpack` config"

**Solução:** Next.js 16 usa Turbopack por padrão, mas MapLibre precisa de alias específico. Configurado both em `next.config.js`:
```javascript
turbopack: {
  resolveAlias: { 'maplibre-gl': 'maplibre-gl/dist/maplibre-gl.js' }
},
webpack: (config) => {
  config.resolve.alias['maplibre-gl'] = 'maplibre-gl/dist/maplibre-gl.js'
  return config
}
```

### Por que centroids em vez de labels no vector tile?

**Problema:** Vector tiles têm múltiplas geometrias por país (ilhas, territórios)
- Brasil aparecia 4-6 vezes
- EUA aparecia 3x (continente + Alasca + Havaí)

**Solução:** Arquivo `src/config/countryCentroids.js` com exatamente 1 ponto por país.

### Por que go.mod em cada função Lambda?

**Erro SST:** "package stats is not in std"

**Solução:** SST requer go.mod individual em cada diretório de função:
```
packages/functions/webhook/go.mod
packages/functions/stats/go.mod
packages/functions/seed/go.mod
```

Cada um com `replace github.com/mundotalendo/functions => ..`

### Por que remover "type": "commonjs" do package.json?

**Erro:** "Specified module format (CommonJs) is not matching"

**Solução:** Next.js 16 espera ESM. Remover essa linha resolve conflito.

### Por que @tailwindcss/postcss?

**Erro:** "PostCSS plugin has moved to a separate package"

Tailwind CSS v4 mudou arquitetura:
- Antes: `tailwindcss` plugin
- Agora: `@tailwindcss/postcss` plugin
- globals.css: `@import "tailwindcss"` em vez de `@tailwind`

## 🐛 Bugs Resolvidos (Referência)

### 1. Labels duplicadas no mapa

**Sintoma:** Brasil aparecia 4-12 vezes no mapa

**Causa:** Vector tiles (`demotiles.maplibre.org`) têm múltiplas features por país

**Fix:** Criado `countryCentroids.js` com GeoJSON custom

### 2. Cores desaparecendo do mapa

**Sintoma:** "A cor carrega e some"

**Causa:** Closure problem - `applyCountryColors` tinha referência stale de `countries`

**Fix:** Wrapped em `useCallback` com dependência `[countries]`

### 3. MapLibre match expression com array vazio

**Erro:** "Expected at least 4 arguments, but found only 2"

**Causa:** `countries.length === 0` → match expression inválido

**Fix:** Check antes de criar expression:
```javascript
if (countries.length === 0) {
  map.current.setPaintProperty('country-fills', 'fill-color', '#F5F5F5')
  return
}
```

### 4. Lambda não encontra table name

**Sintoma:** "Member must have length greater than or equal to 1"

**Causa:** SST resource linking quebrado pelo bug do deploy

**Fix:** Adicionar variável de ambiente manualmente:
```bash
aws lambda update-function-configuration \
  --environment "Variables={...,SST_Resource_Leituras_name=mundotalendo-danielbalieiro-LeiturasTable-hdkkstmu}"
```

## 🔧 Workaround Manual do Bug SST

### Comando completo para recriar integração

Se precisar refazer ou criar para novo stage:

```bash
# 1. Pegar ARNs dos Lambdas
WEBHOOK_ARN=$(aws lambda get-function --function-name mundotalendo-danielbalieiro-WebhookFunctionFunction-snobkmoh --region us-east-2 --query 'Configuration.FunctionArn' --output text)
STATS_ARN=$(aws lambda get-function --function-name mundotalendo-danielbalieiro-StatsFunctionFunction-zdvhcmhx --region us-east-2 --query 'Configuration.FunctionArn' --output text)
SEED_ARN=$(aws lambda get-function --function-name mundotalendo-danielbalieiro-SeedFunctionFunction-kzdzkknw --region us-east-2 --query 'Configuration.FunctionArn' --output text)

# 2. Criar integrações
WEBHOOK_INT=$(aws apigatewayv2 create-integration --region us-east-2 \
  --api-id q9f0i3fp0d \
  --integration-type AWS_PROXY \
  --integration-uri $WEBHOOK_ARN \
  --payload-format-version 2.0 \
  --query 'IntegrationId' --output text)

STATS_INT=$(aws apigatewayv2 create-integration --region us-east-2 \
  --api-id q9f0i3fp0d \
  --integration-type AWS_PROXY \
  --integration-uri $STATS_ARN \
  --payload-format-version 2.0 \
  --query 'IntegrationId' --output text)

SEED_INT=$(aws apigatewayv2 create-integration --region us-east-2 \
  --api-id q9f0i3fp0d \
  --integration-type AWS_PROXY \
  --integration-uri $SEED_ARN \
  --payload-format-version 2.0 \
  --query 'IntegrationId' --output text)

# 3. Criar rotas
aws apigatewayv2 create-route --region us-east-2 \
  --api-id q9f0i3fp0d \
  --route-key "POST /webhook" \
  --target "integrations/$WEBHOOK_INT"

aws apigatewayv2 create-route --region us-east-2 \
  --api-id q9f0i3fp0d \
  --route-key "GET /stats" \
  --target "integrations/$STATS_INT"

aws apigatewayv2 create-route --region us-east-2 \
  --api-id q9f0i3fp0d \
  --route-key "POST /test/seed" \
  --target "integrations/$SEED_INT"

# 4. Permissões
aws lambda add-permission --region us-east-2 \
  --function-name mundotalendo-danielbalieiro-WebhookFunctionFunction-snobkmoh \
  --statement-id apigateway-webhook \
  --action lambda:InvokeFunction \
  --principal apigateway.amazonaws.com \
  --source-arn "arn:aws:execute-api:us-east-2:219024422667:q9f0i3fp0d/*/*"

aws lambda add-permission --region us-east-2 \
  --function-name mundotalendo-danielbalieiro-StatsFunctionFunction-zdvhcmhx \
  --statement-id apigateway-stats \
  --action lambda:InvokeFunction \
  --principal apigateway.amazonaws.com \
  --source-arn "arn:aws:execute-api:us-east-2:219024422667:q9f0i3fp0d/*/*"

aws lambda add-permission --region us-east-2 \
  --function-name mundotalendo-danielbalieiro-SeedFunctionFunction-kzdzkknw \
  --statement-id apigateway-seed \
  --action lambda:InvokeFunction \
  --principal apigateway.amazonaws.com \
  --source-arn "arn:aws:execute-api:us-east-2:219024422667:q9f0i3fp0d/*/*"

# 5. Atualizar variáveis de ambiente nos Lambdas
TABLE_NAME=$(aws dynamodb list-tables --region us-east-2 --query 'TableNames[?contains(@, `mundotalendo`) && contains(@, `Leituras`)]' --output text)

aws lambda update-function-configuration --region us-east-2 \
  --function-name mundotalendo-danielbalieiro-StatsFunctionFunction-zdvhcmhx \
  --environment "Variables={SST_RESOURCE_App='{\"name\":\"mundotalendo\",\"stage\":\"danielbalieiro\"}',SST_KEY=wkaOlOssBKiBSw1dtJA1PbTJjoCZTHQbnzkirsTdAQw=,SST_KEY_FILE=resource.enc,SST_Resource_Leituras_name=$TABLE_NAME}"

aws lambda update-function-configuration --region us-east-2 \
  --function-name mundotalendo-danielbalieiro-WebhookFunctionFunction-snobkmoh \
  --environment "Variables={SST_RESOURCE_App='{\"name\":\"mundotalendo\",\"stage\":\"danielbalieiro\"}',SST_KEY=wkaOlOssBKiBSw1dtJA1PbTJjoCZTHQbnzkirsTdAQw=,SST_KEY_FILE=resource.enc,SST_Resource_Leituras_name=$TABLE_NAME}"

aws lambda update-function-configuration --region us-east-2 \
  --function-name mundotalendo-danielbalieiro-SeedFunctionFunction-kzdzkknw \
  --environment "Variables={SST_RESOURCE_App='{\"name\":\"mundotalendo\",\"stage\":\"danielbalieiro\"}',SST_KEY=wkaOlOssBKiBSw1dtJA1PbTJjoCZTHQbnzkirsTdAQw=,SST_KEY_FILE=resource.enc,SST_Resource_Leituras_name=$TABLE_NAME}"
```

## 📁 Estrutura de Arquivos Importantes

### Frontend

```
src/
├── app/
│   ├── layout.js           # Layout raiz, sem MapLibre CSS aqui
│   ├── page.js             # Página principal, título "Mundo Tá Lendo 2026"
│   ├── globals.css         # @import "tailwindcss" + maplibre CSS
│   └── api/stats/route.js  # Mock API para dev local
├── components/
│   └── Map.jsx             # Mapa MapLibre GL JS com centroids, cores vibrantes
├── config/
│   ├── countries.js        # ISO → Nome PT-BR (193 países)
│   ├── countryCentroids.js # ISO → [lng, lat] (1 ponto exato por país)
│   └── months.js           # 12 meses → cores vibrantes → países
└── hooks/
    └── useStats.js         # SWR polling 15s com fallback /api local
```

### Backend

```
packages/functions/
├── types/
│   └── types.go           # Structs compartilhados
├── mapping/
│   └── countries.go       # Nome país → ISO code
├── webhook/
│   ├── go.mod            # Module individual (replace ..)
│   └── main.go           # Processa webhook Maratona.app
├── stats/
│   ├── go.mod
│   └── main.go           # Query DynamoDB, retorna ISOs únicos
└── seed/
    ├── go.mod
    └── main.go           # Popula DB com países aleatórios
```

### Config

```
├── sst.config.ts          # SST Ion config
├── next.config.js         # Turbopack + Webpack aliases
├── postcss.config.js      # @tailwindcss/postcss
├── package.json           # dev:local usa --webpack
└── .env.local            # NEXT_PUBLIC_API_URL (AWS ou /api)
```

## 🎨 Design Decisions

### Cores

**Oceano:** `#0077BE` (azul vibrante)

**Países não explorados:** `#F5F5F5` (cinza claro)

**12 Meses:** Cores vibrantes definidas em `src/config/months.js`
- Janeiro: #FF1744 (vermelho vibrante)
- Fevereiro: #00E5FF (ciano brilhante)
- Março: #FFEA00 (amarelo limão)
- etc.

### Mapa

**Zoom:**
- Inicial: 1.5
- Min: 1
- Max: 6 (previne divisões estaduais)

**Labels:**
- Português PT-BR obrigatório
- 1 label por país (via centroids)
- Font: Noto Sans Bold 12px
- Halo branco para legibilidade

**Interatividade:**
- Hover: cursor pointer
- Click: exibe nome do país (futuro)
- Tooltip com nome PT-BR

## 🔮 Próximos Passos Sugeridos

### Curto Prazo

1. **Resolver bug SST ou fazer deploy manual completo do Next.js**
   - Opções: aguardar SST fix, testar versão anterior, ou deploy manual S3+CloudFront

2. **Configurar CloudFront para Next.js**
   - Apontar para S3 bucket do build
   - Configurar cache policies
   - Conectar com API Gateway

3. **Re-habilitar domínios Cloudflare**
   - Descomentar config no sst.config.ts
   - Configurar DNS records

### Médio Prazo

4. **Implementar cache no stats endpoint**
   - Lambda muito requisitado (polling 15s)
   - Considerar cache DynamoDB DAX ou Lambda cache layer

5. **Adicionar telemetria de participantes**
   - Mostrar quem leu cada país
   - Lista de leituras por país
   - Timeline de descobertas

6. **Melhorar mapa**
   - Animações de transição quando país é explorado
   - Popup com detalhes ao clicar
   - Filtro por mês

### Longo Prazo

7. **Dashboard admin**
   - Moderação de leituras
   - Estatísticas agregadas
   - Gestão de usuários

8. **Notificações**
   - WebSocket para updates em tempo real
   - Celebração quando país é explorado pela primeira vez

9. **Gamification leve**
   - "Badges" por regiões completadas
   - Progresso coletivo

## 🔧 Comandos Make e STAGE

### Ambientes (STAGE)

O projeto tem dois ambientes:
- **dev** (padrão) - Desenvolvimento e testes
- **prod** - Produção com dados reais

### Comandos que suportam STAGE=prod

Para executar comandos em produção, use `STAGE=prod`:

```bash
# DEV (padrão)
make stats

# PROD
make stats STAGE=prod
```

**Comandos que suportam STAGE=prod:**
- `make stats STAGE=prod` - Ver estatísticas de produção
- `make users STAGE=prod` - Ver localizações de usuários em produção
- `make get-api-key STAGE=prod` - Pegar API key de produção
- `make fix-env STAGE=prod` - Fixar env vars de produção
- `make update-secret STAGE=prod` - Atualizar SST Secret com API key de produção
- `make unlock STAGE=prod` - Desbloquear deploy travado em produção

**Comandos DEV-ONLY (bloqueados em prod por segurança):**
- `make seed` - Popular database (apenas DEV)
- `make clear` - Limpar database (apenas DEV)
- `make webhook-test` - Testar webhook (apenas DEV)
- `make webhook-full` - Gerar todos os países (apenas DEV)

Se tentar usar comandos DEV-ONLY com `STAGE=prod`, você receberá erro:
```bash
make seed STAGE=prod
# Error: seed command is DEV-only for safety.
```

### Referência Rápida de Comandos

**Deploy e Infraestrutura:**
```bash
make deploy-dev          # Deploy completo para DEV
make deploy-prod         # Deploy completo para PROD (pede confirmação)
make unlock              # Desbloquear deploy travado (DEV)
make unlock STAGE=prod   # Desbloquear deploy travado (PROD)
```

**Consultas e Testes:**
```bash
make stats               # Ver estatísticas (DEV)
make stats STAGE=prod    # Ver estatísticas (PROD)
make users               # Ver localizações de usuários (DEV)
make users STAGE=prod    # Ver localizações de usuários (PROD)
make seed                # Popular BD com dados de teste (DEV only)
make clear               # Limpar BD (DEV only)
make webhook-full        # Gerar webhook com todos os países (DEV only)
```

**API Keys:**
```bash
make get-api-key                    # Pegar API key (DEV)
make get-api-key STAGE=prod         # Pegar API key (PROD)
make create-api-key name=test       # Criar API key (DEV)
make create-api-key-prod name=test  # Criar API key (PROD)
make list-api-keys                  # Listar API keys (DEV)
make list-api-keys-prod             # Listar API keys (PROD)
```

**Monitoring:**
```bash
make logs-all            # Ver logs em tempo real (DEV)
make logs-all-prod       # Ver logs em tempo real (PROD)
make alarms              # Ver status dos alarmes (DEV)
make alarms-prod         # Ver status dos alarmes (PROD)
make info                # Ver recursos AWS (DEV)
make info-prod           # Ver recursos AWS (PROD)
```

## 🧪 Testes

### Testar API manualmente

**DEV (padrão):**
```bash
# Stats - ver países sendo lidos
make stats

# Users - ver localizações de usuários
make users

# Seed - adicionar países aleatórios
make seed

# Clear - limpar todos os dados
make clear
```

**PROD:**
```bash
# Stats - ver estatísticas reais
make stats STAGE=prod

# Users - ver localizações reais
make users STAGE=prod

# ⚠️ seed, clear, webhook-test NÃO funcionam em prod
# (são comandos apenas para ambiente de desenvolvimento)
```

**Testar com curl direto:**
```bash
# DEV
curl https://api.dev.mundotalendo.com.br/stats \
  -H "X-API-Key: $(make get-api-key -s)"

# PROD
curl https://api.mundotalendo.com.br/stats \
  -H "X-API-Key: $(STAGE=prod make get-api-key -s)"
```

### Testar frontend localmente

```bash
# Terminal 1: Start dev server
npm run dev:local

# Terminal 2: Verificar carregamento
curl http://localhost:3000 | grep "Mundo Tá Lendo"

# Browser: http://localhost:3000
# Deve mostrar mapa com países coloridos
```

## 📊 Dados em Produção

**🔴 ATENÇÃO:** O banco de dados agora contém **dados reais de participantes**.

### Comandos seguros para produção

✅ **Comandos de LEITURA** (seguros em produção):
```bash
# Ver estatísticas reais
make stats STAGE=prod

# Ver localizações de usuários
make users STAGE=prod

# Ver recursos AWS
make info-prod

# Ver logs
make logs-all-prod
```

🚫 **Comandos BLOQUEADOS** em produção (DEV-only):
- `make seed` - Popular database (protegido)
- `make clear` - Limpar database (protegido)
- `make webhook-test` - Testar webhook (protegido)
- `make webhook-full` - Gerar webhook completo (protegido)

Estes comandos têm **proteção de segurança** e retornam erro se tentar usar com `STAGE=prod`.

### Verificar dados DynamoDB

```bash
# Contar leituras em produção
PROD_TABLE=$(aws dynamodb list-tables --region us-east-2 \
  --query 'TableNames[?contains(@, `mundotalendo-prod-DataTable`)]' --output text)
aws dynamodb scan --table-name $PROD_TABLE --select COUNT --region us-east-2
```

### Para testes locais

- Use ambiente local com mock data
- Configure `NEXT_PUBLIC_API_URL=/api` no `.env.local`
- Ou use ambiente DEV: `make stats` (sem STAGE=prod)

## ⚠️ Avisos Importantes

1. **🔴 SISTEMA EM PRODUÇÃO** - Dados reais de participantes, mudanças exigem extremo cuidado
2. **Não usar "conquista" ou "conquered"** - projeto é sobre descoberta cultural
3. **Labels devem estar em português** - sempre PT-BR
4. **1 label por país** - usar centroids, não vector tiles
5. **Cores vibrantes** - oceano azul, países com cores dos meses
6. **SST 3.17.25 tem bug** - workaround manual necessário
7. **Não usar react-map-gl** - implementação direta MapLibre
8. **Webpack, não Turbopack** - via npm run dev:local
9. **Tailwind CSS v4** - nova sintaxe com @tailwindcss/postcss
10. **🔒 Comandos DEV-only protegidos** - `make seed`, `make clear`, `make webhook-test`, `make webhook-full` são bloqueados em produção por segurança (não funcionam com `STAGE=prod`)
11. **✅ Use STAGE=prod para consultas** - `make stats STAGE=prod` e `make users STAGE=prod` são seguros para consultar dados de produção

## 🔗 Links Úteis

- API Dev: https://api.dev.mundotalendo.com.br
- Frontend Dev: https://dev.mundotalendo.com.br
- Local Dev: http://localhost:3000
- Vector Tiles: https://demotiles.maplibre.org
- SST Docs: https://sst.dev

## 📞 Integração Webhook

O endpoint `/webhook` espera payload do Maratona.app com estrutura:
- `perfil.nome` (string) - nome do participante
- `desafios[]` (array) - lista de desafios
  - `descricao` (string) - nome do país
  - `categoria` (string) - mês/categoria
  - `concluido` (boolean) - se foi completado
  - `tipo` (string) - "leitura" ou "atividade"
  - `vinculados[]` (array) - leituras vinculadas
    - `completo` (boolean) - se a leitura foi completa
    - `updatedAt` (timestamp) - quando foi atualizado

**Processamento:**
1. Filtra apenas `tipo === "leitura" && concluido === true`
2. Verifica se há `vinculados[].completo === true`
3. Converte nome do país para ISO code via mapping
4. Salva no DynamoDB com timestamp RFC3339

---

## 📝 Notas Finais

**🔴 LEMBRETE CRÍTICO:** Este projeto está **EM PRODUÇÃO** com **dados reais de participantes**.

Qualquer mudança no código, schema de dados, ou lógica de processamento pode impactar:
- Leituras já registradas no DynamoDB
- Experiência de usuários ativos
- Integridade dos dados históricos
- Funcionamento do webhook em produção

**Antes de fazer qualquer alteração:**
1. Leia atentamente a seção "🚨 ATENÇÃO: SISTEMA EM PRODUÇÃO" acima
2. Teste exaustivamente em ambiente local
3. Valide compatibilidade com dados existentes
4. Documente mudanças neste arquivo
5. Tenha um plano de rollback preparado

**Esta documentação** deve ser mantida atualizada conforme o projeto evolui. É a fonte de verdade para contexto técnico em futuras sessões.
