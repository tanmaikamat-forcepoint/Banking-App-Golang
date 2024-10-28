package controller

import (
	"bankManagement/components/client/service"
	errorsUtils "bankManagement/utils/errors"
	"bankManagement/utils/log"
	"net/http"

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
	subRouter.HandleFunc("/beneficiaries", ctrl.Todo).Methods(http.MethodPost)
	subRouter.HandleFunc("/beneficiaries", ctrl.Todo).Methods(http.MethodGet)
	subRouter.HandleFunc("/beneficiaries/{beneficiary_id}", ctrl.Todo).Methods(http.MethodGet)
	subRouter.HandleFunc("/beneficiaries/{beneficiary_id}", ctrl.Todo).Methods(http.MethodPut)
	subRouter.HandleFunc("/beneficiaries/{beneficiary_id}", ctrl.Todo).Methods(http.MethodDelete)
	subRouter.HandleFunc("/beneficiaries/{beneficiary_id}/make_payment", ctrl.Todo).Methods(http.MethodPost)
}

func (ctrl *PaymentController) Todo(w http.ResponseWriter, r *http.Request) {
	errorsUtils.SendErrorWithCustomMessage(w, "Route Not Implemented", 400)
}
