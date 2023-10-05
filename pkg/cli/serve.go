package cli

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"

	r8auth "github.com/anotherjesse/r8im/pkg/auth"
	"github.com/anotherjesse/r8im/pkg/serve"
)

var (
	port int
)

func newServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "serve -p port",
		Short:  "http api for serving",
		Hidden: false,

		RunE: serveCommmand,
	}

	cmd.Flags().StringVarP(&sToken, "token", "t", "", "replicate cog token")
	cmd.Flags().StringVarP(&sRegistry, "registry", "r", "r8.im", "registry host")

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "port to serve on")

	return cmd
}

func serveCommmand(cmd *cobra.Command, args []string) error {
	if sToken == "" {
		sToken = os.Getenv("COG_TOKEN")
	}
	var auth authn.Authenticator

	if sToken == "" {
		auth = authn.Anonymous
	} else {
		u, err := r8auth.VerifyCogToken(sRegistry, sToken)
		if err != nil {
			fmt.Fprintln(os.Stderr, "authentication error, invalid token or registry host error")
			return err
		}
		auth = authn.FromConfig(authn.AuthConfig{Username: u, Password: sToken})
	}

	err := serve.Start(port, auth)

	if err != nil {
		return err
	}

	return nil
}
