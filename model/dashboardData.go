package model


type DashboardData struct {
	MonitoredURLsCount int            `json:"monitored_urls_count"`
	NewJobsTodayCount  int            `json:"new_jobs_today_count"`
	AlertsSentCount    int            `json:"alerts_sent_count"`
	LatestJobs         []Job          `json:"latest_jobs"`
	UserMonitoredURLs  []MonitoredURL `json:"user_monitored_urls"`
}

type MonitoredURL struct {
	SiteName string `json:"site_name"`
	BaseURL  string `json:"base_url"`
}

type PaginatedJobs struct {
	Jobs       []Job `json:"jobs"`
	TotalCount int   `json:"total_count"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
}

type AdminDashboardData struct {
	TotalRevenue   float64          `json:"total_revenue"`
	ActiveUsers    int              `json:"active_users"`
	MonitoredSites int              `json:"monitored_sites"`
	ScrapingErrors int              `json:"scraping_errors"`
	RecentErrors   []ScrapingError  `json:"recent_errors"`
}

type ScrapingError struct {
	ID           int    `json:"id"`
	SiteName     string `json:"site_name"`
	ErrorMessage string `json:"error_message"`
	CreatedAt    string `json:"created_at"`
}