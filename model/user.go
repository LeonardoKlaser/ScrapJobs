package model

type User struct {
	Id           int     `json:"id"`
	Name         string  `json:"user_name"`
	Email        string  `json:"email"`
	Password     string  `json:"-"`
	Tax          *string `json:"tax,omitempty" db:"tax"`
	Cellphone    *string `json:"cellphone,omitempty" db:"cellphone"`
	IsAdmin      bool    `json:"is_admin"`
	CurriculumId *int    `json:"curriculum_id,omitempty"`
	PlanID       *int    `json:"plan_id,omitempty"`
	Plan         *Plan   `json:"plan,omitempty"`
}
