package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/richardmorrey/flap/pkg/model"
	"fmt"
)

type restAPI interface {
	init(*mux.Router,string) error
	release() error
}

type adminRestAPI struct {
	engine *model.Engine
}

// init configures handlers for all methods of the admin rest api
func (self *adminRestAPI) init(r *mux.Router,configfile string) error {
	var err error

	api := r.PathPrefix("/admin/v1").Subrouter()
	
	self.engine,err = model.NewEngine(configfile)
	if err != nil {
		return logError(err)
	}

	api.HandleFunc("/destroy",
		func (w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/text")
			err := self.engine.Reset(true)
			if err != nil {
				logError(err)
				http.Error(w, fmt.Sprintf("\nFailed to destroy model with error '%s'\n",err), http.StatusInternalServerError)
			}
			w.Write([]byte("Done"))
		})

	return err
}	

// release releases all state associated with admin rest api
func (self *adminRestAPI) release() error {
	self.engine.Release()
	return nil
}


/*

	case "destroy":
			engine,err := model.NewEngine(*configfile)
			if err != nil {
				fmt.Printf("\nFailed to initialize model engine with error '%s'\n",err)
			} else {
				defer engine.Release()
	 			err := engine.Reset(true)
	 			if err != nil {
		 			fmt.Printf("\nFailed to build model with error '%s'\n",err)
				}
			}
		case "reset":
			engine,err := model.NewEngine(*configfile)
			if err != nil {
				fmt.Printf("\nFailed to initialize model engine with error '%s'\n",err)
			} else {
				defer engine.Release()
	 			err := engine.Reset(false)
	 			if err != nil {
		 			fmt.Printf("\nFailed to build model with error '%s'\n",err)
				}
			}
		case "build":
			engine,err := model.NewEngine(*configfile)
			if err != nil {
				fmt.Printf("\nFailed to initialize model engine with error '%s'\n",err)
			} else {
				defer engine.Release()
	 			err := engine.Build()
	 			if err != nil {
		 			fmt.Printf("\nFailed to build model with error '%s'\n",err)
				}
			}
		case "warm":
			engine,err := model.NewEngine(*configfile)
			if err != nil {
				fmt.Printf("\nFailed to initialize model engine with error '%s'\n",err)
			} else {
				defer engine.Release()

*/
