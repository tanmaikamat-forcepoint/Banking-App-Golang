package controller

import (
	"bankManagement/components/user/service"
	"bankManagement/utils/log"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type UserController struct {
	UserService *service.UserService
	log         log.WebLogger
}

func NewUserController(UserServcice *service.UserService, log log.WebLogger) *UserController {
	return &UserController{
		log:         log,
		UserService: UserServcice,
	}

}

func (controller *UserController) RegisterRoutes(router *mux.Router) {
	userRouter := router.PathPrefix("/users").Subrouter()
	userRouter.HandleFunc("/createSuperAdmin", controller.CreateSuperAdmin).Methods("POST")
}

func (controller *UserController) CreateSuperAdmin(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if requestData.Username == "" || requestData.Password == "" || requestData.Email == "" || requestData.Name == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// service call to create a SuperAdmin
	err := controller.UserService.CreateSuperAdmin(requestData.Username, requestData.Password, requestData.Name, requestData.Email)
	if err != nil {
		if err.Error() == "SuperAdmin already exists" {
			http.Error(w, "Only one SuperAdmin can exist in the system", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("SuperAdmin created successfully"))
}
