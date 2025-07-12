package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"web-scrapper/model"
)

type UserSiteRepository struct {
	connection *sql.DB
}

func NewUserSiteRepository(db *sql.DB) UserSiteRepository{
	return UserSiteRepository{
		connection: db,
	}
}

func (dep *UserSiteRepository) GetUsersBySiteId(siteId int) ([]model.UserSiteCurriculum, error){
	query := `
        SELECT
            I.id,
            I.user_name,
            I.email,
			c.experience, 
			c.education, 
			c.skills, 
			c.languages, 
			c.summary,
			U.filters
        FROM
            users I
        INNER JOIN
            user_sites U ON I.id = U.user_id
		LEFT JOIN
			curriculum c ON I.id = c.user_id
        WHERE
            U.site_id = $1`
	rows, err := dep.connection.Query(query, siteId)
	if err != nil {
		return []model.UserSiteCurriculum{}, fmt.Errorf("error to querie userdata from database: %w", err)
	}

	defer rows.Close()
	var users []model.UserSiteCurriculum
	var targetWordsJSON sql.NullString

	for rows.Next(){
		var user model.UserSiteCurriculum
		var skills, languages, summary sql.NullString
		var experienceJSON, educationJSON sql.NullString
		err := rows.Scan(
			&user.UserId,
			&user.Name,
			&user.Email,
			&experienceJSON,
            &educationJSON,
            &skills,
            &languages,
            &summary,
            &targetWordsJSON,
		)

		if err != nil {
			if err == sql.ErrNoRows{
				return []model.UserSiteCurriculum{}, fmt.Errorf("error to get user data: %w", err)
			}
			return []model.UserSiteCurriculum{}, err
		}

		if targetWordsJSON.Valid {
            if err := json.Unmarshal([]byte(targetWordsJSON.String), &user.TargetWords); err != nil {
                return nil, fmt.Errorf("error unmarshalling targetWords: %w", err)
            }
        }

		if skills.Valid { // Checa se um dos campos obrigatórios do currículo não é nulo
            curriculum := &model.Curriculum{}
            curriculum.Skills = skills.String
            curriculum.Languages = languages.String
            curriculum.Summary = summary.String

            if educationJSON.Valid {
                if err := json.Unmarshal([]byte(educationJSON.String), &curriculum.Educations); err != nil {
                    return nil, fmt.Errorf("error unmarshalling education: %w", err)
                }
            }
            if experienceJSON.Valid {
                if err := json.Unmarshal([]byte(experienceJSON.String), &curriculum.Experiences); err != nil {
                    return nil, fmt.Errorf("error unmarshalling experience: %w", err)
                }
            }
            user.Curriculum = curriculum
        }


		users = append(users, user)
	}

	return users, nil
}