package serve

import (
	"net/http"
	"fmt"

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
	// yes, one slash, not sure why:
	r.HandleFunc("/https:/replicate.com/{user}/{model}/versions/{version}", replicateComHandler)
	http.Handle("/", r)
	return http.ListenAndServe(":8080", nil)
}

func replicateComHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	model := vars["model"]
	version := vars["version"]
	newUrl := fmt.Sprintf("/%s/%s:%s", user, model, version)
	http.Redirect(w, r, newUrl, http.StatusFound)
}
