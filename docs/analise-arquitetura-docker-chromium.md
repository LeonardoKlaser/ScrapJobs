# Análise de Arquitetura Docker: Chromium no ScrapJobs

> Data: 2026-03-10
> Contexto: Necessidade de adicionar `chromium` à imagem Docker para suportar web scraping headless via `chromedp`.
> Ambiente de deploy: **Railway** (não AWS EC2/ECS)

---

## Situação Atual

| Item | Valor |
|------|-------|
| Imagem base | `alpine:3.21` |
| Tamanho estimado atual | ~40-50 MB |
| Serviços na mesma imagem | api, worker, scheduler, archive-monitor |
| Seleção de serviço | `SERVICE_CMD` via `entrypoint.sh` |
| Quem precisa do Chromium | **Apenas o Worker** (scraping headless) |
| Railway config | `railway.json` aponta para um único `Dockerfile` |

---

## Cenário 1: Imagem Única (Chromium para todos)

### O que muda no Dockerfile

```dockerfile
# Apenas adicionar esta linha na imagem final:
RUN apk --no-cache add ca-certificates curl procps chromium
```

### Impacto no tamanho da imagem

| Componente | Tamanho estimado |
|-----------|-----------------|
| alpine:3.21 base | ~7 MB |
| ca-certs + curl + procps | ~5 MB |
| 4 binários Go + migrate | ~30 MB |
| **chromium + dependências** | **~250-280 MB** |
| **Total estimado** | **~300 MB** |

O pacote `chromium` no Alpine puxa `nss`, `freetype`, `harfbuzz`, `fontconfig`, `ttf-freefont`, entre outras libs gráficas. Isso é inevitável — não tem como ter Chromium funcional sem essas dependências.

### Prós
- **Zero mudança na arquitetura**: 1 linha no Dockerfile
- **Zero mudança no CI/CD**: `railway.json`, `deploy.yml` e `docker-compose.yml` permanecem iguais
- **Zero mudança no Railway**: cada serviço continua usando a mesma imagem, diferenciando apenas pelo `SERVICE_CMD`
- **Tempo de implementação**: ~5 minutos

### Contras
- API, Scheduler e Archive-Monitor carregam Chromium sem precisar
- Imagem ~6x maior que o necessário para serviços que não usam headless
- Pull/deploy ligeiramente mais lento para todos os serviços

---

## Cenário 2: Multi-Target Build (Imagens especializadas)

### O que muda no Dockerfile

```dockerfile
# ---- Camada base (sem Chromium) ----
FROM alpine:3.21 AS base-release
RUN apk --no-cache add ca-certificates curl procps
WORKDIR /app
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
COPY --from=builder /go/bin/migrate /app/migrate
COPY --from=builder /app/api /app/api
COPY --from=builder /app/scheduler /app/scheduler
COPY --from=builder /app/archive-monitor /app/archive-monitor
COPY --from=builder /app/migrations /app/migrations
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh && chown -R nonroot:nonroot /app
EXPOSE 8080
USER nonroot
ENTRYPOINT ["/app/entrypoint.sh"]

# ---- Camada worker (com Chromium) ----
FROM base-release AS worker-release
USER root
RUN apk --no-cache add chromium
USER nonroot
ENV CHROME_BIN=/usr/bin/chromium-browser
COPY --from=builder /app/worker /app/worker
ENV SERVICE_CMD=/app/worker
```

### O que muda no Railway

O Railway **não suporta nativamente** múltiplos targets de build de um mesmo Dockerfile em um único projeto. Para usar este cenário, seria necessário:

1. **Criar serviços separados no Railway**, cada um apontando para um Dockerfile diferente (ou usar build args)
2. **Alternativa**: criar `Dockerfile.worker` separado e configurar cada serviço Railway com seu próprio `dockerfilePath`
3. Mudanças no `railway.json` **por serviço** (Railway permite configuração por serviço via UI ou `railway.toml`)

### O que muda no CI/CD (AWS deploy.yml)

O `deploy.yml` precisaria:
- Fazer **2 builds** (`docker build --target base-release` e `docker build --target worker-release`)
- Push de **2 imagens** para o ECR
- Atualizar o `docker-compose.yml` para usar `IMAGE_NAME_BASE` e `IMAGE_NAME_WORKER`
- O `SSM send-command` precisaria referenciar as imagens corretas

### Impacto no docker-compose.yml

```yaml
# Antes: todos usam ${IMAGE_NAME}
# Depois:
api:
  image: ${IMAGE_NAME_BASE}
worker:
  image: ${IMAGE_NAME_WORKER}
scheduler:
  image: ${IMAGE_NAME_BASE}
```

### Prós
- Imagens otimizadas por serviço (~40 MB vs ~300 MB)
- Princípio de responsabilidade única — cada serviço carrega só o que precisa
- Melhor prática da indústria para projetos grandes

### Contras
- **Complexidade significativa**: mudanças em Dockerfile, docker-compose, deploy.yml, e configuração do Railway
- **2 builds no CI/CD**: tempo de build dobra (~2x)
- **Gerenciar 2 imagens**: mais tags, mais espaço no registry, mais pontos de falha
- **Railway**: configuração não-trivial para multi-target

