# Monitoramento e Alertas - Mundo Tá Lendo 2026 (PRODUÇÃO)

Este documento contém todas as queries CloudWatch Logs Insights e instruções de uso do sistema de monitoramento.

**⚠️ IMPORTANTE:** O sistema de monitoramento com alertas está configurado **APENAS para PRODUÇÃO**. DEV possui apenas Log Groups para debug.

## 📊 Recursos Implementados

### ✅ Log Groups com Retenção
- **Retenção:** 14 dias (configurável em `sst.config.ts`)
- **Redução de custos:** Logs antigos são excluídos automaticamente
- **5 Log Groups:** webhook, stats, users, seed, clear

### ✅ Metric Filters (9 métricas)
Transformam logs plain text em métricas quantificáveis:
- `CountryNotFoundErrorCount` - Países não mapeados
- `UnmarshalErrorCount` - Erros de parsing JSON
- `DynamoPutErrorCount` - Falhas de escrita no DynamoDB (CRÍTICO)
- `PayloadTooLargeCount` - Payloads >1MB
- `StatsQueryErrorCount` - Erros no endpoint /stats
- `UsersQueryErrorCount` - Erros no endpoint /users/locations
- `WebhookAuthFailureCount` - Falhas de autenticação webhook
- `StatsAuthFailureCount` - Falhas de autenticação stats
- `UsersAuthFailureCount` - Falhas de autenticação users

**Namespace:** `MundoTaLendo`

### ✅ CloudWatch Alarms (8 alarmes)

| # | Alarme | Threshold | Descrição |
|---|--------|-----------|-----------|
| 1 | DynamoPutError | 1 erro | 🔴 CRÍTICO: Perda de dados |
| 2 | CountryNotFound | **1 erro/5min** | ⚠️ País não mapeado detectado |
| 3 | UnmarshalError | **1 erro/5min** | ⚠️ Erro de parsing JSON |
| 4 | StatsQueryError | 5 erros/5min | 🔴 CRÍTICO: Stats endpoint com falhas |
| 5 | UsersQueryError | 5 erros/5min | 🔴 CRÍTICO: Users endpoint com falhas |
| 6 | DynamoReadThrottle | 5 throttles | 🔴 Capacidade insuficiente |
| 7 | DynamoWriteThrottle | 1 throttle | 🔴 Capacidade insuficiente |
| 8 | AuthFailure | 20 falhas/5min | 🔒 Possível ataque |

**Email de notificação:** daniel@balieiro.com

### ✅ CloudWatch Dashboard
**Nome:** `mundotalendo-prod-dashboard`

**Widgets:**
1. API Error Rates (5 min) - Taxa de erros total
2. API Latency (Average) - Latência média das Lambdas
3. Webhook Error Breakdown - Breakdown por tipo de erro
4. DynamoDB Health - Throttling do DynamoDB

**Acesso:** AWS Console → CloudWatch → Dashboards

---

## 🔧 Comandos Makefile

### Ver status dos alarmes (PROD)
```bash
make alarms-prod
```

**Output esperado:**
```
CloudWatch Alarms Status - PROD:
-----------------------------------------------------------------------
| Name                                          | State | Reason     |
-----------------------------------------------------------------------
| mundotalendo-prod-webhook-dynamo-put-error    | OK    | ...        |
| mundotalendo-prod-stats-query-error           | OK    | ...        |
...
```

### Ver métricas custom (PROD)
```bash
make metrics-prod
```

**Output esperado:**
```
Custom Metrics:
------------------------------------------------------------
| Namespace      | MetricName                      | ...  |
------------------------------------------------------------
| MundoTaLendo   | CountryNotFoundErrorCount       | ...  |
| MundoTaLendo   | UnmarshalErrorCount             | ...  |
...
```

### Tail de todos os logs em tempo real (PROD)
```bash
make logs-all-prod
```

**Output esperado:**
```
Tailing PROD Lambda logs...
2025-12-23T10:00:00 /aws/lambda/mundotalendo-prod-webhook Received webhook request: {...}
2025-12-23T10:00:01 /aws/lambda/mundotalendo-prod-stats Fetching stats from DynamoDB
...
```

---

## 📝 CloudWatch Logs Insights Queries

Acesse: **AWS Console → CloudWatch → Logs Insights**

### Query 1: Todos erros webhook (última hora)
**Quando usar:** Debugar problemas no webhook em produção

```sql
fields @timestamp, @message
| filter @message like /ERROR|Error|error/
| filter @logStream like /webhook/
| sort @timestamp desc
| limit 100
```

### Query 2: Países não mapeados (precisa adicionar)
**Quando usar:** Descobrir quais países estão faltando no `mapping/countries.go`

```sql
fields @timestamp, @message
| filter @message like "Country not found:"
| parse @message "Country not found: * (original: *)" as cleaned, original
| stats count() by cleaned, original
| sort count desc
```

**Output esperado:**
```
| count | cleaned    | original   |
|-------|------------|------------|
| 5     | zzzinvalid | ZZZ-Invalid|
| 2     | test       | Test       |
```

**Ação:** Adicionar países faltantes em `packages/functions/mapping/countries.go`

### Query 3: DynamoDB throttling
**Quando usar:** Sistema está lento ou com erros de throttling

```sql
fields @timestamp, @message
| filter @message like /throttl/i
| sort @timestamp desc
| limit 50
```

**Ação:** Aumentar capacidade do DynamoDB ou habilitar auto-scaling

### Query 4: Lambdas lentas (>1s)
**Quando usar:** Frontend está lento, investigar performance

