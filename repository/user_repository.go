package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"web-scrapper/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository (DB *sql.DB) *UserRepository{
	return &UserRepository{
		db: DB,
	}
}

func (usr *UserRepository) CreateUser(user model.User) (model.User, error){
	query := `INSERT INTO users (user_name, email, user_password) VALUES ($1, $2, $3) RETURNING id, user_name, email, user_password`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return model.User{}, fmt.Errorf("error to prepare database query: %w", err)
	}

	var userToReturn model.User
	err = queryPrepare.QueryRow(user.Name, user.Email, user.Password).Scan(
		&userToReturn.Id,
		&userToReturn.Name,
		&userToReturn.Email,
		&userToReturn.Password,
	)
	if(err != nil){
		return model.User{}, fmt.Errorf("error to insert new user in database: %w", err)	
	}

	queryPrepare.Close()

	return userToReturn, nil

}

func (usr *UserRepository) GetUserByEmail(userEmail string) (model.User, error){
	query := `SELECT id, user_name, email, user_password, curriculum_id FROM users WHERE email = $1`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return model.User{}, fmt.Errorf("error to prepare database query: %w", err)
	}

	var userToReturn model.User
	err = queryPrepare.QueryRow(userEmail).Scan(
		&userToReturn.Id,
		&userToReturn.Name,
		&userToReturn.Email,
		&userToReturn.Password,
		&userToReturn.CurriculumId,
	)
	if(err != nil){
		if errors.Is(err, sql.ErrNoRows) {
			
			return model.User{}, nil 
		}
		return model.User{}, fmt.Errorf("error to get user from database: %w", err)	
	}

	queryPrepare.Close()

	return userToReturn, nil

}


func (usr *UserRepository) GetUserById(Id int) (model.User, error){
	query := `SELECT id, user_name, email, user_password, curriculum_id FROM users WHERE id = $1`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return model.User{}, fmt.Errorf("error to prepare database query: %w", err)
	}

	var userToReturn model.User
	err = queryPrepare.QueryRow(Id).Scan(
		&userToReturn.Id,
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


