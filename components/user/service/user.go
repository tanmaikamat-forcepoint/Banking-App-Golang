package service

import (
	"bankManagement/repository"
	"bankManagement/utils/log"

	"github.com/jinzhu/gorm"
)

type UserService struct {
	DB         *gorm.DB
	repository *repository.Repository
	log        log.WebLogger
}

func NewUserService(
	DB *gorm.DB,
	repository *repository.Repository,
	log log.WebLogger,
) *UserService {
	return &UserService{
		DB:         DB,
		repository: repository,
		log:        log,
	}
}
