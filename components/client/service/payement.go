package service

import (
	"bankManagement/repository"
	"bankManagement/utils/log"

	"github.com/jinzhu/gorm"
)

type PaymentService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.WebLogger
}

func NewPaymentService(
	DB *gorm.DB,
	repository repository.Repository,
	log log.WebLogger,
) *PaymentService {
	return &PaymentService{
		DB,
		repository,
		log,
	}
}
