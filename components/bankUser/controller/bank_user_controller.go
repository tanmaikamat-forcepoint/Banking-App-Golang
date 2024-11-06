package controller

import (
	"bankManagement/components/bankUser/service"
	"bankManagement/constants"
	"bankManagement/middlewares/auth"
	"bankManagement/models/client"
	"bankManagement/models/document"
	"bankManagement/utils/encrypt"
	errorsUtils "bankManagement/utils/errors"
	"bankManagement/utils/log"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type BankUserController struct {
	BankUserService *service.BankUserService
	log             log.WebLogger
}

func NewBankUserController(bankUserService *service.BankUserService, log log.WebLogger) *BankUserController {
	return &BankUserController{
		BankUserService: bankUserService,
		log:             log,
	}
}

func (controller *BankUserController) RegisterRoutes(router *mux.Router) {
	clientRouter := router.PathPrefix("/banks/{bank_id}/clients").Subrouter()
	clientRouter.Use(auth.AuthenticationMiddleware, auth.ValidateBankPermissionsMiddleware) // BankUSer middleware  (BANK_USER can only CRUD on Client and ClientUser)
	clientRouter.HandleFunc("/", controller.CreateClient).Methods("POST")
	clientRouter.HandleFunc("/{id}", controller.GetClientByID).Methods("GET")
	clientRouter.HandleFunc("/", controller.GetAllClients).Methods("GET")
	clientRouter.HandleFunc("/{id}", controller.UpdateClientByID).Methods("PUT")
	clientRouter.HandleFunc("/{id}", controller.DeleteClientByID).Methods("DELETE")

	documentRouter := router.PathPrefix("/banks/{bank_id}/documents").Subrouter()
	documentRouter.Use(auth.AuthenticationMiddleware, auth.ValidateBankPermissionsMiddleware) // BankUSer middleware  (BANK_USER can only UPLOAD DOCUMENTS)
	documentRouter.HandleFunc("/", controller.UploadDocument).Methods("POST")
	documentRouter.HandleFunc("/{id}", controller.GetDocumentByID).Methods("GET")
	documentRouter.HandleFunc("/", controller.GetAllDocuments).Methods("GET")
	documentRouter.HandleFunc("/{id}", controller.DeleteDocumentByID).Methods("DELETE")

	paymentRouter := router.PathPrefix("/banks/{bank_id}/payment_requests").Subrouter()
	paymentRouter.Use(auth.AuthenticationMiddleware, auth.ValidateBankPermissionsMiddleware)
	paymentRouter.HandleFunc("/{payment_request_id}/approve", controller.ApprovePaymentRequest).Methods(http.MethodPost)
	paymentRouter.HandleFunc("/{payment_request_id}/reject", controller.RejectPaymentRequest).Methods(http.MethodPost)

	paymentRouter.HandleFunc("/{id}", controller.GetPaymentRequest).Methods(http.MethodGet)

	transactionRouter := router.PathPrefix("/transactions").Subrouter()
	transactionRouter.HandleFunc("/{client_id}/reports", controller.GenerateTransactionReport).Methods(http.MethodGet)

}

// // CREATE CLIENT
func (controller *BankUserController) CreateClient(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreateClient called")
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)

	clientDTO := client.ClientDTO{}
	if err := json.NewDecoder(r.Body).Decode(&clientDTO); err != nil {
		http.Error(w, "Invalid input format", http.StatusBadRequest)
		return
	}

	// validation on all attribures - contrller level
	if err := validateClientDTO(clientDTO); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	clientDTO.BankID = claims.BankId
	//service call
	if err := controller.BankUserService.CreateClient(clientDTO); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Client and ClientUser created successfully.")
	w.WriteHeader(http.StatusCreated)
}

// // VALIDATION
func validateClientDTO(dto client.ClientDTO) error {
	if dto.ClientName == "" {
		return errors.New("client name is required")
	}
	if dto.ClientEmail == "" {
		return errors.New("client email is required")
	}
	if !strings.Contains(dto.ClientEmail, "@") {
		return errors.New("invalid email format")
	}
	if dto.Balance < 0 {
		return errors.New("balance cannot be negative")
	}
	if dto.Username == "" {
		return errors.New("username is required for client user")
	}
	if dto.Password == "" {
		return errors.New("password is required for client user")
	}
	return nil
}

