package service

import (
	"bankManagement/models/bank"
	"bankManagement/models/client"
	"bankManagement/models/document"
	"bankManagement/models/payments"
	"bankManagement/models/transaction"
	"bankManagement/models/user"
	"bankManagement/repository"
	"bankManagement/utils/log"
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type BankUserService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.WebLogger
}

func NewBankUserService(db *gorm.DB, repo repository.Repository, log log.WebLogger) *BankUserService {
	return &BankUserService{
		DB:         db,
		repository: repo,
		log:        log,
	}
}

func (s *BankUserService) CreateClient(clientDTO client.ClientDTO) error {

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// validation
	if err := s.ValidateClientDTO(clientDTO); err != nil {
		return err
	}

	// Validate if the bank_id exists in the banks table - if wrong BankID then Error
	if err := s.ValidateBankID(clientDTO.BankID); err != nil {
		return err
	}

	clientEntity := &client.Client{
		ClientName:         clientDTO.ClientName,
		ClientEmail:        clientDTO.ClientEmail,
		Balance:            clientDTO.Balance,
		IsActive:           clientDTO.IsActive,
		VerificationStatus: clientDTO.VerificationStatus,
		BankID:             clientDTO.BankID,
	}

	if err := s.repository.Add(uow, clientEntity); err != nil {
		return err
	}

	var clientUserRole user.Role
	if err := s.DB.Where("role_name = ?", "CLIENT_USER").First(&clientUserRole).Error; err != nil {
		return fmt.Errorf("CLIENT_USER role not found: %w", err)
	}

	hashedPassword, err := HashPassword(clientDTO.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	clientUser := &user.User{
		Username: clientDTO.Username,
		Password: hashedPassword,
		Name:     clientDTO.ClientName,
		Email:    clientDTO.ClientEmail,
		IsActive: true,
		RoleID:   clientUserRole.ID, // 1 , 2 , 3
	}
	if err := s.repository.Add(uow, clientUser); err != nil {
		return fmt.Errorf("failed to create client user: %w", err)
	}

	// Link -- ClientID & UserID in ClientUSer Table
	client_user := &client.ClientUser{
		UserID:   clientUser.ID,
		ClientID: clientEntity.ID,
	}
	if err := s.repository.Add(uow, client_user); err != nil {
		return fmt.Errorf("failed to create cientUSer table: %w", err)
	}

	uow.Commit()
	return nil
}

// // VALIDATION - already exists check
func (s *BankUserService) ValidateClientDTO(clientDTO client.ClientDTO) error {

	var existingClient client.Client
	clientName := clientDTO.ClientName
	clientEmail := clientDTO.ClientEmail
	if err := s.DB.Where("client_name = ? ", clientName).First(&existingClient).Error; err == nil {
		return fmt.Errorf("client name already exists")
	}
	if err := s.DB.Where("client_email = ?", clientEmail).First(&existingClient).Error; err == nil {
		return fmt.Errorf("client email already exists")
	}

	existingUser := user.User{}
	clientUserUsername := clientDTO.Username
	clientUserEmail := clientDTO.ClientEmail
	if err := s.DB.Where("username = ?", clientUserUsername).First(&existingUser).Error; err == nil {
		return fmt.Errorf("username already exists")
	}

	if err := s.DB.Where("email = ?", clientUserEmail).First(&existingUser).Error; err == nil {
		return fmt.Errorf("email already exists")
	}

	return nil
}

// //// HASHING PASSWORD
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *BankUserService) ValidateBankID(bankID uint) error {
	var bankEntity bank.Bank
	if err := s.DB.First(&bankEntity, bankID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("bank with ID %d does not exist", bankID)
		}
		return err // other errors
	}
	return nil
}

// GET CLIENT BY ID
func (s *BankUserService) GetClientByID(id uint) (*client.Client, error) {
	fmt.Println("Getting ClientID in Service...")

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	clientEntity := client.Client{}
	err := s.repository.GetByID(uow, &clientEntity, id)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// Custom err msg
			return nil, fmt.Errorf("client with ID %d not found or has been deleted", id)
		}
		return nil, err // other errs directly
	}

	fmt.Println("Client found and mapped to DTO")
	uow.Commit()
	return &clientEntity, nil
}

// / GET ALL CLIENTS
func (s *BankUserService) GetAllClients() ([]client.Client, error) {
	fmt.Println("GetAllClients service called")

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var clients []client.Client
	err := s.repository.GetAll(uow, &clients)
	if err != nil {
		fmt.Println("Error in retrieving clients: ", err)
		return nil, err
	}

	fmt.Println("Successfully retrieved all clients")
	uow.Commit()
	return clients, nil
}

// get ClientUser by clientID
func (s *BankUserService) GetClientUserByClientID(clientID uint) *user.User {

	var clientUser client.ClientUser
	err := s.DB.Where("client_id = ?", clientID).First(&clientUser).Error
	if err != nil {
		return nil
	}

	var userEntity user.User
	s.DB.Where("id = ?", clientUser.UserID).First(&userEntity)
	return &userEntity
}

