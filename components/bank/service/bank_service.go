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
	if err := s.ValidateBankDTO(uow, bankAndUserEntityDTO); err != nil {
		return err
	}
	//Create Bank
	bankEntity := &bank.Bank{
		BankName:         bankAndUserEntityDTO.BankName,
		BankAbbreviation: bankAndUserEntityDTO.BankAbbreviation,
		IsActive:         true,
	}
	if err := s.repository.Add(uow, bankEntity); err != nil {
		return fmt.Errorf("bank is soft-deleted, create Bank with different name: --- %w", err)
		// return err
	}

	hashedPassword := encrypt.HashPassword(bankAndUserEntityDTO.Password)

	//Create User for BankUser
	userEntity := &user.User{
		Username: bankAndUserEntityDTO.Username,
		Password: hashedPassword,
		Name:     bankAndUserEntityDTO.BankName, //  BankUSer Name == Bank Name
		Email:    bankAndUserEntityDTO.Email,    //  BankUserEmail == Bank Email
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
func (s *BankService) ValidateBankDTO(uow *repository.UOW, bankDTO bank.BankAndUserDTO) error {

	var existingBank bank.Bank
	name := bankDTO.BankName
	abbreviation := bankDTO.BankAbbreviation

	// Check if bank name or abbreviation already exists
	if err := s.repository.GetFirstWhere(uow, &existingBank, "bank_name = ? OR bank_abbreviation = ?", name, abbreviation); err == nil {
		return fmt.Errorf("bank name or abbreviation already exists")
	} else if !gorm.IsRecordNotFoundError(err) {
		return err // other database errors
	}

	var existingUser user.User
	// Check if username already exists
	if err := s.repository.GetFirstWhere(uow, &existingUser, "username = ?", bankDTO.Username); err == nil {
		return fmt.Errorf("username already exists")
	} else if !gorm.IsRecordNotFoundError(err) {
		return err // other database errors
	}

	return nil

	///.............. using DB ......
	//// Check if bank name or abbreviation already exists || Check if username already exists
	// if err := s.DB.Where("bank_name = ? OR bank_abbreviation = ?", name, abbreviation).First(&existingBank).Error; err == nil { return fmt.Errorf("bank name or abbreviation already exists")}
	// if err := s.DB.Where("username = ?", bankDTO.Username).First(&existingUser).Error; err == nil { return fmt.Errorf("username already exists") }
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

// DELETE  [ Bank and BankUser both are deleted - if No Client is Associated.]
func (s *BankService) DeleteBank(bankID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// checking if Bank Exists or not - May be already deleted
	bankEntity := bank.Bank{}
	if err := s.repository.GetByID(uow, &bankEntity, bankID); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("bank with ID %d does not exist or already deleted", bankID)
		}
		return fmt.Errorf("error checking for bank existence: %w", err)
	}

	// 1. No client should be assocaited with Bank - Only then we can delete Bank
	var clients []client.Client
	if err := s.repository.GetAll(uow, &clients, s.repository.Filter("bank_id = ?", bankID)); err != nil {
		return fmt.Errorf("failed to check associated clients: %w", err)
	}
	if len(clients) > 0 {
		return fmt.Errorf("cannot delete bank with ID %d: associated clients exist", bankID)
	} else {
		fmt.Println("No associated found.")
	}

	// 2: Delete associated BankUser (from Joining `bankmanagement.bank_users` Table) and its corresponding User (in `users` Table)
	var bankUser bank.BankUser
	if err := s.repository.GetFirstWhere(uow, &bankUser, "bank_id = ?", bankID); err == nil {
		// Delete the entry in `bank_users` table first
		if err := s.repository.DeleteById(uow, &bankUser, bankUser.UserID); err != nil {
			return fmt.Errorf("failed to delete associated BankUser: %w", err)
		}
		// now delete the associated User from `users` table
		userEntity := user.User{}
		userEntity.ID = bankUser.UserID
		if err := s.repository.DeleteById(uow, &userEntity, userEntity.ID); err != nil {
			return fmt.Errorf("failed to delete associated User: %w", err)
		}
	} else if !gorm.IsRecordNotFoundError(err) {
		return fmt.Errorf("error checking for BankUser in bank_users table: %w", err)
	}

	// 3. Delete the Bank itself from the `banks` table
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

	//1. Retrieves the Bank by bankID and checks  - Exists or not
	existingBank := bank.Bank{}
	if err := s.repository.GetByID(uow, &existingBank, bankID); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("bank with ID %d does not exist", bankID)
		}
		return fmt.Errorf("error checking for bank existence: %w", err)
	}

	// Validate i/p json data
	if err := s.ValidateBankDTO(uow, bankDTO); err != nil {
		return err
	}

	// 2. Updates Bank fields from bankDTO ---- Updating fields if BankUser wants that, otherwise I am keeping Old Values
	if bankDTO.BankName != "" {
		existingBank.BankName = bankDTO.BankName
	}
	if bankDTO.BankAbbreviation != "" {
		existingBank.BankAbbreviation = bankDTO.BankAbbreviation
	}

	// 3. Save the updated User entity
	if err := s.repository.Update(uow, &existingBank); err != nil {
		return fmt.Errorf("failed to update bank: %w", err)
	}

	// ----------------  ASSOCIATED bANKUser deletion -----------------

	//finding and updating associated BankUser

	associatedUser := user.User{} // here i don't have UserID of the Bank User to be Updated  -- need to find that from Joining Table - 'bank_users'  [this user with  UserId  is automatically created with Bank]
	bankUser := bank.BankUser{}   // Joining Table Struct

	// 1 : Fetch the associated BankUser by bankId :  ----- and getting the UserID from 'bank_users' Table - basically by Populating(automatic) BankUser Struct
	if err := s.repository.GetFirstWhere(uow, &bankUser, "bank_id = ?", bankID); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("no associated bank user found for bank ID %d", bankID)
		}
		return fmt.Errorf("error fetching associated BankUser: %w", err)
	}

	//2 : Retrieve the associated 'User'  by 'UserID'(BankUser.UserID)
	associatedUser.ID = bankUser.UserID
	if err := s.repository.GetByID(uow, &associatedUser, associatedUser.ID); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("user associated with bank UserID  %d does not exist", bankUser.UserID)
		}
		return fmt.Errorf("error fetching associated User: %w", err)
	}

	// 3 : Update User fields related to Bank info (Name and Email)
	if bankDTO.BankName != "" {
		associatedUser.Name = bankDTO.BankName
		associatedUser.Email = generateBankUserEmail(bankDTO.BankName)

		// email := strings.ReplaceAll(strings.ToLower(dto.BankName), " ", "") + "@gmail.com"
	}

	// 4 : save to DB
	if err := s.repository.Update(uow, associatedUser); err != nil {
		return fmt.Errorf("failed to update associated BankUser: %w", err)
	}

	// -----------------------------------------------------------

	uow.Commit()
	return nil
}

func generateBankUserEmail(updatedBankName string) string {
	return strings.ReplaceAll(strings.ToLower(updatedBankName), " ", "") + "@gmail.com"
}
