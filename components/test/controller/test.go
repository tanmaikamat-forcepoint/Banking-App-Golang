package controller

import (
	"bankManagement/components/test/service"
	"bankManagement/middlewares/auth"
	"bankManagement/utils/log"
	"net/http"

	"github.com/gorilla/mux"
)

type TestController struct {
	TestService *service.TestService
	log         log.WebLogger
}

func NewTestController(
	TestServcice *service.TestService,
	log log.WebLogger,
) *TestController {
	return &TestController{
		log:         log,
		TestService: TestServcice,
	}

}

func (controller *TestController) RegisterRoutes(router *mux.Router) {
	subRouter := router.PathPrefix("/test").Subrouter()
	subRouter.Use(auth.AuthenticationMiddleware)
	subRouter.HandleFunc("/", controller.TestApi).Methods(http.MethodGet)
	subRouter.HandleFunc("/", controller.TestApi).Methods(http.MethodPost)
}

func (controller *TestController) TestApi(w http.ResponseWriter, r *http.Request) {

}
