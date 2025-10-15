package usecase

import (
	"web-scrapper/model"
	"web-scrapper/interfaces"
)


type CurriculumUsecase struct {
	CurriculumRepository interfaces.CurriculumRepositoryInterface
	
}

func NewCurriculumUsecase (repository interfaces.CurriculumRepositoryInterface) *CurriculumUsecase{
	return &CurriculumUsecase{
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

func (cur *CurriculumUsecase) GetCurriculumByUserId(userId int) ([]model.Curriculum, error){
	res, err := cur.CurriculumRepository.FindCurriculumByUserID(userId)
	if err != nil {
		return []model.Curriculum{}, err
	}

	return res, nil
}

func (cur *CurriculumUsecase) UpdateCurriculum(curriculum model.Curriculum) (model.Curriculum, error){
	res, err := cur.CurriculumRepository.UpdateCurriculum(curriculum)
	if err != nil {
		return model.Curriculum{}, err
	}

	return res, nil
}

func (cur *CurriculumUsecase) SetActiveCurriculum(userID int, curriculumID int) error {
	err := cur.CurriculumRepository.SetActiveCurriculum(userID, curriculumID)
	if err != nil {
		return err
	}
	return nil
}