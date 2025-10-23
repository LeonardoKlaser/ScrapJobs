package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"web-scrapper/model"

	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository (DB *sql.DB) *UserRepository{
	return &UserRepository{
		db: DB,
	}
}

func (usr *UserRepository) CreateUser(user model.User) (error){
	query := `INSERT INTO users (user_name, email, user_password, cellphone, tax, plan) VALUES ($1, $2, $3, $4, $5, $6)`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("error to prepare database query: %w", err)
	}

	err = queryPrepare.QueryRow(user.Name, user.Email, user.Password, user.Cellphone, user.Tax, user.Plan.ID).Scan()
	if(err != nil){
		return fmt.Errorf("error to insert new user in database: %w", err)	
	}

	queryPrepare.Close()

	return nil

}

func (usr *UserRepository) GetUserByEmail(userEmail string) (model.User, error){
	query := `
        SELECT u.id, u.user_name, u.email, u.user_password, u.curriculum_id,
               p.id, p.name, p.price, p.max_sites, p.max_ai_analyses, p.features
        FROM users u
        LEFT JOIN plans p ON u.plan_id = p.id
        WHERE u.email = $1`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return model.User{}, fmt.Errorf("error to prepare database query: %w", err)
	}

	var userToReturn model.User
	var plan model.Plan
	var features pq.StringArray
	err = queryPrepare.QueryRow(userEmail).Scan(
		&userToReturn.Id,
		&userToReturn.Name,
		&userToReturn.Email,
		&userToReturn.Password,
		&userToReturn.CurriculumId,
		&plan.ID,
		&plan.Name,
		&plan.Price,
		&plan.MaxSites,
		&plan.MaxAIAnalyses,
		&features,
	)
	if(err != nil){
		if errors.Is(err, sql.ErrNoRows) {
			
			return model.User{}, nil 
		}
		return model.User{}, fmt.Errorf("error to get user from database: %w", err)	
	}

	plan.Features = features
	userToReturn.Plan = &plan

	queryPrepare.Close()

	return userToReturn, nil

}


func (usr *UserRepository) GetUserById(Id int) (model.User, error){
	query := `
        SELECT u.id, u.user_name, u.email, u.user_password, u.curriculum_id,
               p.id, p.name, p.price, p.max_sites, p.max_ai_analyses, p.features
        FROM users u
        LEFT JOIN plans p ON u.plan_id = p.id
        WHERE u.id = $1`

	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return model.User{}, fmt.Errorf("error to prepare database query: %w", err)
	}

	var userToReturn model.User
	var plan model.Plan
	var features pq.StringArray
	err = queryPrepare.QueryRow(Id).Scan(
		&userToReturn.Id,
		&userToReturn.Name,
		&userToReturn.Email,
		&userToReturn.Password,
		&userToReturn.CurriculumId,
		&plan.ID,
		&plan.Name,
		&plan.Price,
		&plan.MaxSites,
		&plan.MaxAIAnalyses,
		&features,
	)

	if(err != nil){
		return model.User{}, fmt.Errorf("error to get user from database: %w", err)	
	}

	plan.Features = features
	userToReturn.Plan = &plan

	queryPrepare.Close()

	return userToReturn, nil

}


