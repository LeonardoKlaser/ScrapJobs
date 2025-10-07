package model

import "github.com/lib/pq"

type Plan struct {
	ID            int            `json:"id"`
	Name          string         `json:"name"`
	Price         float64        `json:"price"`
	MaxSites      int            `json:"max_sites"`
	MaxAIAnalyses int            `json:"max_ai_analyses"`
	Features      pq.StringArray `json:"features" gorm:"type:text[]"`
}