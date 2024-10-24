package main

import (
	"bankManagement/repository"
	"bankManagement/utils/log"
	"sync"

	"github.com/jinzhu/gorm"
)

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
