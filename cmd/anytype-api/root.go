package main

import (
	"github.com/jaxkodex/anytype-api-cli/internal/anytype"
	"github.com/spf13/cobra"
)

// version is overridable at build time with -ldflags "-X main.version=...".
var version = "dev"

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "anytype-api",
		Short: "A command line interface for Anytype",
		Long: `anytype-api is a command line interface for the Anytype API.

Authentication uses an API token read from the ANYTYPE_API_KEY environment
variable. The CLI talks to the local Anytype API server (` + "http://127.0.0.1:31009" + `
by default); override it with ANYTYPE_API_URL.`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newTypesCmd())
	cmd.AddCommand(newFilesCmd())
	cmd.AddCommand(newListsCmd())
	cmd.AddCommand(newPropertiesCmd())
	cmd.AddCommand(newTagsCmd())
	cmd.AddCommand(newObjectsCmd())

	return cmd
}

// newClient resolves configuration from the environment and builds an
// authenticated API client, shared by all commands.
func newClient() (*anytype.Client, error) {
	cfg, err := anytype.ConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return anytype.NewClient(cfg)
}
