package modules

import (
	"bankManagement/app"
	"bankManagement/models/bank"
	"bankManagement/models/beneficiary"
	"bankManagement/models/client"
	"bankManagement/models/document"
	"bankManagement/models/employee"
	"bankManagement/models/payments"
	"bankManagement/models/salaryDisbursement"
	"bankManagement/models/transaction"
	"bankManagement/models/user"
	"bankManagement/utils/encrypt"

	"github.com/gorilla/mux"
)

type Controller interface {
	RegisterRoutes(router *mux.Router)
}

type ModuleConfig interface {
	TableMigration()
}

func RegisterAllModules(appObj *app.App) {
	RegisterAuthModule(appObj)
	RegisterTestModule(appObj)
	RegisterUserModule(appObj)
	RegisterClientModule(appObj)
}

func RegisterTableMigrations(appObj *app.App) {

	roleConfig := user.RoleConfig{DB: appObj.DB}
	userConfig := user.UserConfig{DB: appObj.DB}
	bankConfig := bank.BankConfig{DB: appObj.DB}
	bankUserConfig := bank.BankUserConfig{DB: appObj.DB}
	clientConfig := client.ClientConfig{DB: appObj.DB}
	clientUserConfig := client.ClientUserConfig{DB: appObj.DB}
	employeeConfig := employee.EmployeeConfig{DB: appObj.DB}
	transactionConfig := transaction.TransactionConfig{DB: appObj.DB}
	beneficiaryConfig := beneficiary.BeneficiaryConfig{DB: appObj.DB}
	paymentConfig := payments.PaymentConfig{DB: appObj.DB}
	salaryDisbursementConfig := salaryDisbursement.SalaryDisbursementConfig{DB: appObj.DB}
	documentConfig := document.DocumentConfig{DB: appObj.DB}

	// Register each module's configuration in the correct order
	registerAllConfigs(appObj, []ModuleConfig{
		&roleConfig,
		&userConfig,
		&bankConfig,
		&bankUserConfig,
		&clientConfig,
		&clientUserConfig,
		&employeeConfig,
		&transactionConfig,
		&beneficiaryConfig,
		&documentConfig,
		&paymentConfig,
		&salaryDisbursementConfig,
	})

}

func SeedData(appObj *app.App) {
	data := []interface{}{
		// &user.Role{RoleName: "SuperAdmin"},
		// &user.Role{RoleName: "Bank User"},
		// &user.Role{RoleName: "Client User"},
		// &user.User{
		// 	Username: "bank_user",
		// 	Password: encrypt.HashPassword("admin@123"),
		// 	Email:    "bank@gmail.com",
		// 	IsActive: true,
		// 	Name:     "Bank User",
		// 	RoleID:   uint(encrypt.BankUserRoleID),
		// },
		// &user.User{
		// 	Username: "client_user",
		// 	Password: encrypt.HashPassword("admin@123"),
		// 	Email:    "client@gmail.com",
		// 	IsActive: true,
		// 	Name:     "Client User",
		// 	RoleID:   uint(encrypt.ClientUserRoleID),
		// },

		&user.User{
			Username: "client_user_tcs",
			Password: encrypt.HashPassword("admin@123"),
			Email:    "client_tcs@gmail.com",
			IsActive: true,
			Name:     "Client User TCS",
			RoleID:   uint(encrypt.ClientUserRoleID),
		},
	}

	for _, entry := range data {
		appObj.Log.Info(appObj.DB.Create(entry).Error)
	}
	appObj.Log.Info("Database Seeding Completed")
}

func registerAllRoutes(appObj *app.App, controllers []Controller) {
	for i := 0; i < len(controllers); i++ {
		controllers[i].RegisterRoutes(appObj.Router)
	}
}

func registerAllConfigs(appObj *app.App, configs []ModuleConfig) {
	for i := 0; i < len(configs); i++ {
		configs[i].TableMigration()
	}
}
