package repository

import (
	"database/sql"
	"web-scrapper/model"
	"encoding/json"
	"fmt"
)

type CurriculumRepository struct {
	connection *sql.DB
}

func NewCurriculumRepository(db *sql.DB) *CurriculumRepository {
	return &CurriculumRepository{
		connection: db,
	}
}

// INSERT A NEW CURRICULUM INTO THE DATABASE
func (cur *CurriculumRepository) CreateCurriculum(curriculum model.Curriculum) (model.Curriculum, error) {
	query := `INSERT INTO curriculum (user_id, experience, education, skills, languages, summary) VALUES ($1, $2, $3, $4, $5, $6) RETURNING user_id , experience, education, skills, languages, summary;`
	queryPrepare, err := cur.connection.Prepare(query)

	if err != nil {
		return model.Curriculum{}, err
	}
	var curriculumCreated model.Curriculum

	var experienceJSON, educationJSON []byte
	experiences, err := json.Marshal(curriculum.Experiences)
	if err != nil {
		return curriculumCreated, fmt.Errorf("error to serialize experiences: %v", err)
	}

	educations, err := json.Marshal(curriculum.Educations)
	if err != nil {
		return curriculumCreated, fmt.Errorf("error to serialize educations: %v", err)
	}

	err = queryPrepare.QueryRow(curriculum.UserID, experiences, educations, curriculum.Skills, curriculum.Languages, curriculum.Summary).Scan(
		&curriculumCreated.UserID,
		&experienceJSON,
		&educationJSON,
		&curriculumCreated.Skills,
		&curriculumCreated.Languages,
		&curriculumCreated.Summary,
	)
	if err != nil{
		if err == sql.ErrNoRows{
			return model.Curriculum{}, fmt.Errorf("error to insert new curriculum: %w", err)
		}
		return model.Curriculum{}, err
	}

	if len(educationJSON) > 0 {
		if err := json.Unmarshal(educationJSON, &curriculumCreated.Educations); err != nil {
			return model.Curriculum{}, fmt.Errorf("error to get education informations: %w", err )
		}
	}

	if len(experienceJSON) > 0 {
		if err := json.Unmarshal(experienceJSON, &curriculumCreated.Experiences); err != nil {
			return model.Curriculum{}, fmt.Errorf("error to get experiences informations: %w", err )
		}
	}

	queryPrepare.Close()
	return curriculumCreated, nil
	
}


func (cur *CurriculumRepository) FindCurriculumByUserID(userID int) (model.Curriculum, error) {
	query := `SELECT experience, education, skills, languages, summary FROM curriculum WHERE user_id = $1`
	queryPrepare, err := cur.connection.Prepare(query)

	var curriculum model.Curriculum
	var experienceJSON, educationJSON []byte
	
	if err != nil{
		return curriculum, err
	}

	err = queryPrepare.QueryRow(userID).Scan(
		&experienceJSON,
		&educationJSON,
		&curriculum.Skills,
		&curriculum.Languages,
		&curriculum.Summary,
	)
	if err != nil{
		if err == sql.ErrNoRows{
			return model.Curriculum{}, fmt.Errorf("curriculum for this user_id: %d not found: %w", userID, err)
		}
		return model.Curriculum{}, err
	}

	if len(educationJSON) > 0 {
		if err := json.Unmarshal(educationJSON, &curriculum.Educations); err != nil {
			return model.Curriculum{}, fmt.Errorf("error to get education informations: %w", err )
		}
	}

	if len(experienceJSON) > 0 {
		if err := json.Unmarshal(experienceJSON, &curriculum.Experiences); err != nil {
			return model.Curriculum{}, fmt.Errorf("error to get experiences informations: %w", err )
		}
	}

	queryPrepare.Close()
	return curriculum, nil

}