// / GET CLIENT BY ID
func (controller *BankUserController) GetClientByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetClientByID called..")
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)

	clientID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// service layer call .. (bankId also added) -- onlyy clients specified BankId are fetched
	clientEntity, err := controller.BankUserService.GetClientByID(uint(clientID), claims.BankId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, fmt.Sprintf("Client with ID %d not found", clientID), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//// do in this also

	// Fetch ClientUser Username  (validating  ClientUser belongsto specified ClientId and BankId)
	clientUser, err := controller.BankUserService.GetClientUserByClientID(clientEntity.ID, claims.BankId)
	var clientUserUsername string
	if err != nil {
		fmt.Println("Error fetching client user:", err)
	} else if clientUser != nil {
		clientUserUsername = clientUser.Username
	}

	// response data - not passing all fields in response
	response := &client.ClientResponseDTO{
		ID:                 clientEntity.ID,
		ClientName:         clientEntity.ClientName,
		ClientEmail:        clientEntity.ClientEmail,
		Balance:            clientEntity.Balance,
		IsActive:           clientEntity.IsActive,
		VerificationStatus: clientEntity.VerificationStatus,
		BankID:             clientEntity.BankID,
		Username:           clientUserUsername,
	}

	fmt.Println("GetClientByID Finished..")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response) //return client data
}

// // GET ALL CLIENTS
func (controller *BankUserController) GetAllClients(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetAllClients controller called ...")

	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)

	clients, err := controller.BankUserService.GetAllClients(claims.BankId)
	if err != nil {
		http.Error(w, "Failed to retrieve clients", http.StatusInternalServerError) //err.Error()
		return
	}

	///// implementing DTO for response - maping each client entity to a ClientResponseDTO before sending to response (exclude bank details and other fields)
	var clientResponses []client.ClientResponseDTO //slice of client.ClientResponseDTO

	for _, clientEntity := range clients {

		// Username - Fetch associated ClientUser Username
		clientUser, err := controller.BankUserService.GetClientUserByClientID(clientEntity.ID, claims.BankId)
		var clientUserUsername string
		if err != nil {
			fmt.Println("Error fetching client user:", err)
		} else if clientUser != nil {
			clientUserUsername = clientUser.Username
		}

		// Create the response DTO
		clientResponse := client.ClientResponseDTO{
			ID:                 clientEntity.ID,
			ClientName:         clientEntity.ClientName,
			ClientEmail:        clientEntity.ClientEmail,
			Balance:            clientEntity.Balance,
			IsActive:           clientEntity.IsActive,
			VerificationStatus: clientEntity.VerificationStatus,
			BankID:             clientEntity.BankID,
			Username:           clientUserUsername,
		}
		clientResponses = append(clientResponses, clientResponse)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientResponses)
	fmt.Println("GetAllClients response sent successfully.")

}

// UPDATE CLIENT ()
func (controller *BankUserController) UpdateClientByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UpdateClient controller called ...")

	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)

	vars := mux.Vars(r) /// path varibales from req
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr) // string to int
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	clientID := uint(id)

	updatedData := client.Client{}
	if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil { // Decododing request body into client.Client struct

		http.Error(w, "Invalid input format; please check the JSON structure", http.StatusBadRequest)
		return
	}

	// Validation
	if err := validateClientUpdateInput(updatedData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := controller.BankUserService.UpdateClientByID(clientID, claims.BankId, updatedData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Update ClientByID Controller Finished Successfullyy..")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Client updated successfully"))
}

func validateClientUpdateInput(updatedData client.Client) error {
	// check if ClientName, ClientEmai - > not same as some Client. Update BAnkID check
	allowedStatuses := map[string]bool{
		"Pending":  true,
		"Verified": true,
		"Rejected": true,
	}

	if updatedData.Balance != 0 && updatedData.Balance < 1000 {
		return errors.New("balance must be at least 1000")
	}

	if updatedData.VerificationStatus != "" {
		if _, ok := allowedStatuses[updatedData.VerificationStatus]; !ok {
			return fmt.Errorf("invalid verification status; allowed values are: Pending, Verified, Rejected")
		}
	}

	return nil
}

// / DELETE CLIENT BY ID

func (controller *BankUserController) DeleteClientByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeleteClient controller called..")
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	clientID := uint(id)

	err = controller.BankUserService.DeleteClientByID(clientID, claims.BankId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, fmt.Sprintf("Client with ID %d not found or has already been deleted", clientID), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("DeleteClient controller finished..")
	w.WriteHeader(http.StatusNoContent)
}

// Verify client
func (controller *BankUserController) VerifyClient(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Verified client controlle called ...")

	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)

	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid client ID. Check that ClientID is valid positive number", http.StatusBadRequest)
		return
	}
	clientID := uint(id)

	err = controller.BankUserService.VerifyClient(clientID, claims.BankId)
	if err != nil {
		if strings.Contains(err.Error(), "Client not found. Check ClientId") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Client verification status updated successfully"))
}

//-------------------------------------------------------------------------------------------------------
/////////////  Payment Approval Functions  //////// Controller /////

// Approve Payment Request
func (controller *BankUserController) ApprovePaymentRequest(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["payment_request_id"]
	id, err := strconv.Atoi(idStr)
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	if err != nil {
		http.Error(w, "Invalid payment request ID", http.StatusBadRequest)
		return
	}
	if err := controller.BankUserService.ApprovePaymentRequest(uint(id), claims.UserId); err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payment request approved"))
}

