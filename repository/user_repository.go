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

func NewUserRepository(DB *sql.DB) *UserRepository {
	return &UserRepository{
		db: DB,
	}
}

func (usr *UserRepository) CreateUser(user model.User) (model.User, error) {
	query := `INSERT INTO users (user_name, email, user_password, cellphone, tax, plan_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, user_name, email`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return model.User{}, fmt.Errorf("error to prepare database query: %w", err)
	}
	defer queryPrepare.Close()

	var created model.User
	err = queryPrepare.QueryRow(user.Name, user.Email, user.Password, user.Cellphone, user.Tax, user.PlanID).Scan(
		&created.Id,
		&created.Name,
		&created.Email,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("error to insert new user in database: %w", err)
	}

	created.PlanID = user.PlanID
	return created, nil
}

func (usr *UserRepository) GetUserByEmail(userEmail string) (model.User, error) {
	query := `
        SELECT u.id, u.user_name, u.email, u.user_password, u.tax, u.cellphone, u.is_admin, u.curriculum_id,
               p.id, p.name, p.price, p.max_sites, p.max_ai_analyses, p.features
        FROM users u
        LEFT JOIN plans p ON u.plan_id = p.id
        WHERE u.email = $1`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return model.User{}, fmt.Errorf("error to prepare database query: %w", err)
	}
	defer queryPrepare.Close()

	var userToReturn model.User
	var plan model.Plan
	var planID sql.NullInt64
	var planName sql.NullString
	var planPrice sql.NullFloat64
	var planMaxSites sql.NullInt64
	var planMaxAI sql.NullInt64
	var features pq.StringArray

	err = queryPrepare.QueryRow(userEmail).Scan(
		&userToReturn.Id,
		&userToReturn.Name,
		&userToReturn.Email,
		&userToReturn.Password,
		&userToReturn.Tax,
		&userToReturn.Cellphone,
		&userToReturn.IsAdmin,
		&userToReturn.CurriculumId,
		&planID,
		&planName,
		&planPrice,
		&planMaxSites,
		&planMaxAI,
		&features,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, nil
		}
		return model.User{}, fmt.Errorf("error to get user from database: %w", err)
	}

	if planID.Valid {
		id := int(planID.Int64)
		plan.ID = id
		plan.Name = planName.String
		plan.Price = planPrice.Float64
		plan.MaxSites = int(planMaxSites.Int64)
		plan.MaxAIAnalyses = int(planMaxAI.Int64)
		plan.Features = features
		userToReturn.Plan = &plan
		userToReturn.PlanID = &id
	}

	return userToReturn, nil
}

func (usr *UserRepository) GetUserById(Id int) (model.User, error) {
	query := `
        SELECT u.id, u.user_name, u.email, u.user_password, u.tax, u.cellphone, u.is_admin, u.curriculum_id,
               p.id, p.name, p.price, p.max_sites, p.max_ai_analyses, p.features
        FROM users u
        LEFT JOIN plans p ON u.plan_id = p.id
        WHERE u.id = $1`

	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return model.User{}, fmt.Errorf("error to prepare database query: %w", err)
	}
	defer queryPrepare.Close()

	var userToReturn model.User
	var plan model.Plan
	var planID sql.NullInt64
	var planName sql.NullString
	var planPrice sql.NullFloat64
	var planMaxSites sql.NullInt64
	var planMaxAI sql.NullInt64
	var features pq.StringArray

	err = queryPrepare.QueryRow(Id).Scan(
		&userToReturn.Id,
		&userToReturn.Name,
		&userToReturn.Email,
		&userToReturn.Password,
		&userToReturn.Tax,
		&userToReturn.Cellphone,
		&userToReturn.IsAdmin,
		&userToReturn.CurriculumId,
		&planID,
		&planName,
		&planPrice,
		&planMaxSites,
		&planMaxAI,
		&features,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("error to get user from database: %w", err)
	}

	if planID.Valid {
		id := int(planID.Int64)
		plan.ID = id
		plan.Name = planName.String
		plan.Price = planPrice.Float64
		plan.MaxSites = int(planMaxSites.Int64)
		plan.MaxAIAnalyses = int(planMaxAI.Int64)
		plan.Features = features
		userToReturn.Plan = &plan
		userToReturn.PlanID = &id
	}

	return userToReturn, nil
}

func (usr *UserRepository) UpdateUserProfile(userId int, name string, cellphone *string, tax *string) error {
	query := `UPDATE users SET user_name = $1, cellphone = $2, tax = $3 WHERE id = $4`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("error to prepare database query: %w", err)
	}
	defer queryPrepare.Close()

	_, err = queryPrepare.Exec(name, cellphone, tax, userId)
	if err != nil {
		return fmt.Errorf("error to update user profile: %w", err)
	}

	return nil
}

func (usr *UserRepository) UpdateUserPassword(userId int, hashedPassword string) error {
	query := `UPDATE users SET user_password = $1 WHERE id = $2`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("error to prepare database query: %w", err)
	}
	defer queryPrepare.Close()

	_, err = queryPrepare.Exec(hashedPassword, userId)
	if err != nil {
		return fmt.Errorf("error to update user password: %w", err)
	}

	return nil
}
