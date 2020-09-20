package main

import (
	"log"
	"net/http"
	"encoding/json"
	"os"
	oidc "github.com/coreos/go-oidc"
	"golang.org/x/net/context"
	//"golang.org/x/oauth2"
	//"github.com/gorilla/handlers"
)

var (
	clientID     = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
)

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

