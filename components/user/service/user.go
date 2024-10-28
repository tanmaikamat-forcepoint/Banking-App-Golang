package service

import (
	"bankManagement/models/user"
	"bankManagement/repository"
	"bankManagement/utils/log"
	"fmt"

	"github.com/jinzhu/gorm"
)

type UserService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.WebLogger
}

func NewUserService(
	DB *gorm.DB,
	repository repository.Repository,
	log log.WebLogger,
) *UserService {
	return &UserService{
		DB:         DB,
		repository: repository,
		log:        log,
	}
}

// CreateSuperAdmin function
func (s *UserService) CreateSuperAdmin(username, password string, name, email string) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// 1. Retrieve the SUPER_ADMIN role or create it if it doesn't exist   [ role table ]
	var superAdminRole user.Role
	if err := s.DB.Where("role_name = ?", "SUPER_ADMIN").First(&superAdminRole).Error; err != nil {
		return fmt.Errorf("SUPER_ADMIN role not found. Plz check seeder should have executed from main.go): %w", err)
	}

	// 2: Check if a user with the SUPER_ADMIN role already exists in db [user table]
	var existingSuperAdmin user.User
	if err := s.DB.Where("role_id = ?", superAdminRole.ID).First(&existingSuperAdmin).Error; err == nil {
		return fmt.Errorf("SuperAdmin already exists")
	} else if !gorm.IsRecordNotFoundError(err) {
		return fmt.Errorf("error checking for existing SuperAdmin: %w", err)
	}

	// Create SuperAdmin user
	superAdmin := &user.User{
		Username: username,
		Password: password,
		Name:     name,
		Email:    email,
		IsActive: true,
		RoleID:   superAdminRole.ID,
	}
	if err := s.repository.Add(uow, superAdmin); err != nil {
		return fmt.Errorf("failed to create SuperAdmin: %w", err)
	}
	uow.Commit()
	return nil

}
