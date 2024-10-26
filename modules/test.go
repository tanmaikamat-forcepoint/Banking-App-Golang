package modules

import (
	"bankManagement/app"
	"bankManagement/components/test/controller"
	"bankManagement/components/test/service"
)

func RegisterTestModule(appObj *app.App) {
	TestService := service.NewTestService(appObj.DB, appObj.Repository, appObj.Log)
	testContoller := controller.NewTestController(TestService, appObj.Log)
	testContoller.RegisterRoutes(appObj.Router)
}
