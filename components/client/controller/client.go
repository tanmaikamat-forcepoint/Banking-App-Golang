package controller

import (
	"bankManagement/components/client/service"
	"bankManagement/constants"
	"bankManagement/middlewares/auth"
	"bankManagement/models/employee"
	"bankManagement/models/reports"
	"bankManagement/utils/encrypt"
	errorsUtils "bankManagement/utils/errors"
	"bankManagement/utils/log"
	"bankManagement/utils/web"
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
	subRouter.Use(auth.AuthenticationMiddleware, auth.ValidateClientPermissionsMiddleware)
	subRouter.HandleFunc("/employees", ctrl.CreateEmployee).Methods(http.MethodPost)
	subRouter.HandleFunc("/employees", ctrl.GetAllEmployees).Methods(http.MethodGet)
	subRouter.HandleFunc("/employees/{employee_id}", ctrl.GetAllEmployees).Methods(http.MethodGet)
	subRouter.HandleFunc("/employees/{employee_id}", ctrl.UpdateEmployee).Methods(http.MethodPut)
	subRouter.HandleFunc("/employees/{employee_id}", ctrl.DeleteEmployeeById).Methods(http.MethodDelete)
	subRouter.HandleFunc("/disburse_salary", ctrl.DisburseSalary).Methods(http.MethodPost)
	subRouter.HandleFunc("/reports/salary_report", ctrl.GetSalaryReport).Methods(http.MethodPost)
	subRouter.HandleFunc("/reports/payment_report", ctrl.GetPaymentReport).Methods(http.MethodPost)
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
	web.SendResponse(w,
		web.WebResponse{
			StatusCode: http.StatusCreated,
			Message:    "Employee Created Successfully",
			Data:       employeeDetails,
		})

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

	web.SendResponse(w,
		web.WebResponse{
			StatusCode: http.StatusAccepted,
			Message:    "Employee Update Successfully",
			Data:       employeeDetails,
		})
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
	web.SendResponse(w,
		web.WebResponse{
			StatusCode: http.StatusOK,
			Message:    "Employee Deleted Successfully",
		})
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
	web.SendResponse(w,
		web.WebResponse{
			StatusCode: http.StatusOK,
			Message:    "Employee Retrieved Successfully",
			Data:       employeeDetails,
		})
	// json.NewEncoder(w).Encode(employeeDetails)

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
	web.SendResponse(w,
		web.WebResponse{
			StatusCode: http.StatusCreated,
			Message:    "Salary Disbursed Successfully",
		})

}

func (ctrl *ClientController) GetSalaryReport(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	client_id := claims.ClientId
	if client_id == 0 {
		errorsUtils.SendErrorWithCustomMessage(w, "Invalid JWT Token for Access", http.StatusBadRequest)
	}
	var salaryReport reports.SalaryReport
	err := ctrl.ClientService.GetSalaryReport(uint(client_id), &salaryReport)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusBadRequest)
		return
	}
	web.SendResponse(w,
		web.WebResponse{
			StatusCode: http.StatusCreated,
			Message:    "Salary Report Generated",
			Data:       salaryReport,
		})

}

func (ctrl *ClientController) GetPaymentReport(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	client_id := claims.ClientId
	if client_id == 0 {
		errorsUtils.SendErrorWithCustomMessage(w, "Invalid JWT Token for Access", http.StatusBadRequest)
	}
	var paymentReport reports.PaymentReport
	err := ctrl.ClientService.GetPaymentReport(uint(client_id), &paymentReport)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusBadRequest)
		return
	}
	web.SendResponse(w,
		web.WebResponse{
			StatusCode: http.StatusCreated,
			Message:    "Payment Report Generated",
			Data:       paymentReport,
		})

}
