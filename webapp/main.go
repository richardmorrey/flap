package main

import (
	"log"
	"net/http"
	"encoding/json"
	"os"
	"time"
	//"fmt"
	oidc "github.com/coreos/go-oidc"
	"golang.org/x/net/context"
	//"golang.org/x/oauth2"
	"github.com/gorilla/mux"
)

var (
	clientID     = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
)

// middlewareIdToken function validates openid id token
func middlewareIdToken(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		 	pathParams := mux.Vars(r)
			if raw, ok := pathParams["token"]; !ok {
				http.Error(w,"Missing argument: id", http.StatusForbidden)
			} else {
				parsed,err:= validateIDToken(raw)
				if err == nil {
					log.Printf("Authenticated user %s\n", parsed)
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
func flightHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("flight history will go here"))
}

// main is main
func main() {

	// Top level router
	r := mux.NewRouter()
	
	// REST API router
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/flighthistory/id/{token}", flightHistory).Methods(http.MethodGet)
	api.Use(middlewareIdToken)

	// Static content
	r.PathPrefix("/app/").Handler(http.FileServer(http.Dir(".")))
	r.PathPrefix("/doc/").Handler(http.FileServer(http.Dir("..")))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../website")))

	// Start serving up content
	srv := &http.Server{Handler:r,Addr:":8080",WriteTimeout: 15 * time.Second,ReadTimeout:  15 * time.Second}
	log.Fatal(srv.ListenAndServe())
}

