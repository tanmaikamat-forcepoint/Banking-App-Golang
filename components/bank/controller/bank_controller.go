package controller

import (
	"bankManagement/components/bank/service"
	"bankManagement/middlewares/auth"
	"bankManagement/utils/log"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"bankManagement/models/bank"

	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type BankController struct {
	BankService *service.BankService
	log         log.WebLogger
}

func NewBankController(UserServcice *service.BankService, log log.WebLogger) *BankController {
	return &BankController{
		log:         log,
		BankService: UserServcice,
	}
}

func (controller *BankController) RegisterRoutes(router *mux.Router) {
	bankRouter := router.PathPrefix("/banks").Subrouter()
	bankRouter.Use(auth.AuthenticationMiddleware, auth.ValidateAdminPermissionsMiddleware) // BankUSer middleware  (BANK_USER can only CRUD on Client and ClientUser)
	// bankRouter.Use(middleware.ValidateAdminPermissionsMiddleware)                         // SuperAdmin middleware
	bankRouter.HandleFunc("/", controller.CreateBank).Methods("POST")
	bankRouter.HandleFunc("/", controller.GetAllBanks).Methods("GET")
	bankRouter.HandleFunc("/{id}", controller.GetBankByID).Methods("GET")
	bankRouter.HandleFunc("/{id}", controller.DeleteBank).Methods("DELETE")
	bankRouter.HandleFunc("/{id}", controller.UpdateBank).Methods("PUT")

	// POST - /api/v1/bankManagement/banks
}

// ------------ Super Admin: Manages banks and generates various reports [SRS] --------

// / CREATE BANK
func (controller *BankController) CreateBank(w http.ResponseWriter, r *http.Request) {

	fmt.Println("CreateBank called")

	bankEntityDTO := bank.BankAndUserDTO{}

	if err := json.NewDecoder(r.Body).Decode(&bankEntityDTO); err != nil {
		http.Error(w, "Invalid input format", http.StatusBadRequest)
		return
	}

	// Validate
	email, err := validateBankDTO(bankEntityDTO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bankEntityDTO.Email = email //// Set generated email in DTO for service layer use

	// service call --  create bank and bank user
	if err := controller.BankService.CreateBank(bankEntityDTO); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("bank  and associated bankuser created successfully"))
}

// validation on BankDTO (includes Bank + User info)
func validateBankDTO(dto bank.BankAndUserDTO) (string, error) {
	if dto.BankName == "" {
		return "", errors.New("bank  name is required")
	}
	if len(dto.BankName) < 3 {
		return "", errors.New("bank  name must be at least 3 characters long")
	}

	if dto.BankAbbreviation == "" {
		return "", errors.New("bank  abbreviation is required")
	}
	if len(dto.BankAbbreviation) > len(dto.BankName) {
		return "", errors.New("bank  abbreviation cannot be longer than Bank Name")
	}

	if dto.Username == "" {
		return "", errors.New("username is required")
	}
	if len(dto.Username) < 5 {
		return "", errors.New("username must be at least 4 characters long")
	}

	if dto.Password == "" {
		return "", errors.New("password is required")
	}
	if len(dto.Password) < 5 {
		return "", errors.New("password must be at least 8 characters long")
	}

	email := strings.ReplaceAll(strings.ToLower(dto.BankName), " ", "") + "@gmail.com"

	return email, nil
}

// GET BY ID
func (controller *BankController) GetBankByID(w http.ResponseWriter, r *http.Request) {

	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid bank ID", http.StatusBadRequest)
		return
	}

	bankEntity, err := controller.BankService.GetBankByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bankEntity)
}

// GET ALL
func (controller *BankController) GetAllBanks(w http.ResponseWriter, r *http.Request) {

	banks, err := controller.BankService.GetAllBanks()
	if err != nil {
		http.Error(w, "Failed to retrieve banks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banks)
}

// DELETE
func (controller *BankController) DeleteBank(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeleteBank controller called...")

	bankIDStr := mux.Vars(r)["id"]
	bankID, err := strconv.Atoi(bankIDStr)
	if err != nil {
		http.Error(w, "Invalid bank ID", http.StatusBadRequest)
		return
	}

	if err := controller.BankService.DeleteBank(uint(bankID)); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(fmt.Sprintf("Bank and associated BankUser with ID %d successfully deleted", bankID)))

}

// // UPDATE BANK
func (controller *BankController) UpdateBank(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UpdateBank called")

	//bank ID-- URL
	bankIDStr := mux.Vars(r)["id"]
	bankID, err := strconv.Atoi(bankIDStr)
	if err != nil {
		http.Error(w, "Invalid bank ID", http.StatusBadRequest)
		return
	}
	// JSON decoder -  updated data
	var bankDTO bank.BankAndUserDTO
	if err := json.NewDecoder(r.Body).Decode(&bankDTO); err != nil {
		http.Error(w, "Invalid input format", http.StatusBadRequest)
		return
	}
	// service call -- update(Bank & BankUser)
	if err := controller.BankService.UpdateBank(uint(bankID), bankDTO); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, fmt.Sprintf("Bank with ID %d not found", bankID), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Bank with ID %d and associated BankUser updated successfully", bankID)))
}
