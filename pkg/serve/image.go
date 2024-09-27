package serve

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/anotherjesse/r8im/pkg/images"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/gorilla/mux"
)

func imageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	model := vars["model"]
	version := vars["version"]

	imageName := fmt.Sprintf("r8.im/%s/%s@sha256:%s", user, model, version)

	user_image, err := crane.Pull(imageName, crane.WithAuth(global_auth))
	if err != nil {
		fmt.Fprintf(w, "Error pulling image: %s", err)
	}

	layers, err := images.LayersForImage(user_image)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	cfg, err := user_image.ConfigFile()
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
	}
	env := cfg.Config.Env

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1>%s</h1>", imageName)

	fmt.Fprintf(w, "<a href=\"https://replicate.com/\">replicate.com</a>/")
	fmt.Fprintf(w, "<a href=\"https://replicate.com/%s\">%s</a>/", user, user)
	fmt.Fprintf(w, "<a href=\"https://replicate.com/%s/%s\">%s</a>/", user, model, model)
	fmt.Fprintf(w, "<a href=\"https://replicate.com/%s/%s/versions\">versions</a>/", user, model)
	fmt.Fprintf(w, "<a href=\"https://replicate.com/%s/%s/versions/%s\">%s</a>", user, model, version, version)

	fmt.Fprintf(w, "<h2>Environment</h2>")
	fmt.Fprintf(w, "<table>")
	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, "<th>Key</th>")
	fmt.Fprintf(w, "<th>Value</th>")
	fmt.Fprintf(w, "</tr>")
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		fmt.Fprintf(w, "<tr>")
		fmt.Fprintf(w, "<td>%s</td>", parts[0])
		fmt.Fprintf(w, "<td>%s</td>", parts[1])
		fmt.Fprintf(w, "</tr>")
	}

	fmt.Fprintf(w, "<h2>Layers</h2>")

	fmt.Fprintf(w, "<table>")
	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, "<th>Command</th>")
	fmt.Fprintf(w, "<th>Size</th>")
	fmt.Fprintf(w, "</tr>")

	for i := len(layers) - 1; i >= 0; i-- {
		layer := layers[i]
		layerUrl := fmt.Sprintf("/%s/%s:%s/%s", user, model, version, layer.Digest)
		fmt.Fprintf(w, "<tr>")
		fmt.Fprintf(w, "<td><a href='%s'>%s</a></td>", layerUrl, layer.Command)
		fmt.Fprintf(w, "<td>%s</td>", ByteCountSI(layer.Size))
		fmt.Fprintf(w, "</tr>")
	}
	fmt.Fprintf(w, "</table>")

}

func LayersForImage(image v1.Image) {
	panic("unimplemented")
}
