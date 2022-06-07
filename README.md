<p align="center">
  <img alt="JB Designs logo" src="./assets/jb-icon.jpg" height="150"/>
  <h3 align="center">middleware</h3>
  <p align="center">Lightweight middleware package that utilizes 
  <a href="https://www.auth0.com/">Auth0</a>
  to validate API endpoints with Bearer Token Authentication.</p>
</p>

# Badges

[![Go Report Card](https://goreportcard.com/badge/github.com/jobaldw/middleware?style=plastic)](https://goreportcard.com/report/github.com/jobaldw/middleware) [![GitHub issues](https://img.shields.io/github/issues/jobaldw/middleware?style=plastic)](https://github.com/jobaldw/middleware/issues) [![Release Version](https://img.shields.io/github/v/release/jobaldw/middleware?style=plastic)](https://img.shields.io/github/v/release/jobaldw/middleware)

# How To Use

Below is a standard configuration utilizing the share/v2/router without any middleware wrappers.

``` go
package main

import (
  "encoding/json"
  "net/http"

  "github.com/jobaldw/shared/v2/router"
)

func main() {
  srv, r := router.New(3001, nil)

  r.HandleFunc("/endpoint", helloWorld()).Methods(http.MethodGet)

  srv.Handler = r
  srv.ListenAndServe()
}

func helloWorld() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        router.Respond(w, json.Marshal, 200, router.Message{MSG: "hello world"})
    }
}
```

In this example, we create the authentication client and wrap hellWorld() handler to be authenticated.

``` go
package main

import (
  "encoding/json"
  "net/http"

  "github.com/jobaldw/middleware/authentication"
  "github.com/jobaldw/shared/v2/config"
  "github.com/jobaldw/shared/v2/router"
  "github.com/rs/cors"
)

func main() {
  // configure authentication
  conf := authentication.Config{
    ID:     "client auth0 id",
    Secret: "client auth0 secret",
    Client: config.Client{
      Headers: map[string][]string{},
      Health:  "/health",
      Timeout: 5,
      URL:     "www.url.com",
    },
  }

  // create new Auth0 authentication client
  authentication.New("myApp", conf)

  srv, r := router.New(3001, nil)

  // wrap a handler function with with Auth0's middleware
  r.HandleFunc("/endpoint", authentication.Auth0(helloWorld())).Methods(http.MethodGet)

  srv.Handler = cors.AllowAll().Handler(middleware.Handler(r))
  srv.ListenAndServe()
}

func helloWorld() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        router.Respond(w, json.Marshal, 200, router.Message{MSG: "hello world"})
    }
}
```
