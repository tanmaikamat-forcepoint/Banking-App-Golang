package controller

import (
	"bankManagement/components/client/service"
	"bankManagement/constants"
	"bankManagement/middlewares/auth"
	"bankManagement/models/beneficiary"
	"bankManagement/models/payments"
	"bankManagement/utils/encrypt"
	errorsUtils "bankManagement/utils/errors"
	"bankManagement/utils/log"
	"bankManagement/utils/web"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PaymentController struct {
	PaymentService *service.PaymentService
	log            log.WebLogger
}

func NewPaymentController(
	PaymentService *service.PaymentService,
	log log.WebLogger,
) *PaymentController {
	return &PaymentController{
		PaymentService: PaymentService,
		log:            log,
	}
}

func (ctrl *PaymentController) RegisterRoutes(
	router *mux.Router) {
	subRouter := router.PathPrefix("/clients/{client_id}/").Subrouter()
	subRouter.Use(auth.AuthenticationMiddleware, auth.ValidateClientPermissionsMiddleware)
	subRouter.HandleFunc("/beneficiaries", ctrl.CreateBeneficiary).Methods(http.MethodPost)
	subRouter.HandleFunc("/beneficiaries", ctrl.GetAllBeneficiaries).Methods(http.MethodGet)
	subRouter.HandleFunc("/beneficiaries/{beneficiary_id}", ctrl.Todo).Methods(http.MethodGet)
	subRouter.HandleFunc("/beneficiaries/{beneficiary_id}", ctrl.Todo).Methods(http.MethodPut)
	subRouter.HandleFunc("/beneficiaries/{beneficiary_id}", ctrl.DeleteBeneficiaryById).Methods(http.MethodDelete)
	subRouter.HandleFunc("/make_payment", ctrl.CreatePaymentRequest).Methods(http.MethodPost)
	subRouter.HandleFunc("/payments_requests", ctrl.GetAllPaymentRequestForClient).Methods(http.MethodGet)
	subRouter.HandleFunc("/payments", ctrl.GetAllPaymentRequestForClient).Methods(http.MethodGet)
}

func (ctrl *PaymentController) Todo(w http.ResponseWriter, r *http.Request) {
	errorsUtils.SendErrorWithCustomMessage(w, "Route Not Implemented", 400)
}

func (ctrl *PaymentController) CreateBeneficiary(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	client_id := claims.ClientId
	user_id := claims.UserId
	if client_id == 0 {
		errorsUtils.SendErrorWithCustomMessage(w, "Invalid JWT Token for Access", http.StatusBadRequest)
		return
	}

	tempBeneficiary := beneficiary.Beneficiary{}
	err := web.UnMarshalJSON(r, &tempBeneficiary)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	err = web.GetValidator().Struct(tempBeneficiary)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, web.GetValidationError(err), http.StatusBadRequest)
		return
	}
	tempBeneficiary.ClientID = client_id
	tempBeneficiary.CreatedByUserID = user_id
	tempBeneficiary.ApprovedByUserID = user_id

	err = ctrl.PaymentService.CreateBeneficiary(&tempBeneficiary)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	web.SendResponse(w, web.WebResponse{
		StatusCode: http.StatusCreated,
		Message:    "Beneficiary Successfully Created",
		Data:       tempBeneficiary,
	})
}

func (ctrl *PaymentController) GetAllBeneficiaries(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	client_id := claims.ClientId
	if client_id == 0 {
		errorsUtils.SendErrorWithCustomMessage(w, "Invalid JWT Token for Access", http.StatusBadRequest)
		return
	}
	var allBeneficiaries []beneficiary.Beneficiary
	err := ctrl.PaymentService.GetAllBeneficiariesForClient(client_id, &allBeneficiaries)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusBadRequest)
		return
	}
	web.SendResponse(w, web.WebResponse{
		StatusCode: http.StatusOK,
		Message:    "Beneficiary Successfully  Retrieved",
		Data:       allBeneficiaries,
	})
}

func (ctrl *PaymentController) DeleteBeneficiaryById(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	client_id := claims.ClientId
	// user_id := claims.UserId
	if client_id == 0 {
		errorsUtils.SendErrorWithCustomMessage(w, "Invalid JWT Token for Access", http.StatusBadRequest)
		return
	}

	tempParser := web.NewParser(r)
	id, ok := tempParser.Params["beneficiary_id"]
	if !ok {
		errorsUtils.SendErrorWithCustomMessage(w, "Malformed URL: Beneficiary Id Not Found", http.StatusBadRequest)
		return
	}
	beneficiary_id, err := strconv.Atoi(id)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, "Malformed URL: Beneficiary Id Should be Integer", http.StatusBadRequest)
		return
	}

	err = ctrl.PaymentService.DeleteBeneficiaryById(client_id, uint(beneficiary_id))
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	web.SendResponse(w, web.WebResponse{
		StatusCode: http.StatusAccepted,
		Message:    "Beneficiary Successfully Deleted",
	})
}

func (ctrl *PaymentController) CreatePaymentRequest(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	client_id := claims.ClientId
	user_id := claims.UserId
	if client_id == 0 {
		errorsUtils.SendErrorWithCustomMessage(w, "Invalid JWT Token for Access", http.StatusBadRequest)
		return
	}

	paymentRequestDTO := payments.PaymentRequestDTO{}
	err := web.UnMarshalJSON(r, &paymentRequestDTO)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	err = web.GetValidator().Struct(paymentRequestDTO)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, web.GetValidationError(err), http.StatusBadRequest)
		return
	}

	err = ctrl.PaymentService.CreatePaymentRequest(client_id, user_id, &paymentRequestDTO)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	web.SendResponse(w, web.WebResponse{
		StatusCode: http.StatusCreated,
		Message:    "Payment Request Successfully Created",
		Data:       paymentRequestDTO,
	})
}

func (ctrl *PaymentController) GetAllPaymentRequestForClient(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(constants.ClaimKey).(*encrypt.Claims)
	client_id := claims.ClientId
	if client_id == 0 {
		errorsUtils.SendErrorWithCustomMessage(w, "Invalid JWT Token for Access", http.StatusBadRequest)
		return
	}
	var allPaymentRequest []payments.PaymentRequest
	err := ctrl.PaymentService.GetAllPaymentRequest(client_id, &allPaymentRequest)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	web.SendResponse(w, web.WebResponse{
		StatusCode: http.StatusCreated,
		Message:    "Payment Request Successfully retrieved",
		Data:       allPaymentRequest,
	})
}
