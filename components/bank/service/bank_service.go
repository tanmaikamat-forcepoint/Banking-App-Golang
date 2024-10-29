package service

import (
	"bankManagement/models/bank"
	"bankManagement/models/client"
	"bankManagement/models/user"
	"bankManagement/repository"
	"bankManagement/utils/encrypt"
	"bankManagement/utils/log"
	"strings"

	"fmt"

	"github.com/jinzhu/gorm"
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
func (s *BankService) CreateBank(bankAndUserEntityDTO bank.BankAndUserDTO) error {

	fmt.Println("Bank Service called ...")

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

	hashedPassword := encrypt.HashPassword(bankAndUserEntityDTO.Password)

	//Create User for BankUser
	userEntity := &user.User{
		Username: bankAndUserEntityDTO.Username,
		Password: hashedPassword,
		Name:     bankAndUserEntityDTO.BankName, //  BankUSer Name == Name
		Email:    bankAndUserEntityDTO.Email,    //  BankUserEmail == Email
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

func (s *BankService) DeleteBank(bankID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// checking if clients exist for this bank
	var clients []client.Client
	if err := s.repository.GetAll(uow, &clients, s.repository.Filter("bank_id = ?", bankID)); err != nil {
		return fmt.Errorf("failed to check associated clients: %w", err)
	}
	if len(clients) > 0 {
		return fmt.Errorf("cannot delete bank with ID %d: associated clients exist", bankID)
	} else {
		fmt.Println("No associated found.")
	}

	// Finding and delete associated BankUser (created during Bank Creation)
	var bankUser bank.BankUser
	if err := s.repository.GetFirstWhere(uow, &bankUser, "bank_id=?", bankID); err == nil { //Filter("bank_id = ?", bankID)); err == nil {
		if err := s.repository.DeleteById(uow, &bankUser, bankUser.UserID); err != nil {
			return fmt.Errorf("failed to delete associated BankUser: %w", err)
		}
	} else if !gorm.IsRecordNotFoundError(err) {
		return fmt.Errorf("error checking for BankUser: %w", err)
	}

	// Delete the Bank itself
	bankEntity := bank.Bank{Model: gorm.Model{ID: bankID}}
	if err := s.repository.DeleteById(uow, &bankEntity, bankID); err != nil {
		return fmt.Errorf("failed to delete bank with ID %d: %w", bankID, err)
	}

	uow.Commit()
	return nil
}

// // UPDATE BANK (and BankUser - related details also)
func (s *BankService) UpdateBank(bankID uint, bankDTO bank.BankAndUserDTO) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// checking if bank to be updated exists by id
	var existingBank bank.Bank
	if err := s.repository.GetByID(uow, &existingBank, bankID); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("bank with ID %d not found", bankID)
		}
		return err
	}

	// Validate i/p json data
	if err := s.ValidateBankDTO(bankDTO); err != nil {
		return err
	}

	// Updating fields if BankUser wants that, otherwise I am keeping Old Values
	if bankDTO.BankName != "" {
		existingBank.BankName = bankDTO.BankName
	}
	if bankDTO.BankAbbreviation != "" {
		existingBank.BankAbbreviation = bankDTO.BankAbbreviation
	}

	if err := s.repository.Update(uow, &existingBank); err != nil {
		return fmt.Errorf("failed to update bank: %w", err)
	}

	// ----------------  ASSOCIATED bANKUser deletion -----------------

	// // finding and updating associated BankUser
	/*
		bankUser := &user.User{}
		if err := s.repository.GetFirstWhere(uow, bankUser, "bank_id=?", bankID); err != nil {
			return fmt.Errorf("error retrieving associated BankUser: %w", err)
		}

		// Update BankUser fields if bank fields have changed
		if bankDTO.BankName != "" {
			bankUser.Name = bankDTO.BankName
		}
		if bankDTO.BankAbbreviation != "" {
			bankUser.Email = generateBankUserEmail(bankDTO.BankAbbreviation) // assumes a method to format email based on abbreviation
		}
		if err := s.repository.Update(uow, bankUser); err != nil {
			return fmt.Errorf("failed to update associated BankUser: %w", err)
		}
	*/
	// -----------------------------------------------------------

	uow.Commit()
	return nil
}

func generateBankUserEmail(abbreviation string) string {
	return strings.ToLower(strings.ReplaceAll(abbreviation, " ", "")) + "@gmail.com"
}
