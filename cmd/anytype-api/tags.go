package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
	"github.com/spf13/cobra"
)

func newTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Manage the tags defined on a select/multi-select property",
		Long: `Work with the tags (the selectable values) of a select or multi-select
property within a space.

Tags are nested under a property, so every subcommand requires both a --space id
and a --property id.`,
	}

	cmd.AddCommand(newTagsListCmd())
	cmd.AddCommand(newTagsGetCmd())
	cmd.AddCommand(newTagsCreateCmd())
	cmd.AddCommand(newTagsUpdateCmd())
	cmd.AddCommand(newTagsDeleteCmd())

	return cmd
}

func newTagsListCmd() *cobra.Command {
	var (
		spaceID    string
		propertyID string
		jsonOut    bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the tags defined on a property",
		Args:  cobra.NoArgs,
		Example: `  # List every tag on a property
  anytype-api tags list --space bafyre... --property bafyre...

  # Emit raw JSON for scripting
  anytype-api tags list --space bafyre... --property bafyre... --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.ListTags(cmd.Context(), spaceID, propertyID)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			return printTagsTable(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the property belongs to (required)")
	cmd.Flags().StringVarP(&propertyID, "property", "p", "", "Property id to list tags from (required)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")
	cmd.MarkFlagRequired("property")

	return cmd
}

func newTagsGetCmd() *cobra.Command {
	var (
		spaceID    string
		propertyID string
		jsonOut    bool
	)

	cmd := &cobra.Command{
		Use:   "get [tag-id]",
		Short: "Show a single tag by id",
		Args:  cobra.ExactArgs(1),
		Example: `  # Show one tag's details
  anytype-api tags get bafyre... --space bafyre... --property bafyre...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.GetTag(cmd.Context(), spaceID, propertyID, args[0])
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			return printTagDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the property belongs to (required)")
	cmd.Flags().StringVarP(&propertyID, "property", "p", "", "Property id the tag belongs to (required)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")
	cmd.MarkFlagRequired("property")

	return cmd
}

func newTagsCreateCmd() *cobra.Command {
	var (
		spaceID    string
		propertyID string
		file       string
		name       string
		color      string
		key        string
		jsonOut    bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tag on a property",
		Args:  cobra.NoArgs,
		Long: `Create a new tag on a select or multi-select property.

The tag definition may be supplied as a JSON payload via --file (use "-" to read
from stdin) and/or with the convenience flags below. Flags take precedence over
fields in the payload, so you can use a file as a template and tweak it.

A tag requires a name and a color. Color is one of: grey, yellow, orange, red,
pink, purple, blue, ice, teal, lime.`,
		Example: `  # Create a tag with convenience flags
  anytype-api tags create --space bafyre... --property bafyre... \
    --name Urgent --color red

  # Create from a JSON payload on stdin
  cat tag.json | anytype-api tags create --space bafyre... --property bafyre... --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var body api.CreateTagRequest
			if file != "" {
				if err := readPayload(cmd, file, &body); err != nil {
					return err
				}
			}

			if name != "" {
				body.Name = name
			}
			if color != "" {
				body.Color = api.Color(color)
			}
			if cmd.Flags().Changed("key") {
				body.Key = &key
			}

			if body.Name == "" {
				return fmt.Errorf("a name is required: pass --name or include it in --file")
			}
			if body.Color == "" {
				return fmt.Errorf("a color is required: pass --color or include it in --file")
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.CreateTag(cmd.Context(), spaceID, propertyID, body)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Created tag:")
			return printTagDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the property belongs to (required)")
	cmd.Flags().StringVarP(&propertyID, "property", "p", "", "Property id to create the tag on (required)")
	cmd.Flags().StringVarP(&file, "file", "f", "", `JSON payload file ("-" for stdin)`)
	cmd.Flags().StringVar(&name, "name", "", "Name of the tag")
	cmd.Flags().StringVar(&color, "color", "", "Color of the tag (e.g. grey, yellow, red, blue)")
	cmd.Flags().StringVar(&key, "key", "", "Optional custom key for the tag")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")
	cmd.MarkFlagRequired("property")

	return cmd
}

func newTagsUpdateCmd() *cobra.Command {
	var (
		spaceID    string
		propertyID string
		file       string
		name       string
		color      string
		key        string
		jsonOut    bool
	)

	cmd := &cobra.Command{
		Use:   "update [tag-id]",
		Short: "Update an existing tag",
		Args:  cobra.ExactArgs(1),
		Long: `Update an existing tag. Only the fields you supply are changed.

Provide changes as a JSON payload via --file (use "-" for stdin) and/or with the
convenience flags below. Flags take precedence over fields in the payload.`,
		Example: `  # Rename a tag and recolor it
  anytype-api tags update bafyre... --space bafyre... --property bafyre... \
    --name "High priority" --color orange`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var body api.UpdateTagRequest
			if file != "" {
				if err := readPayload(cmd, file, &body); err != nil {
					return err
				}
			}

			if name != "" {
				body.Name = &name
			}
			if color != "" {
				c := api.Color(color)
				body.Color = &c
			}
			if cmd.Flags().Changed("key") {
				body.Key = &key
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.UpdateTag(cmd.Context(), spaceID, propertyID, args[0], body)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Updated tag:")
			return printTagDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the property belongs to (required)")
	cmd.Flags().StringVarP(&propertyID, "property", "p", "", "Property id the tag belongs to (required)")
	cmd.Flags().StringVarP(&file, "file", "f", "", `JSON payload file ("-" for stdin)`)
	cmd.Flags().StringVar(&name, "name", "", "New name of the tag")
	cmd.Flags().StringVar(&color, "color", "", "New color of the tag (e.g. grey, yellow, red, blue)")
	cmd.Flags().StringVar(&key, "key", "", "New key for the tag")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")
	cmd.MarkFlagRequired("property")

	return cmd
}

func newTagsDeleteCmd() *cobra.Command {
	var (
		spaceID    string
		propertyID string
		yes        bool
		jsonOut    bool
	)

	cmd := &cobra.Command{
		Use:   "delete [tag-id]",
		Short: "Delete a tag",
		Args:  cobra.ExactArgs(1),
		Long: `Delete a tag from a property.

You are asked to confirm before the tag is deleted; pass --yes to skip the
prompt when scripting.`,
		Example: `  # Delete a tag, with confirmation
  anytype-api tags delete bafyre... --space bafyre... --property bafyre...

  # Delete without prompting
  anytype-api tags delete bafyre... --space bafyre... --property bafyre... --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tagID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Delete tag %s?", tagID))
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

			result, err := client.DeleteTag(cmd.Context(), spaceID, propertyID, tagID)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Deleted tag:")
			return printTagDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the property belongs to (required)")
	cmd.Flags().StringVarP(&propertyID, "property", "p", "", "Property id the tag belongs to (required)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")
	cmd.MarkFlagRequired("property")

	return cmd
}

func printTagsTable(cmd *cobra.Command, result *api.PaginatedResponseTag) error {
	out := cmd.OutOrStdout()

	if result == nil || len(deref(result.Data)) == 0 {
		fmt.Fprintln(out, "No tags found on this property.")
		return nil
	}

	tags := deref(result.Data)
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCOLOR\tKEY\tID")
	for _, t := range tags {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			orDash(deref(t.Name)),
			orDash(string(deref(t.Color))),
			orDash(deref(t.Key)),
			shortID(t.Id),
		)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if p := result.Pagination; p != nil {
		shown := len(tags)
		total := derefInt(p.Total)
		fmt.Fprintf(out, "\nShowing %d of %d", shown, total)
		if derefBool(p.HasMore) {
			fmt.Fprintf(out, " (use --offset %d for more)", derefInt(p.Offset)+shown)
		}
		fmt.Fprintln(out)
	}

	return nil
}

func printTagDetail(cmd *cobra.Command, result *api.TagResponse) error {
	out := cmd.OutOrStdout()

	if result == nil || result.Tag == nil {
		fmt.Fprintln(out, "Tag not found.")
		return nil
	}

	t := result.Tag
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", orDash(deref(t.Name)))
	fmt.Fprintf(w, "Color:\t%s\n", orDash(string(deref(t.Color))))
	fmt.Fprintf(w, "Key:\t%s\n", orDash(deref(t.Key)))
	fmt.Fprintf(w, "ID:\t%s\n", orDash(deref(t.Id)))
	return w.Flush()
}
