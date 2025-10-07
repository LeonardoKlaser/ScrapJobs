package model
type User struct{
	Id int `json:"id"`
	Name string `json:"user_name"`
	Email string `json:"email"`
	Password string `json:"user_password"`
	CurriculumId *int `json:"curriculum_id,omitempty"`
	Plan         *Plan   `json:"plan,omitempty"`
}