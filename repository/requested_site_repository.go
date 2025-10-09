package repository

import (
	"database/sql"
	"fmt"
)

type RequestedSiteRepository struct {
	connection *sql.DB
}

func NewRequestedSiteRepository(db *sql.DB) *RequestedSiteRepository {
	return &RequestedSiteRepository{
		connection: db,
	}
}

func (r *RequestedSiteRepository) Create(userID int, url string) error {
	query := `INSERT INTO requested_sites (user_id, url) VALUES ($1, $2)`
	_, err := r.connection.Exec(query, userID, url)
	if err != nil {
		return fmt.Errorf("erro ao inserir nova solicitação de site: %w", err)
	}
	return nil
}