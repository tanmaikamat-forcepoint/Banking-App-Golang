package modules

import (
	"bankManagement/app"
	"bankManagement/components/user/controller"
	"bankManagement/components/user/service"
)

func RegisterUserModule(appObj *app.App) {
	userService := service.NewUserService(appObj.DB, appObj.Repository, appObj.Log)
	userController := controller.NewUserController(userService, appObj.Log)
	userController.RegisterRoutes(appObj.Router)
}
