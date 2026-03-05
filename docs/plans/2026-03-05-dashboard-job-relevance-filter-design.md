# Design: Filtro de Relevancia no Dashboard

**Data:** 2026-03-05
**Status:** Aprovado

## Contexto

O dashboard mostra TODAS as vagas dos sites que o usuario esta inscrito, sem aplicar os `target_words` (filters) configurados por ele. Exemplo: 110 vagas aparecem, mas so 10 sao relevantes. Isso polui a interface e dificulta encontrar vagas que importam.

## Decisoes

- Filtro de match computado no SQL (campo `matched` via subquery)
- Query param `matched_only` (default `true`) controla o filtro
- Frontend: Switch discreto ao lado do titulo da secao de vagas
- Vagas com match recebem badge verde "Match" quando mostrando todas

## Arquitetura

### Backend

**Query SQL** — `GetLatestJobsPaginated` ganha:
1. `LEFT JOIN user_sites` pra pegar os `filters` (JSON array) de cada par user+site
2. Campo computado `matched` via `EXISTS` + `jsonb_array_elements_text` que checa se `LOWER(j.title)` contem algum filter
3. Quando `matched_only=true`, adiciona filtro `WHERE matched = true`
4. Campo `matched` retornado no JSON de cada job

**Model** — Novo struct `JobWithMatch`:
```go
type JobWithMatch struct {
    Job
    Matched bool `json:"matched"`
}
```

`PaginatedJobs.Jobs` muda de `[]Job` para `[]JobWithMatch`.

**Controller** — Le novo query param `matched_only` (default `"true"`), passa pro repository.

**Repository** — `GetLatestJobsPaginated` recebe novo parametro `matchedOnly bool`.

### Frontend

**Model** — `LatestJob` ganha `matched: boolean`

**Hook** — `useLatestJobs` aceita `matched_only?: boolean` nos params

**UI (Home.tsx):**
- Switch ao lado do `SectionHeader` das vagas
- Label: "So relevantes"
- Default: checked (true) — mostra so matched
- Quando unchecked: passa `matched_only=false` pro hook
- Na tabela, quando mostrando todas: vagas com `matched=true` recebem badge verde "Match" ao lado do titulo

### O que NAO muda

- Logica de notificacao assincrona (continua independente)
- Filtros existentes (search, days) — funcionam combinados com `matched_only`
- Cards de stats (contam todas as vagas, nao so matched)
- Endpoints de `user_sites` (CRUD de filters continua igual)
- `GetDashboardData` (query do resumo/cards)
