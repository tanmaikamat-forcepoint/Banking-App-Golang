package service

import (
	"bankManagement/repository"
	"bankManagement/utils/log"

	"github.com/jinzhu/gorm"
)

type TestService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.WebLogger
}

func NewTestService(
	DB *gorm.DB,
	repository repository.Repository,
	log log.WebLogger,
) *TestService {
	return &TestService{
		DB:         DB,
		repository: repository,
		log:        log,
	}
}
