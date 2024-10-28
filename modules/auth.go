package modules

import (
	"bankManagement/app"
	"bankManagement/components/auth/controller"
	"bankManagement/components/auth/service"
)

func RegisterAuthModule(appObj *app.App) {
	authService := service.NewAuthService(appObj.DB, appObj.Repository, appObj.Log)
	authController := controller.NewAuthController(authService, appObj.Log)
	authController.RegisterRoutes(appObj.Router)
}
