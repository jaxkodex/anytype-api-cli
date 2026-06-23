package main

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/jaxkodex/anytype-api-cli/internal/anytype"
	"github.com/jaxkodex/anytype-api-cli/internal/api"
	"github.com/spf13/cobra"
)

func newTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types",
		Short: "Inspect the object types defined in a space",
		Long: `Work with the object types (e.g. Page, Task, Bookmark) defined within a space.

Types are scoped to a space, so every subcommand requires a --space id. List the
spaces accessible to you in the Anytype app, or via the API, to find one.`,
	}

	cmd.AddCommand(newTypesListCmd())
	cmd.AddCommand(newTypesGetCmd())

	return cmd
}

func newTypesListCmd() *cobra.Command {
	var (
		spaceID string
		limit   int
		offset  int
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the types defined in a space",
		Args:  cobra.NoArgs,
		Example: `  # List every type in a space
  anytype-api types list --space bafyre...

  # Emit raw JSON for scripting
  anytype-api types list --space bafyre... --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.ListTypes(cmd.Context(), anytype.ListTypesOptions{
				SpaceID: spaceID,
				Limit:   limit,
				Offset:  offset,
			})
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			return printTypesTable(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id to list types from (required)")
	cmd.Flags().IntVarP(&limit, "limit", "L", 100, "Maximum number of results to return (max 1000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newTypesGetCmd() *cobra.Command {
	var (
		spaceID string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "get [type-id]",
		Short: "Show a single type by id",
		Args:  cobra.ExactArgs(1),
		Example: `  # Show one type's details
  anytype-api types get bafyre... --space bafyre...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.GetType(cmd.Context(), spaceID, args[0])
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			return printTypeDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the type belongs to (required)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func printTypesTable(cmd *cobra.Command, result *api.PaginatedResponseType) error {
	out := cmd.OutOrStdout()

	if result == nil || len(deref(result.Data)) == 0 {
		fmt.Fprintln(out, "No types found in this space.")
		return nil
	}

	types := deref(result.Data)
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tKEY\tLAYOUT\tID")
	for _, t := range types {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			orDash(deref(t.Name)),
			orDash(deref(t.Key)),
			orDash(string(deref(t.Layout))),
			shortID(t.Id),
		)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if p := result.Pagination; p != nil {
		shown := len(types)
		total := derefInt(p.Total)
		fmt.Fprintf(out, "\nShowing %d of %d", shown, total)
		if derefBool(p.HasMore) {
			fmt.Fprintf(out, " (use --offset %d for more)", derefInt(p.Offset)+shown)
		}
		fmt.Fprintln(out)
	}

	return nil
}

func printTypeDetail(cmd *cobra.Command, result *api.TypeResponse) error {
	out := cmd.OutOrStdout()

	if result == nil || result.Type == nil {
		fmt.Fprintln(out, "Type not found.")
		return nil
	}

	t := result.Type
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", orDash(deref(t.Name)))
	fmt.Fprintf(w, "Plural:\t%s\n", orDash(deref(t.PluralName)))
	fmt.Fprintf(w, "Key:\t%s\n", orDash(deref(t.Key)))
	fmt.Fprintf(w, "Layout:\t%s\n", orDash(string(deref(t.Layout))))
	fmt.Fprintf(w, "Archived:\t%t\n", deref(t.Archived))
	fmt.Fprintf(w, "ID:\t%s\n", orDash(deref(t.Id)))
	if props := deref(t.Properties); len(props) > 0 {
		fmt.Fprintf(w, "Properties:\t%d\n", len(props))
	}
	return w.Flush()
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// encodeJSON writes any value as indented JSON to the command's output.
func encodeJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
