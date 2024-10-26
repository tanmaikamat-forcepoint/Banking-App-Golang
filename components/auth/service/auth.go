package service

import (
	"bankManagement/models/user"
	"bankManagement/repository"
	"bankManagement/utils/encrypt"
	"bankManagement/utils/log"
	"errors"

	"github.com/jinzhu/gorm"
)

type AuthService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.WebLogger
}

func NewAuthService(
	DB *gorm.DB,
	repository repository.Repository,
	log log.WebLogger,
) *AuthService {
	return &AuthService{
		DB,
		repository,
		log,
	}
}

func (service *AuthService) LoginRequest(requestedUserCredentials *user.UserLoginParamDTO, tempUser *user.User) error {
	uow := repository.NewUnitOfWork(service.DB)
	defer uow.RollBack()
	err := service.repository.GetByID(uow, &tempUser, "username = ?", requestedUserCredentials.Username)
	if err != nil {
		service.log.Error(err)
		return err
	}
	if tempUser.ID == 0 {
		service.log.Error("User Not Found")
		return errors.New("Invalid User Credentials")
	}
	if !encrypt.CheckHashWithPassword(requestedUserCredentials.Password, tempUser.Password) {
		service.log.Error("Wrong Password")
		return errors.New("Invalid User Credentials")
	}

	uow.Commit()
	return nil
}

func (service *AuthService) CreateNewAdmin(admin *user.User) error {
	uow := repository.NewUnitOfWork(service.DB)
	defer uow.RollBack()
	var tempUser = &user.User{
		Name:     admin.Name,
		Username: admin.Username,
		Password: encrypt.HashPassword(admin.Password),
		Email:    admin.Email,
		IsActive: true,
		RoleID:   uint(encrypt.AdminUserRoleID),
	}
	err := service.repository.Add(uow, &tempUser)
	if err != nil {
		service.log.Error(err)
		return err
	}
	*admin = *tempUser
	uow.Commit()
	return nil
}
