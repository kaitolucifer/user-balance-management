package main

import (
	"log"
	"net/http"
)

const portNumber = ":8080"

func main() {
	handler := NewHandler()
	log.Printf("Starting application on port %s\n", portNumber)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: handler,
	}

	log.Fatal(srv.ListenAndServe())
}
