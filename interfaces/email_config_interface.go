package interfaces

import "web-scrapper/model"

type EmailConfigRepository interface {
	GetAll() ([]model.EmailProviderConfig, error)
	Update(configs []model.EmailProviderConfig, updatedBy int) error
}
