package interfaces

import (
	"time"
	"web-scrapper/model"
)

type PasswordResetRepositoryInterface interface {
	CreateToken(userID int, ttl time.Duration) (string, error)
	FindValidToken(token string) (*model.PasswordResetToken, error)
	MarkUsed(tokenID int) error
	DeleteExpiredTokens() (int64, error)
}
