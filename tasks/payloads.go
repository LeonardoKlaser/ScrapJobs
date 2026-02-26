package tasks

import "web-scrapper/model"

const (
	TypeScrapSite           = "scrape:site"
	TypeCompleteRegistration = "payment:complete_registration"
	TypeMatchUser           = "match:user"
	TypeSendDigest          = "digest:send"
)

type ScrapeSitePayload struct {
	SiteID             int
	SiteScrapingConfig model.SiteScrapingConfig
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
