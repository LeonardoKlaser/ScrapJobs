package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"web-scrapper/logging"
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
            -- Card 1: Sites monitorados pelo USUÁRIO
            (SELECT COUNT(*) FROM user_sites us
             JOIN site_scraping_config sc ON us.site_id = sc.id
             WHERE us.user_id = $1 AND sc.is_active = TRUE) AS monitored_urls_count,

            -- Card 2: Vagas novas nos sites do USUÁRIO
            (SELECT COUNT(*) FROM jobs j
             JOIN user_sites us ON j.site_id = us.site_id
             WHERE us.user_id = $1 AND j.created_at >= current_date) AS new_jobs_today_count,

            -- Card 3: Total de alertas enviados para o usuário específico
            (SELECT COUNT(*) FROM job_notifications WHERE user_id = $1) AS alerts_sent_count,

            -- Card 4: Últimas 5 vagas dos sites do USUÁRIO
            (
                SELECT json_agg(j)
                FROM (
                    SELECT j.id, j.title, j.location, j.company, j.job_link, j.requisition_id, j.last_seen_at
                    FROM jobs j
                    JOIN user_sites us ON j.site_id = us.site_id
                    WHERE us.user_id = $1
                    ORDER BY j.last_seen_at DESC
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

func (dr *DashboardRepository) GetLatestJobsPaginated(userID, page, limit, days int, search string, matchedOnly bool) (model.PaginatedJobs, error) {
	var result model.PaginatedJobs
	result.Page = page
	result.Limit = limit

	offset := (page - 1) * limit

	matchedExpr := `
		CASE WHEN us.filters IS NULL OR us.filters::text = '[]' THEN TRUE
		ELSE EXISTS (
			SELECT 1 FROM json_array_elements_text(us.filters) AS f
			WHERE LOWER(j.title) LIKE '%%' || LOWER(f.value) || '%%'
		) END`

	baseFrom := `
		FROM jobs j
		JOIN user_sites us ON j.site_id = us.site_id AND us.user_id = $1
		WHERE 1=1`

	args := []interface{}{userID}
	argIdx := 2

	if days > 0 {
		baseFrom += fmt.Sprintf(" AND j.created_at >= NOW() - INTERVAL '1 day' * $%d", argIdx)
		args = append(args, days)
		argIdx++
	}

	if search != "" {
		baseFrom += fmt.Sprintf(" AND LOWER(j.title) LIKE '%%' || LOWER($%d) || '%%'", argIdx)
		args = append(args, search)
		argIdx++
	}

	if matchedOnly {
		baseFrom += fmt.Sprintf(` AND (%s)`, matchedExpr)
	}

	// Count query
	countQuery := "SELECT COUNT(DISTINCT j.id) " + baseFrom
	err := dr.connection.QueryRow(countQuery, args...).Scan(&result.TotalCount)
	if err != nil {
		return result, fmt.Errorf("erro ao contar vagas: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(
		`SELECT DISTINCT j.id, j.site_id, j.title, j.location, j.company, j.job_link, j.requisition_id, COALESCE(j.description, ''), (%s) AS matched %s ORDER BY j.created_at DESC LIMIT $%d OFFSET $%d`,
		matchedExpr, baseFrom, argIdx, argIdx+1,
	)
	args = append(args, limit, offset)

	rows, err := dr.connection.Query(dataQuery, args...)
	if err != nil {
		return result, fmt.Errorf("erro ao buscar vagas paginadas: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var job model.JobWithMatch
		if err := rows.Scan(&job.ID, &job.SiteID, &job.Title, &job.Location, &job.Company, &job.JobLink, &job.RequisitionID, &job.Description, &job.Matched); err != nil {
			return result, fmt.Errorf("erro ao ler vaga: %w", err)
		}
		result.Jobs = append(result.Jobs, job)
	}

	if result.Jobs == nil {
		result.Jobs = []model.JobWithMatch{}
	}

	return result, rows.Err()
}

func (dr *DashboardRepository) GetAdminDashboardData() (model.AdminDashboardData, error) {
	var data model.AdminDashboardData

	query := `
        SELECT
            COALESCE((SELECT SUM(p.price) FROM users u JOIN plans p ON u.plan_id = p.id WHERE p.price > 0), 0) AS total_revenue,
            (SELECT COUNT(*) FROM users) AS active_users,
            (SELECT COUNT(*) FROM site_scraping_config WHERE is_active = TRUE) AS monitored_sites,
            (SELECT COUNT(*) FROM scraping_errors WHERE created_at >= NOW() - INTERVAL '24 hours') AS scraping_errors
    `

	err := dr.connection.QueryRow(query).Scan(
		&data.TotalRevenue,
		&data.ActiveUsers,
		&data.MonitoredSites,
		&data.ScrapingErrors,
	)
	if err != nil {
		return model.AdminDashboardData{}, fmt.Errorf("erro ao buscar dados do admin dashboard: %w", err)
	}

	errQuery := `SELECT id, site_name, error_message, created_at FROM scraping_errors ORDER BY created_at DESC LIMIT 10`
	rows, err := dr.connection.Query(errQuery)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to query scraping errors")
		return data, nil
	}
	defer rows.Close()

	for rows.Next() {
		var se model.ScrapingError
		if err := rows.Scan(&se.ID, &se.SiteName, &se.ErrorMessage, &se.CreatedAt); err != nil {
			logging.Logger.Warn().Err(err).Msg("failed to scan scraping error row")
			continue
		}
		data.RecentErrors = append(data.RecentErrors, se)
	}

	return data, nil
}

func (dr *DashboardRepository) RecordScrapingError(siteID int, siteName string, errorMessage string, taskID string) error {
	query := `INSERT INTO scraping_errors (site_id, site_name, error_message, task_id) VALUES ($1, $2, $3, $4)`
	_, err := dr.connection.Exec(query, siteID, siteName, errorMessage, taskID)
	if err != nil {
		return fmt.Errorf("erro ao registrar erro de scraping: %w", err)
	}
	return nil
}