package repository

import (
	"database/sql"
	"fmt"
	"web-scrapper/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository (DB *sql.DB) UserRepository{
	return UserRepository{
		db: DB,
	}
}

func (usr *UserRepository) CreateUser (user model.User) (model.User, error){
	query := `INSERT INTO users (user_name, email, user_password, curriculum_id) VALUES ($1, $2, $3, $4) RETURNING user_name, email, user_password, curriculum_id`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return model.User{}, fmt.Errorf("error to prepare database query: %w", err)
	}

	var userToReturn model.User
	err = queryPrepare.QueryRow(user.Name, user.Email, user.Password, user.CurriculumId).Scan(
		&userToReturn.Name,
		&userToReturn.Email,
		&userToReturn.Password,
		&userToReturn.CurriculumId,
	)
	if(err != nil){
		return model.User{}, fmt.Errorf("error to insert new user in database: %w", err)	
	}

	queryPrepare.Close()

	return userToReturn, nil

}

func (usr *UserRepository) GetUserByEmail (userEmail string) (model.User, error){
	query := `SELECT user_name, email, user_password, curriculum_id FROM users WHERE email = $1`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return model.User{}, fmt.Errorf("error to prepare database query: %w", err)
	}

	var userToReturn model.User
	err = queryPrepare.QueryRow(userEmail).Scan(
		&userToReturn.Name,
		&userToReturn.Email,
		&userToReturn.Password,
		&userToReturn.CurriculumId,
	)
	if(err != nil){
		return model.User{}, fmt.Errorf("error to get user from database: %w", err)	
	}

	queryPrepare.Close()

	return userToReturn, nil

}


