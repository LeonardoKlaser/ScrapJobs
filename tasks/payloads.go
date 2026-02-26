package tasks

import "web-scrapper/model"

const (
	TypeScrapSite           = "scrape:site"
	TypeNotifyUser          = "notify:user"
	TypeAnalyzeUserJob      = "analyze:resume"
	TypeCompleteRegistration = "payment:complete_registration"
	TypeMatchUser           = "match:user"
	TypeSendDigest          = "digest:send"
)

type ScrapeSitePayload struct {
	SiteID             int
	SiteScrapingConfig model.SiteScrapingConfig
}

type NotifyUserPayload struct {
	User model.UserSiteCurriculum
	Job *model.Job
	Analysis model.ResumeAnalysis
}

type AnalyzeUserJobPayload struct {
	User model.UserSiteCurriculum
	Job *model.Job
}

type CompleteRegistrationPayload struct {
	PendingRegistrationID string `json:"pending_registration_id"`
	CustomerEmail         string `json:"customer_email"`
}

type MatchUserPayload struct {
	UserID int `json:"user_id"`
}

type SendDigestPayload struct {
	UserID int `json:"user_id"`
}