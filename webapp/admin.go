package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/richardmorrey/flap/pkg/model"	
	"github.com/richardmorrey/flap/pkg/flap"
	"fmt"
	"time"
	"io"
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
			err := self.engine.Reset(true)
			if err != nil {
				logError(err)
				http.Error(w, fmt.Sprintf("\nFailed to destroy model with error '%s'\n",err), http.StatusInternalServerError)
				return
			}
			io.WriteString(w, "Done\n")
		})

	api.HandleFunc("/reset",
		func (w http.ResponseWriter, r *http.Request) {
			err := self.engine.Reset(false)
			if err != nil {
				logError(err)
				http.Error(w, fmt.Sprintf("\nFailed to reset model with error '%s'\n",err), http.StatusInternalServerError)
				return
			}
			io.WriteString(w, "Done\n")
		})

	api.HandleFunc("/build",
		func (w http.ResponseWriter, r *http.Request) {
			err := self.engine.Build()
			if err != nil {
				logError(err)
				http.Error(w, fmt.Sprintf("\nFailed to build model with error '%s'\n",err), http.StatusInternalServerError)
				return
			}
			io.WriteString(w, "Done\n")
		})

	api.HandleFunc("/warm/startday/{date}",
		func (w http.ResponseWriter, r *http.Request) {
			pathParams := mux.Vars(r)
			raw, ok := pathParams["date"]
			if !ok {
				http.Error(w,"Missing argument: startday", http.StatusForbidden)
				return
			}

		        startDayTime,err := time.Parse("2006-01-02",raw)
			if err != nil {
				http.Error(w,"Invalid argument: startdata", http.StatusInternalServerError)
				return
			}
			
			err = self.engine.Run(true,flap.EpochTime(startDayTime.Unix()))
			if err != nil {
				logError(err)
				http.Error(w, fmt.Sprintf("\nFailed to warm model with error '%s'\n",err), http.StatusInternalServerError)
				return
			}
			io.WriteString(w, "Done\n")
		})
	
	api.HandleFunc("/runoneday",
		func (w http.ResponseWriter, r *http.Request) {
			dayToRun := flap.EpochTime(time.Now().Unix())
			dayToRun -= dayToRun % flap.SecondsInDay
			err := self.engine.RunOneDay(dayToRun)
			if err != nil {
				logError(err)
				http.Error(w, fmt.Sprintf("\nFailed to run model for %s  with error '%s'\n",dayToRun.ToTime(),err), http.StatusInternalServerError)
				return
			}
			io.WriteString(w, "Done\n")
		})
	return err
}

// release releases all state associated with admin rest api
func (self *adminRestAPI) release() error {
	self.engine.Release()
	return nil
}

