# Design: Redis Pooling & Observability (Fase 12.6 + 12.7)

## Contexto

Dois itens restantes da Fase 12 de performance do backend:
- **12.6**: Redis clients sem pool — PaymentUsecase cria/destroi client por chamada
- **12.7**: Sem metricas estruturadas — apenas logs texto via zerolog

## 12.6 — Redis Client Centralizado

### Problema
`PaymentUsecase.getRedisClient()` chama `redisOpt.MakeRedisClient()` a cada operacao, criando e fechando conexao imediatamente. Sem pool, sem reuso.

### Solucao
Criar `infra/redis/client.go` com client compartilhado e pool configurado.

**Novo pacote:** `infra/redis/`
```go
// infra/redis/client.go
func NewRedisClient(addr string) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:         addr,
        PoolSize:     10,
        MinIdleConns: 2,
        PoolTimeout:  5 * time.Second,
    })
}
```

**Mudancas no PaymentUsecase:**
- Campo `redisOpt asynq.RedisConnOpt` → `redisClient redis.UniversalClient`
- Remover `getRedisClient()` — usar campo direto
- Remover `defer redisClient.Close()` de cada metodo (lifecycle gerenciado pelo caller)

**Mudancas no cmd/api/main.go:**
- Criar `redisClient := redispkg.NewRedisClient(secrets.RedisAddr)` uma vez
- Injetar no `NewPaymentUsecase`
- `defer redisClient.Close()` no main
- Simplificar `connectRedis()` → usar o mesmo client para ping

**Mudancas no HealthController:**
- Aceitar `redis.UniversalClient` para health check direto (ao inves de depender do asynqClient.Ping)

## 12.7 — Prometheus Metrics

### Problema
Sem metricas queryaveis. Logs zerolog tem latency_ms mas nao sao agregaveis.

### Solucao
Adicionar Prometheus via `prometheus/client_golang` com middleware Gin custom.

**Novo pacote:** `infra/metrics/`
```
infra/metrics/
  prometheus.go    — middleware Gin + registro de metricas
  collectors.go    — DB pool stats + Redis pool stats collectors
```

**Metricas registradas:**
1. `http_requests_total` (Counter) — labels: method, path, status
2. `http_request_duration_seconds` (Histogram) — labels: method, path
3. `db_pool_open_connections` (Gauge) — do sql.DBStats
4. `db_pool_idle_connections` (Gauge)
5. `db_pool_wait_count_total` (Counter)
6. `redis_pool_hits_total` (Counter) — do redis.PoolStats
7. `redis_pool_misses_total` (Counter)
8. `redis_pool_active_connections` (Gauge)
9. `redis_pool_idle_connections` (Gauge)

**Endpoint:** `GET /metrics` (handler padrao Prometheus, sem auth)

**Mudancas no cmd/api/main.go:**
- Registrar middleware metrics nos route groups
- Adicionar rota `/metrics`
- Passar `*sql.DB` e `*redis.Client` para collectors

### Path normalization
Para evitar cardinalidade alta nos labels de path (ex: `/curriculum/123`), normalizar paths com parametros:
- `/curriculum/:id` ao inves de `/curriculum/123`
- Usar `c.FullPath()` do Gin que ja retorna o pattern registrado

## Dependencias Novas
- `github.com/prometheus/client_golang` — unica dependencia nova

## Arquivos Modificados
1. `infra/redis/client.go` (NOVO)
2. `infra/metrics/prometheus.go` (NOVO)
3. `infra/metrics/collectors.go` (NOVO)
4. `usecase/payment_usecase.go` (MODIFICADO — trocar redisOpt por redisClient)
5. `cmd/api/main.go` (MODIFICADO — injetar redis client, registrar metrics)
6. `controller/health_controller.go` (MODIFICADO — aceitar redis client)

## Testes
- `infra/redis/client_test.go` — verifica criacao com pool config
- `infra/metrics/prometheus_test.go` — verifica middleware registra metricas
- `usecase/payment_usecase_test.go` — atualizar se existir mock de redis
