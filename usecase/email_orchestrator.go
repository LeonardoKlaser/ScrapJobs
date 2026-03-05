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
