package service

import (
	"bankManagement/models/bank"
	"bankManagement/models/client"
	"bankManagement/models/document"
	"bankManagement/models/payments"
	"bankManagement/models/transaction"
	"bankManagement/models/user"
	"bankManagement/repository"
	"bankManagement/utils/email"
	"bankManagement/utils/encrypt"
	"bankManagement/utils/log"
	"errors"
	"fmt"
	"strconv"

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
	if err := s.ValidateClientDTO(uow, clientDTO); err != nil {
		return err
	}

	// Validate if the bank_id exists in the banks table - if wrong BankID then Error
	if err := s.ValidateBankID(uow, clientDTO.BankID); err != nil {
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
	if err := s.repository.GetFirstWhere(uow, &clientUserRole, "role_name = ?", "CLIENT_USER"); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("CLIENT_USER role not found: %w", err)
		}
		return err

	}

	hashedPassword := encrypt.HashPassword(clientDTO.Password)

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
func (s *BankUserService) ValidateClientDTO(uow *repository.UOW, clientDTO client.ClientDTO) error { //using existing uow created in CreateClient - as validation part of the same transactional flow (this only performs the validation and doesn’t commit or rollback the transaction—those decisions are left to the calling function -- CreateClient )

	var existingClient client.Client
	clientName := clientDTO.ClientName
	clientEmail := clientDTO.ClientEmail
	if err := s.repository.GetFirstWhere(uow, &existingClient, "client_name = ? ", clientName); err == nil {
		return fmt.Errorf("client name already exists")
	} else if !gorm.IsRecordNotFoundError(err) {
		return err
	}
	if err := s.repository.GetFirstWhere(uow, &existingClient, "client_email = ? ", clientEmail); err == nil {
		return fmt.Errorf("client email already exists")
	} else if !gorm.IsRecordNotFoundError(err) {
		return err
	}
	// if err := s.DB.Where("client_name = ? ", clientName).First(&existingClient).Error; err == nil { return fmt.Errorf("client name already exists") }
	// if err := s.DB.Where("client_email = ?", clientEmail).First(&existingClient).Error; err == nil { return fmt.Errorf("client email already exists") }

	existingUser := user.User{}
	clientUserUsername := clientDTO.Username
	clientUserEmail := clientDTO.ClientEmail
	if err := s.repository.GetFirstWhere(uow, &existingUser, "username = ?", clientUserUsername); err == nil {
		return fmt.Errorf("username already exists")
	} else if !gorm.IsRecordNotFoundError(err) {
		return err
	}
	if err := s.repository.GetFirstWhere(uow, &existingUser, "email = ?", clientUserEmail); err == nil {
		return fmt.Errorf("email already exists")
	} else if !gorm.IsRecordNotFoundError(err) {
		return err
	}
	// if err := s.DB.Where("username = ?", clientUserUsername).First(&existingUser).Error; err == nil { return fmt.Errorf("username already exists") }
	// if err := s.DB.Where("email = ?", clientUserEmail).First(&existingUser).Error; err == nil { return fmt.Errorf("email already exists") }
	return nil
}

// //// HASHING PASSWORD
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *BankUserService) ValidateBankID(uow *repository.UOW, bankID uint) error {
	var bankEntity bank.Bank
	if err := s.repository.GetByID(uow, &bankEntity, bankID); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("bank with ID %d does not exist", bankID)
		}
		return err // other errors
	}
	// if err := s.DB.First(&bankEntity, bankID).Error; err != nil {}
	return nil
}

// GET CLIENT BY ID
func (s *BankUserService) GetClientByID(id uint, bankId uint) (*client.Client, error) {
	fmt.Println("Getting ClientID in Service...")

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	clientEntity := client.Client{}

	err := s.repository.GetByID(uow, &clientEntity, s.repository.Filter("id = ?", id)) // s.repository.Filter("bank_id = ?", bankId),

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
func (s *BankUserService) GetAllClients(bankId uint) ([]client.Client, error) {
	fmt.Println("GetAllClients service called")

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var clients []client.Client
	err := s.repository.GetAll(uow, &clients,
		s.repository.Filter("bank_id=?", bankId),
	)
	if err != nil {
		fmt.Println("Error in retrieving clients: ", err)
		return nil, err
	}

	fmt.Println("Successfully retrieved all clients")
	uow.Commit()
	return clients, nil
}

// get ClientUser by clientID & and BankID
func (s *BankUserService) GetClientUserByClientID(clientID uint, bankID uint) (*user.User, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var clientUser client.ClientUser
	// err := s.DB.Where("client_id = ?", clientID).First(&clientUser).Error
	if err := s.repository.GetFirstWhere(
		uow,
		&clientUser,
		s.repository.Filter("client_id=? AND bank_id=?", clientID, bankID),
	); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, fmt.Errorf("no client user found for client ID %d and bank ID %d", clientID, bankID)
		}
		return nil, err
	}

	var userEntity user.User
	if err := s.repository.GetFirstWhere(uow, &userEntity, s.repository.Filter("id=?", clientUser.UserID)); err != nil {
		return nil, fmt.Errorf("user associated with client ID %d not found", clientID)
	}

	uow.Commit()
	return &userEntity, nil
}

