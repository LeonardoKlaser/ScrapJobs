package interfaces

import(
	"web-scrapper/model"
)

type CurriculumRepositoryInterface interface {
	CreateCurriculum(curriculum model.Curriculum) (model.Curriculum, error)
	FindCurriculumByUserID(userId int) (model.Curriculum, error)
}