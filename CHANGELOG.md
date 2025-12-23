# Changelog

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/),
e este projeto adere ao [Semantic Versioning](https://semver.org/lang/pt-BR/).

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

### v1.0.1
Sistema de monitoramento proativo em produção com alertas automáticos por email. Detecção imediata de erros críticos (países não mapeados, parsing JSON, falhas DynamoDB). Log Groups com retenção controlada para redução de custos.

### v1.0.0
Lançamento inicial do sistema em produção. Dashboard funcional com mapa interativo, marcadores de usuários, e infraestrutura serverless otimizada.
