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

func NewHandler() http.Handler {
	flag.Parse()
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		*dbHost, *dbPort, *dbName, *dbUser, *dbPassword, *dbSSL)
	db := injector.InjectDatabase(dsn)
	repo := injector.InjectRepository(db)
	usecase := injector.InjectUsecase(repo)

	var app presentation.App

	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	handler := injector.InjectHandler(usecase, &app)
	return presentation.Routes(handler)
}
