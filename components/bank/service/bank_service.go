package service

import (
	"bankManagement/models/bank"
	"bankManagement/models/user"
	"bankManagement/repository"
	"bankManagement/utils/encrypt"
	"bankManagement/utils/log"

	"fmt"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type BankService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.WebLogger
}

func NewBankService(
	DB *gorm.DB,
	repository repository.Repository,
	log log.WebLogger,
) *BankService {
	return &BankService{
		DB:         DB,
		repository: repository,
		log:        log,
	}
}

// / create bank
// // func (b *BankService) CreateBank(bankDTO bank.BankDTO) error {
// func (s *BankService) CreateBank(bankDTO bank.BankDTO, bankUserDTO user.UserDTO) error {

// 	fmt.Println("Bank Service called ...")
// 	uow := repository.NewUnitOfWork(b.DB)
// 	defer uow.RollBack()

// 	// Validation input details - name,email
// 	if err := b.ValidateBankDTO(bankDTO); err != nil {
// 		return err
// 	}

// 	// Create Bank
// 	bankEntity := &bank.Bank{
// 		BankName:         bankDTO.BankName,
// 		BankAbbreviation: bankDTO.BankAbbreviation,
// 	}

// 	if err := s.repository.Add(uow, bankEntity); err != nil {
// 		return err
// 	}

// 	hashedPassword, err := HashPassword(bankUserDTO.Password)
// 	if err != nil {
// 		return fmt.Errorf("failed to hash password: %w", err)
// 	}

// 	// Create User for BankUser
// 	bankUserRole, err := s.getRoleByName("BANK_USER")
// 	if err != nil {
// 		return fmt.Errorf("BANK_USER role not found: %w", err)
// 	}

// 	bankUserEntity := &user.User{
// 		Username: bankUserDTO.Username,
// 		Password: hashedPassword, // Ensure password is hashed
// 		Name:     bankUserDTO.Name,
// 		Email:    bankUserDTO.Email,
// 		IsActive: true,
// 		RoleID:   bankUserRole.ID,
// 	}
// 	if err := s.repository.Add(uow, bankUserEntity); err != nil {
// 		return fmt.Errorf("failed to create BankUser: %w", err)
// 	}

// 	// Link BankUser to Bank
// 	bankUserLink := &bank.BankUser{
// 		UserID: bankUserEntity.ID,
// 		BankID: bankEntity.ID,
// 	}
// 	if err := s.repository.Add(uow, bankUserLink); err != nil {
// 		return fmt.Errorf("failed to link BankUser to Bank: %w", err)
// 	}

// 	fmt.Println("Bank Service Finished ...")
// 	uow.Commit()
// 	return nil
// }

func (s *BankService) CreateBank(bankAndUserEntityDTO bank.BankAndUserDTO) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Validate
	if err := s.ValidateBankDTO(bankAndUserEntityDTO); err != nil {
		return err
	}
	//Create Bank
	bankEntity := &bank.Bank{
		BankName:         bankAndUserEntityDTO.BankName,
		BankAbbreviation: bankAndUserEntityDTO.BankAbbreviation,
		IsActive:         true,
	}
	if err := s.repository.Add(uow, bankEntity); err != nil {
		return err
	}

	hashedPassword, err := HashPassword(bankAndUserEntityDTO.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	//Create User for BankUser
	userEntity := &user.User{
		Username: bankAndUserEntityDTO.Username,
		Password: hashedPassword,
		Name:     bankAndUserEntityDTO.BankName, // Using the BankName as User's Name
		Email:    bankAndUserEntityDTO.Email,    // Derived email
		IsActive: true,
		RoleID:   uint(encrypt.BankUserRoleID),
	}
	if err := s.repository.Add(uow, userEntity); err != nil {
		return fmt.Errorf("failed to create BankUser: %w", err)
	}

	//BankUser table  creation -- link BAnk and USer
	bankUserLink := &bank.BankUser{
		UserID: userEntity.ID,
		BankID: bankEntity.ID,
	}
	if err := s.repository.Add(uow, bankUserLink); err != nil {
		return fmt.Errorf("failed to add recored to BankUser Table  - Links UserID to BankID: %w", err)
	}

	uow.Commit()
	return nil
}

// VALIDATION
// checks if ClientName or ClientEmail already exists
func (s *BankService) ValidateBankDTO(bankDTO bank.BankAndUserDTO) error {
	var existingBank bank.Bank
	name := bankDTO.BankName
	abbreviation := bankDTO.BankAbbreviation

	if err := s.DB.Where("bank_name = ? OR bank_abbreviation = ?", name, abbreviation).First(&existingBank).Error; err == nil {
		return fmt.Errorf("bank name or abbreviation already exists")
	}

	var existingUser user.User
	if err := s.DB.Where("username = ?", bankDTO.Username).First(&existingUser).Error; err == nil {
		return fmt.Errorf("username already exists")
	}

	return nil
}

// HASHING PASSWORD
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// GET BANK BY ID
func (s *BankService) GetBankByID(id uint) (*bank.Bank, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var bankEntity bank.Bank
	err := s.repository.GetByID(uow, &bankEntity, id)
	if err != nil {
		return nil, fmt.Errorf("bank not found or deleted: %w", err)
	}

	uow.Commit()
	return &bankEntity, nil
}

// GET ALL BANKS
func (s *BankService) GetAllBanks() ([]bank.Bank, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var banks []bank.Bank
	err := s.repository.GetAll(uow, &banks)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch banks: %w", err)
	}

	uow.Commit()
	return banks, nil
}

// DELETE BANK
func (s *BankService) DeleteBank(id uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	bankEntity := &bank.Bank{Model: gorm.Model{ID: id}}
	// Set IsActive to false and delete
	bankEntity.IsActive = false
	if err := s.DB.Save(bankEntity).Error; err != nil {
		return fmt.Errorf("failed to update IsActive: %w", err)
	}
	err := s.repository.DeleteById(uow, bankEntity, id)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}
