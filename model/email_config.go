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
