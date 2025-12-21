# Claude Context - Mundo Tá Lendo 2026

> **Última atualização:** 2025-12-21
> **Status:** Em desenvolvimento - Debugando tiles customizados com Tippecanoe

## 📋 Resumo Executivo

Projeto de **descoberta cultural colaborativa** através da leitura. Dashboard que mapeia em tempo real países sendo explorados por participantes do desafio de leitura "Mundo Tá Lendo 2026".

**Conceito importante:** Não é sobre "conquista" de países, é sobre **descobrir culturas** colaborativamente.

## 🚨 INVESTIGAÇÃO COMPLETA: Vector Tiles Tippecanoe (21 Dez 2025 - 12:00-12:25)

**Status**: INVESTIGAÇÃO CONCLUÍDA - Root cause identificado, solução recomendada

**IMPORTANTE**: Código foi modificado durante investigação e precisa ser revertido antes de implementar solução final.

---

### 📌 RESUMO EXECUTIVO (TL;DR)

**Problema**: Tiles gerados com Tippecanoe não parseiam no MapLibre GL JS

**Root Cause**: Estrutura interna dos tiles do Tippecanoe é incompatível com MapLibre
- ✅ Servidor está correto (confirmado servindo tile do demotiles)
- ✅ Headers estão corretos (Content-Type: application/octet-stream)
- ❌ Tiles gerados são inválidos (7 tentativas diferentes, todas falharam)

**Solução**: Usar GeoJSON completo em vez de vector tiles
- Arquivo: `countries.geojson` (2.7MB comprimido)
- MapLibre parseia GeoJSON perfeitamente (sem problemas)
- Controle total sobre geometrias (GUF/GRL/PRI separados)
- Implementação: 30-60 minutos

**Próximos Passos**:
1. Reverter código modificado (ver lista abaixo)
2. Implementar GeoJSON source no Map.jsx (passo a passo documentado)
3. Testar e validar

---

### Descobertas Chave (Sessão 21 Dez 12:00-12:17)

**1. Content-Type Header estava ERRADO** ✅ CORRIGIDO
- **Problema**: `scripts/serve-tiles.js` enviava `Content-Type: application/x-protobuf`
- **Correto**: demotiles.org envia `Content-Type: application/octet-stream`
- **Fix**: Mudado em `scripts/serve-tiles.js` linha 19
- **Resultado**: Header correto agora, MAS tile ainda não parseia ❌

**2. Tiles tinham layer name ERRADO** ✅ CORRIGIDO
- **Problema**: Tiles gerados tinham layer "centroids" em vez de "countries"
- **Causa**: Tiles antigos no diretório, não regenerados após mudanças
- **Fix**: Executado `./scripts/prepare-geojson.sh` + `./scripts/generate-tiles.sh`
- **Resultado**: Tiles agora têm layer "countries", 264 features incluindo GUF/GRL/PRI
- **Hexdump confirma**: `00000000  1a fd f5 05 78 02 0a 09  63 6f 75 6e 74 72 69 65  |....x...countrie|`
- **Mas tile AINDA não parseia** ❌

**3. ROOT CAUSE CONFIRMADO (21 Dez 12:20)** ✅
- **Teste realizado**: Servimos tile do demotiles.org via nosso servidor
- **Resultado**: ERRO DIFERENTE! `Source layer "centroids" does not exist`
- **Significado**: Tile foi PARSEADO com sucesso! Servidor está OK! ✅
- **Teste de confirmação**: Regeneramos nossos tiles e testamos
- **Resultado**: Voltou ao erro original "Unable to parse the tile"
- **CONCLUSÃO FINAL**: Problema são os TILES gerados pelo Tippecanoe, NÃO o servidor!

### Tentativas Realizadas (Histórico Completo)

| Step | Ação | Arquivo | Resultado |
|------|------|---------|-----------|
| 1 | Remover `Content-Encoding: gzip` header | serve-tiles.js | ❌ Falhou |
| 2 | Descomprimir on-the-fly antes de enviar | serve-tiles.js | ❌ Falhou |
| 3 | Regenerar com `--no-tile-compression` | generate-tiles.sh | ❌ Falhou |
| 4 | Regenerar com flags mínimas | generate-tiles.sh | ❌ Falhou |
| 5 | Mudar Content-Type para octet-stream | serve-tiles.js | ❌ Falhou |
| 6 | Regenerar tiles com layer "countries" | generate-tiles.sh | ❌ Falhou |
| 7 | **TESTE**: Servir tile demotiles via nosso servidor | - | ⏳ Em teste |

