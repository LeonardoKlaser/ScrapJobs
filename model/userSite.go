package model
type UserSite struct{
	UserId int `json:"user_id"`
	Name string `json:"user_name"`
	Email string `json:"email"`
	CurriculumId *int `json:"curriculum_id,omitempty"`
	TargetWords []string `db:"target_words"`
}