// / UPDATE CLIENT
func (s *BankUserService) UpdateClientByID(id uint, bankId uint, updatedData client.Client) error {
	fmt.Println("UpdateClient service called ...")

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Check if it exists & fetch it
	var existingClient client.Client
	if err := s.repository.GetFirstWhere(uow, &existingClient, "id = ? AND bank_id = ?", id, bankId); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("client with ID %d not found or does not belong to the specified bank", id)
		}
		return err
	}

	// Updating fileds if BankUser wants that, otherwise i am keeping Old Values
	if updatedData.ClientName != "" {
		existingClient.ClientName = updatedData.ClientName
	}
	if updatedData.ClientEmail != "" {
		existingClient.ClientEmail = updatedData.ClientEmail
	}
	if updatedData.Balance > 0 { // balance cannot be negative, will not update
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
	if err := s.repository.Update(uow, &existingClient); err != nil {
		return err
	}

	fmt.Println("Update ClientByID Controller Finished Successfullyy..")
	uow.Commit()
	return nil
}

// // DELETE CLIENT
func (s *BankUserService) DeleteClientByID(id uint, bankId uint) error {
	fmt.Println("Deelte Client service called ...")

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	// Fetch the client to check if it belongs to the correct bank
	var tmpclient client.Client
	if err := s.repository.GetByID(uow, &tmpclient, id); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("client with ID %d not found or has already been deleted", id)
		}
		return err
	}

	// Validate bank ID
	if tmpclient.BankID != bankId {
		return fmt.Errorf("client with ID %d does not belong to bank with ID %d", id, bankId)
	}

	// Check for dependent ClientUser records
	var clientUser client.ClientUser
	if err := s.repository.GetFirstWhere(uow, &clientUser, "client_id = ?", id); err == nil {
		return fmt.Errorf("cannot delete client with ID %d: associated client user exists", id)
	} else if !gorm.IsRecordNotFoundError(err) {
		return fmt.Errorf("error checking for dependent client user records: %w", err)
	}

	if err := s.repository.DeleteById(uow, &tmpclient, id); err != nil {
		return fmt.Errorf("failed to delete client with ID %d: %w", id, err)
	}

	fmt.Println("Delete Client service finished ...")
	uow.Commit()
	return nil
}

func (s *BankUserService) VerifyClient(id uint, bankId uint) error {
	clientEntity, err := s.GetClientByID(id, bankId)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}
	clientEntity.VerificationStatus = "Verified"

	return s.UpdateClientByID(id, bankId, *clientEntity)
}

//--------------------------------------------------------------------------------------------------------------------------

/////////////  Payment Approval Functions  //////////// Service /////