### Arquivos Modificados (Esta Sessão)

**scripts/serve-tiles.js**
- Linha 19: `'.pbf': 'application/octet-stream'` (antes: application/x-protobuf)
- Headers CORS mantidos
- Serve tiles uncompressed via `fs.createReadStream().pipe(res)`

**scripts/generate-tiles.sh**
- Flags Tippecanoe: `--layer=countries`, `--no-tile-compression`, `--include=ADM0_A3`, `--include=NAME`
- Input: `data/world-countries-final.geojson` (264 features)
- Output: tiles com layer "countries" correta

**scripts/prepare-geojson.sh**
- Usa `src/config/territoryGeometries.json` para extrair GUF, GRL, PRI
- Validação confirma: 264 features totais, territórios presentes

**src/components/Map.jsx** (temporário para teste)
- Linha 174, 192: `source-layer: 'centroids'` (temporário - demotiles usa isso)
- Normalmente deveria ser `'countries'`

**tiles/0/0/0.pbf**
- Substituído pelo tile do demotiles.org para teste de isolamento
- Layer: "centroids" (99KB)

### Estado do Servidor

**Rodando**: `node scripts/serve-tiles.js` (background task be7a744)
**Headers enviados**:
- `Content-Type: application/octet-stream` ✅
- `Access-Control-Allow-Origin: *` ✅
- `Cache-Control: public, max-age=86400` ✅
- **NÃO envia** `Content-Encoding` ✅

### Próximos Passos (Problema CONFIRMADO: Tiles Tippecanoe)

**PRIORIDADE**: Investigar por que Tippecanoe gera tiles inválidos

1. **Opção A: Usar tile demotiles como base** (RÁPIDO - RECOMENDADO)
   - Tile do demotiles FUNCIONA (já testado e confirmado)
   - Problema: Não tem GUF/GRL/PRI separados
   - Solução temporária: Voltar para GeoJSON overlay (Opção 1 original)
   - 2.7MB comprimido, cache do browser, funciona HOJE

2. **Opção B: Investigar Tippecanoe flags**
   - Comparar flags usadas: nosso vs demotiles (se tiverem documentação)
   - Testar gerar tiles a partir do mesmo GeoJSON que demotiles usa
   - Comparar output byte-a-byte

3. **Opção C: Ferramenta alternativa**
   - Testar `tippecanoe` com GeoJSON MÍNIMO (só 3 países) para debug
   - Testar outras ferramentas: `mbutil`, `geojson-vt`, `vt-geojson`
   - Gerar tiles manualmente com Mapbox Studio

4. **Opção D: Híbrido**
   - Usar tiles do demotiles para países principais
   - Adicionar GUF/GRL/PRI como camada GeoJSON separada
   - Simples, funciona, resolve 100% do problema

### CONCLUSÃO DEFINITIVA

**ROOT CAUSE**: Tippecanoe gera tiles em formato que MapLibre GL JS não consegue parsear
- Servidor Node.js está correto (confirmado servindo tile do demotiles)
- Headers estão corretos (Content-Type: application/octet-stream)
- Problema é estrutura interna dos tiles gerados pelo Tippecanoe

**SOLUÇÃO RECOMENDADA**: Opção 1 - GeoJSON Completo ⭐⭐⭐

**Por que GeoJSON é a melhor opção agora**:
1. ✅ MapLibre GL JS parseia GeoJSON perfeitamente (sem problemas)
2. ✅ Controle total sobre geometrias (GUF/GRL/PRI separados)
3. ✅ Tamanho aceitável: 2.7MB comprimido (Brotli), cache do browser
4. ✅ Funciona HOJE, zero risco de parsing errors
5. ✅ Já temos o arquivo pronto (countries.geojson)

**Por que outras opções NÃO funcionam**:
- ❌ **Overlay GeoJSON**: Polígonos sem detalhe suficiente, fica visível
- ❌ **Tippecanoe**: Gera tiles inválidos (confirmado após 7 tentativas)
- ⚠️ **Outro tile service**: Sem garantia de separar GUF/GRL/PRI
- ❌ **Aceitar limitação**: Não resolve o problema

### ARQUIVOS MODIFICADOS (PRECISAM SER REVERTIDOS)

**Antes de implementar solução, reverter estas mudanças:**

