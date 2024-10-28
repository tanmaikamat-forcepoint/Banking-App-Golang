package service

import (
	"bankManagement/models/bank"
	"bankManagement/models/client"
	"bankManagement/models/user"
	"bankManagement/repository"
	"bankManagement/utils/encrypt"
	"bankManagement/utils/log"
	"errors"
	"time"

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

func (service *AuthService) LoginRequest(requestedUserCredentials *user.UserLoginParamDTO, tempUser *user.User, permissions *user.UserPermissionDTO, loginSessionId *uint) error {
	uow := repository.NewUnitOfWork(service.DB)
	defer uow.RollBack()
	err := service.repository.GetFirstWhere(uow, &tempUser, "username = ?", requestedUserCredentials.Username)
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
	if tempUser.RoleID == uint(encrypt.AdminUserRoleID) {
		permissions.IsSuperAdmin = true
	} else if tempUser.RoleID == uint(encrypt.BankUserRoleID) {
		tempBankUser := bank.BankUser{}
		err = service.repository.GetFirstWhere(uow, &tempBankUser, "user_id=?", tempUser.ID)
		if err != nil || tempBankUser.UserID == 0 {
			service.log.Error(err)
			return errors.New("Bank User donot have access to Any Bank")
		}
		permissions.BankId = tempBankUser.BankID
	} else if tempUser.RoleID == uint(encrypt.ClientUserRoleID) {
		tempClientUser := client.ClientUser{}
		err = service.repository.GetFirstWhere(uow, &tempClientUser, "user_id=?", tempUser.ID)
		if err != nil || tempClientUser.UserID == 0 {
			service.log.Error(err)
			return errors.New("Client User donot have access to Any Client")
		}
		permissions.ClientId = tempClientUser.ClientID
	}

	tempSession := user.UserLoginInfo{
		UserId:    tempUser.ID,
		UserName:  tempUser.Username,
		IsActive:  true,
		RoleID:    tempUser.RoleID,
		LoginTime: time.Now(),
	}
	err = service.repository.Add(uow, &tempSession)
	if err != nil {
		return err
	}
	service.log.Info(tempSession.ID)
	*loginSessionId = tempSession.ID

	service.log.Info(*loginSessionId)
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
