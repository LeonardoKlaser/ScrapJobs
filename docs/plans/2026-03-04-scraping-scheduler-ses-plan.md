# Scraping Configs + Scheduler Fix + SES Validation — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix scheduler ticker drift bug, add SES startup validation to Worker, and generate scraping configs for 20 companies.

**Architecture:** Two surgical code changes in Go binaries (scheduler + worker), plus a research-driven document generation for scraping configurations. Backend changes are independent of each other and of the research task.

**Tech Stack:** Go 1.24, Asynq, AWS SES SDK v2, zerolog, WebFetch/WebSearch for research

---

### Task 1: Scheduler Fix — Extract Constants and Derive TTL

**Files:**
- Modify: `cmd/scheduler/main.go`

**Step 1: Add constants block after imports**

Add this block between the import closing paren (line 21) and `func main()` (line 23):

```go
const (
	scrapingInterval   = 120 * time.Minute
	matchInterval      = 4 * time.Hour
	matchUniqueTTL     = matchInterval - 10*time.Minute // 3h50m — expires before next tick
	digestInterval     = 8 * time.Hour
	digestUniqueTTL    = digestInterval - 10*time.Minute // 7h50m — expires before next tick
	deleteJobsInterval = 24 * time.Hour
)
```

**Step 2: Replace ticker literals with constants**

In `func main()`, replace these 4 lines:

| Line | Before | After |
|------|--------|-------|
| 88 | `time.NewTicker(120 * time.Minute)` | `time.NewTicker(scrapingInterval)` |
| 91 | `time.NewTicker(24 * time.Hour)` | `time.NewTicker(deleteJobsInterval)` |
| 94 | `time.NewTicker(4 * time.Hour)` | `time.NewTicker(matchInterval)` |
| 97 | `time.NewTicker(8 * time.Hour)` | `time.NewTicker(digestInterval)` |

**Step 3: Replace Unique TTL in enqueueMatchTasks**

Line 212: replace `asynq.Unique(4*time.Hour)` with `asynq.Unique(matchUniqueTTL)`

**Step 4: Replace Unique TTL in enqueueDigestTasks**

Line 250: replace `asynq.Unique(8*time.Hour)` with `asynq.Unique(digestUniqueTTL)`

**Step 5: Verify it compiles**

Run: `cd ScrapJobs && go build ./cmd/scheduler/...`
Expected: no errors, no output

**Step 6: Commit**

```bash
cd ScrapJobs
git add cmd/scheduler/main.go
git commit -m "fix(scheduler): derive asynq.Unique TTL from ticker interval with 10min margin

Fixes 'task already exists' error caused by Go ticker millisecond drift.
The Unique lock now expires 10 minutes before the next ticker cycle,
preventing the race condition where the lock hasn't expired yet."
```

---

### Task 2: SES Startup Validation in Worker

**Files:**
- Modify: `cmd/worker/main.go`

**Step 1: Enhance SES logging after config load**

Replace lines 63-76 (the SES config block) with:

```go
	// Carrega configuracao AWS para SES (email)
	awsCfg, err := ses.LoadAWSConfig(context.Background())
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("could not load aws config — email via SES nao estara disponivel")
	}

	senderEmail := os.Getenv("SES_SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "noreply@scrapjobs.com.br"
		logging.Logger.Warn().Msg("SES_SENDER_EMAIL nao definida — usando fallback noreply@scrapjobs.com.br")
	}

	clientSES := ses.LoadAWSClient(awsCfg)
	mailSender := ses.NewSESMailSender(clientSES, senderEmail)
	emailService := usecase.NewSESSenderAdapter(mailSender)

	if err == nil {
		logging.Logger.Info().
			Str("sender_email", senderEmail).
			Str("aws_region", awsCfg.Region).
			Msg("SES configurado com sucesso")
	}
```

The changes are:
1. Added `Warn` log when `SES_SENDER_EMAIL` is empty (line 71 area — currently no warning)
2. Added `Info` log after successful config showing `sender_email` and `aws_region` for Railway debugging

**Step 2: Verify it compiles**

Run: `cd ScrapJobs && go build ./cmd/worker/...`
Expected: no errors, no output

