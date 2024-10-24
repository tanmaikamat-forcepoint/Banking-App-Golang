package controller

import (
	"bankManagement/components/user/service"
	"bankManagement/utils/log"

	"github.com/gorilla/mux"
)

type UserController struct {
	UserService *service.UserService
	log         log.WebLogger
}

func NewUserController(
	UserServcice *service.UserService,
	log log.WebLogger,
) *UserController {
	return &UserController{
		log:         log,
		UserService: UserServcice,
	}

}

func (controller *UserController) RegisterRoutes(router *mux.Router) {

}
