package model

type Experience struct{
	Company string `json:"company"`
	Title string `json:"title"`
	Description string `json:"description"`
}

type Education struct {
	Institution string `json:"institution"`
	Degree string `json:"degree"`
	Year string `json:"year"`
}

type Curriculum struct {
	Id int `json:"id"`
	Title string `json:"title"`
	IsActive bool `json:"is_active"`
	UserID int `json:"user_id"`
	Experiences []Experience `json:"experiences"`
	Skills string `json:"skills"`
	Summary string `json:"summary"`
	Educations []Education `json:"educations"`
	Languages string `json:"languages"`
}