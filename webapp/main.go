package main

import (
	"log"
	"os"
	"errors"
	"net/http"
	"time"
	//"fmt"
	"flag"
	"github.com/gorilla/mux"
)

var ERESTAPIMODENOTSPECIFIED = errors.New("Rest API mode not specified.")

// main is main
func main() {
	
	// Parse command-line
	configfile := flag.String("configfile","./config.yaml","File path of yaml config file to use")
	flag.Parse()

	// Create logger
	NewLogger(llInfo)

	// Create top level router
	r := mux.NewRouter()

	// Initialize REST API
	var api restAPI
	err := ERESTAPIMODENOTSPECIFIED
	switch flag.Arg(0){

		case "admin":
			api = new(adminRestAPI)

		break

		case "user":
			api = new(userRestAPI)
		break

		default:
		break

	}
	if api != nil {
		err = api.init(r,*configfile)
	}

	// Exit on failure
	if err != nil {
		logError(err)
		os.Exit(0)
	}

	// Static content
	r.PathPrefix("/app/").Handler(http.FileServer(http.Dir(".")))
	r.PathPrefix("/doc/").Handler(http.FileServer(http.Dir("..")))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../website")))

	// Start serving up content
	srv := &http.Server{Handler:r,Addr:":8080",WriteTimeout: 15 * time.Second,ReadTimeout:  15 * time.Second}
	log.Fatal(srv.ListenAndServe())
}

