package main

import (
	"net/http"
)

const portNumber = ":8080"

func main() {
	app := NewApp()
	app.InfoLog.Printf("Starting application on port %s\n", portNumber)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: app.handler,
	}

	app.ErrorLog.Fatal(srv.ListenAndServe())
}
