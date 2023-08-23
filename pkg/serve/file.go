package serve

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"

	"github.com/anotherjesse/r8im/pkg/images"
	"github.com/gorilla/mux"
)

func fileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	model := vars["model"]
	version := vars["version"]
	layerDigest := vars["layer"]
	fileName := vars["file"]

	imageName := fmt.Sprintf("r8.im/%s/%s@sha256:%s", user, model, version)
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

	l := layer.Raw
	rc, err := l.Uncompressed()
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	showFile(w, rc, fileName)
}

func showFile(w http.ResponseWriter, rc io.ReadCloser, fileName string) {
	tr := tar.NewReader(rc)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if header.Typeflag == tar.TypeReg && header.Name == fileName {
			if _, err := io.Copy(w, tr); err != nil {
				fmt.Fprintf(w, "Error: %s", err)
				return
			}
			break
		}
	}

	rc.Close()
}
