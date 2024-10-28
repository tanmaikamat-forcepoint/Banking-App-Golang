package modules

import (
	"bankManagement/app"
	"bankManagement/components/bank/controller"
	"bankManagement/components/bank/service"
)

func RegisterBankModule(appObj *app.App) {
	userService := service.NewBankService(appObj.DB, appObj.Repository, appObj.Log)

	userController := controller.NewBankController(userService, appObj.Log)

	userController.RegisterRoutes(appObj.Router)
}