```sql
fields @timestamp, @duration, @message
| filter @duration > 1000
| sort @duration desc
| limit 20
```

**Ação:** Otimizar queries DynamoDB ou adicionar paginação

### Query 5: Erros agregados por hora
**Quando usar:** Analisar padrões de erro ao longo do tempo

```sql
fields @timestamp, @message
| filter @message like /ERROR|Error|error/
| stats count() by bin(1h) as hour
| sort hour desc
```

**Output esperado:**
```
| count | hour              |
|-------|-------------------|
| 23    | 2025-12-23 10:00  |
| 45    | 2025-12-23 09:00  |
```

### Query 6: Requests por Lambda (últimas 24h)
**Quando usar:** Entender distribuição de carga entre endpoints

```sql
fields @timestamp
| stats count() by bin(1h) as hour, @logStream
| sort hour desc
```

### Query 7: Auth failures por endpoint
**Quando usar:** Investigar possível ataque brute force

```sql
fields @timestamp, @message
| filter @message like "Unauthorized: invalid API key"
| stats count() by bin(5m) as time_window
| sort time_window desc
| limit 100
```

---

## 🚨 Como Salvar Queries

1. Acesse **AWS Console → CloudWatch → Logs Insights**
2. Cole a query no editor
3. Selecione os log groups relevantes:
   - `/aws/lambda/mundotalendo-prod-webhook`
   - `/aws/lambda/mundotalendo-prod-stats`
   - `/aws/lambda/mundotalendo-prod-users`
4. Clique em **Run query**
5. Clique em **Save** → **Save as query**
6. Crie uma pasta: **"Mundo Tá Lendo - Debug Queries"**
7. Nomeie a query (ex: "Webhook Errors - Last Hour")

**Resultado:** Queries ficam salvas e reutilizáveis!

---

## 📧 Configurando Alertas por Email

### Primeira vez após deploy:

1. **Cheque seu email:** daniel@balieiro.com
2. **Procure por:** "AWS Notification - Subscription Confirmation"
3. **Clique no link** de confirmação
4. **Status muda para:** "Confirmed"

**Após confirmar:**
- Você receberá emails automáticos quando alarmes dispararem
- Formato do email:
  ```
  Subject: ALARM: mundotalendo-dev-webhook-dynamo-put-error
  Body: Alarm Description: CRITICAL: DynamoDB writes failing - DATA LOSS RISK
        Threshold: 1.0
        Current Value: 3.0
  ```

### Re-enviar email de confirmação:

```bash
aws sns list-subscriptions-by-topic \
  --topic-arn <ARN_DO_TOPIC> \
  --region us-east-2
```

---

## 🔍 Troubleshooting

### Problema: Alarmes não disparam

**Verificar:**
```bash
make alarms
```

**Status esperado:** OK ou INSUFFICIENT_DATA

**Se status = ALARM mas sem email:**
1. Verificar subscription confirmada (AWS Console → SNS)
2. Verificar spam na caixa de email
3. Verificar ARN do topic no alarme

### Problema: Métricas não incrementam

**Verificar logs:**
```bash
make logs-all
```

**Procurar por:** Pattern do metric filter (ex: "Country not found:")

**Se pattern não encontrado:**
1. Ajustar pattern no `sst.config.ts`
2. Redeploy: `npx sst deploy --stage dev`

### Problema: Dashboard vazio

**Causa:** Métricas ainda não têm dados (sistema sem erros)

**Teste forçar erro:**
```bash
# Enviar país inválido para webhook
curl -X POST https://api.dev.mundotalendo.com.br/webhook \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $(make -s get-api-key)" \
  -d '{
    "perfil": {"nome": "Test User"},
    "maratona": {"identificador": "mundotalendo-2026"},
    "desafios": [{
      "descricao": "ZZZ_Invalid_Country",
      "categoria": "Janeiro",
      "concluido": true,
      "tipo": "leitura",
      "vinculados": [{"progresso": 100, "updatedAt": "2024-12-23T10:00:00Z"}]
    }]
  }'
```

**Aguardar:** ~5 minutos para métrica aparecer no dashboard

---

## 💰 Custos

### Ano 1 (Free Tier): $0/mês ✅

| Recurso | Uso | Free Tier | Custo |
|---------|-----|-----------|-------|
| CloudWatch Logs | 300 MB/mês | 5 GB | $0 |
| Custom Metrics | 9 métricas | 10 | $0 |
| Alarms | 10 alarmes | 10 | $0 |
| Dashboard | 1 dashboard | 3 | $0 |
| SNS Emails | ~100/mês | 1.000 | $0 |

### Após Free Tier (12 meses): ~$6.85/mês

- Logs: ~$0.15/mês
- Metrics: ~$2.70/mês
- Alarms: ~$1.00/mês
- Dashboard: ~$3.00/mês

---

## 🎯 Próximos Passos

### Opcional (futuro):
1. **Structured Logging** - Migrar de plain text para JSON
2. **X-Ray Tracing** - Rastreamento end-to-end de requests
3. **Lambda Insights** - Métricas detalhadas de memória/CPU
4. **Alarmes adicionais** - Se ficar abaixo de 10 alarmes
5. **Dashboards** - Criar dashboard específico por endpoint

### Quando implementar:
- Structured logging: Quando volume de logs > 1 GB/mês
- X-Ray: Quando precisar debugar latência entre serviços
- Lambda Insights: Quando suspeitar de problemas de memória
- Alarmes extras: Quando identificar novos padrões de erro

---

**Documentação gerada em:** 2025-12-23
**Versão:** 1.0.0
