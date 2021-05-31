package main

import (
	"net/http"
)

const portNumber = ":8080"

func main() {
	configApp()
	defer db.Close()
	handler.App.InfoLog.Printf("starting application on port %s\n", portNumber)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: mux,
	}

	handler.App.ErrorLog.Fatal(srv.ListenAndServe())
}
