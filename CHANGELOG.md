# Changelog

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/),
e este projeto adere ao [Semantic Versioning](https://semver.org/lang/pt-BR/).

## [1.0.2] - 2025-12-25

### 🎯 Otimização Massiva de Storage (99% de redução!)

#### Adicionado
- **Arquitetura UUID para eventos e payloads**
  - Chave `EVENT#LEITURA#<uuid>` agrupa todos os países de um webhook
  - Chave `WEBHOOK#PAYLOAD#<uuid>` armazena payload original UMA VEZ
  - Chave `ERROR#<uuid>` para rastreamento único de erros
  - Eliminação de 195x duplicação de payload (2.9 GB → 35 MB esperado)

- **Auto-limpeza de dados antigos**
  - Função `deleteOldUserReadings()` com query via GSI UserIndex
  - Delete automático de leituras antigas quando novo webhook chega
  - Mantém apenas a última interação válida por usuário
  - Não precisa mais de limpeza manual

- **Validação completa de países**
  - Teste `iso_validation_test.go` validando 195 países do Maratona.app 2026
  - Relatório detalhado de cobertura por mês
  - 100% de mapeamento (203 variações de nomes)

- **GSI UserIndex para queries eficientes**
  - hashKey: `user` (nome do participante)
  - rangeKey: `PK` (partition key do evento)
  - projection: `all` (todos os atributos)
  - Permite deletar todos os registros de um usuário com 1 query

#### Alterado
- **Threshold de progresso: 0% → 1%**
  - Países com 0% agora aparecem como não explorados (cinza)
  - Apenas países com progresso ≥ 1% são coloridos no mapa
  - Melhora experiência visual (menos poluição no mapa)

- **Schema DynamoDB otimizado**
  - `LeituraItem`: removido campo `Metadata` (eliminada duplicação!)
  - `LeituraItem`: PK agora inclui UUID (`EVENT#LEITURA#<uuid>`)
  - `LeituraItem`: SK agora usa ISO (`COUNTRY#<iso3>`)
  - `WebhookItem`: novo struct para payload único
  - `FalhaItem`: PK agora usa UUID (`ERROR#<uuid>`)

- **Mapeamento de países**
  - Adicionados 4 aliases para variações do Maratona.app:
    - `"Azerbajão"` → AZE (variação de grafia)
    - `"Cabo verde"` → CPV (capitalização diferente)
    - `"Irlanda do norte"` → GBR (capitalização diferente)
    - `"Suiça"` → CHE (sem acento)

#### Técnico
- **packages/functions/types/types.go**
  - Removido `Metadata string` de `LeituraItem`
  - Adicionado `WebhookItem` struct com campos: PK, SK, User, Payload
  - Atualizado `FalhaItem` para usar UUID no PK
  - Comentários detalhados sobre estratégia UUID

- **packages/functions/webhook/main.go**
  - Nova função `saveWebhookPayload()` - salva payload 1x com UUID
  - Nova função `deleteOldUserReadings()` - query GSI + delete batch
  - Handler modificado para gerar UUID uma vez por execução
  - Item creation usa UUID no PK (`EVENT#LEITURA#<uuid>`)
  - Sem marshaling de metadata (economiza CPU + storage)
  - `saveToFalhas()` atualizado para usar UUID

- **packages/functions/stats/main.go**
  - Filtro adicionado: `if progress >= 1` antes de incluir na response
  - Países com 0% não retornam no endpoint (frontend mostra como cinza)

- **packages/functions/mapping/iso.go**
  - 4 novos aliases com comentários indicando origem (Maratona.app)

- **packages/functions/mapping/iso_validation_test.go** (NOVO)
  - Lista completa dos 195 países organizados por mês
  - Test function `TestValidateAllMaratonaCountries()`
  - Test function `TestCountryVariations()` para variações conhecidas
  - Relatório visual com box drawing (╔═╗║╚╝)

- **sst.config.ts**
  - Campo `user: "string"` adicionado aos fields do DynamoDB
  - `globalIndexes.UserIndex` configurado com hashKey=user, rangeKey=PK
  - Comentários atualizados nos fields (PK, SK) com novos padrões UUID

### 📊 Impacto Esperado

**Storage (exemplo com 100 usuários, 3 webhooks cada):**
- **ANTES:** 58.500 items × 50KB = 2.9 GB de payloads duplicados
- **DEPOIS:** 19.500 items países + 300 payloads = ~35 MB
- **ECONOMIA:** 99% (2.9 GB → 35 MB)

**Performance:**
- Queries mais rápidas (items menores: <1KB vs ~50KB)
- Menos writes no DynamoDB (1 payload vs 195 payloads)
- Menos read capacity consumido (items compactos)

**Custo:**
- Storage: ~99% de redução
- Write operations: ~99% de redução (1 webhook write vs 195)
- Read operations: melhor cache efficiency (items menores)

### ⚠️ Breaking Changes

**DynamoDB Schema:**
- Novos webhooks usam padrão `EVENT#LEITURA#<uuid>` no PK
- Webhooks antigos mantêm padrão `EVENT#LEITURA` (queries continuam funcionando)
- Primeiro webhook de cada usuário após deploy fará limpeza de dados antigos

**API Response:**
- `/stats` não retorna mais países com 0% de progresso
- Frontend trata ausência como não explorado (comportamento correto)

### 🧪 Testes
- ✅ Todos os testes Go passando
- ✅ 100% cobertura de países (203/203 variações)
- ✅ Funções UUID validadas
- ✅ GSI será criado automaticamente no deploy

### 📚 Referências
- Plano completo: `~/.claude/plans/modular-tumbling-stallman.md`
- Issue original: País "Emirados Árabes" não mapeado (resolvido em contexto)

## [1.0.1] - 2025-12-23

### Adicionado
- **Sistema de Monitoramento CloudWatch (Produção)**
  - 8 CloudWatch Alarms com alertas automáticos por email
  - 9 Metric Filters para transformar logs em métricas
  - 1 CloudWatch Dashboard (`mundotalendo-prod-dashboard`)
  - SNS Topic para notificações por email (daniel@balieiro.com)
  - Log Groups com retenção de 14 dias para controle de custos

- **Alarmes Configurados**
  - DynamoPutError (threshold: 1) - Perda de dados crítica
  - CountryNotFoundAlarm (threshold: 1) - País não mapeado detectado
  - UnmarshalErrorAlarm (threshold: 1) - Erro de parsing JSON
  - StatsQueryErrorAlarm (threshold: 5) - Falhas no endpoint /stats
  - UsersQueryErrorAlarm (threshold: 5) - Falhas no endpoint /users/locations
  - DynamoReadThrottleAlarm (threshold: 5) - Capacidade insuficiente
  - DynamoWriteThrottleAlarm (threshold: 1) - Capacidade insuficiente
  - AuthFailureAlarm (threshold: 20) - Possível ataque brute force

- **Comandos Makefile para Produção**
  - `make alarms-prod` - Ver status dos alarmes
  - `make metrics-prod` - Ver métricas custom
  - `make logs-all-prod` - Tail de logs em tempo real
  - `make info-prod` - Ver recursos AWS

- **Documentação**
  - `MONITORING.md` - Guia completo de monitoramento com queries CloudWatch Logs Insights

### Alterado
- Monitoramento configurado **apenas para PRODUÇÃO** (stage: prod)
- DEV mantém apenas Log Groups para debug (sem alarmes/alertas)
- `sst.config.ts` com condição `isProduction` para recursos de monitoramento

### Técnico
- Thresholds sensíveis para detecção imediata de problemas (1 erro = alerta)
- Namespace de métricas: `MundoTaLendo`
- Retenção de logs: 14 dias (redução de custos)
- Região: us-east-2 (Ohio)

## [1.0.0] - 2025-12-21

### Adicionado
- **Sistema completo de descoberta cultural colaborativa**
  - Dashboard em tempo real com mapa interativo
  - Marcadores GPS de usuários no mapa
  - Backend Go (5 Lambdas ARM64/Graviton)
  - Frontend Next.js 16 com MapLibre GL JS

- **Features Principais**
  - User markers com avatares circulares
  - Posicionamento inteligente para múltiplos usuários
  - Tooltip mostrando usuário e livro sendo lido
  - Proxy de imagens para resolver CORS
  - Feature flag `NEXT_PUBLIC_SHOW_USER_MARKERS`

- **Infraestrutura**
  - DynamoDB Single Table com backups PITR
  - API Gateway V2 com domínios custom
  - Lambda concurrency limits
  - CORS configurado
  - Polling otimizado (60s stats + users)

- **Endpoints API**
  - `POST /webhook` - Recebe eventos do Maratona.app
  - `GET /stats` - Retorna países sendo explorados
  - `GET /users/locations` - Retorna localizações de usuários
  - `POST /test/seed` - Popular dados de teste
  - `POST /clear` - Limpar dados (com proteção)

- **Qualidade**
  - 26 testes Go passando
  - Error Boundary no frontend
  - Retry logic automático
  - Input validation robusto
  - Security headers configurados

### Deploy
- **DEV:** https://dev.mundotalendo.com.br
- **DEV API:** https://api.dev.mundotalendo.com.br
- **PROD:** https://mundotalendo.com.br
- **PROD API:** https://api.mundotalendo.com.br

---

## Notas de Versão

### v1.0.2
Otimização massiva de storage com arquitetura UUID (99% de redução: 2.9 GB → 35 MB). Auto-limpeza de dados antigos (mantém apenas última interação por usuário). Validação 100% dos países do Maratona.app 2026 (203 variações). Threshold de progresso ajustado para ≥1% (melhora experiência visual no mapa). GSI UserIndex para queries eficientes.

### v1.0.1
Sistema de monitoramento proativo em produção com alertas automáticos por email. Detecção imediata de erros críticos (países não mapeados, parsing JSON, falhas DynamoDB). Log Groups com retenção controlada para redução de custos.

### v1.0.0
Lançamento inicial do sistema em produção. Dashboard funcional com mapa interativo, marcadores de usuários, e infraestrutura serverless otimizada.
