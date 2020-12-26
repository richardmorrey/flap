package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/richardmorrey/flap/pkg/model"	
	"github.com/richardmorrey/flap/pkg/flap"
	"fmt"
	"time"
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
			w.Header().Set("Content-Type", "text/html")
			err := self.engine.Reset(true)
			if err != nil {
				logError(err)
				http.Error(w, fmt.Sprintf("\nFailed to destroy model with error '%s'\n",err), http.StatusInternalServerError)
			}
			w.Write([]byte("Done"))
		})

	api.HandleFunc("/reset",
		func (w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			err := self.engine.Reset(false)
			if err != nil {
				logError(err)
				http.Error(w, fmt.Sprintf("\nFailed to reset model with error '%s'\n",err), http.StatusInternalServerError)
			}
			w.Write([]byte("Done"))
		})

	api.HandleFunc("/build",
		func (w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			err := self.engine.Build()
			if err != nil {
				logError(err)
				http.Error(w, fmt.Sprintf("\nFailed to build model with error '%s'\n",err), http.StatusInternalServerError)
			}
			w.Write([]byte("Done"))
		})

	api.HandleFunc("/warm",
		func (w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			err := self.engine.Run(true)
			if err != nil {
				logError(err)
				http.Error(w, fmt.Sprintf("\nFailed to warm model with error '%s'\n",err), http.StatusInternalServerError)
			}
			w.Write([]byte("Done"))
		})
	
	api.HandleFunc("/runoneday",
		func (w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			dayToRun := flap.EpochTime(time.Now().Unix())
			dayToRun -= dayToRun % flap.SecondsInDay
			err := self.engine.RunOneDay(dayToRun)
			if err != nil {
				logError(err)
				http.Error(w, fmt.Sprintf("\nFailed to run model for %s  with error '%s'\n",dayToRun.ToTime(),err), http.StatusInternalServerError)
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

