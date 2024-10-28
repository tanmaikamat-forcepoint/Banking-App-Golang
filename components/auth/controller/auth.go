package controller

import (
	"bankManagement/components/auth/service"
	"bankManagement/models/user"
	"bankManagement/utils/encrypt"
	errorsUtils "bankManagement/utils/errors"
	"bankManagement/utils/log"
	"bankManagement/utils/web"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type AuthController struct {
	AuthService *service.AuthService
	log         log.WebLogger
}

func NewAuthController(
	AuthService *service.AuthService,
	log log.WebLogger,
) *AuthController {
	return &AuthController{
		AuthService: AuthService,
		log:         log,
	}
}

func (ctrl *AuthController) RegisterRoutes(
	router *mux.Router) {
	subRouter := router.NewRoute().Subrouter()
	subRouter.HandleFunc("/login", ctrl.LoginApi).Methods(http.MethodPost)
	subRouter.HandleFunc("/register-admin", ctrl.RegisterAdmin).Methods(http.MethodPost)
}

func (ctrl *AuthController) LoginApi(w http.ResponseWriter, r *http.Request) {
	//validations
	loginCreds := &user.UserLoginParamDTO{}
	err := web.UnMarshalJSON(r, loginCreds)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), 400)
		return
	}
	err = web.GetValidator().Struct(loginCreds)

	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, web.GetValidationError(err), 400)
		return
	}
	//
	var authenticatedUser = &user.User{}
	err = ctrl.AuthService.LoginRequest(loginCreds, authenticatedUser)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), 400)
		return
	}
	token, err := encrypt.GetJwtFromData(authenticatedUser.ID, authenticatedUser.RoleID)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), 400)
		return
	}
	w.Header().Set("authorization", token)
	json.NewEncoder(w).Encode(authenticatedUser)
}

func (ctrl *AuthController) RegisterAdmin(w http.ResponseWriter, r *http.Request) {
	//validations
	adminDetails := &user.User{}
	err := web.UnMarshalJSON(r, adminDetails)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), 400)
		return
	}
	err = web.GetValidator().Struct(adminDetails)

	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, web.GetValidationError(err), 400)
		return
	}
	err = ctrl.AuthService.CreateNewAdmin(adminDetails)
	if err != nil {
		errorsUtils.SendErrorWithCustomMessage(w, err.Error(), 400)
		return
	}

}
