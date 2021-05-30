package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kaitolucifer/user-balance-management/injector"
	"github.com/kaitolucifer/user-balance-management/presentation"
)

var dbHost = flag.String("dbhost", "localhost", "database host")
var dbName = flag.String("dbname", "user_balance", "database name")
var dbUser = flag.String("dbuser", "admin", "database user name")
var dbPassword = flag.String("dbpass", "password", "database password")
var dbPort = flag.String("dbport", "5432", "database port number")
var dbSSL = flag.String("dbssl", "disable", "use database ssl tunnel or not")

type App struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	handler  http.Handler
}

func NewApp() *App {
	flag.Parse()
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		*dbHost, *dbPort, *dbName, *dbUser, *dbPassword, *dbSSL)
	db := injector.InjectDatabase(dsn)
	repo := injector.InjectRepository(db)
	usecase := injector.InjectUsecase(repo)
	handler := injector.InjectHandler(usecase)

	var app App

	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.handler = presentation.Routes(handler)
	return &app
}
