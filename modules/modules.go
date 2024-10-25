package modules

import (
	"bankManagement/app"
	"bankManagement/models/bank"
	"bankManagement/models/beneficiary"
	"bankManagement/models/client"
	"bankManagement/models/employee"
	"bankManagement/models/transaction"
	"bankManagement/models/user"

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
<<<<<<< HEAD
	// RegisterAllConfigs(appObj)

}

func RegisterTableMigrations(appObj *app.App) {
	userConfig := user.UserConfig{DB: appObj.DB}
	paymentConfig := payments.PaymentConfig{DB: appObj.DB}
	transactionConfig := transaction.TransactionConfig{DB: appObj.DB}
	salaryDisbursementConfig := salaryDisbursement.SalaryDisbursementConfig{DB: appObj.DB}
	documentConfig := document.DocumentConfig{DB: appObj.DB}

	RegisterAllConfigs(appObj, []ModuleConfig{&userConfig, &paymentConfig, &transactionConfig, &salaryDisbursementConfig, &documentConfig})
=======
}

func RegisterTableMigrations(appObj *app.App) {

	roleConfig := user.RoleConfig{DB: appObj.DB}
	userConfig := user.UserConfig{DB: appObj.DB}
	clientConfig := client.ClientConfig{DB: appObj.DB}
	clientUserConfig := client.ClientUserConfig{DB: appObj.DB}
	bankConfig := bank.BankConfig{DB: appObj.DB}
	bankUserConfig := bank.BankUserConfig{DB: appObj.DB}
	employeeConfig := employee.EmployeeConfig{DB: appObj.DB}
	transactionConfig := transaction.TransactionConfig{DB: appObj.DB}
	beneficiaryConfig := beneficiary.BeneficiaryConfig{DB: appObj.DB}

	// Register each module's configuration in the correct order
	RegisterAllConfigs(appObj, []ModuleConfig{
		&roleConfig,
		&userConfig,
		&clientConfig,
		&clientUserConfig,
		&bankConfig,
		&bankUserConfig,
		&employeeConfig,
		&transactionConfig,
		&beneficiaryConfig,
	})

>>>>>>> d304374cac5be24d0b51881eb4113c755e132b92
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