**Step 3: Commit**

```bash
cd ScrapJobs
git add cmd/worker/main.go
git commit -m "fix(worker): add SES config validation logging on startup

Logs sender_email and aws_region at startup to diagnose
Railway env var mismatches between API and Worker services."
```

---

### Task 3: Investigate ATS for 13 New Companies

**Research task — no code changes.**

For each company (Creditas, Stone, Magazine Luiza, Grupo Boticario, TOTVS, Vale, Natura, Embraer, Cielo, Wildlife Studios, PicPay, C6 Bank, Sicredi):

1. WebSearch: `"{company name}" careers page site:gupy.io OR site:greenhouse.io OR site:lever.co`
2. WebFetch the career page URL to confirm the ATS
3. If Greenhouse: verify slug via `https://boards-api.greenhouse.io/v1/boards/{slug}/jobs` (should return JSON)
4. If Gupy: confirm the subdomain via `https://{slug}.gupy.io/`
5. If other ATS: inspect the page for API endpoints or HTML structure

**Dispatch as parallel subagents** — 3 groups:
- Group A: Creditas, Stone, Magazine Luiza, Grupo Boticario, TOTVS
- Group B: Vale, Natura, Embraer, Cielo, Wildlife Studios
- Group C: PicPay, C6 Bank, Sicredi

Each subagent returns: `{ company, ats, slug_or_url, scraping_type, notes }`

---

### Task 4: Generate scraping_configs_inputs_v2.md

**Files:**
- Create: `scraping_configs_inputs_v2.md` (workspace root)

**Depends on:** Task 3 results

**Step 1: Write the file**

Structure:
1. Header with date and purpose
2. ATS patterns section (Gupy template, Greenhouse template, Eightfold template)
3. One JSON block per company (20 total), following the `SiteScrapingConfig` model fields:
   - `SiteName`, `BaseURL`, `IsActive`, `ScrapingType`
   - HTML fields: `JobListItemSelector`, `TitleSelector`, `LinkSelector`, `LinkAttribute`, `LocationSelector`, `NextPageSelector`, `JobDescriptionSelector`, `JobRequisitionIdSelector`
   - API fields: `APIEndpointTemplate`, `APIMethod`, `APIHeadersJSON`, `APIPayloadTemplate`, `JSONDataMappings`
4. Summary table

**Rules enforcement:**
- Gupy companies: `ScrapingType: "HTML"`, NO `_next/data` URLs, use provided CSS selector template
- Greenhouse companies: `ScrapingType: "API"`, endpoint `boards-api.greenhouse.io`, standard JSONDataMappings
- v1 Gupy companies (Itau, Ambev, Vivo): change from API to HTML

**Step 2: Commit**

```bash
git add scraping_configs_inputs_v2.md
git commit -m "docs: add scraping configs v2 for 20 companies

Includes configs for Nubank, Itau, Ambev, Mercado Livre, XP, Vivo,
QuintoAndar, Creditas, Stone, Magazine Luiza, Grupo Boticario, TOTVS,
Vale, Natura, Embraer, Cielo, Wildlife Studios, PicPay, C6 Bank, Sicredi.

Gupy companies now use HTML scraping instead of _next/data API."
```

---

### Task 5: Final Verification

**Step 1: Build all binaries**

```bash
cd ScrapJobs && go build ./...
```

Expected: no errors

**Step 2: Run tests**

```bash
cd ScrapJobs && go test ./...
```

Expected: all tests pass

**Step 3: Verify scraping configs file exists and has 20 entries**

```bash
grep -c '"SiteName"' ../scraping_configs_inputs_v2.md
```

Expected: `20`

---

## Execution Order

Tasks 1 and 2 are independent — can run in parallel.
Task 3 is research only — can run in parallel with 1 and 2.
Task 4 depends on Task 3.
Task 5 depends on Tasks 1, 2, and 4.

```
[Task 1: Scheduler Fix]  ──────────────────────┐
[Task 2: SES Validation] ──────────────────────┤
[Task 3: ATS Research] → [Task 4: Gen Configs] ┤
                                                └→ [Task 5: Verify]
```
