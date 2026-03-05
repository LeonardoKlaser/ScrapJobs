# Multi-Provider Email (Resend + SES) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add Resend as email provider with admin-configurable priority, fallback, and toggle alongside existing SES.

**Architecture:** Extract a `MailSender` interface from the existing `SESMailSender`, create `ResendMailSender` implementing the same interface, and add an `EmailOrchestrator` that reads provider config from Postgres to decide send order with automatic fallback. Admin UI section added to existing dashboard page.

**Tech Stack:** Go (resend-go/v2 SDK), PostgreSQL (new migration), React + shadcn/ui (Switch, Badge, Card)

---

### Task 1: Database Migration — `email_provider_config` table

**Files:**
- Create: `ScrapJobs/migrations/036_create_email_provider_config.up.sql`
- Create: `ScrapJobs/migrations/036_create_email_provider_config.down.sql`

**Step 1: Create up migration**

```sql
CREATE TABLE email_provider_config (
    id SERIAL PRIMARY KEY,
    provider_name VARCHAR(20) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 1,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by INTEGER REFERENCES users(id)
);

INSERT INTO email_provider_config (provider_name, is_active, priority)
VALUES ('resend', true, 1), ('ses', true, 2);
```

**Step 2: Create down migration**

```sql
DROP TABLE IF EXISTS email_provider_config;
```

**Step 3: Run migration locally**

Run: `cd ScrapJobs && migrate -path migrations/ -database "$DATABASE_URL" up`
Expected: migration 036 applied successfully

**Step 4: Commit**

```bash
cd ScrapJobs
git add migrations/036_*
git commit -m "feat: add email_provider_config migration"
```

---

### Task 2: `MailSender` Interface

**Files:**
- Create: `ScrapJobs/interfaces/mail_sender_interface.go`

**Step 1: Create the interface**

```go
package interfaces

import "context"

// MailSender is the low-level contract for sending raw emails.
// Both SESMailSender and ResendMailSender implement this.
type MailSender interface {
	SendEmail(ctx context.Context, to string, subject string, bodyText string, bodyHtml string) error
}
```

**Step 2: Commit**

```bash
git add interfaces/mail_sender_interface.go
git commit -m "feat: add MailSender interface"
```

---

### Task 3: Make `SESMailSender` implicitly satisfy `MailSender`

No code changes needed — `SESMailSender.SendEmail` already matches the signature `(ctx context.Context, to string, subject string, bodyText string, bodyHtml string) error`. Just verify.

**Step 1: Add compile-time check**

Modify: `ScrapJobs/infra/ses/aws_ses.go` — add after the struct definition (after line 15):

```go
// compile-time check
var _ interfaces.MailSender = (*SESMailSender)(nil)
```

Add `"web-scrapper/interfaces"` to imports.

**Step 2: Run build**

Run: `cd ScrapJobs && go build ./...`
Expected: compiles successfully

**Step 3: Commit**

```bash
git add infra/ses/aws_ses.go
git commit -m "feat: SESMailSender satisfies MailSender interface"
```

---

### Task 4: `ResendMailSender` Implementation

**Files:**
- Create: `ScrapJobs/infra/resend/resend_sender.go`

**Step 1: Add resend-go dependency**

Run: `cd ScrapJobs && go get github.com/resend/resend-go/v2`

**Step 2: Create ResendMailSender**

```go
package resend

import (
	"context"
	"fmt"
	"web-scrapper/interfaces"
	"web-scrapper/logging"

	resendSdk "github.com/resend/resend-go/v2"
)

var _ interfaces.MailSender = (*ResendMailSender)(nil)

type ResendMailSender struct {
	client *resendSdk.Client
	from   string
}

func NewResendMailSender(apiKey, from string) *ResendMailSender {
	client := resendSdk.NewClient(apiKey)
	return &ResendMailSender{
		client: client,
		from:   from,
	}
}

func (r *ResendMailSender) SendEmail(ctx context.Context, to string, subject string, bodyText string, bodyHtml string) error {
	params := &resendSdk.SendEmailRequest{
		From:    r.from,
		To:      []string{to},
		Subject: subject,
		Html:    bodyHtml,
		Text:    bodyText,
	}

	sent, err := r.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend send failed: %w", err)
	}

	logging.Logger.Info().
		Str("subject", subject).
		Str("to", to).
		Str("resend_id", sent.Id).
		Msg("E-mail enviado via Resend")

	return nil
}
```

