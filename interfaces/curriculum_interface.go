package interfaces

import(
	"web-scrapper/model"
)

type CurriculumRepositoryInterface interface {
	CreateCurriculum(curriculum model.Curriculum) (model.Curriculum, error)
	FindCurriculumByUserID(userId int) ([]model.Curriculum, error)
	UpdateCurriculum(curriculum model.Curriculum) (model.Curriculum, error)
	DeleteCurriculum(userId int, curriculumId int) error
	CountCurriculumsByUserID(userId int) (int, error)
}