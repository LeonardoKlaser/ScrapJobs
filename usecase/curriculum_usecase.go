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

func (cur *CurriculumUsecase) CreateCurriculum(curriculum model.Curriculum) (model.Curriculum, error){
	res, err := cur.CurriculumRepository.CreateCurriculum(curriculum)
	if err != nil {
		return model.Curriculum{}, err
	}

	return res, nil
}

func (cur *CurriculumUsecase) GetCurriculumByUserId(userId int) (model.Curriculum, error){
	res, err := cur.CurriculumRepository.FindCurriculumByUserID(userId)
	if err != nil {
		return model.Curriculum{}, err
	}

	return res, nil
}