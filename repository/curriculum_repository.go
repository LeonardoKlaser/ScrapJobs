package repository

import (
	"database/sql"
	"web-scrapper/model"
)

type CurriculumRepository struct {
	connection *sql.DB
}

func NewCurriculumRepository(db *sql.DB) CurriculumRepository {
	return CurriculumRepository{
		connection: db,
	}
}

// INSERT A NEW CURRICULUM INTO THE DATABASE
func (cur *CurriculumRepository) CreateCurriculum(curriculum model.Curriculum) (model.Curriculum, error) {
	query := `INSERT INTO curriculums (user_id, experience, education, skills, languages, summary) VALUES ($1, $2, $3, $4, $5, $6) RETURNING experience, education, skills, languages, summary;`
	queryPrepare, err := cur.connection.Prepare(query)

	if err != nil {
		return model.Curriculum{}, err
	}
	var curriculumCreated model.Curriculum

	err = queryPrepare.QueryRow(curriculum.UserID, curriculum.Experiences, curriculum.Educations, curriculum.Skills, curriculum.Languages, curriculum.Summary).Scan(&curriculumCreated)
	if err != nil {
		return model.Curriculum{}, err
	}

	queryPrepare.Close()
	return curriculumCreated, nil
}


func (cur *CurriculumRepository) FindCurriculumByUserID(userID int) (model.Curriculum, error) {
	query := `SELECT experience, education, skills, languages, summary FROM curriculum WHERE user_id = $1`
	queryPrepare, err := cur.connection.Prepare(query)

	var curriculum model.Curriculum
	
	if err != nil{
		return curriculum, err
	}

	err = queryPrepare.QueryRow(userID).Scan(&curriculum)
	if err != nil{
		return curriculum, err
	}

	return curriculum, nil

}