---

## Comparativo Direto

| Critério | Cenário 1 (Única) | Cenário 2 (Multi-Target) |
|----------|-------------------|--------------------------|
| **Mudanças necessárias** | 1 linha no Dockerfile | Dockerfile + compose + CI/CD + Railway config |
| **Tempo de implementação** | 5 min | 2-4 horas |
| **Tamanho da imagem (API)** | ~300 MB | ~40 MB |
| **Tamanho da imagem (Worker)** | ~300 MB | ~300 MB |
| **Tempo de build CI** | Igual ao atual (+30s) | ~2x (dois builds) |
| **Tempo de pull/deploy** | +10-15s por serviço | Igual para worker, menor para outros |
| **Risco de quebrar algo** | Quase zero | Médio (pipeline inteiro muda) |
| **Complexidade de manutenção** | Mesma de hoje | Maior (2 targets, 2 imagens) |

---

## Análise Financeira: Railway

O Railway cobra por **uso de recursos (CPU/RAM)** e não por tamanho de imagem armazenada. O impacto financeiro real entre os dois cenários é:

| Fator | Impacto |
|-------|---------|
| **Armazenamento de imagem** | Railway faz build internamente — você não paga storage separado de registry |
| **Tempo de build** | Cenário 2 dobra o tempo de build (mais minutos cobrados se houver limite) |
| **RAM em runtime** | Chromium instalado mas **não rodando** não consome RAM extra. O binário Go da API usa a mesma RAM independente de ter Chromium no filesystem |
| **Custo extra real** | **Praticamente zero** — a diferença entre ter ou não Chromium no disco não afeta o custo de runtime |
| **Tempo de cold start** | +5-10s para pull da imagem maior. Em Railway com deploys infrequentes, impacto negligível |

### Sobre AWS (deploy.yml atual)

| Fator | Cenário 1 | Cenário 2 |
|-------|-----------|-----------|
| **ECR Storage** | ~300 MB × $0.10/GB/mês = ~$0.03/mês | ~340 MB total (base+worker) = ~$0.034/mês |
| **Data Transfer** | 300 MB × 4 serviços × deploys | 40 MB × 3 + 300 MB × 1 × deploys |
| **Diferença mensal estimada** | — | Economia de ~$0.01-0.05/mês |

**Conclusão financeira**: A diferença de custo é irrelevante para um projeto MVP. Estamos falando de centavos por mês.

---

## Análise para Railway Especificamente

O Railway tem particularidades que favorecem o **Cenário 1**:

1. **Build interno**: O Railway faz o build da imagem no próprio servidor. Não há push/pull externo — o overhead de uma imagem maior é só no build time.

2. **Um Dockerfile por serviço**: Para multi-target, você precisaria configurar cada serviço Railway separadamente com `dockerfilePath` ou `buildCommand` diferentes. Isso é possível mas adiciona complexidade na UI do Railway.

3. **Sem registry separado**: Diferente de AWS ECR onde você paga por GB armazenado, o Railway gerencia o registry internamente.

4. **Deploy atômico**: Todos os serviços usarem a mesma imagem garante consistência. Com imagens separadas, existe o risco (pequeno) de uma buildar e a outra falhar.

---

## Recomendação Final

### Para agora (fase MVP): **Cenário 1 — Imagem Única**

**Motivos:**

1. **Simplicidade acima de tudo**: Você está em fase MVP. Cada hora gasta otimizando infraestrutura é uma hora não gasta em features que geram valor.

2. **O "custo" é cosmético**: 300 MB em vez de 40 MB parece ruim no papel, mas na prática:
   - Não afeta custo financeiro significativamente
   - Não afeta performance em runtime (Chromium no disco ≠ Chromium na RAM)
   - Cold start +10s é aceitável para deploys que acontecem poucas vezes por dia

3. **Risco zero**: É 1 linha de mudança. Não quebra nada. Você pode testar em 5 minutos.

4. **Railway-friendly**: Uma imagem, um Dockerfile, configuração simples.

### Para o futuro (quando escalar): Migrar para Cenário 2

Considere migrar quando:
- O projeto tiver **mais de 5 serviços** com necessidades distintas
- O tempo de deploy se tornar um **gargalo real** (não percebido, mas medido)
- A equipe crescer e a separação de concerns justificar a complexidade extra
- Você migrar para **Kubernetes/ECS** onde multi-image é padrão

### Implementação sugerida (Cenário 1)

Adicionar ao Dockerfile, na imagem final:

```dockerfile
FROM alpine:3.21

RUN apk --no-cache add ca-certificates curl procps chromium

ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/
```

E garantir no código Go (`chromedp`) que o path do Chromium é lido de `CHROME_BIN` ou use a detecção automática do `chromedp`.

---

## Referências

- [Railway Docs - Dockerfiles](https://docs.railway.com/guides/dockerfiles)
- [Alpine Packages - Chromium](https://pkgs.alpinelinux.org/package/v3.21/community/x86_64/chromium)
- [chromedp - Headless Chrome in Go](https://github.com/chromedp/chromedp)
