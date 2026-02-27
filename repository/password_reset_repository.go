package repository

import (
	"database/sql"
	"fmt"
	"time"
	"web-scrapper/model"
)

type PasswordResetRepository struct {
	db *sql.DB
}

func NewPasswordResetRepository(db *sql.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) CreateToken(userID int, ttl time.Duration) (string, error) {
	query := `INSERT INTO password_reset_tokens (user_id, expires_at) VALUES ($1, NOW() + $2::interval) RETURNING token`
	var token string
	err := r.db.QueryRow(query, userID, fmt.Sprintf("%d seconds", int(ttl.Seconds()))).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("error creating password reset token: %w", err)
	}
	return token, nil
}

func (r *PasswordResetRepository) FindValidToken(token string) (*model.PasswordResetToken, error) {
	query := `SELECT id, user_id, token, expires_at, used_at, created_at FROM password_reset_tokens WHERE token = $1 AND expires_at > NOW() AND used_at IS NULL`
	var t model.PasswordResetToken
	err := r.db.QueryRow(query, token).Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding password reset token: %w", err)
	}
	return &t, nil
}

func (r *PasswordResetRepository) MarkUsed(tokenID int) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, tokenID)
	if err != nil {
		return fmt.Errorf("error marking token as used: %w", err)
	}
	return nil
}
