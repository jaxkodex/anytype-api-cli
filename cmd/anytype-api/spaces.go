package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/jaxkodex/anytype-api-cli/internal/anytype"
	"github.com/jaxkodex/anytype-api-cli/internal/api"
	"github.com/spf13/cobra"
)

func newSpacesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spaces",
		Short: "Manage the spaces accessible to you",
		Long: `Work with the spaces (workspaces) accessible to the authenticated user.

A space groups objects, types and lists together. Use these subcommands to list
your spaces, inspect a single one, or create and update spaces.`,
	}

	cmd.AddCommand(newSpacesListCmd())
	cmd.AddCommand(newSpacesGetCmd())
	cmd.AddCommand(newSpacesCreateCmd())
	cmd.AddCommand(newSpacesUpdateCmd())

	return cmd
}

func newSpacesListCmd() *cobra.Command {
	var (
		limit   int
		offset  int
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the spaces accessible to you",
		Args:  cobra.NoArgs,
		Example: `  # List every space you can access
  anytype-api spaces list

  # Emit raw JSON for scripting
  anytype-api spaces list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.ListSpaces(cmd.Context(), anytype.ListSpacesOptions{
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			return printSpacesTable(cmd, result)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 100, "Maximum number of results to return (max 1000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")

	return cmd
}

func newSpacesGetCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "get [space-id]",
		Short: "Show a single space by id",
		Args:  cobra.ExactArgs(1),
		Example: `  # Show one space's details
  anytype-api spaces get bafyre...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.GetSpace(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			return printSpaceDetail(cmd, result)
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")

	return cmd
}

func newSpacesCreateCmd() *cobra.Command {
	var (
		file        string
		name        string
		description string
		jsonOut     bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new space",
		Args:  cobra.NoArgs,
		Long: `Create a new space.

The space definition may be supplied as a JSON payload via --file (use "-" to
read from stdin) and/or with the convenience flags below. Flags take precedence
over fields in the payload, so you can use a file as a template and tweak it.

A space requires a name.`,
		Example: `  # Create a space with convenience flags
  anytype-api spaces create --name "My space" --description "Notes and tasks"

  # Create from a JSON payload on stdin
  cat space.json | anytype-api spaces create --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var body api.CreateSpaceRequest
			if file != "" {
				if err := readPayload(cmd, file, &body); err != nil {
					return err
				}
			}

			if name != "" {
				body.Name = name
			}
			if cmd.Flags().Changed("description") {
				body.Description = &description
			}

			if body.Name == "" {
				return fmt.Errorf("a name is required: pass --name or include it in --file")
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.CreateSpace(cmd.Context(), body)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Created space:")
			return printSpaceDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", `JSON payload file ("-" for stdin)`)
	cmd.Flags().StringVar(&name, "name", "", "Name of the space")
	cmd.Flags().StringVar(&description, "description", "", "Description of the space")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")

	return cmd
}

func newSpacesUpdateCmd() *cobra.Command {
	var (
		file        string
		name        string
		description string
		jsonOut     bool
	)

	cmd := &cobra.Command{
		Use:   "update [space-id]",
		Short: "Update an existing space",
		Args:  cobra.ExactArgs(1),
		Long: `Update the name and/or description of an existing space.

Provide changes as a JSON payload via --file (use "-" for stdin) and/or with the
convenience flags below. Flags take precedence over fields in the payload.`,
		Example: `  # Rename a space
  anytype-api spaces update bafyre... --name "New name"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var body api.UpdateSpaceRequest
			if file != "" {
				if err := readPayload(cmd, file, &body); err != nil {
					return err
				}
			}

			if name != "" {
				body.Name = &name
			}
			if cmd.Flags().Changed("description") {
				body.Description = &description
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.UpdateSpace(cmd.Context(), args[0], body)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Updated space:")
			return printSpaceDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", `JSON payload file ("-" for stdin)`)
	cmd.Flags().StringVar(&name, "name", "", "New name of the space")
	cmd.Flags().StringVar(&description, "description", "", "New description of the space")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")

	return cmd
}

func printSpacesTable(cmd *cobra.Command, result *api.PaginatedResponseSpace) error {
	out := cmd.OutOrStdout()

	if result == nil || len(deref(result.Data)) == 0 {
		fmt.Fprintln(out, "No spaces found.")
		return nil
	}

	spaces := deref(result.Data)
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tNETWORK\tID")
	for _, s := range spaces {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			orDash(deref(s.Name)),
			orDash(deref(s.NetworkId)),
			shortID(s.Id),
		)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if p := result.Pagination; p != nil {
		shown := len(spaces)
		total := derefInt(p.Total)
		fmt.Fprintf(out, "\nShowing %d of %d", shown, total)
		if derefBool(p.HasMore) {
			fmt.Fprintf(out, " (use --offset %d for more)", derefInt(p.Offset)+shown)
		}
		fmt.Fprintln(out)
	}

	return nil
}

func printSpaceDetail(cmd *cobra.Command, result *api.SpaceResponse) error {
	out := cmd.OutOrStdout()

	if result == nil || result.Space == nil {
		fmt.Fprintln(out, "Space not found.")
		return nil
	}

	s := result.Space
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", orDash(deref(s.Name)))
	fmt.Fprintf(w, "Description:\t%s\n", orDash(deref(s.Description)))
	fmt.Fprintf(w, "Network:\t%s\n", orDash(deref(s.NetworkId)))
	fmt.Fprintf(w, "Gateway:\t%s\n", orDash(deref(s.GatewayUrl)))
	fmt.Fprintf(w, "ID:\t%s\n", orDash(deref(s.Id)))
	return w.Flush()
}
