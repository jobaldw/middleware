package authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/auth0-community/go-auth0"
	"github.com/gorilla/handlers"

	"github.com/jobaldw/shared/v2/client"
	"github.com/jobaldw/shared/v2/config"
	"github.com/jobaldw/shared/v2/router"

	"gopkg.in/square/go-jose.v2"
)

const (
	// package logging key
	packageKey = "client"

	// all options values
	all = "*"

	// basic http request methods
	basicMethods = "DELETE, GET, POST, PUT"
)

// Middleware configurations that stems from the shared package config.Client.
type Config struct {
	// Auth0 clientID
	ID string `json:"id"`

	// Auth0 clientSecret
	Secret string `json:"secret"`

	// shared config.Client configurations
	config.Client
}

// The configurations of the Auth0 client used to handle authentication.
type Authentication struct {
	// identifier needed for retrieving the applications access token
	clientID string

	// secret needed for retrieving the applications access token
	clientSecret string

	// issuer of middleware authentication
	domain string

	// audience who will be using the middleware authentication
	identifier string

	// shared config.Client configurations
	client client.Client
}

// New
// 	Configures a new struct to be used for Auth0 authentication.
// 	* @param app: name of the application
// 	* @param conf: authentication client configuration
func New(app string, conf Config) (*Authentication, error) {
	newClient, err := client.New(conf.Client)
	if err != nil {
		return nil, err
	}

	return &Authentication{
		domain:       conf.URL,
		identifier:   "https://" + app,
		client:       *newClient,
		clientID:     conf.ID,
		clientSecret: conf.Secret,
	}, err
}

// Middleware
//	Handler function that authenticates handler.
// 	* @param next: handler function that needs to be authenticated
func (a *Authentication) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set writer header options
		w.Header().Set("Access-Control-Allow-Origin", all)
		w.Header().Set("Access-Control-Allow-Methods", basicMethods)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")

		// create clients for handling and validation JSON Web Keys for Auth0 tokens.
		client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: a.domain + "/.well-known/jwks.json"}, nil)
		configuration := auth0.NewConfiguration(client, []string{a.identifier}, a.domain+"/", jose.RS256)
		validator := auth0.NewValidator(configuration, nil)

		// retrieve authorized access bearer token
		token, err := a.getToken(r.Context())
		if err != nil {
			router.RespondError(w, json.Marshal, http.StatusGatewayTimeout, err)
			return
		}
		r.Header.Add("Authorization", "Bearer "+token)

		// validate token
		_, err = validator.ValidateRequest(r)
		if err != nil {
			router.RespondError(w, json.Marshal, http.StatusUnauthorized, err)
			return
		}

		next(w, r)
	})
}

// Handler
// 	Configure middleware CORS options.
// 	* @param r: http response handler
func Handler(r http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		handlers.AllowedMethods([]string{http.MethodDelete, http.MethodGet, http.MethodPost, http.MethodPut}),
		handlers.AllowedOrigins([]string{all}),
	)(r)
}

/********** Helper functions **********/

// getToken
// 	Executes a client credentials exchange to retrieve an access token for the Auth0 client application.
// 	* @param ctx: context used to handle any cancellations
func (a *Authentication) getToken(ctx context.Context) (string, error) {
	// build request and call api
	payload := struct {
		Audience     string `json:"audience"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		GrantType    string `json:"grant_type"`
	}{
		Audience:     a.identifier,
		ClientID:     a.clientID,
		ClientSecret: a.clientSecret,
		GrantType:    "client_credentials",
	}
	resp, err := a.client.PostWithContext(ctx, "oauth/token", nil, payload)
	if err != nil {
		return "", fmt.Errorf("%s: %s, could not call client", packageKey, err)
	}

	// check response status
	switch resp.StatusCode {
	case http.StatusOK:
		access := struct {
			Token string `json:"access_token,omitempty"`
		}{}
		if err := json.Unmarshal(resp.GetBodyBytes(), &access); err != nil {
			return "", fmt.Errorf("%s: %s, could not unmarshal", packageKey, err)
		}
		return access.Token, nil
	default:
		return "", fmt.Errorf("%s: %s", packageKey, err)
	}
}
