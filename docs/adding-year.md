# Task: Implementação do Filtro Global de Ano no Painel Lateral (Sidebar) — Mundo Tá Lendo

## 🎯 Objetivo
Implementar um filtro por ano localizado no Painel Lateral (Sidebar) do `mundotalendo.com.br`. O filtro deve selecionar o ano atual por padrão e permitir a alternância entre anos anteriores que possuam dados de leitura registrados, atualizando o mapa e as estatísticas de forma reativa.

---

## 🛠️ Requisitos Funcionais

### 1. Seletor de Ano no Painel Lateral (UI)
- **Localização:** No topo da Sidebar / Painel de Estatísticas, logo abaixo do cabeçalho do painel ou perfil do usuário, em posição de destaque como "Filtro Global".
- **Componente:** Um `<Select />` ou Dropdown estilizado e acessível (ex: Shadcn/UI, Radix ou Tailwind CSS).
- **Opção Padrão:** O ano atual (ex: `2026`), pré-selecionado no primeiro carregamento.
- **Lista de Anos:** Populada estritamente pelos anos em que o usuário possui registros de leitura.

### 2. Integração de Dados & API
- **Endpoint de Anos com Dados (`GET /api/v1/readings/years`):**
  - Buscar a lista de anos que possuem pelo menos um livro/registro de leitura cadastrado.
  - Anos sem dados não devem aparecer nas opções do dropdown para evitar telas vazias desnecessárias.
- **Endpoint de Leituras/Mapa (`GET /api/v1/readings?year=YYYY`):**
  - Atualizar a busca de marcadores do mapa e métricas do painel enviando o parâmetro do ano selecionado.

### 3. Sincronização de Estado via URL (URL Search Params)
- O ano selecionado deve ser refletido na URL como Query Parameter:
  - Exemplo para 2025: `https://mundotalendo.com.br/?year=2025`
- **Navegação & Compartilhamento:**
  - O estado inicial da página deve ler o parâmetro `year` da URL. Se estiver ausente ou for inválido, assume o ano atual.
  - Deve ser possível compartilhar a URL direta de um ano específico com outra pessoa.

### 4. Estados de Loading, Erro e Feedback Visual
- **Transição de Carregamento:** Ao alterar o ano no dropdown, o mapa e o painel não devem piscar em branco (*flash of unstyled/empty content*).
  - Usar um indicador discreto de carregamento (Skeleton no painel e overlay/fade suave nos pinos do mapa).
- **Empty State:** Caso a rota seja acessada diretamente via URL com um ano sem dados cadastrados, exibir uma mensagem amigável no painel: *"Nenhuma leitura registrada para este ano."*

---

## 📐 Requisitos Técnicos & Arquitetura (Next.js / Frontend)

1. **Gerenciamento de Estado:**
   - Sincronizar o estado do filtro com o Next.js `useSearchParams` / `useRouter`.
   - Garantir que a troca de ano acione um *fetch* suave (ex: usando `SWR`, `React Query` ou `useTransition` para evitar re-renderizações pesadas do componente de mapa).

2. **Responsividade (Mobile & Desktop):**
   - **Desktop:** O filtro deve ficar visível no topo da Sidebar fixada.
   - **Mobile:** Se a Sidebar virar um Drawer / Bottom Sheet no mobile, o filtro deve continuar no topo do Drawer para que o usuário possa trocar de ano sem precisar fechar o painel.

3. **Acessibilidade (a11y):**
   - O seletor deve ter rótulo acessível (`aria-label="Filtrar leituras por ano"`).
   - Suporte completo a navegação via teclado (`Tab`, setas e `Enter`).

---

## ✅ Critérios de Aceite (Definition of Done)

- [ ] Ao abrir a aplicação sem parâmetros na URL, o ano atual é selecionado automaticamente.
- [ ] O dropdown exibe apenas anos que possuem dados gravados na base.
- [ ] Mudar o ano no dropdown atualiza os pinos no mapa e as estatísticas do painel simultaneamente.
- [ ] A URL é atualizada instantaneamente para refletir o ano escolhido (`?year=YYYY`).
- [ ] Abrir uma URL diretamente com `?year=2025` carrega a página já filtrada para 2025.
- [ ] Design adaptado perfeitamente para desktop e dispositivos móveis.