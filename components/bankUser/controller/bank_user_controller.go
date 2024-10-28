package controller

import (
	"bankManagement/components/bankUser/service"
	"bankManagement/middleware"
	"bankManagement/models/client"
	"bankManagement/models/document"
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
	clientRouter := router.PathPrefix("/clients").Subrouter()
	clientRouter.Use(middleware.ValidateBankUserPermissionsMiddleware) // BankUSer middleware  (BANK_USER can only CRUD on Client and ClientUser)
	clientRouter.HandleFunc("/", controller.CreateClient).Methods("POST")
	clientRouter.HandleFunc("/{id}", controller.GetClientByID).Methods("GET")
	clientRouter.HandleFunc("/", controller.GetAllClients).Methods("GET")
	clientRouter.HandleFunc("/{id}", controller.UpdateClientByID).Methods("PUT")
	clientRouter.HandleFunc("/{id}", controller.DeleteClientByID).Methods("DELETE")

	docRouter := router.PathPrefix("/banks").Subrouter()
	docRouter.Use(middleware.ValidateBankUserPermissionsMiddleware) // BankUSer middleware  (BANK_USER can only UPLOAD DOCUMENTS)
	docRouter.HandleFunc("/documents", controller.UploadDocument).Methods("POST")
	docRouter.HandleFunc("/{bankId}/documents/{documentId}", controller.GetDocumentByID).Methods("GET")
	docRouter.HandleFunc("/{bankId}/documents", controller.GetAllDocuments).Methods("GET")
	docRouter.HandleFunc("/documents/{documentId}", controller.DeleteDocumentByDocumentID).Methods("DELETE")

	paymentRouter := router.PathPrefix("/payments").Subrouter()
	paymentRouter.HandleFunc("/{id}/approve", controller.ApprovePaymentRequest).Methods(http.MethodPost)
	paymentRouter.HandleFunc("/{id}/reject", controller.RejectPaymentRequest).Methods(http.MethodPost)
	paymentRouter.HandleFunc("/{id}", controller.GetPaymentRequest).Methods(http.MethodGet)

	transactionRouter := router.PathPrefix("/transactions").Subrouter()
	transactionRouter.HandleFunc("/{client_id}/report", controller.GenerateTransactionReport).Methods(http.MethodGet)

}

// // CREATE CLIENT
func (controller *BankUserController) CreateClient(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreateClient called")

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

	clientID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	clientEntity, err := controller.BankUserService.GetClientByID(uint(clientID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, fmt.Sprintf("Client with ID %d not found", clientID), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch  ClientUser - Username
	clientUser := controller.BankUserService.GetClientUserByClientID(clientEntity.ID)
	var clientUserUsername string
	if clientUser != nil {
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

	clients, err := controller.BankUserService.GetAllClients()
	if err != nil {
		http.Error(w, "Failed to retrieve clients", http.StatusInternalServerError) //err.Error()
		return
	}

	///// implementing DTO for response - maping each client entity to a ClientResponseDTO before sending to response (exclude bank details and other fields)
	var clientResponses []client.ClientResponseDTO //slice of client.ClientResponseDTO

	for _, clientEntity := range clients {

		// Username - Fetch associated ClientUser Username
		clientUser := controller.BankUserService.GetClientUserByClientID(clientEntity.ID)
		var clientUserUsername string
		if clientUser != nil {
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

// ///// UPDATE CLIENT
func (controller *BankUserController) UpdateClientByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UpdateClient controller called ...")

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

	if err := controller.BankUserService.UpdateClientByID(clientID, updatedData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Update ClientByID Controller Finished Successfullyy..")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Client updated successfully"))
}

func validateClientUpdateInput(updatedData client.Client) error {
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

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	clientID := uint(id)

	err = controller.BankUserService.DeleteClientByID(clientID)
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

	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid client ID. Check that ClientID is valid positive number", http.StatusBadRequest)
		return
	}
	clientID := uint(id)

	err = controller.BankUserService.VerifyClient(clientID)
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

// /////////// Document Management Functions  /////// Controller //////

// UPLOAD DOCUMENT
func (controller *BankUserController) UploadDocument(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UPLOAD DOCUMENT Controller called ...")

	var newDocument document.Document
	if err := json.NewDecoder(r.Body).Decode(&newDocument); err != nil {
		http.Error(w, "Invalid document data", http.StatusBadRequest)
		return
	}

	fmt.Printf("file_name: %s\nfile_type: %s\nfile_url: %s\nuploaded_by_user_id: %d\nclient_id: %d\n",
		newDocument.FileName, newDocument.FileType, newDocument.FileURL, newDocument.UploadedByUserId, newDocument.ClientId)

	if newDocument.FileName == "" || newDocument.FileType == "" || newDocument.FileURL == "" || newDocument.UploadedByUserId == 0 || newDocument.ClientId == 0 {
		http.Error(w, "All fields (file_name, file_type, file_url, uploaded_by_user_id, client_id) are required", http.StatusBadRequest)
		return
	}
	// checks that client exists or not
	if _, err := controller.BankUserService.GetClientByID(newDocument.ClientId); err != nil {
		http.Error(w, "Client does not exist", http.StatusBadRequest)
		return
	}
	// service Call
	if err := controller.BankUserService.UploadDocument(newDocument); err != nil {
		http.Error(w, "Error uploading document: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Document Uploaded Controller Finished...")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Document uploaded successfully"))
}

// GET DOCUEMNT BY DOCUMENTID
func (controller *BankUserController) GetDocumentByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bankID, _ := strconv.Atoi(vars["bankId"])
	docID, _ := strconv.Atoi(vars["documentId"])

	doc, err := controller.BankUserService.GetDocumentByID(uint(bankID), uint(docID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

// / GET ALL DOCUMENTS
func (controller *BankUserController) GetAllDocuments(w http.ResponseWriter, r *http.Request) {

	bankIDStr := mux.Vars(r)["bankId"]
	clientIDStr := r.URL.Query().Get("clientId") //// bankID and clientID from query parameters

	bankID, _ := strconv.Atoi(bankIDStr)
	clientID, _ := strconv.Atoi(clientIDStr)

	docs, err := controller.BankUserService.GetAllDocuments(uint(bankID), uint(clientID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

// DELETE DOCUEMNT
func (controller *BankUserController) DeleteDocumentByDocumentID(w http.ResponseWriter, r *http.Request) {
	bankIDStr := mux.Vars(r)["bankId"]
	clientIDStr := r.URL.Query().Get("clientId")
	documentIDStr := mux.Vars(r)["documentId"]

	bankID, _ := strconv.Atoi(bankIDStr)
	clientID, _ := strconv.Atoi(clientIDStr)
	documentID, _ := strconv.Atoi(documentIDStr)

	err := controller.BankUserService.DeleteDocumentByDocumentID(uint(bankID), uint(clientID), uint(documentID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/////////////  Payment Approval Functions  //////// Controller /////

// Approve Payment Request
func (controller *BankUserController) ApprovePaymentRequest(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid payment request ID", http.StatusBadRequest)
		return
	}
	if err := controller.BankUserService.ApprovePaymentRequest(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payment request approved"))
}

func (controller *BankUserController) RejectPaymentRequest(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
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