**Step 3: Run build**

Run: `cd ScrapJobs && go build ./...`
Expected: compiles successfully

**Step 4: Commit**

```bash
git add infra/resend/ go.mod go.sum
git commit -m "feat: add ResendMailSender implementation"
```

---

### Task 5: Email Provider Config Repository

**Files:**
- Create: `ScrapJobs/interfaces/email_config_interface.go`
- Create: `ScrapJobs/repository/email_config_repository.go`
- Create: `ScrapJobs/model/email_config.go`

**Step 1: Create model**

```go
package model

import "time"

type EmailProviderConfig struct {
	ID           int       `json:"id"`
	ProviderName string    `json:"provider_name"`
	IsActive     bool      `json:"is_active"`
	Priority     int       `json:"priority"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    *int      `json:"updated_by"`
}
```

**Step 2: Create interface**

```go
package interfaces

import "web-scrapper/model"

type EmailConfigRepository interface {
	GetAll() ([]model.EmailProviderConfig, error)
	Update(configs []model.EmailProviderConfig, updatedBy int) error
}
```

**Step 3: Create repository**

```go
package repository

import (
	"database/sql"
	"web-scrapper/model"
)

type EmailConfigRepo struct {
	db *sql.DB
}

func NewEmailConfigRepo(db *sql.DB) *EmailConfigRepo {
	return &EmailConfigRepo{db: db}
}

