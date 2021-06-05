package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kaitolucifer/user-balance-management/infrastructure"
	"github.com/kaitolucifer/user-balance-management/injector"
	GrpcHandler "github.com/kaitolucifer/user-balance-management/presentation/grpc"
	RestfulHandler "github.com/kaitolucifer/user-balance-management/presentation/restful"
)

var useGrpc = flag.Bool("use_grpc", true, "true to use gRPC API and false to use normal RESTful API")

// DB設定
var dbHost = flag.String("dbhost", "localhost", "database host")
var dbName = flag.String("dbname", "user_balance", "database name")
var dbUser = flag.String("dbuser", "admin", "database user name")
var dbPassword = flag.String("dbpass", "password", "database password")
var dbPort = flag.String("dbport", "5432", "database port number")
var dbSSL = flag.String("dbssl", "disable", "use database ssl tunnel or not")

var db infrastructure.DB
var restfulHandler *RestfulHandler.RestfulUserBalanceHandler
var grpcHandler *GrpcHandler.GrpcUserBalanceHander
var mux http.Handler

func configApp() {
	flag.Parse()

	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		*dbHost, *dbPort, *dbName, *dbUser, *dbPassword, *dbSSL)
	db = injector.InjectDatabase(dsn)
	repo := injector.InjectRepository(db)
	usecase := injector.InjectUsecase(repo)

	if *useGrpc {
		app := new(GrpcHandler.App)
		app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
		app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
		grpcHandler = injector.InjectGrpcHandler(usecase, app)
	} else {
		app := new(RestfulHandler.App)
		app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
		app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
		restfulHandler = injector.InjectRestfulHandler(usecase, app)
		mux = RestfulHandler.Routes(restfulHandler)
	}
}
