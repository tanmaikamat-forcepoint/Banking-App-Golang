package app

import (
	"bankManagement/constants"
	"bankManagement/repository"
	"bankManagement/utils/log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type App struct {
	sync.Mutex
	Name       string
	Router     *mux.Router
	Server     *http.Server
	DB         *gorm.DB
	Log        log.WebLogger
	WG         *sync.WaitGroup
	Repository *repository.Repository
}

func NewApp(name string, db *gorm.DB, log log.WebLogger, wg *sync.WaitGroup, repository *repository.Repository) *App {
	return &App{
		Name:       name,
		DB:         db,
		Log:        log,
		WG:         wg,
		Repository: repository,
	}
}

func (app *App) Init() {
	app.initializeRouter()
	app.initializeServer()
}

func (app *App) StartServer() error {
	err := app.Server.ListenAndServe()
	if err != nil {
		app.Log.Error(err.Error())
		return err
	}
	return nil
}

func (app *App) initializeRouter() {
	app.Log.Info(app.Name + " App Route initializing")
	app.Router = mux.NewRouter().StrictSlash(true)
	app.Router = app.Router.PathPrefix(constants.APIPrefix).Subrouter()
}
func (app *App) initializeServer() {
	headers := handlers.AllowedHeaders([]string{
		"Content-Type", "X-Total-Count", "token",
	})
	methods := handlers.AllowedMethods([]string{
		http.MethodPost, http.MethodPut, http.MethodGet, http.MethodDelete, http.MethodOptions,
	})
	originOption := handlers.AllowedOriginValidator(app.checkOrigin)
	app.Server = &http.Server{
		Addr:         constants.ServerAddress,
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
		IdleTimeout:  time.Second * 60,
		Handler:      handlers.CORS(headers, methods, originOption)(app.Router),
	}
	app.Log.Info("Server Exposed On 4000")
}

func (app *App) checkOrigin(origin string) bool {
	// origin will be the actual origin from which the request is made.

	return true
}