// / UPDATE CLIENT
func (s *BankUserService) UpdateClientByID(id uint, updatedData client.Client) error {
	fmt.Println("UpdateClient service called ...")

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// check if it exists & fetch it
	var existingClient client.Client
	if err := s.repository.GetByID(uow, &existingClient, id); err != nil {
		return errors.New("client not found")
	}

	// Updating fileds if BankUser wants that, otherwise i am keeping Old Values
	if updatedData.ClientName != "" {
		existingClient.ClientName = updatedData.ClientName
	}
	if updatedData.ClientEmail != "" {
		existingClient.ClientEmail = updatedData.ClientEmail
	}

	// during update also bal > 1K
	if updatedData.Balance > 0 {
		if updatedData.Balance < 1000 {
			return errors.New("balance must be at least 1000")
		}
		existingClient.Balance = updatedData.Balance
	}

	// 0 or 1 - isActive
	switch updatedData.IsActive {
	case true, false:
		existingClient.IsActive = updatedData.IsActive
	default:
		return errors.New("invalid value for isActive; only true or false are allowed")
	}

	/// 3 values are allowed for verif. status -- checking that
	allowedStatuses := map[string]bool{
		"Pending":  true,
		"Verified": true,
		"Rejected": true,
	}
	if updatedData.VerificationStatus != "" {
		if _, ok := allowedStatuses[updatedData.VerificationStatus]; !ok {
			return fmt.Errorf("invalid verification status; allowed values are: Pending, Approved, Rejected")
		}
		existingClient.VerificationStatus = updatedData.VerificationStatus
	}

	///updated client record saved
	if err := s.repository.Update(uow, &existingClient, id); err != nil {
		return err
	}

	fmt.Println("Update ClientByID Controller Finished Successfullyy..")
	uow.Commit()
	return nil
}

// // DELETE CLIENT
func (s *BankUserService) DeleteClientByID(id uint) error {
	fmt.Println("Deelte Client service called ...")

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Check for dependent ClientUser records
	var clientUser client.ClientUser
	if err := s.DB.Where("client_id = ?", id).First(&clientUser).Error; err == nil {
		return fmt.Errorf("cannot delete client with ID %d: associated client user exists", id)
	} else if !gorm.IsRecordNotFoundError(err) {
		return fmt.Errorf("error checking for dependent client user records: %w", err)
	}

	// get clientID - checking exists or not - before executing delete query at db level, preventing Rollback() - expensive operation)
	clientEntity := client.Client{}
	err := s.repository.GetByID(uow, &clientEntity, id)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("client with ID %d not found or has already been deleted", id)
		}
		return err //other error
	}

	if err := s.repository.DeleteById(uow, &clientEntity, id); err != nil { ///soft delete
		return err
	}

	fmt.Println("Delete Client service finished ...")
	uow.Commit()
	return nil
}

func (s *BankUserService) VerifyClient(id uint) error {
	clientEntity, err := s.GetClientByID(id)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}
	clientEntity.VerificationStatus = "Verified"

	return s.UpdateClientByID(id, *clientEntity)
}

// /////////// Document Management Functions  /////// Service //////
// CREATE DOCUMENT
func (s *BankUserService) UploadDocument(newDoc document.Document) error {
	fmt.Println("UPLOAD DOCUMENT Service called ...")

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Check if ClientID and UploadedByUserID are associated in ClientUser table
	if err := s.ValidateClientUserRelation(newDoc.ClientId, newDoc.UploadedByUserId); err != nil {
		return fmt.Errorf("the user ID %d does not belong to client ID %d", newDoc.UploadedByUserId, newDoc.ClientId)
	}

	// Create a new document record
	newDocument := &document.Document{
		FileName:         newDoc.FileName,
		FileType:         newDoc.FileType,
		FileURL:          newDoc.FileURL,
		UploadedByUserId: newDoc.UploadedByUserId,
		ClientId:         newDoc.ClientId,
	}

	// Add the document to the repository and check for errors
	if err := s.repository.Add(uow, newDocument); err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	fmt.Println("UPLOAD DOCUMENT Service Finished ...")
	uow.Commit()
	return nil
}

// Checks if the UserID belongs to the specified ClientID (Client & ClientUser validation check)
func (s *BankUserService) ValidateClientUserRelation(clientID, userID uint) error {
	var clientUserRelation client.ClientUser
	if err := s.DB.Where("client_id = ? AND user_id = ?", clientID, userID).First(&clientUserRelation).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("the provided user ID does not belong to the specified client")
		}
		return err
	}
	return nil
}

