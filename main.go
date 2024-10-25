package main

import (
	"bankManagement/app"
	"bankManagement/modules"
	"bankManagement/repository"
	"bankManagement/utils/log"
	"os"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	name := "Bank Management App"
	log := NewLogger()
	db := NewDBConnection(log)
	if db == nil {
		log.Error("DB connection Failed")
	}
	defer func() {
		db.Close()
		log.Error("Database Closed")
	}()
	wg := NewWaitGroup()
	repo := NewRepository()
	appObj := app.NewApp(name, db, log, wg, repo)
	appObj.Init()
	modules.RegisterTableMigrations(appObj)
	modules.RegisterAllModules(appObj)
	appObj.StartServer()
}

func NewDBConnection(log log.WebLogger) *gorm.DB {
	host, b1 := os.LookupEnv("HOST")
	user, b2 := os.LookupEnv("USER")
	password, b3 := os.LookupEnv("PASSWORD")
	database, b4 := os.LookupEnv("DATABASE")
	if !(b1 && b2 && b3 && b4) {
		panic("Error: Enviroment Variables not Initialized")
	}
	db, err := gorm.Open("mysql", user+":"+password+"@("+host+")/"+database+"?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	log.Info("Database Connected")
	db.LogMode(true)
	return db
}

func NewLogger() log.WebLogger {
	return log.GetLogger()
}

func NewWaitGroup() *sync.WaitGroup {
	return &sync.WaitGroup{}
}

func NewRepository() repository.Repository {
	return repository.NewGormRepositoryMySQL()
}
