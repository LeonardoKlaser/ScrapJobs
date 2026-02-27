# Redis Rate Limiter Distribuido — Design

## Objetivo
Substituir rate limiter in-memory por Redis-based, permitindo escalar para multiplas instancias da API com contadores compartilhados.

## Algoritmo: Fixed Window Counter (Lua atomico)
```lua
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local current = redis.call("INCR", key)
if current == 1 then
    redis.call("EXPIRE", key, window)
end
if current > limit then
    return 0
end
return 1
```
- 1 round-trip Redis por request
- Sem cleanup goroutine (Redis TTL cuida da limpeza)
- Key format: `rate_limit:{ip}` com auto-expire

## Componentes

### 1. `middleware/redis_rate_limiter.go`
- `RedisRateLimiter(redisClient *redis.Client, limit int, windowSeconds int) gin.HandlerFunc`
- Resposta 429 identica ao atual

### 2. `cmd/api/main.go`
- Trocar 4 chamadas `middleware.RateLimiter(...)` por `middleware.RedisRateLimiter(redisClient, ...)`
- Manter `RateLimiter()` in-memory para dev/fallback

### 3. `middleware/redis_rate_limiter_test.go`
- miniredis para simular Redis
- Testes: dentro/acima do limite, IPs diferentes, 2 instancias compartilham contador

## Mapeamento de limites

| Uso | Atual | Redis |
|-----|-------|-------|
| Public | 5.0/60.0, burst 2 | 5, 60s |
| Checkout | 10.0/60.0, burst 3 | 10, 60s |
| Private write | 15.0/60.0, burst 10 | 15, 60s |
| Analyze | 3.0/60.0, burst 2 | 3, 60s |
