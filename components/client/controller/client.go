package controller

import (
	"bankManagement/components/client/service"
	"bankManagement/constants"
	"bankManagement/models/employee"
	"bankManagement/utils/encrypt"
	errorsUtils "bankManagement/utils/errors"
	"bankManagement/utils/log"
	"bankManagement/utils/web"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ClientController struct {
	ClientService *service.ClientService
	log           log.WebLogger
}

func NewClientController(
	ClientService *service.ClientService,
	log log.WebLogger,
) *ClientController {
	return &ClientController{
		ClientService: ClientService,
		log:           log,
	}
}

func (ctrl *ClientController) RegisterRoutes(
	router *mux.Router) {
	subRouter := router.PathPrefix("/clients/{client_id}/").Subrouter()
	subRouter.HandleFunc("/employees", ctrl.CreateEmployee).Methods(http.MethodPost)
	subRouter.HandleFunc("/employees", ctrl.GetAllEmployees).Methods(http.MethodGet)
	subRouter.HandleFunc("/employees/{employee_id}", ctrl.GetAllEmployees).Methods(http.MethodGet)
	subRouter.HandleFunc("/employees/{employee_id}", ctrl.UpdateEmployee).Methods(http.MethodPut)
	subRouter.HandleFunc("/employees/{employee_id}", ctrl.DeleteEmployeeById).Methods(http.MethodDelete)
	subRouter.HandleFunc("/disburse_salary", ctrl.DisburseSalary).Methods(http.MethodPost)
}

func (ctrl *ClientController) Todo(w http.ResponseWriter, r *http.Request) {
	errorsUtils.SendErrorWithCustomMessage(w, "Route Not Implemented", 400)
}

func (ctrl *ClientController) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	employeeDetails := &employee.Employee{}
	err := web.UnMarshalJSON(r, employeeDetails)
	if err != nil {
		ctrl.log.Error(err)
		errorsUtils.SendInvalidBodyError(w)
		return
	}
	//validations

	id, ok := mux.Vars(r)["client_id"]
	if !ok {
		errorsUtils.SendErrorWithCustomMessage(w, "Client ID Not Found", http.StatusBadRequest)
		return
	}
	client_id, err := strconv.Atoi(id)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, "Client ID should be a int", http.StatusBadRequest)
		return
	}
	err = web.GetValidator().Struct(employeeDetails)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, web.GetValidationError(err), http.StatusBadRequest)
		return
	}

	employeeDetails.ClientID = uint(client_id)
	err = ctrl.ClientService.CreateEmployee(employeeDetails)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(employeeDetails)

}

func (ctrl *ClientController) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	employeeDetails := &employee.Employee{}
	err := web.UnMarshalJSON(r, employeeDetails)
	if err != nil {
		errorsUtils.SendInvalidBodyError(w)
		return
	}
	id, ok := mux.Vars(r)["employee_id"]
	if !ok {
		errorsUtils.SendErrorWithCustomMessage(w, "Employee  ID Not Found", http.StatusBadRequest)
		return
	}
	finalId, err := strconv.Atoi(id)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, "Employee ID should be a int", http.StatusBadRequest)
		return
	}
	employeeDetails.ID = uint(finalId)
	//validations
	err = web.GetValidator().Struct(employeeDetails)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, web.GetValidationError(err), http.StatusBadRequest)
	}
	err = ctrl.ClientService.Update(employeeDetails)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(employeeDetails)
}

func (ctrl *ClientController) DeleteEmployeeById(w http.ResponseWriter, r *http.Request) {

	id, ok := mux.Vars(r)["employee_id"]
	if !ok {
		errorsUtils.SendErrorWithCustomMessage(w, "Employee  ID Not Found", http.StatusBadRequest)
		return
	}
	finalId, err := strconv.Atoi(id)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, "Employee ID should be a int", http.StatusBadRequest)
		return
	}
	empId := uint(finalId)

	err = ctrl.ClientService.DeleteEmployeeById(empId)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

}

func (ctrl *ClientController) GetAllEmployees(w http.ResponseWriter, r *http.Request) {
	employeeDetails := []employee.Employee{}

	//validations
	id, ok := mux.Vars(r)["client_id"]
	if !ok {
		errorsUtils.SendErrorWithCustomMessage(w, "Client ID Not Found", http.StatusBadRequest)
		return
	}
	client_id, err := strconv.Atoi(id)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, "Client ID should be a int", http.StatusBadRequest)
		return
	}

	err = ctrl.ClientService.GetAllEmployeesByClientId(&employeeDetails, uint(client_id))
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, web.GetValidationError(err), http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(employeeDetails)

}

func (ctrl *ClientController) DisburseSalary(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	user_id := claims.UserId
	client_id := claims.ClientId
	if client_id == 0 || user_id == 0 {
		errorsUtils.SendErrorWithCustomMessage(w, "Invalid JWT Token for Access", http.StatusBadRequest)
	}

	err := ctrl.ClientService.DisburseSalaryAllEmployees(uint(client_id), uint(user_id))
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, web.GetValidationError(err), http.StatusBadRequest)
	}

}
