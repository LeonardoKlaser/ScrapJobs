package mocks

import (
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockCurriculumRepository struct {
	mock.Mock
}

func (m *MockCurriculumRepository) CreateCurriculum(curriculum model.Curriculum) (model.Curriculum, error) {
	args := m.Called(curriculum)
	return args.Get(0).(model.Curriculum), args.Error(1)
}

func (m *MockCurriculumRepository) FindCurriculumByUserID(userId int) ([]model.Curriculum, error) {
	args := m.Called(userId)
	return args.Get(0).([]model.Curriculum), args.Error(1)
}

func (m *MockCurriculumRepository) UpdateCurriculum(curriculum model.Curriculum) (model.Curriculum, error) {
	args := m.Called(curriculum)
	return args.Get(0).(model.Curriculum), args.Error(1)
}

func (m *MockCurriculumRepository) SetActiveCurriculum(userID int, curriculumID int) error {
	args := m.Called(userID, curriculumID)
	return args.Error(0)
}
