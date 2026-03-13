package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
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
	query := `INSERT INTO users (user_name, email, user_password, cellphone, tax, plan_id, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, user_name, email`
	queryPrepare, err := usr.db.Prepare(query)
	if err != nil {
		return model.User{}, fmt.Errorf("error to prepare database query: %w", err)
	}
	defer queryPrepare.Close()

	var created model.User
	err = queryPrepare.QueryRow(user.Name, user.Email, user.Password, user.Cellphone, user.Tax, user.PlanID, user.ExpiresAt).Scan(
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
               u.expires_at, u.deleted_at, u.weekdays_only,
               p.id, p.name, p.price, p.max_sites, p.max_ai_analyses, p.features
        FROM users u
        LEFT JOIN plans p ON u.plan_id = p.id
        WHERE u.email = $1 AND u.deleted_at IS NULL`
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
	var expiresAt sql.NullTime
	var deletedAt sql.NullTime

	err = queryPrepare.QueryRow(userEmail).Scan(
		&userToReturn.Id,
		&userToReturn.Name,
		&userToReturn.Email,
		&userToReturn.Password,
		&userToReturn.Tax,
		&userToReturn.Cellphone,
		&userToReturn.IsAdmin,
		&userToReturn.CurriculumId,
		&expiresAt,
		&deletedAt,
		&userToReturn.WeekdaysOnly,
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

	if expiresAt.Valid {
		userToReturn.ExpiresAt = &expiresAt.Time
	}
	if deletedAt.Valid {
		userToReturn.DeletedAt = &deletedAt.Time
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
               u.expires_at, u.deleted_at, u.weekdays_only,
               p.id, p.name, p.price, p.max_sites, p.max_ai_analyses, p.features
        FROM users u
        LEFT JOIN plans p ON u.plan_id = p.id
        WHERE u.id = $1 AND u.deleted_at IS NULL`

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
	var expiresAt sql.NullTime
	var deletedAt sql.NullTime

	err = queryPrepare.QueryRow(Id).Scan(
		&userToReturn.Id,
		&userToReturn.Name,
		&userToReturn.Email,
		&userToReturn.Password,
		&userToReturn.Tax,
		&userToReturn.Cellphone,
		&userToReturn.IsAdmin,
		&userToReturn.CurriculumId,
		&expiresAt,
		&deletedAt,
		&userToReturn.WeekdaysOnly,
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

	if expiresAt.Valid {
		userToReturn.ExpiresAt = &expiresAt.Time
	}
	if deletedAt.Valid {
		userToReturn.DeletedAt = &deletedAt.Time
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

func (usr *UserRepository) GetUserMeData(userID int) (model.UserMeData, error) {
	query := `
		SELECT
			u.user_name, u.cellphone, u.tax, u.is_admin, u.expires_at, u.weekdays_only,
			p.id, p.name, p.price, p.max_sites, p.max_ai_analyses, p.features,
			(SELECT COUNT(*) FROM user_sites us
			 JOIN site_scraping_config sc ON us.site_id = sc.id
			 WHERE us.user_id = $1 AND sc.is_active = TRUE) AS monitored_sites_count,
			(SELECT COUNT(*) FROM job_notifications
			 WHERE user_id = $1 AND analysis_result IS NOT NULL
			   AND notified_at >= date_trunc('month', CURRENT_DATE)) AS monthly_analysis_count
		FROM users u
		LEFT JOIN plans p ON u.plan_id = p.id
		WHERE u.id = $1 AND u.deleted_at IS NULL`

	var data model.UserMeData
	var plan model.Plan
	var planID sql.NullInt64
	var planName sql.NullString
	var planPrice sql.NullFloat64
	var planMaxSites sql.NullInt64
	var planMaxAI sql.NullInt64
	var features pq.StringArray
	var expiresAt sql.NullTime

	err := usr.db.QueryRow(query, userID).Scan(
		&data.UserName,
		&data.Cellphone,
		&data.Tax,
		&data.IsAdmin,
		&expiresAt,
		&data.WeekdaysOnly,
		&planID,
		&planName,
		&planPrice,
		&planMaxSites,
		&planMaxAI,
		&features,
		&data.MonitoredSitesCount,
		&data.MonthlyAnalysisCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.UserMeData{}, fmt.Errorf("user %d: %w", userID, model.ErrUserNotFound)
		}
		return model.UserMeData{}, fmt.Errorf("error fetching user me data: %w", err)
	}

	if expiresAt.Valid {
		data.ExpiresAt = &expiresAt.Time
	}

	if planID.Valid {
		plan.ID = int(planID.Int64)
		plan.Name = planName.String
		plan.Price = planPrice.Float64
		plan.MaxSites = int(planMaxSites.Int64)
		plan.MaxAIAnalyses = int(planMaxAI.Int64)
		plan.Features = features
		data.Plan = &plan
	}

	return data, nil
}

func (usr *UserRepository) UpdateUserProfile(userId int, name string, cellphone *string, tax *string) error {
	query := `UPDATE users SET user_name = $1, cellphone = $2, tax = $3 WHERE id = $4 AND deleted_at IS NULL`
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

func (usr *UserRepository) CheckUserExists(email string, tax string) (bool, bool, error) {
	query := `SELECT
		EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL),
		EXISTS(SELECT 1 FROM users WHERE tax = $2 AND tax IS NOT NULL AND tax != '' AND deleted_at IS NULL)`
	var emailExists, taxExists bool
	err := usr.db.QueryRow(query, email, tax).Scan(&emailExists, &taxExists)
	return emailExists, taxExists, err
}

func (usr *UserRepository) UpdateUserPassword(userId int, hashedPassword string) error {
	query := `UPDATE users SET user_password = $1 WHERE id = $2 AND deleted_at IS NULL`
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

func (usr *UserRepository) GetUserBasicInfo(userID int) (string, string, error) {
	query := `SELECT user_name, email FROM users WHERE id = $1 AND deleted_at IS NULL`

	var name, email string
	err := usr.db.QueryRow(query, userID).Scan(&name, &email)
	if err != nil {
		return "", "", fmt.Errorf("error fetching basic info for user %d: %w", userID, err)
	}
	return name, email, nil
}

func (usr *UserRepository) SoftDeleteUser(userId int) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := usr.db.Exec(query, userId)
	if err != nil {
		return fmt.Errorf("error soft-deleting user %d: %w", userId, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user %d not found or already deleted", userId)
	}
	return nil
}

func (usr *UserRepository) UpdateExpiresAt(userId int, expiresAt time.Time) error {
	query := `UPDATE users SET expires_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := usr.db.Exec(query, expiresAt, userId)
	if err != nil {
		return fmt.Errorf("error updating expires_at for user %d: %w", userId, err)
	}
	return nil
}

func (usr *UserRepository) UpdateWeekdaysOnly(userID int, value bool) error {
	_, err := usr.db.Exec("UPDATE users SET weekdays_only = $1 WHERE id = $2", value, userID)
	return err
}
