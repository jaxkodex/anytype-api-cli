package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/jaxkodex/anytype-api-cli/internal/anytype"
	"github.com/jaxkodex/anytype-api-cli/internal/api"
	"github.com/spf13/cobra"
)

func newObjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "objects",
		Short: "Work with the objects stored in a space",
		Long: `Work with the objects (pages, notes, tasks, ...) stored within a space.

Objects are scoped to a space, so every subcommand requires a --space id. List
the spaces accessible to you in the Anytype app, or via the API, to find one.`,
	}

	cmd.AddCommand(newObjectsListCmd())
	cmd.AddCommand(newObjectsGetCmd())
	cmd.AddCommand(newObjectsCreateCmd())
	cmd.AddCommand(newObjectsUpdateCmd())
	cmd.AddCommand(newObjectsDeleteCmd())

	return cmd
}

func newObjectsListCmd() *cobra.Command {
	var (
		spaceID string
		limit   int
		offset  int
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the objects in a space",
		Args:  cobra.NoArgs,
		Example: `  # List objects in a space
  anytype-api objects list --space bafyre...

  # Emit raw JSON for scripting
  anytype-api objects list --space bafyre... --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.ListSpaceObjects(cmd.Context(), anytype.ListSpaceObjectsOptions{
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
			return printObjectsTable(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id to list objects from (required)")
	cmd.Flags().IntVarP(&limit, "limit", "L", 100, "Maximum number of results to return (max 1000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newObjectsGetCmd() *cobra.Command {
	var (
		spaceID string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "get [object-id]",
		Short: "Show a single object by id",
		Args:  cobra.ExactArgs(1),
		Example: `  # Show one object's details
  anytype-api objects get bafyre... --space bafyre...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.GetObject(cmd.Context(), spaceID, args[0])
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			return printObjectDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the object belongs to (required)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newObjectsCreateCmd() *cobra.Command {
	var (
		spaceID string
		file    string
		typeKey string
		name    string
		body    string
		icon    string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new object in a space",
		Args:  cobra.NoArgs,
		Long: `Create a new object in a space.

The object definition may be supplied as a JSON payload via --file (use "-" to
read from stdin) and/or with the convenience flags below. Flags take precedence
over fields in the payload, so you can use a file as a template and tweak it.

An object requires a type key (see "anytype-api types list" for the keys of the
types defined in a space).`,
		Example: `  # Create an object with convenience flags
  anytype-api objects create --space bafyre... \
    --type-key page --name "My page" --icon 📄

  # Create from a JSON payload on stdin
  cat object.json | anytype-api objects create --space bafyre... --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var reqBody api.CreateObjectRequest
			if file != "" {
				if err := readPayload(cmd, file, &reqBody); err != nil {
					return err
				}
			}

			if typeKey != "" {
				reqBody.TypeKey = typeKey
			}
			if name != "" {
				reqBody.Name = &name
			}
			if body != "" {
				reqBody.Body = &body
			}
			if icon != "" {
				ic, err := emojiIcon(icon)
				if err != nil {
					return err
				}
				reqBody.Icon = ic
			}

			if reqBody.TypeKey == "" {
				return fmt.Errorf("a type key is required: pass --type-key or include it in --file")
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.CreateObject(cmd.Context(), spaceID, reqBody)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Created object:")
			return printObjectDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id to create the object in (required)")
	cmd.Flags().StringVarP(&file, "file", "f", "", `JSON payload file ("-" for stdin)`)
	cmd.Flags().StringVar(&typeKey, "type-key", "", "Key of the type of object to create")
	cmd.Flags().StringVar(&name, "name", "", "Name of the object")
	cmd.Flags().StringVar(&body, "body", "", "Markdown body of the object")
	cmd.Flags().StringVar(&icon, "icon", "", "Emoji icon for the object")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newObjectsUpdateCmd() *cobra.Command {
	var (
		spaceID  string
		file     string
		typeKey  string
		name     string
		markdown string
		icon     string
		jsonOut  bool
	)

	cmd := &cobra.Command{
		Use:   "update [object-id]",
		Short: "Update an existing object",
		Args:  cobra.ExactArgs(1),
		Long: `Update an existing object. Only the fields you supply are changed.

Provide changes as a JSON payload via --file (use "-" for stdin) and/or with the
convenience flags below. Flags take precedence over fields in the payload.`,
		Example: `  # Rename an object
  anytype-api objects update bafyre... --space bafyre... --name "New name"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var reqBody api.UpdateObjectRequest
			if file != "" {
				if err := readPayload(cmd, file, &reqBody); err != nil {
					return err
				}
			}

			if typeKey != "" {
				reqBody.TypeKey = &typeKey
			}
			if name != "" {
				reqBody.Name = &name
			}
			if markdown != "" {
				reqBody.Markdown = &markdown
			}
			if icon != "" {
				ic, err := emojiIcon(icon)
				if err != nil {
					return err
				}
				reqBody.Icon = ic
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.UpdateObject(cmd.Context(), spaceID, args[0], reqBody)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Updated object:")
			return printObjectDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the object belongs to (required)")
	cmd.Flags().StringVarP(&file, "file", "f", "", `JSON payload file ("-" for stdin)`)
	cmd.Flags().StringVar(&typeKey, "type-key", "", "New type key for the object")
	cmd.Flags().StringVar(&name, "name", "", "New name of the object")
	cmd.Flags().StringVar(&markdown, "body", "", "New markdown body of the object")
	cmd.Flags().StringVar(&icon, "icon", "", "New emoji icon for the object")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newObjectsDeleteCmd() *cobra.Command {
	var (
		spaceID string
		yes     bool
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "delete [object-id]",
		Short: "Delete (archive) an object",
		Args:  cobra.ExactArgs(1),
		Long: `Delete an object. This archives the object rather than permanently removing it.

You are asked to confirm before the object is archived; pass --yes to skip the
prompt when scripting.`,
		Example: `  # Delete an object, with confirmation
  anytype-api objects delete bafyre... --space bafyre...

  # Delete without prompting
  anytype-api objects delete bafyre... --space bafyre... --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			objectID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Archive object %s? This can be undone in the Anytype app.", objectID))
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

			result, err := client.DeleteObject(cmd.Context(), spaceID, objectID)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Deleted object:")
			return printObjectDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the object belongs to (required)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func printObjectsTable(cmd *cobra.Command, result *api.PaginatedResponseObject) error {
	out := cmd.OutOrStdout()

	if result == nil || len(deref(result.Data)) == 0 {
		fmt.Fprintln(out, "No objects found in this space.")
		return nil
	}

	objects := deref(result.Data)
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tID")
	for _, o := range objects {
		typeName := "-"
		if o.Type != nil {
			typeName = orDash(deref(o.Type.Name))
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			orDash(deref(o.Name)),
			typeName,
			shortID(o.Id),
		)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if p := result.Pagination; p != nil {
		shown := len(objects)
		total := derefInt(p.Total)
		fmt.Fprintf(out, "\nShowing %d of %d", shown, total)
		if derefBool(p.HasMore) {
			fmt.Fprintf(out, " (use --offset %d for more)", derefInt(p.Offset)+shown)
		}
		fmt.Fprintln(out)
	}

	return nil
}

func printObjectDetail(cmd *cobra.Command, result *api.ObjectResponse) error {
	out := cmd.OutOrStdout()

	if result == nil || result.Object == nil {
		fmt.Fprintln(out, "Object not found.")
		return nil
	}

	o := result.Object
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", orDash(deref(o.Name)))
	if o.Type != nil {
		fmt.Fprintf(w, "Type:\t%s\n", orDash(deref(o.Type.Name)))
	}
	fmt.Fprintf(w, "Layout:\t%s\n", orDash(deref(o.Layout)))
	fmt.Fprintf(w, "Archived:\t%t\n", deref(o.Archived))
	fmt.Fprintf(w, "Space:\t%s\n", orDash(deref(o.SpaceId)))
	fmt.Fprintf(w, "ID:\t%s\n", orDash(deref(o.Id)))
	if props := deref(o.Properties); len(props) > 0 {
		fmt.Fprintf(w, "Properties:\t%d\n", len(props))
	}
	if md := deref(o.Markdown); md != "" {
		fmt.Fprintf(w, "Body:\t%d chars\n", len(md))
	}
	return w.Flush()
}
