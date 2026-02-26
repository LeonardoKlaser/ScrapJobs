# Design: Refatoração para Arquitetura Batch/Digest Desacoplada

**Data:** 2026-02-25
**Status:** Aprovado

## Problema

O pipeline atual é síncrono e acoplado: `ScrapeSiteTask` -> enfileira `ProcessResultsTask` (com `[]*model.Job` inteiro) -> enfileira `NotifyNewJobsTask` (com `UserSiteCurriculum` + `[]*model.Job`). Isso trafega structs pesadas no Redis/Asynq e acopla 3 responsabilidades distintas.

## Solução

Quebrar em 3 processos independentes orquestrados por cronjobs separados no Scheduler:

1. **Ingestor (Scraping)** — a cada 2h, scrape + upsert, morre após persistir
2. **Match Engine** — a cada 4h, batch por usuário, gera notificações `PENDING`
3. **Carteiro (Digest)** — a cada 8h, envia email consolidado, marca `SENT`

Payloads trafegam apenas IDs. Sem análise AI no fluxo batch.

## Fase 0: Migration

**Arquivo:** `migrations/031_add_status_to_job_notifications.up.sql`

```sql
ALTER TABLE job_notifications
ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'SENT';

CREATE INDEX idx_job_notifications_status
ON job_notifications(user_id, status)
WHERE status = 'PENDING';
```

- Default `'SENT'` para registros existentes (sem spam retroativo)
- Indice parcial para queries do digest

**Down:**
```sql
DROP INDEX IF EXISTS idx_job_notifications_status;
ALTER TABLE job_notifications DROP COLUMN IF EXISTS status;
```

## Fase 1: Refatoração do Ingestor (Scraping)

### Mudanças

- `HandleScrapeSiteTask` em `processor/processor.go`: remover linhas 65-86 (marshal + enqueue de `TypeProcessResults`). Após `ScrapeAndStoreJobs`, retorna `nil`.
- Deletar `HandleFindMatchesTask` e `HandleNotifyNewJobsTask` do `processor/processor.go`
- Remover registros `TypeProcessResults` e `TypeNotifyNewJobs` do `cmd/worker/main.go`
- Remover de `tasks/payloads.go`: constantes `TypeProcessResults`, `TypeNotifyNewJobs` e structs `ProcessResultsPayload`, `NotifyNewJobsPayload`

### Sem mudança

- `ScrapeAndStoreJobs` continua igual
- `ScrapeSitePayload` mantém `SiteScrapingConfig` (é config, não dados)
- Cleanup de jobs velhos (24h ticker) continua

## Fase 2: Motor de Match (Batch por Usuário)

### Novos tipos

```go
// tasks/payloads.go
const TypeMatchUser = "match:user"

type MatchUserPayload struct {
    UserID int `json:"user_id"`
}
```

### Scheduler (novo ticker 4h)

1. Busca `SELECT DISTINCT user_id FROM user_sites`
2. Enfileira `MatchUserJob` por usuário com apenas `user_id`

### Worker: `HandleMatchUserTask`

1. Recebe `user_id`
2. Query otimizada (sem N+1):

```sql
SELECT j.id, j.title, j.location, j.company, j.job_link, us.filters
FROM jobs j
INNER JOIN user_sites us ON j.site_id = us.site_id AND us.user_id = $1
WHERE j.last_seen_at >= NOW() - INTERVAL '24 hours'
  AND NOT EXISTS (
      SELECT 1 FROM job_notifications jn
      WHERE jn.user_id = $1 AND jn.job_id = j.id
  )
```

3. Aplica `matchJobWithFilters` em Go (substring case-insensitive no titulo)
4. Bulk insert `PENDING`:

```sql
INSERT INTO job_notifications (user_id, job_id, status)
VALUES ($1, $2, 'PENDING'), ...
ON CONFLICT (user_id, job_id) DO NOTHING
```

### Novas interfaces

- `UserSiteRepositoryInterface.GetActiveUserIDs() ([]int, error)`
- `NotificationRepositoryInterface.BulkInsertPendingNotifications(userID int, jobIDs []int) error`
- Nova query: `GetUnnotifiedJobsForUser(userID int) ([]JobWithFilters, error)`

### Novo use case

```go
func (s *NotificationsUsecase) MatchJobsForUser(ctx context.Context, userID int) error
```

## Fase 3: Carteiro / Digest Email

### Novos tipos

```go
// tasks/payloads.go
const TypeSendDigest = "digest:send"

type SendDigestPayload struct {
    UserID int `json:"user_id"`
}
```

### Scheduler (novo ticker 8h)

1. Busca `SELECT DISTINCT user_id FROM job_notifications WHERE status = 'PENDING'`
2. Enfileira `SendDigestJob` por usuário com apenas `user_id`

### Worker: `HandleSendDigestTask`

1. Recebe `user_id`
2. Busca vagas pendentes:

```sql
SELECT jn.id, j.id, j.title, j.company, j.location, j.job_link
FROM job_notifications jn
INNER JOIN jobs j ON jn.job_id = j.id
WHERE jn.user_id = $1 AND jn.status = 'PENDING'
ORDER BY j.company, j.title
```

3. Busca nome/email do usuario
4. Monta email digest HTML (agrupado por empresa)
5. Envia via SES (reutiliza `SendNewJobsEmail` ou novo `SendDigestEmail`)
6. Bulk update apos envio com sucesso:

```sql
UPDATE job_notifications
SET status = 'SENT', notified_at = NOW()
WHERE user_id = $1 AND status = 'PENDING' AND job_id = ANY($2)
```

### Novas interfaces

- `NotificationRepositoryInterface.GetUserIDsWithPendingNotifications() ([]int, error)`
- `NotificationRepositoryInterface.GetPendingJobsForUser(userID int) ([]model.NotificationWithJob, error)`
- `NotificationRepositoryInterface.BulkUpdateNotificationStatus(userID int, jobIDs []int, status string) error`

### Novo use case

```go
func (s *NotificationsUsecase) SendDigestForUser(ctx context.Context, userID int) error
```

## Codigo morto a remover

- `FindMatches` em `notifications_usecase.go`
- `ProcessNewJobsNotification` em `notifications_usecase.go`
- `HandleFindMatchesTask` em `processor.go`
- `HandleNotifyNewJobsTask` em `processor.go`
- Structs: `ProcessResultsPayload`, `NotifyNewJobsPayload`
- Constantes: `TypeProcessResults`, `TypeNotifyNewJobs`

## Regras estritas

- Payloads Asynq trafegam apenas IDs (int)
- Zero queries dentro de loops (usar `WHERE id IN (...)` ou JOINs)
- `ON CONFLICT DO NOTHING` para idempotencia
- Sem analise AI no fluxo batch

## Diagrama do novo fluxo

```
Scheduler (2h)  ──► ScrapeSiteTask ──► Upsert jobs ──► FIM

Scheduler (4h)  ──► MatchUserJob(user_id) ──► Query + Filter + Bulk Insert PENDING ──► FIM

Scheduler (8h)  ──► SendDigestJob(user_id) ──► Query PENDING + Email + Bulk Update SENT ──► FIM
```
