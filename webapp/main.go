package main

import (
	"io"
	"log"
	"os"
	"net/http"
)

// handleStatic services requests to retrieve static content from the "app" folder
func handleStatic(w http.ResponseWriter, r *http.Request) {
	page, err := os.Open(r.URL.Path[1:])
	if (err != nil) {
		return
	}
	io.Copy(w,page)
}

// main is main
func main() {
	http.HandleFunc("/app/", handleStatic)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
