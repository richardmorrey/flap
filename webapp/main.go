package main

import (
	"log"
	"net/http"
)

// main is main
func main() {
	
	// Handler for app static content requests
	appfs := http.FileServer(http.Dir("."))
	http.Handle("/app/", appfs)

	// Handler for website requests
	websitefs := http.FileServer(http.Dir("../website"))
	http.Handle("/", websitefs)

	// Start serving up content
	log.Fatal(http.ListenAndServe(":8080", nil))
}

