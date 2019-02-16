package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/battlesnakeio/exporter/handlers"
	"github.com/gorilla/mux"
)

//  main function
func main() {
	router := mux.NewRouter()
	handlers.SetupRoutes(router)
	fmt.Println("Serving on 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
