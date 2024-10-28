package modules

import (
	"bankManagement/app"
	"bankManagement/components/client/controller"
	"bankManagement/components/client/service"
)

func RegisterClientModule(appObj *app.App) {
	clientService := service.NewClientService(appObj.DB, appObj.Repository, appObj.Log)
	ClientController := controller.NewClientController(clientService, appObj.Log)
	ClientController.RegisterRoutes(appObj.Router)

	payementservice := service.NewPaymentService(appObj.DB, appObj.Repository, appObj.Log)
	paymentController := controller.NewPaymentController(payementservice, appObj.Log)
	paymentController.RegisterRoutes(appObj.Router)
}
