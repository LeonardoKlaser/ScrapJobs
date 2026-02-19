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
	query := `INSERT INTO curriculum (user_id, experience, education, skills, languages, summary, title) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING user_id , experience, education, skills, languages, summary;`
	queryPrepare, err := cur.connection.Prepare(query)
	if err != nil {
		return model.Curriculum{}, err
	}
	defer queryPrepare.Close()

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

	err = queryPrepare.QueryRow(curriculum.UserID, experiences, educations, curriculum.Skills, curriculum.Languages, curriculum.Summary, curriculum.Title).Scan(
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

	return curriculumCreated, nil
}


func (cur *CurriculumRepository) FindCurriculumByUserID(userID int) ([]model.Curriculum, error) {
	var curriculumList []model.Curriculum

	query := `SELECT id, title, is_active, experience, education, skills, languages, summary FROM curriculum WHERE user_id = $1`
	rows, err := cur.connection.Query(query, userID)

	if err != nil{
		return curriculumList, err
	}
	defer rows.Close()

	for rows.Next() {
		var curriculum model.Curriculum
		var experienceJSON, educationJSON []byte
		err = rows.Scan(
			&curriculum.Id,
			&curriculum.Title,
			&curriculum.IsActive,
			&experienceJSON,
			&educationJSON,
			&curriculum.Skills,
			&curriculum.Languages,
			&curriculum.Summary,
		)

		if err != nil {
			return []model.Curriculum{}, err
		}

		if len(educationJSON) > 0 {
			if err := json.Unmarshal(educationJSON, &curriculum.Educations); err != nil {
				return []model.Curriculum{}, fmt.Errorf("error to get education informations: %w", err)
			}
		}

		if len(experienceJSON) > 0 {
			if err := json.Unmarshal(experienceJSON, &curriculum.Experiences); err != nil {
				return []model.Curriculum{}, fmt.Errorf("error to get experiences informations: %w", err)
			}
		}

		curriculumList = append(curriculumList, curriculum)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return curriculumList, nil

}

func (cur *CurriculumRepository) UpdateCurriculum(curriculum model.Curriculum) (model.Curriculum, error) {
	query := `UPDATE curriculum SET title = $1, experience = $2, education = $3, skills = $4, languages = $5, summary = $6 WHERE id = $7 AND user_id = $8 RETURNING id;`

	experiences, err := json.Marshal(curriculum.Experiences)
	if err != nil {
		return model.Curriculum{}, fmt.Errorf("error to serialize experiences: %v", err)
	}

	educations, err := json.Marshal(curriculum.Educations)
	if err != nil {
		return model.Curriculum{}, fmt.Errorf("error to serialize educations: %v", err)
	}

	_, err = cur.connection.Exec(query, curriculum.Title, experiences, educations, curriculum.Skills, curriculum.Languages, curriculum.Summary, curriculum.Id, curriculum.UserID)
	if err != nil {
		return model.Curriculum{}, fmt.Errorf("error to update curriculum: %w", err)
	}

	return curriculum, nil
}

// SET A CURRICULUM AS ACTIVE
func (cur *CurriculumRepository) SetActiveCurriculum(userID int, curriculumID int) error {
	tx, err := cur.connection.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	_, err = tx.Exec("UPDATE curriculum SET is_active = FALSE WHERE user_id = $1", userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deactivating other curriculums: %w", err)
	}

	_, err = tx.Exec("UPDATE curriculum SET is_active = TRUE WHERE id = $1 AND user_id = $2", curriculumID, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error activating curriculum: %w", err)
	}

	return tx.Commit()
}