// READ DOCUMENT
func (s *BankUserService) GetDocumentByID(id uint, bankUserID uint) (*document.Document, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var documentEntity document.Document

	// get document by ID
	if err := s.repository.GetByID(uow, &documentEntity, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("document with ID %d not found", id)
		}
		return nil, err
	}

	// Check if the document ClientID is associated with the BankUser’s BankID
	clientID := documentEntity.ClientId
	if err := s.ValidateBankUserAssociation(bankUserID, clientID); err != nil {
		return nil, fmt.Errorf("the bank user with ID %d is not authorized to access document for client ID %d", bankUserID, clientID)
	}

	uow.Commit()
	return &documentEntity, nil
}

// Validation -- ensure BankUser and Client association
func (s *BankUserService) ValidateBankUserAssociation(bankUserID, clientID uint) error {
	var bankUserAssociation bank.BankUser
	if err := s.DB.Where("user_id = ? AND client_id = ?", bankUserID, clientID).First(&bankUserAssociation).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("the bank user is not authorized to access this client’s document")
		}
		return err
	}
	return nil
}

// / GET ALL DOCUMENTS BY BANK ID
func (s *BankUserService) GetAllDocuments(bankUserID uint, bankID uint) ([]document.Document, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Validate
	if err := s.ValidateBankUserAccess(bankUserID, bankID); err != nil {
		return nil, fmt.Errorf("bank user with ID %d is not authorized to access documents for bank ID %d", bankUserID, bankID)
	}

	// get all documents of clients - for specified bank
	var documents []document.Document
	err := s.DB.Joins("JOIN clients ON clients.id = documents.client_id").
		Where("clients.bank_id = ?", bankID).
		Find(&documents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents for bank ID %d: %w", bankID, err)
	}

	uow.Commit()
	return documents, nil
}

// Validate --  BankUser is authized to access BankID (bank's docs)
func (s *BankUserService) ValidateBankUserAccess(bankUserID, bankID uint) error {
	var bankUser bank.BankUser
	if err := s.DB.Where("user_id = ? AND bank_id = ?", bankUserID, bankID).First(&bankUser).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("the bank user is not authorized to access documents for this bank")
		}
		return err
	}
	return nil
}

// DELETE DOCUMENT BY DOCUMENT ID WITH VALIDATION
func (s *BankUserService) DeleteDocumentByDocumentID(documentID, bankUserID, bankID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var documentEntity document.Document
	if err := s.repository.GetByID(uow, &documentEntity, documentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("document with ID %d not found", documentID)
		}
		return err
	}

	// Validate BankUser
	if err := s.ValidateBankUserAccessForDocument(bankUserID, bankID, documentEntity.ClientId); err != nil {
		return err
	}

	if err := s.repository.DeleteById(uow, &documentEntity, documentID); err != nil {
		return fmt.Errorf("failed to delete document with ID %d: %w", documentID, err)
	}

	uow.Commit()
	return nil
}

// VALIDATION TO CHECK BANKUSER ACCESS FOR DOCUMENT DELETION [checks bankID matches client’s bank]
func (s *BankUserService) ValidateBankUserAccessForDocument(bankUserID, bankID, clientID uint) error {

	var client client.Client
	if err := s.DB.Where("id = ? AND bank_id = ?", clientID, bankID).First(&client).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("client does not belong to the specified bank")
		}
		return err
	}

	if err := s.ValidateBankUserAccess(bankUserID, bankID); err != nil {
		return errors.New("bank user is not authorized to delete documents for this client")
	}
	return nil
}

/////////////  Payment Approval Functions  //////////// Service /////

// Approve Payment Request
func (s *BankUserService) ApprovePaymentRequest(id uint) error {
	paymentRequest, err := s.GetPaymentRequest(id)
	if err != nil {
		return err
	}

	paymentRequest.Status = "Approved"
	return s.UpdatePaymentRequest(*paymentRequest)
}

// Reject payment Request
func (s *BankUserService) RejectPaymentRequest(id uint) error {
	paymentRequest, err := s.GetPaymentRequest(id)
	if err != nil {
		return err
	}

	paymentRequest.Status = "Rejected"
	return s.UpdatePaymentRequest(*paymentRequest)
}

// Get Payment Request - Helper
func (s *BankUserService) GetPaymentRequest(id uint) (*payments.PaymentRequest, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	paymentRequest := payments.PaymentRequest{}
	err := s.repository.GetByID(uow, &paymentRequest, id)
	if err != nil {
		return nil, err
	}

	uow.Commit()
	return &paymentRequest, nil
}

// Update Payment Request
func (s *BankUserService) UpdatePaymentRequest(paymentRequest payments.PaymentRequest) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	err := s.repository.Update(uow, &paymentRequest, paymentRequest.ID)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}

///////////// Transaction Report Functions  ///////// Service ///////

func (s *BankUserService) GenerateTransactionReport(clientID uint) ([]transaction.Transaction, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var transactions []transaction.Transaction
	queryProcessors := []repository.QueryProcessor{
		s.repository.Filter("client_id = ?", clientID),
	}
	err := s.repository.GetAll(uow, &transactions, queryProcessors...)
	if err != nil {
		return nil, err
	}

	uow.Commit()
	return transactions, nil
}
