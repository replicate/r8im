package serve

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/google/go-containerregistry/pkg/authn"
)

var global_auth authn.Authenticator

func Start(port int, auth authn.Authenticator) error {
	global_auth = auth
	r := mux.NewRouter()
	r.HandleFunc("/{user}/{model}:{version}", imageHandler)
	r.HandleFunc("/{user}/{model}:{version}/{layer}", layerHandler)
	r.HandleFunc("/{user}/{model}:{version}/{layer}/{file:.*}", fileHandler)
	http.Handle("/", r)
	return http.ListenAndServe(":8080", nil)
}
