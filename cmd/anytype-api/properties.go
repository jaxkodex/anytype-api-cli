package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/jaxkodex/anytype-api-cli/internal/anytype"
	"github.com/jaxkodex/anytype-api-cli/internal/api"
	"github.com/spf13/cobra"
)

func newPropertiesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "properties",
		Short: "Inspect the properties defined in a space",
		Long: `Work with the properties (e.g. Status, Due date, Tags) defined within a space.

Properties are scoped to a space, so every subcommand requires a --space id.
List the spaces accessible to you in the Anytype app, or via the API, to find
one.`,
	}

	cmd.AddCommand(newPropertiesListCmd())
	cmd.AddCommand(newPropertiesGetCmd())
	cmd.AddCommand(newPropertiesCreateCmd())
	cmd.AddCommand(newPropertiesUpdateCmd())
	cmd.AddCommand(newPropertiesDeleteCmd())

	return cmd
}

func newPropertiesListCmd() *cobra.Command {
	var (
		spaceID string
		limit   int
		offset  int
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the properties defined in a space",
		Args:  cobra.NoArgs,
		Example: `  # List every property in a space
  anytype-api properties list --space bafyre...

  # Emit raw JSON for scripting
  anytype-api properties list --space bafyre... --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.ListProperties(cmd.Context(), anytype.ListPropertiesOptions{
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
			return printPropertiesTable(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id to list properties from (required)")
	cmd.Flags().IntVarP(&limit, "limit", "L", 100, "Maximum number of results to return (max 1000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newPropertiesGetCmd() *cobra.Command {
	var (
		spaceID string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "get [property-id]",
		Short: "Show a single property by id",
		Args:  cobra.ExactArgs(1),
		Example: `  # Show one property's details
  anytype-api properties get bafyre... --space bafyre...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.GetProperty(cmd.Context(), spaceID, args[0])
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			return printPropertyDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the property belongs to (required)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newPropertiesCreateCmd() *cobra.Command {
	var (
		spaceID string
		file    string
		name    string
		key     string
		format  string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new property in a space",
		Args:  cobra.NoArgs,
		Long: `Create a new property in a space.

The property definition may be supplied as a JSON payload via --file (use "-" to
read from stdin) and/or with the convenience flags below. Flags take precedence
over fields in the payload, so you can use a file as a template and tweak it.

A property requires a name and a format. Format is one of: text, number, select,
multi_select, date, files, checkbox, url, email, phone, objects.`,
		Example: `  # Create a property with convenience flags
  anytype-api properties create --space bafyre... \
    --name "Due date" --format date

  # Create from a JSON payload on stdin
  cat property.json | anytype-api properties create --space bafyre... --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var body api.CreatePropertyRequest
			if file != "" {
				if err := readPayload(cmd, file, &body); err != nil {
					return err
				}
			}

			if name != "" {
				body.Name = name
			}
			if cmd.Flags().Changed("key") {
				body.Key = &key
			}
			if format != "" {
				body.Format = api.PropertyFormat(format)
			}

			if body.Name == "" {
				return fmt.Errorf("a name is required: pass --name or include it in --file")
			}
			if body.Format == "" {
				return fmt.Errorf("a format is required: pass --format (text, number, select, multi_select, date, files, checkbox, url, email, phone, objects) or include it in --file")
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.CreateProperty(cmd.Context(), spaceID, body)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Created property:")
			return printPropertyDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id to create the property in (required)")
	cmd.Flags().StringVarP(&file, "file", "f", "", `JSON payload file ("-" for stdin)`)
	cmd.Flags().StringVar(&name, "name", "", "Name of the property")
	cmd.Flags().StringVar(&key, "key", "", "Key of the property (snake_case)")
	cmd.Flags().StringVar(&format, "format", "", "Format: text, number, select, multi_select, date, files, checkbox, url, email, phone or objects")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newPropertiesUpdateCmd() *cobra.Command {
	var (
		spaceID string
		file    string
		name    string
		key     string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "update [property-id]",
		Short: "Update an existing property",
		Args:  cobra.ExactArgs(1),
		Long: `Update an existing property. Only the fields you supply are changed.

Provide changes as a JSON payload via --file (use "-" for stdin) and/or with the
convenience flags below. Flags take precedence over fields in the payload.`,
		Example: `  # Rename a property
  anytype-api properties update bafyre... --space bafyre... --name "New name"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var body api.UpdatePropertyRequest
			if file != "" {
				if err := readPayload(cmd, file, &body); err != nil {
					return err
				}
			}

			if name != "" {
				body.Name = name
			}
			if cmd.Flags().Changed("key") {
				body.Key = &key
			}

			if body.Name == "" {
				return fmt.Errorf("a name is required: pass --name or include it in --file")
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.UpdateProperty(cmd.Context(), spaceID, args[0], body)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Updated property:")
			return printPropertyDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the property belongs to (required)")
	cmd.Flags().StringVarP(&file, "file", "f", "", `JSON payload file ("-" for stdin)`)
	cmd.Flags().StringVar(&name, "name", "", "New name of the property")
	cmd.Flags().StringVar(&key, "key", "", "New key of the property (snake_case)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newPropertiesDeleteCmd() *cobra.Command {
	var (
		spaceID string
		yes     bool
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "delete [property-id]",
		Short: "Delete (archive) a property",
		Args:  cobra.ExactArgs(1),
		Long: `Delete a property. This archives the property rather than permanently removing it.

You are asked to confirm before the property is archived; pass --yes to skip the
prompt when scripting.`,
		Example: `  # Delete a property, with confirmation
  anytype-api properties delete bafyre... --space bafyre...

  # Delete without prompting
  anytype-api properties delete bafyre... --space bafyre... --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			propertyID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Archive property %s? This can be undone in the Anytype app.", propertyID))
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.DeleteProperty(cmd.Context(), spaceID, propertyID)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Deleted property:")
			return printPropertyDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the property belongs to (required)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func printPropertiesTable(cmd *cobra.Command, result *api.PaginatedResponseProperty) error {
	out := cmd.OutOrStdout()

	if result == nil || len(deref(result.Data)) == 0 {
		fmt.Fprintln(out, "No properties found in this space.")
		return nil
	}

	props := deref(result.Data)
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tKEY\tFORMAT\tID")
	for _, p := range props {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			orDash(deref(p.Name)),
			orDash(deref(p.Key)),
			orDash(string(deref(p.Format))),
			shortID(p.Id),
		)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if pg := result.Pagination; pg != nil {
		shown := len(props)
		total := derefInt(pg.Total)
		fmt.Fprintf(out, "\nShowing %d of %d", shown, total)
		if derefBool(pg.HasMore) {
			fmt.Fprintf(out, " (use --offset %d for more)", derefInt(pg.Offset)+shown)
		}
		fmt.Fprintln(out)
	}

	return nil
}

func printPropertyDetail(cmd *cobra.Command, result *api.PropertyResponse) error {
	out := cmd.OutOrStdout()

	if result == nil || result.Property == nil {
		fmt.Fprintln(out, "Property not found.")
		return nil
	}

	p := result.Property
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", orDash(deref(p.Name)))
	fmt.Fprintf(w, "Key:\t%s\n", orDash(deref(p.Key)))
	fmt.Fprintf(w, "Format:\t%s\n", orDash(string(deref(p.Format))))
	fmt.Fprintf(w, "ID:\t%s\n", orDash(deref(p.Id)))
	return w.Flush()
}