func (r *EmailConfigRepo) GetAll() ([]model.EmailProviderConfig, error) {
	rows, err := r.db.Query(`
		SELECT id, provider_name, is_active, priority, updated_at, updated_by
		FROM email_provider_config
		ORDER BY priority ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []model.EmailProviderConfig
	for rows.Next() {
		var c model.EmailProviderConfig
		if err := rows.Scan(&c.ID, &c.ProviderName, &c.IsActive, &c.Priority, &c.UpdatedAt, &c.UpdatedBy); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

func (r *EmailConfigRepo) Update(configs []model.EmailProviderConfig, updatedBy int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE email_provider_config
		SET is_active = $1, priority = $2, updated_at = NOW(), updated_by = $3
		WHERE provider_name = $4
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range configs {
		if _, err := stmt.Exec(c.IsActive, c.Priority, updatedBy, c.ProviderName); err != nil {
			return err
		}
	}

	return tx.Commit()
}
```

**Step 4: Run build**

Run: `cd ScrapJobs && go build ./...`

**Step 5: Commit**

```bash
git add model/email_config.go interfaces/email_config_interface.go repository/email_config_repository.go
git commit -m "feat: add EmailProviderConfig model, interface, repository"
```

---

### Task 6: `EmailOrchestrator`

**Files:**
- Create: `ScrapJobs/usecase/email_orchestrator.go`

**Step 1: Create orchestrator**

```go
package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/model"
)

type EmailOrchestrator struct {
	senders    map[string]interfaces.MailSender
	configRepo interfaces.EmailConfigRepository

	mu          sync.RWMutex
	cachedOrder []model.EmailProviderConfig
	cacheExpiry time.Time
}

func NewEmailOrchestrator(
	senders map[string]interfaces.MailSender,
	configRepo interfaces.EmailConfigRepository,
) *EmailOrchestrator {
	return &EmailOrchestrator{
		senders:    senders,
		configRepo: configRepo,
	}
}

func (o *EmailOrchestrator) getActiveProviders() ([]model.EmailProviderConfig, error) {
	o.mu.RLock()
	if time.Now().Before(o.cacheExpiry) && len(o.cachedOrder) > 0 {
		result := o.cachedOrder
		o.mu.RUnlock()
		return result, nil
	}
	o.mu.RUnlock()

	configs, err := o.configRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load email provider config: %w", err)
	}

	var active []model.EmailProviderConfig
	for _, c := range configs {
		if c.IsActive {
			active = append(active, c)
		}
	}

	o.mu.Lock()
	o.cachedOrder = active
	o.cacheExpiry = time.Now().Add(5 * time.Minute)
	o.mu.Unlock()

	return active, nil
}

// InvalidateCache forces re-read on next send (called after admin update)
func (o *EmailOrchestrator) InvalidateCache() {
	o.mu.Lock()
	o.cacheExpiry = time.Time{}
	o.mu.Unlock()
}

func (o *EmailOrchestrator) SendEmail(ctx context.Context, to string, subject string, bodyText string, bodyHtml string) error {
	providers, err := o.getActiveProviders()
	if err != nil {
		return err
	}

	if len(providers) == 0 {
		return fmt.Errorf("no active email providers configured")
	}

	var lastErr error
	for i, p := range providers {
		sender, ok := o.senders[p.ProviderName]
		if !ok {
			logging.Logger.Warn().Str("provider", p.ProviderName).Msg("Email provider configured but no sender available — skipping")
			continue
		}

		err := sender.SendEmail(ctx, to, subject, bodyText, bodyHtml)
		if err == nil {
			return nil
		}

		lastErr = err
		logging.Logger.Warn().
			Err(err).
			Str("provider", p.ProviderName).
			Str("to", to).
			Str("subject", subject).
			Msg("Email provider failed")

		if i < len(providers)-1 {
			next := providers[i+1]
			logging.Logger.Warn().
				Str("failed_provider", p.ProviderName).
				Str("fallback_provider", next.ProviderName).
				Msg("Fallback ativado — tentando proximo provider")
		}
	}

	return fmt.Errorf("all email providers failed, last error: %w", lastErr)
}
```

**Step 2: Run build**

Run: `cd ScrapJobs && go build ./...`

**Step 3: Commit**

```bash
git add usecase/email_orchestrator.go
git commit -m "feat: add EmailOrchestrator with fallback and caching"
```

---

### Task 7: Update `SESSenderAdapter` to use `MailSender` interface

**Files:**
- Modify: `ScrapJobs/usecase/emailAdapter.go:208-216`

**Step 1: Change struct and constructor**

Replace lines 208-216 with:

```go
type SESSenderAdapter struct {
	mailSender interfaces.MailSender
}

func NewSESSenderAdapter(mailSender interfaces.MailSender) *SESSenderAdapter {
	return &SESSenderAdapter{
		mailSender: mailSender,
	}
}
```

Update imports: replace `"web-scrapper/infra/ses"` with `"web-scrapper/interfaces"`.

**Step 2: Run tests**

Run: `cd ScrapJobs && go test ./usecase/...`
Expected: all pass (SESMailSender satisfies MailSender, so existing callers still work)

**Step 3: Commit**

```bash
git add usecase/emailAdapter.go
git commit -m "refactor: SESSenderAdapter uses MailSender interface"
```

---

### Task 8: Wire Orchestrator in `cmd/api/main.go`

**Files:**
- Modify: `ScrapJobs/cmd/api/main.go:161-173` (SES setup section)

**Step 1: Update email initialization**

Replace lines 161-173 with:

```go
	// --- Email Providers ---
	senderEmail := os.Getenv("SES_SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "noreply@scrapjobs.com.br"
	}

	emailSenders := make(map[string]interfaces.MailSender)

	// SES sender
	awsCfgSES, sesErr := ses.LoadAWSConfig(context.Background())
	if sesErr != nil {
		logging.Logger.Warn().Err(sesErr).Msg("Falha ao carregar configuração AWS para SES — SES indisponível")
	} else {
		clientSES := ses.LoadAWSClient(awsCfgSES)
		emailSenders["ses"] = ses.NewSESMailSender(clientSES, senderEmail)
		logging.Logger.Info().Msg("SES email sender configurado")
	}

	// Resend sender
	resendKey := os.Getenv("RESEND_API_KEY")
	resendFrom := os.Getenv("RESEND_SENDER_EMAIL")
	if resendFrom == "" {
		resendFrom = senderEmail
	}
	if resendKey != "" {
		emailSenders["resend"] = resend.NewResendMailSender(resendKey, resendFrom)
		logging.Logger.Info().Msg("Resend email sender configurado")
	} else {
		logging.Logger.Warn().Msg("RESEND_API_KEY não definida — Resend indisponível")
	}

	emailConfigRepo := repository.NewEmailConfigRepo(dbConnection)
	orchestrator := usecase.NewEmailOrchestrator(emailSenders, emailConfigRepo)
	emailService := usecase.NewSESSenderAdapter(orchestrator)
```

Add imports: `"web-scrapper/infra/resend"`, `"web-scrapper/interfaces"`.

**Step 2: Run build**

Run: `cd ScrapJobs && go build ./cmd/api/...`

**Step 3: Commit**

```bash
git add cmd/api/main.go
git commit -m "feat: wire EmailOrchestrator in API binary"
```

---

### Task 9: Wire Orchestrator in `cmd/worker/main.go`

**Files:**
- Modify: `ScrapJobs/cmd/worker/main.go:63-84` (SES setup section)

**Step 1: Update email initialization**

Replace lines 63-84 with same pattern as Task 8 (emailSenders map, SES + Resend setup, orchestrator, adapter).

```go
	// --- Email Providers ---
	senderEmail := os.Getenv("SES_SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "noreply@scrapjobs.com.br"
		logging.Logger.Warn().Msg("SES_SENDER_EMAIL nao definida — usando fallback noreply@scrapjobs.com.br")
	}

	emailSenders := make(map[string]interfaces.MailSender)

	awsCfg, err := ses.LoadAWSConfig(context.Background())
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("could not load aws config — SES indisponível")
	} else {
		clientSES := ses.LoadAWSClient(awsCfg)
		emailSenders["ses"] = ses.NewSESMailSender(clientSES, senderEmail)
		logging.Logger.Info().
			Str("sender_email", senderEmail).
			Str("aws_region", awsCfg.Region).
			Msg("SES configurado com sucesso")
	}

	resendKey := os.Getenv("RESEND_API_KEY")
	resendFrom := os.Getenv("RESEND_SENDER_EMAIL")
	if resendFrom == "" {
		resendFrom = senderEmail
	}
	if resendKey != "" {
		emailSenders["resend"] = resend.NewResendMailSender(resendKey, resendFrom)
		logging.Logger.Info().Msg("Resend email sender configurado")
	} else {
		logging.Logger.Warn().Msg("RESEND_API_KEY não definida — Resend indisponível")
	}

	emailConfigRepo := repository.NewEmailConfigRepo(dbConnection)
	orchestrator := usecase.NewEmailOrchestrator(emailSenders, emailConfigRepo)
	emailService := usecase.NewSESSenderAdapter(orchestrator)
```

Add imports: `"web-scrapper/infra/resend"`, `"web-scrapper/interfaces"`.

**Step 2: Run build**

Run: `cd ScrapJobs && go build ./cmd/worker/...`

**Step 3: Commit**

```bash
git add cmd/worker/main.go
git commit -m "feat: wire EmailOrchestrator in Worker binary"
```

---

### Task 10: Admin Email Config Controller + Routes

**Files:**
- Create: `ScrapJobs/controller/email_config_controller.go`
- Modify: `ScrapJobs/cmd/api/main.go:332-343` (admin routes)

**Step 1: Create controller**

```go
package controller

import (
	"net/http"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/model"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type EmailConfigController struct {
	repo         interfaces.EmailConfigRepository
	orchestrator *usecase.EmailOrchestrator
}

func NewEmailConfigController(repo interfaces.EmailConfigRepository, orchestrator *usecase.EmailOrchestrator) *EmailConfigController {
	return &EmailConfigController{repo: repo, orchestrator: orchestrator}
}

func (c *EmailConfigController) GetEmailConfig(ctx *gin.Context) {
	configs, err := c.repo.GetAll()
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao buscar configuração de email")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar configuração"})
		return
	}
	ctx.JSON(http.StatusOK, configs)
}

type UpdateEmailConfigRequest struct {
	Providers []struct {
		ProviderName string `json:"provider_name" binding:"required"`
		IsActive     bool   `json:"is_active"`
		Priority     int    `json:"priority" binding:"required"`
	} `json:"providers" binding:"required"`
}

func (c *EmailConfigController) UpdateEmailConfig(ctx *gin.Context) {
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	var req UpdateEmailConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Validate at least one provider is active
	hasActive := false
	for _, p := range req.Providers {
		if p.IsActive {
			hasActive = true
			break
		}
	}
	if !hasActive {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Pelo menos um provedor deve estar ativo"})
		return
	}

	configs := make([]model.EmailProviderConfig, len(req.Providers))
	for i, p := range req.Providers {
		configs[i] = model.EmailProviderConfig{
			ProviderName: p.ProviderName,
			IsActive:     p.IsActive,
			Priority:     p.Priority,
		}
	}

	uid := userID.(int)
	if err := c.repo.Update(configs, uid); err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao atualizar configuração de email")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar configuração"})
		return
	}

	c.orchestrator.InvalidateCache()

	logging.Logger.Info().Int("updated_by", uid).Msg("Configuração de email atualizada")
	ctx.JSON(http.StatusOK, gin.H{"message": "Configuração atualizada com sucesso"})
}
```

**Step 2: Add routes in `cmd/api/main.go`**

After creating the emailConfigController (near the other controller initializations), add:

```go
emailConfigController := controller.NewEmailConfigController(emailConfigRepo, orchestrator)
```

In admin routes block (after line 342), add:

```go
adminRoutes.GET("/api/admin/email-config", emailConfigController.GetEmailConfig)
adminRoutes.PUT("/api/admin/email-config", emailConfigController.UpdateEmailConfig)
```

**Step 3: Run build**

Run: `cd ScrapJobs && go build ./cmd/api/...`

**Step 4: Commit**

```bash
git add controller/email_config_controller.go cmd/api/main.go
git commit -m "feat: add admin email config endpoints"
```

---

### Task 11: Update `.env.example`

**Files:**
- Modify: `ScrapJobs/.env.example`

**Step 1: Add Resend env vars**

Add after the SES section:

```env
# Resend (provedor de email alternativo)
RESEND_API_KEY=
RESEND_SENDER_EMAIL=
```

**Step 2: Commit**

```bash
git add .env.example
git commit -m "docs: add Resend env vars to .env.example"
```

---

### Task 12: Frontend — Email Config Service + Hook

**Files:**
- Modify: `FrontScrapJobs/src/services/adminDashboardService.ts`
- Modify: `FrontScrapJobs/src/hooks/useAdminDashboard.ts`

**Step 1: Add types and service methods to `adminDashboardService.ts`**

Append to the file:

```typescript
export interface EmailProviderConfig {
  id: number
  provider_name: string
  is_active: boolean
  priority: number
  updated_at: string
  updated_by: number | null
}

export interface UpdateEmailConfigPayload {
  providers: {
    provider_name: string
    is_active: boolean
    priority: number
  }[]
}

export const emailConfigService = {
  getConfig: async (): Promise<EmailProviderConfig[]> => {
    const { data } = await api.get('/api/admin/email-config')
    return data
  },

  updateConfig: async (payload: UpdateEmailConfigPayload): Promise<void> => {
    await api.put('/api/admin/email-config', payload)
  }
}
```

**Step 2: Add hook to `useAdminDashboard.ts`**

Append:

```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { emailConfigService, type UpdateEmailConfigPayload } from '@/services/adminDashboardService'

export function useEmailConfig() {
  return useQuery({
    queryKey: ['email-config'],
    queryFn: emailConfigService.getConfig
  })
}

export function useUpdateEmailConfig() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: UpdateEmailConfigPayload) => emailConfigService.updateConfig(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['email-config'] })
    }
  })
}
```

**Step 3: Commit**

```bash
cd FrontScrapJobs
git add src/services/adminDashboardService.ts src/hooks/useAdminDashboard.ts
git commit -m "feat: add email config service and hooks"
```

---

### Task 13: Frontend — Email Config Section Component

**Files:**
- Create: `FrontScrapJobs/src/components/adminDashboard/email-config-section.tsx`

**Step 1: Create the component**

```tsx
import { useState, useEffect } from 'react'
import { Card } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Mail, ArrowUp, ArrowDown, Loader2 } from 'lucide-react'
import { useEmailConfig, useUpdateEmailConfig } from '@/hooks/useAdminDashboard'
import type { EmailProviderConfig } from '@/services/adminDashboardService'

export function EmailConfigSection() {
  const { data: configs, isLoading } = useEmailConfig()
  const updateMutation = useUpdateEmailConfig()
  const [localConfigs, setLocalConfigs] = useState<EmailProviderConfig[]>([])
  const [isDirty, setIsDirty] = useState(false)

  useEffect(() => {
    if (configs) {
      setLocalConfigs([...configs].sort((a, b) => a.priority - b.priority))
      setIsDirty(false)
    }
  }, [configs])

  function toggleActive(providerName: string) {
    setLocalConfigs(prev => {
      const updated = prev.map(c =>
        c.provider_name === providerName ? { ...c, is_active: !c.is_active } : c
      )
      const activeCount = updated.filter(c => c.is_active).length
      if (activeCount === 0) return prev
      return updated
    })
    setIsDirty(true)
  }

  function movePriority(index: number, direction: 'up' | 'down') {
    setLocalConfigs(prev => {
      const arr = [...prev]
      const swapIdx = direction === 'up' ? index - 1 : index + 1
      if (swapIdx < 0 || swapIdx >= arr.length) return prev
      ;[arr[index], arr[swapIdx]] = [arr[swapIdx], arr[index]]
      return arr.map((c, i) => ({ ...c, priority: i + 1 }))
    })
    setIsDirty(true)
  }

  function handleSave() {
    updateMutation.mutate({
      providers: localConfigs.map(c => ({
        provider_name: c.provider_name,
        is_active: c.is_active,
        priority: c.priority
      }))
    })
  }

  if (isLoading) {
    return (
      <Card className="p-6">
        <div className="flex items-center gap-2 text-muted-foreground">
          <Loader2 className="size-4 animate-spin" />
          <span>Carregando configuração de email...</span>
        </div>
      </Card>
    )
  }

  return (
    <Card className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <Mail className="size-5 text-primary" />
          <h3 className="text-lg font-semibold font-display">Provedores de Email</h3>
        </div>
        {isDirty && (
          <Button
            size="sm"
            onClick={handleSave}
            disabled={updateMutation.isPending}
          >
            {updateMutation.isPending ? (
              <Loader2 className="size-4 animate-spin mr-2" />
            ) : null}
            Salvar
          </Button>
        )}
      </div>

      <div className="space-y-3">
        {localConfigs.map((config, index) => (
          <div
            key={config.provider_name}
            className="flex items-center justify-between p-4 rounded-lg border border-border/50 hover:border-primary/30 transition-colors"
          >
            <div className="flex items-center gap-4">
              <div className="flex flex-col gap-1">
                <button
                  onClick={() => movePriority(index, 'up')}
                  disabled={index === 0}
                  className="text-muted-foreground hover:text-foreground disabled:opacity-30 transition-colors"
                >
                  <ArrowUp className="size-3.5" />
                </button>
                <button
                  onClick={() => movePriority(index, 'down')}
                  disabled={index === localConfigs.length - 1}
                  className="text-muted-foreground hover:text-foreground disabled:opacity-30 transition-colors"
                >
                  <ArrowDown className="size-3.5" />
                </button>
              </div>

              <div>
                <div className="flex items-center gap-2">
                  <span className="font-medium capitalize">{config.provider_name}</span>
                  {index === 0 && config.is_active && (
                    <Badge variant="default" className="text-xs">Primário</Badge>
                  )}
                  {index > 0 && config.is_active && (
                    <Badge variant="secondary" className="text-xs">Fallback</Badge>
                  )}
                  {!config.is_active && (
                    <Badge variant="outline" className="text-xs text-muted-foreground">Desativado</Badge>
                  )}
                </div>
              </div>
            </div>

            <Switch
              checked={config.is_active}
              onCheckedChange={() => toggleActive(config.provider_name)}
            />
          </div>
        ))}
      </div>

      {updateMutation.isSuccess && (
        <p className="text-sm text-primary mt-4">Configuração salva com sucesso.</p>
      )}
      {updateMutation.isError && (
        <p className="text-sm text-destructive mt-4">Erro ao salvar configuração.</p>
      )}
    </Card>
  )
}
```

**Step 2: Commit**

```bash
git add src/components/adminDashboard/email-config-section.tsx
git commit -m "feat: add EmailConfigSection component"
```

---

### Task 14: Frontend — Add Section to Admin Dashboard Page

**Files:**
- Modify: `FrontScrapJobs/src/pages/adminDashboard.tsx`

**Step 1: Import and add section**

Add import:
```tsx
import { EmailConfigSection } from '@/components/adminDashboard/email-config-section'
```

Add after ChartsSection div (after line 26), before ActivityLogs:

```tsx
      <div className="animate-fade-in-up" style={{ animationDelay: '250ms' }}>
        <EmailConfigSection />
      </div>
```

**Step 2: Run dev server**

Run: `cd FrontScrapJobs && npm run build`
Expected: builds successfully

**Step 3: Commit**

```bash
git add src/pages/adminDashboard.tsx
git commit -m "feat: add email config section to admin dashboard"
```

---

### Task 15: Backend Tests — EmailOrchestrator

**Files:**
- Create: `ScrapJobs/usecase/email_orchestrator_test.go`
- Create: `ScrapJobs/repository/mocks/email_config_repo.go`

**Step 1: Create mock**

```go
package mocks

import (
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockEmailConfigRepo struct {
	mock.Mock
}

func (m *MockEmailConfigRepo) GetAll() ([]model.EmailProviderConfig, error) {
	args := m.Called()
	return args.Get(0).([]model.EmailProviderConfig), args.Error(1)
}

func (m *MockEmailConfigRepo) Update(configs []model.EmailProviderConfig, updatedBy int) error {
	args := m.Called(configs, updatedBy)
	return args.Error(0)
}
```

**Step 2: Create mock mail sender**

```go
// Add to repository/mocks/mail_sender.go
package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockMailSender struct {
	mock.Mock
}

func (m *MockMailSender) SendEmail(ctx context.Context, to string, subject string, bodyText string, bodyHtml string) error {
	args := m.Called(ctx, to, subject, bodyText, bodyHtml)
	return args.Error(0)
}
```

**Step 3: Create orchestrator tests**

```go
package usecase

import (
	"context"
	"errors"
	"testing"
	"web-scrapper/interfaces"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEmailOrchestrator_SendEmail_PrimarySuccess(t *testing.T) {
	mockRepo := new(mocks.MockEmailConfigRepo)
	mockResend := new(mocks.MockMailSender)
	mockSES := new(mocks.MockMailSender)

	mockRepo.On("GetAll").Return([]model.EmailProviderConfig{
		{ProviderName: "resend", IsActive: true, Priority: 1},
		{ProviderName: "ses", IsActive: true, Priority: 2},
	}, nil)

	mockResend.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(nil)

	senders := map[string]interfaces.MailSender{
		"resend": mockResend,
		"ses":    mockSES,
	}

	orch := NewEmailOrchestrator(senders, mockRepo)
	err := orch.SendEmail(context.Background(), "to@test.com", "subject", "text", "html")

	assert.NoError(t, err)
	mockResend.AssertExpectations(t)
	mockSES.AssertNotCalled(t, "SendEmail")
}

func TestEmailOrchestrator_SendEmail_FallbackOnPrimaryFailure(t *testing.T) {
	mockRepo := new(mocks.MockEmailConfigRepo)
	mockResend := new(mocks.MockMailSender)
	mockSES := new(mocks.MockMailSender)

	mockRepo.On("GetAll").Return([]model.EmailProviderConfig{
		{ProviderName: "resend", IsActive: true, Priority: 1},
		{ProviderName: "ses", IsActive: true, Priority: 2},
	}, nil)

	mockResend.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(errors.New("resend down"))
	mockSES.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(nil)

	senders := map[string]interfaces.MailSender{
		"resend": mockResend,
		"ses":    mockSES,
	}

	orch := NewEmailOrchestrator(senders, mockRepo)
	err := orch.SendEmail(context.Background(), "to@test.com", "subject", "text", "html")

	assert.NoError(t, err)
	mockResend.AssertExpectations(t)
	mockSES.AssertExpectations(t)
}

func TestEmailOrchestrator_SendEmail_SkipsInactiveProvider(t *testing.T) {
	mockRepo := new(mocks.MockEmailConfigRepo)
	mockResend := new(mocks.MockMailSender)
	mockSES := new(mocks.MockMailSender)

	mockRepo.On("GetAll").Return([]model.EmailProviderConfig{
		{ProviderName: "resend", IsActive: false, Priority: 1},
		{ProviderName: "ses", IsActive: true, Priority: 2},
	}, nil)

	mockSES.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(nil)

	senders := map[string]interfaces.MailSender{
		"resend": mockResend,
		"ses":    mockSES,
	}

	orch := NewEmailOrchestrator(senders, mockRepo)
	err := orch.SendEmail(context.Background(), "to@test.com", "subject", "text", "html")

	assert.NoError(t, err)
	mockResend.AssertNotCalled(t, "SendEmail")
	mockSES.AssertExpectations(t)
}

func TestEmailOrchestrator_SendEmail_AllProvidersFail(t *testing.T) {
	mockRepo := new(mocks.MockEmailConfigRepo)
	mockResend := new(mocks.MockMailSender)
	mockSES := new(mocks.MockMailSender)

	mockRepo.On("GetAll").Return([]model.EmailProviderConfig{
		{ProviderName: "resend", IsActive: true, Priority: 1},
		{ProviderName: "ses", IsActive: true, Priority: 2},
	}, nil)

	mockResend.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(errors.New("resend down"))
	mockSES.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(errors.New("ses down"))

	senders := map[string]interfaces.MailSender{
		"resend": mockResend,
		"ses":    mockSES,
	}

	orch := NewEmailOrchestrator(senders, mockRepo)
	err := orch.SendEmail(context.Background(), "to@test.com", "subject", "text", "html")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all email providers failed")
}
```

**Step 4: Run tests**

Run: `cd ScrapJobs && go test ./usecase/... -v -run TestEmailOrchestrator`
Expected: all 4 tests pass

**Step 5: Commit**

```bash
git add repository/mocks/email_config_repo.go repository/mocks/mail_sender.go usecase/email_orchestrator_test.go
git commit -m "test: add EmailOrchestrator unit tests"
```

---

### Task 16: Run Full Test Suite

**Step 1: Backend tests**

Run: `cd ScrapJobs && go test ./...`
Expected: all pass

**Step 2: Frontend build**

Run: `cd FrontScrapJobs && npm run build`
Expected: builds successfully

**Step 3: Frontend lint**

Run: `cd FrontScrapJobs && npm run lint`
Expected: no errors

**Step 4: Final commit if any fixes needed**

```bash
git commit -m "fix: address test/lint issues"
```
