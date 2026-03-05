package repository

import (
	"database/sql"
	"web-scrapper/model"
)

type EmailConfigRepo struct {
	db *sql.DB
}

func NewEmailConfigRepo(db *sql.DB) *EmailConfigRepo {
	return &EmailConfigRepo{db: db}
}

func (r *EmailConfigRepo) GetAll() ([]model.EmailProviderConfig, error) {
	rows, err := r.db.Query(`
		SELECT id, provider_name, is_active, priority, updated_at, updated_by
		FROM email_provider_config
		ORDER BY priority ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []model.EmailProviderConfig
	for rows.Next() {
		var c model.EmailProviderConfig
		if err := rows.Scan(&c.ID, &c.ProviderName, &c.IsActive, &c.Priority, &c.UpdatedAt, &c.UpdatedBy); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

func (r *EmailConfigRepo) Update(configs []model.EmailProviderConfig, updatedBy int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE email_provider_config
		SET is_active = $1, priority = $2, updated_at = NOW(), updated_by = $3
		WHERE provider_name = $4
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range configs {
		if _, err := stmt.Exec(c.IsActive, c.Priority, updatedBy, c.ProviderName); err != nil {
			return err
		}
	}

	return tx.Commit()
}
