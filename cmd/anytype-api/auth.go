package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/jaxkodex/anytype-api-cli/internal/anytype"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Obtain an API key from the local Anytype app",
		Long: `Bootstrap an API key for the Anytype API.

Authentication is a two-step flow against the local Anytype desktop app:

  1. "auth challenge" asks the app to start a challenge. The app pops up a
     4-digit code and the command prints a challenge id.
  2. "auth api-key" exchanges that challenge id and the 4-digit code for an
     API key, which you then export as ANYTYPE_API_KEY for other commands.

These commands talk to the local Anytype app and do not themselves require an
API key.`,
	}

	cmd.AddCommand(newAuthChallengeCmd())
	cmd.AddCommand(newAuthAPIKeyCmd())

	return cmd
}

func newAuthChallengeCmd() *cobra.Command {
	var (
		appName string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "challenge",
		Short: "Start an authentication challenge",
		Args:  cobra.NoArgs,
		Long: `Start an authentication challenge with the local Anytype app.

The app displays a 4-digit code; this command prints the matching challenge id.
Pass both to "auth api-key" to obtain an API key.`,
		Example: `  # Start a challenge for an app named my-cli
  anytype-api auth challenge --app my-cli`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := anytype.NewUnauthenticatedClient()
			if err != nil {
				return err
			}

			result, err := client.CreateChallenge(cmd.Context(), appName)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}

			out := cmd.OutOrStdout()
			w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "Challenge ID:\t%s\n", orDash(deref(result.ChallengeId)))
			if err := w.Flush(); err != nil {
				return err
			}
			fmt.Fprintln(out, "\nEnter the 4-digit code shown in the Anytype app with:")
			fmt.Fprintf(out, "  anytype-api auth api-key --challenge %s --code <code>\n", deref(result.ChallengeId))
			return nil
		},
	}

	cmd.Flags().StringVar(&appName, "app", "anytype-api-cli", "Name of the app requesting access")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")

	return cmd
}

func newAuthAPIKeyCmd() *cobra.Command {
	var (
		challengeID string
		code        string
		jsonOut     bool
	)

	cmd := &cobra.Command{
		Use:   "api-key",
		Short: "Exchange a challenge and code for an API key",
		Args:  cobra.NoArgs,
		Long: `Exchange a challenge id and the 4-digit code shown in the Anytype app for an
API key.

Run "auth challenge" first to obtain a challenge id, then read the 4-digit code
from the Anytype desktop app. Export the returned key as ANYTYPE_API_KEY so the
other commands can authenticate.`,
		Example: `  # Exchange a challenge for an API key
  anytype-api auth api-key --challenge 67647f... --code 1234

  # Capture the key into the environment
  export ANYTYPE_API_KEY=$(anytype-api auth api-key \
    --challenge 67647f... --code 1234 --json | jq -r .api_key)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := anytype.NewUnauthenticatedClient()
			if err != nil {
				return err
			}

			result, err := client.CreateAPIKey(cmd.Context(), challengeID, code)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}

			out := cmd.OutOrStdout()
			w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "API Key:\t%s\n", orDash(deref(result.ApiKey)))
			if err := w.Flush(); err != nil {
				return err
			}
			fmt.Fprintln(out, "\nExport it so other commands can use it:")
			fmt.Fprintf(out, "  export ANYTYPE_API_KEY=%s\n", deref(result.ApiKey))
			return nil
		},
	}

	cmd.Flags().StringVar(&challengeID, "challenge", "", "Challenge id from 'auth challenge' (required)")
	cmd.Flags().StringVar(&code, "code", "", "4-digit code shown in the Anytype app (required)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("challenge")
	cmd.MarkFlagRequired("code")

	return cmd
}
