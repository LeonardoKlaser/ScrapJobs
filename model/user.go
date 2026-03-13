package model

import (
	"errors"
	"time"
)

// ErrUserNotFound is returned when a user is not found (e.g. soft-deleted or invalid ID).
var ErrUserNotFound = errors.New("user not found")

// UserMeData holds data fetched from the database for /api/me.
// Only id, email, and is_admin come from JWT claims; everything else is fresh from DB.
type UserMeData struct {
	UserName             string     `json:"user_name"`
	Cellphone            *string    `json:"cellphone,omitempty"`
	Tax                  *string    `json:"tax,omitempty"`
	ExpiresAt            *time.Time `json:"expires_at,omitempty"`
	WeekdaysOnly         bool       `json:"weekdays_only"`
	Plan                 *Plan      `json:"plan,omitempty"`
	MonitoredSitesCount  int        `json:"monitored_sites_count"`
	MonthlyAnalysisCount int        `json:"monthly_analysis_count"`
}

type User struct {
	Id           int        `json:"id"`
	Name         string     `json:"user_name"`
	Email        string     `json:"email"`
	Password     string     `json:"-"`
	Tax          *string    `json:"tax,omitempty" db:"tax"`
	Cellphone    *string    `json:"cellphone,omitempty" db:"cellphone"`
	IsAdmin      bool       `json:"is_admin"`
	CurriculumId *int       `json:"curriculum_id,omitempty"`
	PlanID       *int       `json:"plan_id,omitempty"`
	Plan         *Plan      `json:"plan,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	DeletedAt    *time.Time `json:"-"`
	WeekdaysOnly bool       `json:"weekdays_only"`
}
