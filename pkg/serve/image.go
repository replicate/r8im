package serve

import (
	"fmt"
	"net/http"

	"github.com/anotherjesse/r8im/pkg/images"
	"github.com/gorilla/mux"
)

func imageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	model := vars["model"]
	version := vars["version"]

	imageName := fmt.Sprintf("r8.im/%s/%s@sha256:%s", user, model, version)
	layers, err := images.Layers(imageName, global_auth)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1>%s</h1>", imageName)

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
