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

func (m *MockCurriculumRepository) DeleteCurriculum(userId int, curriculumId int) error {
	args := m.Called(userId, curriculumId)
	return args.Error(0)
}

func (m *MockCurriculumRepository) CountCurriculumsByUserID(userId int) (int, error) {
	args := m.Called(userId)
	return args.Int(0), args.Error(1)
}

func (m *MockCurriculumRepository) DeleteCurriculumIfNotLast(userId int, curriculumId int) error {
	args := m.Called(userId, curriculumId)
	return args.Error(0)
}
