package usecase

import (
	"fmt"
	"web-scrapper/interfaces"
	"web-scrapper/model"
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

func (cur *CurriculumUsecase) DeleteCurriculum(userId int, curriculumId int) error {
	count, err := cur.CurriculumRepository.CountCurriculumsByUserID(userId)
	if err != nil {
		return err
	}
	if count <= 1 {
		return fmt.Errorf("não é possível excluir o único currículo")
	}
	return cur.CurriculumRepository.DeleteCurriculum(userId, curriculumId)
}