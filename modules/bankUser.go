package modules

import (
	"bankManagement/app"
	"bankManagement/components/bankUser/controller"
	"bankManagement/components/bankUser/service"
)

func RegisterBankUserModule(appObj *app.App) {

	userService := service.NewBankUserService(appObj.DB, appObj.Repository, appObj.Log)
	userController := controller.NewBankUserController(userService, appObj.Log)
	userController.RegisterRoutes(appObj.Router)
}
