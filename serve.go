package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/battlesnakeio/exporter/handlers"
	"github.com/gorilla/mux"
)

//  main function
func main() {
	router := mux.NewRouter()
	handlers.SetupRoutes(router)
	port := "8000"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	fmt.Printf("Serving on %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
