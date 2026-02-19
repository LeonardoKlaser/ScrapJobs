package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"web-scrapper/model"
)

type DashboardRepository struct{
	connection *sql.DB;
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository{
	return &DashboardRepository{
		connection: db,
	}
}

func (dr *DashboardRepository) GetDashboardData(userID int) (model.DashboardData, error) {
	var dashboardData model.DashboardData
	var latestJobsJSON, userURLsJSON []byte

	query := `
        SELECT
            -- Card 1: Total de URLs monitoradas no sistema
            (SELECT COUNT(*) FROM site_scraping_config WHERE is_active = TRUE) AS monitored_urls_count,

            -- Card 2: Vagas novas encontradas hoje
            (SELECT COUNT(*) FROM jobs WHERE created_at >= current_date) AS new_jobs_today_count,

            -- Card 3: Total de alertas enviados para o usuário específico
            (SELECT COUNT(*) FROM job_notifications WHERE user_id = $1) AS alerts_sent_count,

            -- Card 4: Lista das 5 últimas vagas encontradas
            (
                SELECT json_agg(j)
                FROM (
                    SELECT id, title, location, company, job_link, requisition_id, last_seen_at
                    FROM jobs
                    ORDER BY last_seen_at DESC
                    LIMIT 5
                ) j
            ) AS latest_jobs,

            -- Card 5: Lista de URLs que o usuário está monitorando
            (
                SELECT json_agg(s)
                FROM (
                    SELECT sc.site_name, sc.base_url
                    FROM user_sites us
                    JOIN site_scraping_config sc ON us.site_id = sc.id
                    WHERE us.user_id = $1
                ) s
            ) AS user_monitored_urls;
    `

	err := dr.connection.QueryRow(query, userID).Scan(
		&dashboardData.MonitoredURLsCount,
		&dashboardData.NewJobsTodayCount,
		&dashboardData.AlertsSentCount,
		&latestJobsJSON,
		&userURLsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return model.DashboardData{}, nil
		}
		return model.DashboardData{}, fmt.Errorf("erro ao buscar dados do dashboard: %w", err)
	}

	// Faz o unmarshal dos dados JSON retornados pelo banco
	if len(latestJobsJSON) > 0 {
		if err := json.Unmarshal(latestJobsJSON, &dashboardData.LatestJobs); err != nil {
			return model.DashboardData{}, fmt.Errorf("erro ao deserializar últimas vagas: %w", err)
		}
	}

	if len(userURLsJSON) > 0 {
		if err := json.Unmarshal(userURLsJSON, &dashboardData.UserMonitoredURLs); err != nil {
			return model.DashboardData{}, fmt.Errorf("erro ao deserializar URLs do usuário: %w", err)
		}
	}

	return dashboardData, nil
}

func (dr *DashboardRepository) GetAdminDashboardData() (model.AdminDashboardData, error) {
	var data model.AdminDashboardData

	query := `
        SELECT
            COALESCE((SELECT SUM(p.price) FROM users u JOIN plans p ON u.plan_id = p.id WHERE p.price > 0), 0) AS total_revenue,
            (SELECT COUNT(*) FROM users) AS active_users,
            (SELECT COUNT(*) FROM site_scraping_config WHERE is_active = TRUE) AS monitored_sites
    `

	err := dr.connection.QueryRow(query).Scan(
		&data.TotalRevenue,
		&data.ActiveUsers,
		&data.MonitoredSites,
	)
	if err != nil {
		return model.AdminDashboardData{}, fmt.Errorf("erro ao buscar dados do admin dashboard: %w", err)
	}

	data.ScrapingErrors = 0

	return data, nil
}