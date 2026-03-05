# Design: Scraping Configs + Scheduler Fix + SES Validation

Data: 2026-03-04

## Missao 1: Scraping Configs para 20 Empresas

### Objetivo

Gerar arquivo `scraping_configs_inputs_v2.md` com JSONs de configuracao para 20 empresas, prontos para insercao na tabela `site_scraping_config`.

### Regras

- **Gupy**: `ScrapingType: "HTML"`, URL base `{slug}.gupy.io`, CSS selectors padronizados. Proibido usar `_next/data/{buildId}`.
- **Greenhouse**: `ScrapingType: "API"`, endpoint `boards-api.greenhouse.io/v1/boards/{slug}/jobs?content=true`, JSONDataMappings padronizado.
- **Outros ATS**: API se tiver JSON publico limpo, HTML se SSR/complexo.

### Empresas

| # | Empresa | Status |
|---|---------|--------|
| 1 | Nubank | v1 existente (Greenhouse) |
| 2 | Itau Unibanco | v1 existente (Gupy) - trocar para HTML |
| 3 | Ambev | v1 existente (Gupy) - trocar para HTML |
| 4 | Mercado Livre | v1 existente (Eightfold) |
| 5 | XP Inc. | v1 existente (Greenhouse) |
| 6 | Vivo | v1 existente (Gupy) - trocar para HTML |
| 7 | QuintoAndar | v1 existente (Greenhouse) |
| 8-20 | Creditas, Stone, Magazine Luiza, Grupo Boticario, TOTVS, Vale, Natura, Embraer, Cielo, Wildlife Studios, PicPay, C6 Bank, Sicredi | Investigar ao vivo |

### Processo

1. WebSearch/WebFetch para identificar ATS de cada empresa
2. Aplicar template correto (Gupy/Greenhouse/Outro)
3. Validar endpoint/seletores quando possivel
4. Consolidar em arquivo unico com tabela resumo

### Entregavel

Arquivo `scraping_configs_inputs_v2.md` na raiz do workspace.

---

## Missao 2: Scheduler Fix (Ticker Drift)

### Problema

`time.Ticker(4h)` + `asynq.Unique(4h)` causa drift de milissegundos. O segundo ciclo tenta enfileirar a task antes do lock expirar, resultando em `task already exists`.

### Solucao

Extrair intervalos como constantes nomeadas e derivar o TTL do Unique automaticamente com margem de 10 minutos.

### Mudancas em `cmd/scheduler/main.go`

```go
const (
    scrapingInterval   = 120 * time.Minute
    matchInterval      = 4 * time.Hour
    matchUniqueTTL     = matchInterval - 10*time.Minute  // 3h50m
    digestInterval     = 8 * time.Hour
    digestUniqueTTL    = digestInterval - 10*time.Minute  // 7h50m
    deleteJobsInterval = 24 * time.Hour
)
```

- Substituir literais de tempo nos tickers pelas constantes
- `asynq.Unique(4*time.Hour)` -> `asynq.Unique(matchUniqueTTL)`
- `asynq.Unique(8*time.Hour)` -> `asynq.Unique(digestUniqueTTL)`

---

## Missao 3: Validacao SES no Worker

### Problema

Worker no Railway pode ter env vars AWS diferentes da API, causando falha no envio de emails. Erro: `Email address is not verified` mesmo com identidade verificada.

### Causa Provavel

Servicos separados no Railway com credenciais AWS apontando para contas/regioes diferentes.

### Solucao

Adicionar log de validacao no startup do Worker (`cmd/worker/main.go`) apos carregamento do SES:

- Logar `sender_email` e `aws_region` configurados
- Warning se `SES_SENDER_EMAIL` estiver vazio
- Warning se `LoadAWSConfig` falhar
- Nao bloqueia startup (worker continua para tasks sem email)

---

## Verificacao

1. `go build ./cmd/scheduler/...` compila sem erros
2. `go build ./cmd/worker/...` compila sem erros
3. Worker loga configuracao SES no startup
4. Arquivo `scraping_configs_inputs_v2.md` gerado com 20 empresas
