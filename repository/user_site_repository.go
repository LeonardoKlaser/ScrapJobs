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

func (dep *UserSiteRepository) GetUsersBySiteId(siteId int) ([]model.UserSite, error){
	query := `
        SELECT
            I.id,
            I.name,
            I.email,
            I.curriculum_id,
			U.filters
        FROM
            users I
        INNER JOIN
            user_sites U ON I.id = U.user_id
        WHERE
            U.site_id = $1`
	rows, err := dep.connection.Query(query, siteId)
	if err != nil {
		return []model.UserSite{}, fmt.Errorf("error to querie userdata from database: %w", err)
	}

	defer rows.Close()
	var users []model.UserSite
	var targetWordsJSON []byte

	for rows.Next(){
		var user model.UserSite
		err := rows.Scan(
			&user.UserId,
			&user.Name,
			&user.Email,
			&user.CurriculumId,
			&targetWordsJSON,
		)

		if err != nil {
			if err == sql.ErrNoRows{
				return []model.UserSite{}, fmt.Errorf("error to get user data: %w", err)
			}
			return []model.UserSite{}, err
		}

		if len(targetWordsJSON) > 0 {
			if err := json.Unmarshal(targetWordsJSON, &user.TargetWords); err != nil{
				return []model.UserSite{}, fmt.Errorf("error to get targetWords: %w", err)
			}
		}

		users = append(users, user)
	}

	return users, nil
}