// Approve Payment Request
func (s *BankUserService) ApprovePaymentRequest(paymentId uint, approvedByUserId uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()
	//Getting Payment Request
	paymentRequest := payments.PaymentRequest{}
	err := s.repository.GetByID(uow, &paymentRequest, paymentId)
	if err != nil {
		return err
	}
	//Validating if user has access
	bankUser := bank.BankUser{}
	err = s.repository.GetFirstWhere(uow, &bankUser, "bank_id=? AND user_id=?", paymentRequest.AuthorizerBankId, approvedByUserId)
	if err != nil {
		return err
	}
	if bankUser.UserID != approvedByUserId || bankUser.UserID == 0 {
		return errors.New("unauthorized access to approve request")
	}
	//Validating if User has valid balance
	senderClient := client.Client{}
	err = s.repository.GetByID(uow, &senderClient, paymentRequest.SenderClientID)
	if err != nil {
		return err
	}
	if senderClient.Balance < paymentRequest.PaymentAmount {
		return errors.New("cannot approve the payment as balance insufficient. reject the payment or contact client to update balance")
	}

	//Creating New Payment Entry
	paymentEntry := payments.Payment{
		SenderClientID:   paymentRequest.SenderClientID,
		ReceiverClientID: paymentRequest.ReceiverClientID,
		AuthorizedBankID: paymentRequest.AuthorizerBankId,
		// CreditTransactionID :,
		// DebitTransactionID :,
		PaymentAmount:    paymentRequest.PaymentAmount,
		Status:           transaction.TransactionStatusApproved,
		CreatedByUserId:  paymentRequest.CreatedByUserId,
		ApprovedByUserId: approvedByUserId,
	}
	//Debit Transaction
	debitTransaction := transaction.Transaction{
		ClientID:          paymentRequest.SenderClientID,
		TransactionType:   transaction.TransactionDebit,
		PaymentType:       "Transfer",
		TransactionAmount: paymentEntry.PaymentAmount,
		TransactionStatus: transaction.TransactionStatusApproved,
	}
	err = s.repository.Add(uow, &debitTransaction)
	if err != nil {
		return err
	}
	senderClient.Balance -= paymentEntry.PaymentAmount
	err = s.repository.Update(uow, &senderClient)
	if err != nil {
		return err
	}

	//credit Transaction
	receiverClient := client.Client{}
	err = s.repository.GetByID(uow, &receiverClient, paymentRequest.SenderClientID)
	if err != nil {
		return err
	}
	creditTransaction := transaction.Transaction{
		ClientID:          paymentRequest.ReceiverClientID,
		TransactionType:   transaction.TransactionCredit,
		PaymentType:       "Transfer",
		TransactionAmount: paymentEntry.PaymentAmount,
		TransactionStatus: transaction.TransactionStatusApproved,
	}
	err = s.repository.Add(uow, &creditTransaction)
	if err != nil {
		return err
	}
	receiverClient.Balance += paymentEntry.PaymentAmount
	err = s.repository.Update(uow, &receiverClient)
	if err != nil {
		return err
	}

	paymentEntry.CreditTransactionID = creditTransaction.ID
	paymentEntry.DebitTransactionID = debitTransaction.ID

	err = s.repository.Add(uow, &paymentEntry)
	if err != nil {
		return err
	}
	paymentRequest.Resolved = true
	paymentRequest.Status = transaction.TransactionStatusApproved

	err = s.repository.
		Update(uow, &paymentRequest)
	if err != nil {
		return err
	}

	go email.GetSMTPService().SendEmail("Payment Approved", "Your payment with id:"+strconv.Itoa(int(paymentId))+" has been approved", senderClient.ClientEmail)
	// paymentRequest.Status = "Approved"
	uow.Commit()
	return nil
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

	err := s.repository.Update(uow, &paymentRequest)
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

//-------------------------------------------------------------------------------------------------------

// upload document
// Service layer method to create a document record
func (s *BankUserService) CreateDocument(doc document.Document) error {

	fmt.Println(" --------------------CreateDocument Service Function  START--------------------")

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	if err := s.repository.Add(uow, &doc); err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	uow.Commit()

	fmt.Println(" --------------------CreateDocument Service Function - END-------------------")

	return nil
}

// all documents for a specific bank
func (s *BankUserService) GetAllDocuments(bankID uint) ([]document.Document, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var documents []document.Document
	if err := s.repository.GetAll(uow, &documents, s.repository.Filter("bank_id = ?", bankID)); err != nil {
		return nil, fmt.Errorf("error retrieving documents: %w", err)
	}

	uow.Commit()
	return documents, nil
}

// document by ID
func (s *BankUserService) GetDocumentByID(docID uint, bankID uint) (*document.Document, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var document document.Document
	if err := s.repository.GetFirstWhere(uow, &document, "id = ? AND bank_id = ?", docID, bankID); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("error retrieving document: %w", err)
	}

	uow.Commit()
	return &document, nil
}

// / DELETE DOCUMENT - service
func (s *BankUserService) DeleteDocumentByID(docID uint, bankID uint) error {
	fmt.Println("DeleteDocumentByID service called")

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var documentEntity document.Document
	// Checking if document exists and is associated with the given bank ID
	if err := s.repository.GetFirstWhere(uow, &documentEntity,
		"id = ? AND bank_id = ?", docID, bankID); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return fmt.Errorf("document with ID %d not found for bank ID %d", docID, bankID)
		}
		return fmt.Errorf("failed to fetch document: %w", err)
	}

	// // Remove file from the filesystem
	// if err := os.Remove(documentEntity.FileURL); err != nil {
	// 	return fmt.Errorf("failed to delete file from filesystem: %w", err)
	// }

	// Delete the document record from the database
	if err := s.repository.DeleteById(uow, &documentEntity, docID); err != nil {
		return fmt.Errorf("failed to delete document record: %w", err)
	}

	fmt.Println("DeleteDocumentByID service finished")
	uow.Commit()
	return nil
}
