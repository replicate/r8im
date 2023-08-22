package images

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
)

func getCOGWeights(imageRef v1.Image) (string, error) {
	cfg, err := imageRef.ConfigFile()
	if err != nil {
		return "", fmt.Errorf("getting config %w", err)
	}

	for _, envVar := range cfg.Config.Env {
		if strings.HasPrefix(envVar, "COG_WEIGHTS=") {
			return strings.TrimPrefix(envVar, "COG_WEIGHTS="), nil
		}
	}

	return "", nil
}

func addCogWeights(baseImage v1.Image, cogWeights string) (v1.Image, error) {
	cfg, err := baseImage.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("getting config %w", err)
	}

	// Find and update COG_WEIGHTS if it exists, otherwise append it
	found := false
	for i, envVar := range cfg.Config.Env {
		if strings.HasPrefix(envVar, "COG_WEIGHTS=") {
			cfg.Config.Env[i] = "COG_WEIGHTS=" + cogWeights
			found = true
			break
		}
	}
	if !found {
		cfg.Config.Env = append(cfg.Config.Env, "COG_WEIGHTS="+cogWeights)
	}

	// Create a new image with the updated config
	mutant, err := mutate.Config(baseImage, cfg.Config)
	if err != nil {
		return nil, fmt.Errorf("mutating config %w", err)
	}

	return mutant, nil
}

func ReallyRemix(baseRef string, weightsRef string, dest string, auth authn.Authenticator) (string, error) {
	fmt.Fprintln(os.Stderr, "fetching metadata for", weightsRef)
	start := time.Now()
	weightsImage, err := crane.Pull(weightsRef, crane.WithAuth(auth))
	if err != nil {
		return "", fmt.Errorf("pulling %w", err)
	}
	fmt.Fprintln(os.Stderr, "pulling took", time.Since(start))

	fmt.Fprintln(os.Stderr, "fetching metadata for", baseRef)
	start = time.Now()
	baseImage, err := crane.Pull(baseRef, crane.WithAuth(auth))
	if err != nil {
		return "", fmt.Errorf("pulling %w", err)
	}
	fmt.Fprintln(os.Stderr, "pulling took", time.Since(start))

	fmt.Fprintln(os.Stderr, "finding weights layer")

	cogWeights, err := getCOGWeights(weightsImage)
	if err != nil {
		return "", fmt.Errorf("getting cog weights %w", err)
	}

	var mutant v1.Image

	if cogWeights != "" {
		fmt.Println("found cog weights", cogWeights)
		start = time.Now()
		mutant, err = addCogWeights(baseImage, cogWeights)
		if err != nil {
			return "", fmt.Errorf("adding cog weights %w", err)
		}
		fmt.Fprintln(os.Stderr, "adding cog weights took", time.Since(start))
	} else {
		start = time.Now()
		weightsLayer, err := findWeightsLayer(weightsImage)
		if err != nil {
			return "", fmt.Errorf("getting layers %w", err)
		}
		fmt.Fprintln(os.Stderr, "finding weights layer took", time.Since(start))

		start = time.Now()
		mutant, err = appendLayers(baseImage, weightsLayer)
		if err != nil {
			return "", fmt.Errorf("appending layers %w", err)
		}
		fmt.Fprintln(os.Stderr, "appending layers took", time.Since(start))
	}

	fmt.Fprintln(os.Stderr, "mutant image:", mutant)

	// --- pushing image

	start = time.Now()

	err = crane.Push(mutant, dest, crane.WithAuth(auth))
	if err != nil {
		return "", fmt.Errorf("pushing %s: %w", dest, err)
	}

	fmt.Fprintln(os.Stderr, "pushing took", time.Since(start))

	return "mutant.hexdigest", nil
}

func findWeightsLayer(image v1.Image) (v1.Layer, error) {
	cfg, err := image.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("getting config %w", err)
	}
	idx := 0
	for _, h := range cfg.History {
		if h.EmptyLayer {
			continue
		}

		if h.Comment == "weights" {
			layers, err := image.Layers()
			if err != nil {
				return nil, fmt.Errorf("getting layers %w", err)
			}
			return layers[idx], nil
		}
		idx++
	}
	return nil, fmt.Errorf("no weights layer found")
}
