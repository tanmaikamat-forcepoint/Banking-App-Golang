package controller

import (
	"bankManagement/components/bankUser/service"
	"bankManagement/constants"
	"bankManagement/middlewares/auth"
	"bankManagement/models/client"
	"bankManagement/models/document"
	"bankManagement/utils/encrypt"
	"bankManagement/utils/log"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	clientRouter := router.PathPrefix("/bank/{bank_id}/client").Subrouter()
	clientRouter.Use(auth.AuthenticationMiddleware, auth.ValidateBankPermissionsMiddleware) // BankUSer middleware  (BANK_USER can only CRUD on Client and ClientUser)
	clientRouter.HandleFunc("/", controller.CreateClient).Methods("POST")
	clientRouter.HandleFunc("/{id}", controller.GetClientByID).Methods("GET")
	clientRouter.HandleFunc("/", controller.GetAllClients).Methods("GET")
	clientRouter.HandleFunc("/{id}", controller.UpdateClientByID).Methods("PUT")
	clientRouter.HandleFunc("/{id}", controller.DeleteClientByID).Methods("DELETE")

	docRouter := router.PathPrefix("/bank/{bank_id}/document").Subrouter()
	docRouter.Use(auth.AuthenticationMiddleware, auth.ValidateBankPermissionsMiddleware) // BankUSer middleware  (BANK_USER can only UPLOAD DOCUMENTS)
	docRouter.HandleFunc("/", controller.uploadDocumentHandler).Methods(http.MethodPost)
	docRouter.HandleFunc("/{id}", controller.GetDocumentByID).Methods("GET")
	docRouter.HandleFunc("/", controller.GetAllDocuments).Methods("GET")

	paymentRouter := router.PathPrefix("/payment").Subrouter()
	paymentRouter.HandleFunc("/{id}/approve", controller.ApprovePaymentRequest).Methods(http.MethodPost)
	paymentRouter.HandleFunc("/{id}/reject", controller.RejectPaymentRequest).Methods(http.MethodPost)
	paymentRouter.HandleFunc("/{id}", controller.GetPaymentRequest).Methods(http.MethodGet)

	transactionRouter := router.PathPrefix("/transaction").Subrouter()
	transactionRouter.HandleFunc("/{client_id}/report", controller.GenerateTransactionReport).Methods(http.MethodGet)

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

// -------------------------------------------------------------------------------------------------------
const DocumentUploadDir = `C:\Users\sushant.chauhan\Desktop\Banking-App-Golang\document`

// var db *gorm.DB // Global database connection

func (controller *BankUserController) uploadDocumentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bankId, err := strconv.Atoi(vars["bank_id"])
	if err != nil {
		http.Error(w, "Invalid bank ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Parsing form error", http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileName := handler.Filename
	fileType := handler.Header.Get("Content-Type")
	filePath := filepath.Join(DocumentUploadDir, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Unable to save the file", http.StatusInternalServerError)
		return
	}

	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}
	uploadedByUserIdStr := r.FormValue("uploaded_by_user_id")
	clientIdStr := r.FormValue("client_id")
	uploadedByUserId, _ := strconv.Atoi(uploadedByUserIdStr)
	clientId, _ := strconv.Atoi(clientIdStr)

	document1 := document.Document{
		FileName:         fileName,
		FileType:         fileType,
		FileURL:          filePath,
		UploadedByUserId: uint(uploadedByUserId),
		ClientId:         uint(clientId),
		BankId:           uint(bankId),
	}

	fmt.Println("fileName, filePath, fileType ================ >>>>>>>>>> ", fileName, filePath, fileType)

	if result := controller.BankUserService.DB.Create(&document1); result.Error != nil {
		http.Error(w, "Error saving document metadata", http.StatusInternalServerError)
		return
	}

	fmt.Println(w, "File uploaded and saved successfully")
}

// get all documents for a bank
func (controller *BankUserController) GetAllDocuments(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetAllDocuments called")
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

func (controller *BankUserController) GetDocumentByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetDocumentByID called")
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	bankID := claims.BankId // Assuming BankId is extracted from JWT claims

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
