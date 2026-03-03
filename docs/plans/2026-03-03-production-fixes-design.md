# Production Fixes — Design Doc

**Data**: 2026-03-03
**Contexto**: Primeiro deploy em producao no Railway. Auditoria identificou 5 itens a corrigir.

---

## Fix A — Webhook AbacatePay bloqueado pelo CSRF (URGENTE)

**Problema**: Rota `/api/webhooks/abacatepay` esta dentro do grupo `publicRoutes` que aplica `csrfMiddleware`. Webhooks server-to-server nao enviam header `Origin`, resultando em 403.

**Solucao**: Mover a rota do webhook para um grupo dedicado SEM CSRF. O webhook ja possui autenticacao propria (secret query param + HMAC signature), tornando o CSRF redundante.

**Arquivos**: `cmd/api/main.go`

**Mudanca**:
- Remover `publicRoutes.POST("/api/webhooks/abacatepay", ...)` do grupo `publicRoutes`
- Criar grupo `webhookRoutes` sem CSRF, apenas com logging e metrics
- Registrar a rota webhook nesse novo grupo

---

## Fix B — Security headers no nginx (Frontend)

**Problema**: Frontend sem headers de seguranca (XSS, clickjacking, HSTS).

**Solucao**: Adicionar headers no `nginx.conf`.

**Arquivos**: `FrontScrapJobs/nginx.conf`

**Headers**:
- `X-Frame-Options: SAMEORIGIN`
- `X-Content-Type-Options: nosniff`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `Referrer-Policy: strict-origin-when-cross-origin`

---

## Fix C — Rate limit agressivo em rotas publicas

**Problema**: Logs mostram 403 nas primeiras requests GET a `/api/plans`. CSRF ja ignora GET, entao o bloqueio vem do rate limiter (5 req/60s no `publicRoutes`).

**Solucao**: Aumentar rate limit do `publicRoutes` de 5/60s para 15/60s. O rate limit mais restritivo (5/60s) fica apenas no `forgotPasswordRoutes`.

**Arquivos**: `cmd/api/main.go`

---

## Fix D — robots.txt

**Problema**: `/robots.txt` retorna 404.

**Solucao**: Criar `FrontScrapJobs/public/robots.txt`.

**Arquivos**: `FrontScrapJobs/public/robots.txt`

---

## Fix E — Healthcheck no API (manual)

**Problema**: Removido do `railway.json` porque afetava servicos nao-HTTP.

**Solucao**: Configurar manualmente no dashboard Railway: API > Settings > Deploy > Health Check Path: `/health/live`, Timeout: 60s.

**Acao**: Manual no dashboard, sem codigo.
