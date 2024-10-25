package modules

import (
	"bankManagement/app"

	"github.com/gorilla/mux"
)

type Controller interface {
	RegisterRoutes(router *mux.Router)
}

type ModuleConfig interface {
	TableMigration()
}

func RegisterAllModules(appObj *app.App) {
	RegisterUserModule(appObj)
	// RegisterAllConfigs(appObj)

}

func RegisterAllRoutes(appObj *app.App, controllers []Controller) {
	for i := 0; i < len(controllers); i++ {
		controllers[i].RegisterRoutes(appObj.Router)
	}
}

func RegisterAllConfigs(appObj *app.App, configs []ModuleConfig) {
	for i := 0; i < len(configs); i++ {
		configs[i].TableMigration()
	}
}
