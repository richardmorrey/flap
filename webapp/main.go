package main

import (
	"log"
	"net/http"
	//"github.com/gorilla/handlers"
	//"os"
)

// main is main
func main() {
	
	// Handler for app static content requests
	appfs := http.FileServer(http.Dir("."))
	http.Handle("/app/", appfs)
	
	// Handler for document download requests
	docfs := http.FileServer(http.Dir(".."))
	http.Handle("/doc/",docfs)

	// Handler for website requests
	websitefs := http.FileServer(http.Dir("../website"))
	http.Handle("/", websitefs)

	// Start serving up content
	log.Fatal(http.ListenAndServe(":8080", nil))
}