func (controller *BankUserController) RejectPaymentRequest(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["payment_request_id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid payment request ID", http.StatusBadRequest)
		return
	}
	if err := controller.BankUserService.RejectPaymentRequest(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payment request rejected"))
}

func (controller *BankUserController) GetPaymentRequest(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid payment request ID", http.StatusBadRequest)
		return
	}
	paymentRequest, err := controller.BankUserService.GetPaymentRequest(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paymentRequest)
}

///////////// Transaction Report Functions  ///////// Controller /////

func (controller *BankUserController) GenerateTransactionReport(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["client_id"]
	clientID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	transactions, err := controller.BankUserService.GenerateTransactionReport(uint(clientID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

/////////////// --------------------- Upload Document ------------------------------------------ //////////////////////

const DocumentUploadDir = `\C:\Users\sushant.chauhan\Desktop\Banking-App-Golang\documents`

// / UPLOAD DOCUMENT
func (controller *BankUserController) UploadDocument(w http.ResponseWriter, r *http.Request) {

	fmt.Println(" \n----------  UploadDocument called ----------  ")

	bankId, err := strconv.ParseUint(mux.Vars(r)["bank_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid bank ID", http.StatusBadRequest)
		return
	}

	// if err := r.ParseMultipartForm(10 << 20); err != nil {
	// 	http.Error(w, "Parsing form error", http.StatusInternalServerError)
	// 	return
	// }

	fmt.Println(" -----bankId ----------- = ", bankId)

	fmt.Println(" -------------------- 1 -------------------- ")

	// file, fileHeader, err := r.FormFile("file")
	// if err != nil {
	// 	http.Error(w, "Error retrieving the file", http.StatusInternalServerError)
	// 	return
	// }
	// defer file.Close()

	fmt.Println(" -------------------- 2-------------------- ")

	// os.MkdirAll(DocumentUploadDir, os.ModePerm)

	// // Define file path
	// fileName := filepath.Base(fileHeader.Filename)
	// fileType := fileHeader.Header.Get("Content-Type")
	// filePath := filepath.Join(DocumentUploadDir, fileName)

	// // Save the file
	// targetFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	// if err != nil {
	// 	http.Error(w, "Failed to open file for writing: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// defer targetFile.Close()

	fmt.Println(" -------------------- 3 -------------------- ")

	// if _, err := io.Copy(targetFile, file); err != nil {
	// 	http.Error(w, "Failed to copy the file: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// uploadedByUserId, _ := strconv.ParseUint(r.FormValue("uploaded_by_user_id"), 10, 32)
	// clientId, _ := strconv.ParseUint(r.FormValue("client_id"), 10, 32)

	// Create Document record
	documentRecord := document.Document{
		FileName:         "fileName",
		FileType:         "fileType",
		FileURL:          "filePath",
		UploadedByUserId: 7, //uint(uploadedByUserId),
		ClientId:         2, //uint(clientId),
		BankId:           uint(bankId),
	}

	fmt.Println("------- 4--------------------")

	// Service call to save document details
	if err := controller.BankUserService.CreateDocument(documentRecord); err != nil {
		http.Error(w, "Failed to create document record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("----- 5--------------------")

	fmt.Println("File uploaded and saved successfully")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Document uploaded successfully"))
}

// GET DOCUMENTS - from bank
func (controller *BankUserController) GetAllDocuments(w http.ResponseWriter, r *http.Request) {
	fmt.Println("------- GetAllDocuments called ------------")
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	bankID := claims.BankId // Assuming BankId is extracted from JWT claims

	documents, err := controller.BankUserService.GetAllDocuments(bankID)
	if err != nil {
		http.Error(w, "Failed to fetch documents: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

// / GET DOCUMENT BY ID
func (controller *BankUserController) GetDocumentByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetDocumentByID called")
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	bankID := claims.BankId // BankId is extracted from JWT claims

	docIDStr := mux.Vars(r)["id"]
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		http.Error(w, "Invalid document ID format", http.StatusBadRequest)
		return
	}

	document, err := controller.BankUserService.GetDocumentByID(uint(docID), bankID)
	if err != nil {
		http.Error(w, "Failed to fetch document: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(document)
}

// // DELETE DOCUMENT - controller
func (controller *BankUserController) DeleteDocumentByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeleteDocumentByID called")
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	bankID := claims.BankId // Assuming BankId is extracted from JWT claims

	docIDStr := mux.Vars(r)["id"]
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		http.Error(w, "Invalid document ID format", http.StatusBadRequest)
		return
	}

	// Call the service layer to delete the document
	if err := controller.BankUserService.DeleteDocumentByID(uint(docID), bankID); err != nil {
		http.Error(w, "Failed to delete document: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("Document deleted successfully"))
}
