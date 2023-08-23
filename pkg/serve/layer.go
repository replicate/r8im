package serve

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"

	"github.com/anotherjesse/r8im/pkg/images"
	"github.com/gorilla/mux"
)

func layerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	model := vars["model"]
	version := vars["version"]
	layerDigest := vars["layer"]

	imageName := fmt.Sprintf("r8.im/%s/%s@sha256:%s", user, model, version)
	basePath := fmt.Sprintf("/%s/%s:%s/%s", user, model, version, layerDigest)
	layers, err := images.Layers(imageName, global_auth)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	// find layer
	var layer images.Layer
	for _, l := range layers {
		if l.Digest == layerDigest {
			layer = l
		}
	}

	if layer.Digest == "" {
		fmt.Fprintf(w, "Error: layer not found")
		return
	}

	if layer.Size > 30_000_000 {
		fmt.Fprintf(w, "Error: layer too big to show")
		return
	}

	l := layer.Raw
	rc, err := l.Uncompressed()
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	showFiles(w, rc, basePath)
}

func showFiles(w http.ResponseWriter, rc io.ReadCloser, basePath string) {

	tr := tar.NewReader(rc)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintf(w, "<h2>Files</h2>")
	fmt.Fprintf(w, "<table>")
	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, "<th>Name</th>")
	fmt.Fprintf(w, "<th>Size</th>")
	fmt.Fprintf(w, "</tr>")

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		if header.Typeflag == tar.TypeReg {
			fmt.Fprintf(w, "<tr>")
			url := fmt.Sprintf("%s/%s", basePath, header.Name)
			fmt.Fprintf(w, "<td><a href='%s'>%s</a></td>", url, header.Name)
			fmt.Fprintf(w, "<td>%s</td>", ByteCountSI(header.Size))
			fmt.Fprintf(w, "</tr>")
		}
	}

	fmt.Fprintf(w, "</table>")
	rc.Close()

}
