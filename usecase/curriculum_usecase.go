package usecase

import (
	"web-scrapper/model"
	"web-scrapper/repository"
)

type CurriculumUsecase struct {
	CurriculumRepository repository.CurriculumRepository
}

func NewCurriculumUsecase (repository repository.CurriculumRepository) CurriculumUsecase{
	return CurriculumUsecase{
		CurriculumRepository: repository,
	}
}

func (cur *CurriculumUsecase) InsertCurriculum (model.Curriculum) 