package main

import (
	"log"
	"net/http"
	"encoding/json"
	"fmt"
	"io"
	oidc "github.com/coreos/go-oidc"
	"golang.org/x/net/context"
	"github.com/gorilla/mux"
	"github.com/richardmorrey/flap/pkg/flap"
	"github.com/richardmorrey/flap/pkg/model"
	"os"
	"strconv"
	"errors"
	"time"
)

var EMISSINGARGUMENT= errors.New("Missing Argument")

var (
	clientID     = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
)

type userRestAPI struct {
	engine *model.Engine
}

// init configures handlers for all user rest api methods
func (self *userRestAPI) init(r *mux.Router,configfile string) error {
	var err error

	self.engine,err = model.NewEngine(configfile)
	if err != nil {
		return logError(err)
	}

	api := r.PathPrefix("/user/v1").Subrouter()

	api.HandleFunc("/flighthistory/id/{token}/b/{band}/n/{number}", self.flightHistory).Methods(http.MethodGet)
	api.HandleFunc("/transactions/id/{token}/b/{band}/n/{number}", self.transactions).Methods(http.MethodGet)
	api.HandleFunc("/promises/id/{token}/b/{band}/n/{number}", self.promises).Methods(http.MethodGet)
	api.HandleFunc("/account/id/{token}/b/{band}/n/{number}", self.account).Methods(http.MethodGet)
	api.HandleFunc("/dailystats/id/{token}", self.dailyStats)
	api.Use(middlewareIdToken)

	return nil
}

// release releases all resources associated with user rest api
func (self  *userRestAPI) release() error {
	self.engine.Release()
	return nil
}
			
// MiddlewareIdToken function validates openid id token
func middlewareIdToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		 	pathParams := mux.Vars(r)
			if raw, ok := pathParams["token"]; !ok {
				http.Error(w,"Missing argument: id", http.StatusForbidden)
			} else {
				_,err:= validateIDToken(raw)
				if err == nil {
					log.Printf("Authenticated user\n")
					next.ServeHTTP(w, r)
				} else {
					http.Error(w, "Forbidden", http.StatusForbidden)
				}
			}
	})
}

// validateIDToken validates and parses a raw openid id token
func validateIDToken(rawIDToken string) (string,error) {
	
	// Create verifier
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return "",err
	}
	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	// Verify id token
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "",err
	}

	// Parse token to JSON
	parsed := new(json.RawMessage)
	if err := idToken.Claims(parsed); err != nil {
		return "",err
	}

	// Render json as string
	data, err := json.MarshalIndent(parsed, "", "    ")
	if err != nil {
		return  "",err
	}
	return string(data),nil
}

// flightHistory returns flight history for specified user 
func (self* userRestAPI) flightHistory(w http.ResponseWriter, r *http.Request) {

	// Read arguments
	band,number,err := self.extractBandAndNumber(r)
	if err != nil {
		logError(err)
		http.Error(w, fmt.Sprintf("\nFailed to parse arguments '%s'\n",err), http.StatusInternalServerError)
		return
	}

	// Retrieve history for specified traveller
	w.Header().Set("Content-Type", "application/json")
	_,history,err := self.engine.TripHistoryAsJSON(band,number)
	if err != nil {
		logError(err)
		http.Error(w, fmt.Sprintf("\nFailed to retrieve flight history with error '%s'\n",err), http.StatusInternalServerError)
		return
	}
	io.WriteString(w,history)

}

// transactions returns transcation history for specified user 
func (self* userRestAPI) transactions(w http.ResponseWriter, r *http.Request) {

	// Read arguments
	band,number,err := self.extractBandAndNumber(r)
	if err != nil {
		logError(err)
		http.Error(w, fmt.Sprintf("\nFailed to parse arguments '%s'\n",err), http.StatusInternalServerError)
		return
	}

	// Retrieve transactions for specified traveller
	w.Header().Set("Content-Type", "application/json")
	history,err := self.engine.TransactionsAsJSON(band,number)
	if err != nil {
		logError(err)
		http.Error(w, fmt.Sprintf("\nFailed to retrieve transaction history with error '%s'\n",err), http.StatusInternalServerError)
		return
	}
	io.WriteString(w,history)

}

// Read band and band number from arguments
func (self* userRestAPI) extractBandAndNumber(r *http.Request) (uint64, uint64, error) {
	var band,number uint64
	var err error
	pathParams := mux.Vars(r)
	if raw, ok := pathParams["band"]; !ok {
		return 0,0,EMISSINGARGUMENT
	} else {
		band, err = strconv.ParseUint(raw,10,64)
		if err != nil {
			return 0,0,err
		}
	}
	if raw ,ok := pathParams["number"]; !ok {
		return 0,0,EMISSINGARGUMENT
	}  else {
		number, err = strconv.ParseUint(raw,10,64)
		if err != nil {
			return 0,0,err
		}
	}
	return band, number, nil
}

// promises returns promises for specified user 
func (self* userRestAPI) promises(w http.ResponseWriter, r *http.Request) {

	// Read arguments
	band,number,err := self.extractBandAndNumber(r)
	if err != nil {
		logError(err)
		http.Error(w, fmt.Sprintf("\nFailed to parse arguments '%s'\n",err), http.StatusInternalServerError)
		return
	}

	// Retrieve history for specified traveller
	w.Header().Set("Content-Type", "application/json")
	promises,err := self.engine.PromisesAsJSON(band,number)
	if err != nil {
		logError(err)
		http.Error(w, fmt.Sprintf("\nFailed to retrieve promises with error '%s'\n",err), http.StatusInternalServerError)
		return
	}
	io.WriteString(w,promises)

}

// account returns account balance details for specified user 
func (self* userRestAPI) account(w http.ResponseWriter, r *http.Request) {

	// Read arguments
	band,number,err := self.extractBandAndNumber(r)
	if err != nil {
		logError(err)
		http.Error(w, fmt.Sprintf("\nFailed to parse arguments '%s'\n",err), http.StatusInternalServerError)
		return
	}

	// Retrieve history for specified traveller
	w.Header().Set("Content-Type", "application/json")
	promises,err := self.engine.AccountAsJSON(band,number,flap.EpochTime(time.Now().Unix()))
	if err != nil {
		logError(err)
		http.Error(w, fmt.Sprintf("\nFailed to retrieve account with error '%s'\n",err), http.StatusInternalServerError)
		return
	}
	io.WriteString(w,promises)

}

// stats returns daily statistics about the current model
func (self* userRestAPI) dailyStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json,err := self.engine.SummaryStats()
	if err != nil {
		logError(err)
		http.Error(w, fmt.Sprintf("\nFailed to retrieve daily stats with error '%s'\n",err), http.StatusInternalServerError)
		return
	}
	io.WriteString(w,json)
}
