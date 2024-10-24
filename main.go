package main

import (
	"bankManagement/app"
	"bankManagement/repository"
	"bankManagement/utils/log"
	"sync"

	"github.com/jinzhu/gorm"
)

func main() {
	name := "Bank Management App"
	log := NewLogger()
	db := NewDBConnection()
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
	appObj.StartServer()
}

func NewDBConnection() *gorm.DB {

	return nil
}

func NewLogger() log.WebLogger {
	return log.GetLogger()
}

func NewWaitGroup() *sync.WaitGroup {
	return &sync.WaitGroup{}
}

func NewRepository() *repository.Repository {
	return repository.NewRepository()
}