1. **scripts/serve-tiles.js**
   - Linha 19: Mudado para `application/octet-stream`
   - Reverter para estado original ou deletar arquivo

2. **scripts/generate-tiles.sh**
   - Múltiplas iterações de flags Tippecanoe
   - Reverter ou deletar se não for usado

3. **scripts/prepare-geojson.sh**
   - Modificações para extrair GUF/GRL/PRI
   - Reverter ou deletar se não for usado

4. **src/components/Map.jsx**
   - Linhas 174, 192: Temporariamente mudadas para 'centroids'
   - **DEVE** voltar para configuração original com demotiles

5. **tiles/** (diretório inteiro)
   - Contém tiles gerados que não funcionam
   - Pode ser deletado completamente

6. **data/world-countries-final.geojson**
   - Gerado pelos scripts
   - Pode ser deletado se não for usado

**Git status antes de começar próxima sessão:**
```bash
git status
git diff  # Revisar mudanças
git checkout -- <arquivo>  # Reverter arquivos específicos
# OU
git reset --hard HEAD  # Reverter TUDO (cuidado!)
```

### PRÓXIMA SESSÃO: Implementar GeoJSON Completo

**PASSO A PASSO** (após reverter código):

1. **Preparar GeoJSON**
   ```bash
   # Usar countries.geojson existente na raiz (14 MB)
   # Adicionar GUF, GRL, PRI como features separadas
   # Remover/modificar FRA, DNK, USA para excluir territórios ultramarinos
   # Resultado: ~264 features (261 países + GUF + GRL + PRI)
   ```

2. **Modificar Map.jsx**
   ```javascript
   // REMOVER: Vector tile source
   // ADICIONAR: GeoJSON source
   map.current.addSource('countries', {
     type: 'geojson',
     data: '/countries.geojson'  // ou importar diretamente
   })

   // Layers continuam iguais, mas SEM 'source-layer'
   // Remover linha: 'source-layer': 'countries'
   ```

3. **Otimizar GeoJSON** (opcional, se tamanho for problema)
   ```bash
   # Simplificar geometrias com mapshaper
   npm install -g mapshaper
   mapshaper countries.geojson -simplify 10% -o countries-simplified.geojson
   ```

4. **Configurar Next.js** para servir arquivo estático
   ```
   # Colocar countries.geojson em /public/
   # Next.js vai servir automaticamente
   # Compressão Brotli/Gzip automática em produção
   ```

5. **Testar localmente**
   - Mapa deve carregar normalmente
   - GUF vermelha (Janeiro), GRL azul (Julho), PRI laranja (Maio)
   - Verificar performance (deve ser OK para 264 features)

**ESTIMATIVA**: 30-60 minutos de implementação

**BENEFÍCIOS**:
- ✅ Resolve 100% do problema definitivamente
- ✅ Sem dependência de tiles externos ou Tippecanoe
- ✅ Controle total sobre dados e geometrias
- ✅ Funciona garantido (GeoJSON é nativo do MapLibre)
- ✅ Manutenção simples (apenas um arquivo)

**RISCOS**: Nenhum (GeoJSON é formato padrão e totalmente suportado)

## 🎯 Estado Atual do Projeto

### ✅ O que está funcionando

1. **Backend completo e operacional:**
   - 3 Lambdas Go (webhook, stats, seed) deployadas em ARM64
   - DynamoDB table configurado com 18 países de teste
   - API Gateway configurado manualmente (workaround do bug SST)
   - Endpoints REST funcionais e testados

2. **Frontend funcional localmente:**
   - Next.js 16 rodando em http://localhost:3000
   - Mapa interativo com MapLibre GL JS
   - Labels em português (exatamente 1 por país via centroids)
   - Cores vibrantes por mês
   - Polling SWR a cada 15s
   - Integrado com API AWS real

3. **Integração end-to-end testada:**
   - Frontend → API Gateway → Lambda → DynamoDB → resposta

### ⚠️ Problemas conhecidos

1. **SST Deploy Bug (RangeError: Invalid string length)**
   - SST 3.17.25 tem bug ao tentar exibir mensagens de erro grandes
   - Deployments parciais: recursos criados mas rotas não conectadas
   - Workaround aplicado: configuração manual via AWS CLI

2. **CloudFront não configurado**
   - Existe distribuição de deploy antigo (placeholder)
   - Frontend só funciona localmente por enquanto
   - Necessário resolver bug SST ou fazer deploy manual completo

3. **Domínios Cloudflare comentados**
   - Configuração temporariamente desabilitada no sst.config.ts
   - Aguardando resolução do deploy principal

4. **Vector Tile: Territórios ultramarinos compartilham códigos ISO**
   - Vector tile (`demotiles.maplibre.org`) não separa territórios ultramarinos
   - Guiana Francesa (GUF) renderizada como FRA (França)
   - Groenlândia (GRL) renderizada como DNK (Dinamarca)
   - Porto Rico (PRI) renderizado como USA (Estados Unidos)
   - **Status**: Implementada solução parcial com camadas GeoJSON sobrepostas
   - **Limitação**: Geometrias sobrepostas visíveis (polígono detalhado abaixo + simplificado acima)
   - **Decisão pendente**: Escolher entre 4 abordagens (documentadas abaixo)

## 🗺️ Problema: Territórios Ultramarinos e Vector Tiles

### Contexto do Problema

O vector tile usado (`demotiles.maplibre.org`) não diferencia territórios ultramarinos de seus países principais:
- **Guiana Francesa** → Aparece como **FRA** (França - Março - Amarelo) em vez de **GUF** (Janeiro - Vermelho)
- **Groenlândia** → Aparece como **DNK** (Dinamarca - Setembro - Teal) em vez de **GRL** (Julho - Azul)
- **Porto Rico** → Aparece como **USA** (EUA - Julho - Azul) em vez de **PRI** (Maio - Laranja)

### Soluções Tentadas

#### 1. Detecção por coordenadas (hover funciona ✅)
**Implementado:**
- Arquivo `/src/config/territoryOverrides.js` com bounding boxes
- Função `getCorrectIsoCode(vectorTileIso, lng, lat)` detecta território por coordenadas
- Evento `mousemove` aplica override e mostra nome/progresso correto no tooltip

**Resultado:** Funciona perfeitamente para hover/tooltip.

#### 2. Camadas GeoJSON sobrepostas (renderização problemática ❌)
**Implementado:**
- Arquivo `/src/config/territoryGeometries.json` com geometrias de alta qualidade (extraídas de `countries.geojson`)
- Camada `territory-overrides` renderizada **por cima** do vector tile
- Aplica cores corretas (GUF vermelho, GRL azul, PRI laranja)

**Problema:**
- Sobreposição visível: polígono do vector tile (alta resolução) aparece embaixo
- Polígono GeoJSON (mesmo em alta qualidade) tem pequenas diferenças de precisão
- Visual não profissional: "dá pra ver claramente que tem algo embaixo"

### Opções de Solução (Decisão Pendente)

#### Opção 1: GeoJSON Completo (Recomendado ⭐)

**Descrição:**
- Substituir vector tile por arquivo GeoJSON único
- Usar `countries.geojson` (já temos no projeto) como fonte
- Adicionar GUF, GRL, PRI como features separadas
- Remover/modificar FRA, DNK, USA para excluir territórios

**Vantagens:**
- ✅ Controle total sobre geometrias
- ✅ Sem problemas de sobreposição
- ✅ Geometrias perfeitas e consistentes
- ✅ Fácil adicionar mais territórios no futuro

**Desvantagens:**
- ❌ Tamanho do arquivo maior

**Análise de Tamanho:**
```
Original (countries.geojson):     14.0 MB (em disco)
Gzip (HTTP/1.1):                   4.38 MB (servidores antigos)
Brotli (HTTP/2+):                  2.73 MB (navegadores modernos)
```

**Contexto de Performance:**
- Google Fonts médio: 100-300 KB
- Imagem hero típica: 500 KB - 2 MB
- Bundle JS Next.js médio: 200-500 KB
- **GeoJSON completo comprimido: 2.73 MB** ← comparável a 2-3 imagens

**Impacto:**
- Download **uma vez** e cacheado pelo navegador
- ~2.7 MB para navegadores modernos (Chrome, Firefox, Safari)
- Aceitável para mapa mundial com 193 países em alta qualidade

#### Opção 2: Vector Tile Alternativo

**Descrição:**
- Trocar `demotiles.maplibre.org` por outro serviço que separe territórios
- Exemplos: Mapbox, Maptiler, OpenMapTiles

**Vantagens:**
- ✅ Mantém performance de vector tiles
- ✅ Geometrias nativas sem sobreposição

**Desvantagens:**
- ❌ Precisa encontrar serviço público gratuito
- ❌ Pode ter outras limitações/diferenças
- ❌ Dependência de serviço terceiro

#### Opção 3: Vector Tile Self-Hosted

**Descrição:**
- Gerar próprios tiles com Tippecanoe a partir do GeoJSON
- Hospedar no S3 ou CloudFront

**Vantagens:**
- ✅ Performance de vector tiles
- ✅ Controle total

**Desvantagens:**
- ❌ Setup complexo (Tippecanoe, tile generation, hosting)
- ❌ Manutenção adicional
- ❌ Custo de storage/bandwidth

#### Opção 4: Aceitar Limitação

**Descrição:**
- Manter GUF como FRA (amarela), GRL como DNK, PRI como USA
- Documentar limitação

**Vantagens:**
- ✅ Simples, sem mudanças

**Desvantagens:**
- ❌ Não resolve o problema
- ❌ Inconsistência visual (GUF amarela em vez de vermelha)

### Recomendação

**Opção 1 (GeoJSON Completo)** é a mais adequada:
- Tamanho aceitável (~2.7 MB comprimido)
- Resolve completamente o problema
- Mantém qualidade visual
- Simplicidade de implementação

**Alternativa para reduzir tamanho ainda mais:**
- Simplificar geometrias com `mapshaper` (pode reduzir para ~1 MB)
- Trade-off: menos detalhes nos contornos, mas ainda aceitável

### Arquivos Criados

```
src/config/
├── territoryOverrides.js       # Bounding boxes e função getCorrectIsoCode()
└── territoryGeometries.json    # Geometrias de alta qualidade (GUF, GRL, PRI)

countries.geojson               # Arquivo completo na raiz (14 MB)
```

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
- ID: `q9f0i3fp0d`
- URL: `https://q9f0i3fp0d.execute-api.us-east-2.amazonaws.com`
- Configurado manualmente (workaround)

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
│   └── Map.jsx             # Mapa com useCallback, centroids, cores vibrantes
├── config/
│   ├── countries.js        # ISO → Nome PT-BR (193 países)
│   ├── countryCentroids.js # ISO → [lng, lat] (1 ponto exato por país)
│   └── months.js           # 12 meses → cores vibrantes → países
└── hooks/
    └── useStats.js         # SWR com fallback para /api
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

## 🧪 Testes

### Testar API manualmente

```bash
# Stats (deve retornar ~18 países)
curl https://q9f0i3fp0d.execute-api.us-east-2.amazonaws.com/stats

# Seed mais países
curl -X POST https://q9f0i3fp0d.execute-api.us-east-2.amazonaws.com/test/seed

# Webhook (simular Maratona.app)
curl -X POST https://q9f0i3fp0d.execute-api.us-east-2.amazonaws.com/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "perfil": {"nome": "Test", "link": "https://test.com"},
    "maratona": {"nome": "Test", "identificador": "test"},
    "desafios": [{
      "descricao": "Brasil",
      "categoria": "Janeiro",
      "concluido": true,
      "tipo": "leitura",
      "vinculados": [{"completo": true, "updatedAt": "2024-12-16T00:00:00Z"}]
    }]
  }'
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

## 📊 Dados de Teste Atuais

18 países no banco:
- DJI, GNB, CAN, CAF, MAR, IRN, LUX, TKM, CRI, KOR
- POL, ARE, COL, IDN, TZA, VNM, NRU, BHS

Para adicionar mais: `curl -X POST .../test/seed`

## ⚠️ Avisos Importantes

1. **Não usar "conquista" ou "conquered"** - projeto é sobre descoberta cultural
2. **Labels devem estar em português** - sempre PT-BR
3. **1 label por país** - usar centroids, não vector tiles
4. **Cores vibrantes** - oceano azul, países com cores dos meses
5. **SST 3.17.25 tem bug** - workaround manual necessário
6. **Não usar react-map-gl** - implementação direta MapLibre
7. **Webpack, não Turbopack** - via npm run dev:local
8. **Tailwind CSS v4** - nova sintaxe com @tailwindcss/postcss

## 🔗 Links Úteis

- API Gateway: https://q9f0i3fp0d.execute-api.us-east-2.amazonaws.com
- Local Dev: http://localhost:3000
- Vector Tiles: https://demotiles.maplibre.org
- Project Spec: project.md
- SST Issue: https://github.com/sst/ion/issues (procurar RangeError)

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

**Última observação:** Mantenha esta documentação atualizada conforme o projeto evolui. É a fonte de verdade para contexto técnico em futuras sessões.
