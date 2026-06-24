package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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
	cmd.AddCommand(newTypesCreateCmd())
	cmd.AddCommand(newTypesUpdateCmd())
	cmd.AddCommand(newTypesDeleteCmd())

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

func newTypesCreateCmd() *cobra.Command {
	var (
		spaceID string
		file    string
		name    string
		plural  string
		key     string
		layout  string
		icon    string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new type in a space",
		Args:  cobra.NoArgs,
		Long: `Create a new object type in a space.

The type definition may be supplied as a JSON payload via --file (use "-" to
read from stdin) and/or with the convenience flags below. Flags take precedence
over fields in the payload, so you can use a file as a template and tweak it.

A type requires a name, plural name and layout. Layout is one of: basic, note,
profile, action.`,
		Example: `  # Create a type with convenience flags
  anytype-api types create --space bafyre... \
    --name Task --plural Tasks --layout basic --icon ✅

  # Create from a JSON payload on stdin
  cat type.json | anytype-api types create --space bafyre... --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var body api.CreateTypeRequest
			if file != "" {
				if err := readPayload(cmd, file, &body); err != nil {
					return err
				}
			}

			if name != "" {
				body.Name = name
			}
			if plural != "" {
				body.PluralName = plural
			}
			if cmd.Flags().Changed("key") {
				body.Key = &key
			}
			if layout != "" {
				body.TypeLayoutKind = api.TypeLayoutKind(layout)
			}
			if icon != "" {
				ic, err := emojiIcon(icon)
				if err != nil {
					return err
				}
				body.Icon = ic
			}

			if body.Name == "" {
				return fmt.Errorf("a name is required: pass --name or include it in --file")
			}
			if body.PluralName == "" {
				return fmt.Errorf("a plural name is required: pass --plural or include it in --file")
			}
			if body.TypeLayoutKind == "" {
				return fmt.Errorf("a layout is required: pass --layout (basic, note, profile, action) or include it in --file")
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.CreateType(cmd.Context(), spaceID, body)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Created type:")
			return printTypeDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id to create the type in (required)")
	cmd.Flags().StringVarP(&file, "file", "f", "", `JSON payload file ("-" for stdin)`)
	cmd.Flags().StringVar(&name, "name", "", "Name of the type")
	cmd.Flags().StringVar(&plural, "plural", "", "Plural name of the type")
	cmd.Flags().StringVar(&key, "key", "", "Key of the type (snake_case)")
	cmd.Flags().StringVar(&layout, "layout", "", "Layout: basic, note, profile or action")
	cmd.Flags().StringVar(&icon, "icon", "", "Emoji icon for the type")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newTypesUpdateCmd() *cobra.Command {
	var (
		spaceID string
		file    string
		name    string
		plural  string
		key     string
		layout  string
		icon    string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "update [type-id]",
		Short: "Update an existing type",
		Args:  cobra.ExactArgs(1),
		Long: `Update an existing type. Only the fields you supply are changed.

Provide changes as a JSON payload via --file (use "-" for stdin) and/or with the
convenience flags below. Flags take precedence over fields in the payload.`,
		Example: `  # Rename a type
  anytype-api types update bafyre... --space bafyre... --name "New name"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			// UpdateTypeRequest.Icon is serialized without omitempty, so a nil
			// icon would clear the type's icon on every update. Seed the body
			// with the existing icon so partial updates preserve it; a --file or
			// --icon that supplies an icon still overrides it below.
			existing, err := client.GetType(cmd.Context(), spaceID, args[0])
			if err != nil {
				return err
			}

			var body api.UpdateTypeRequest
			if existing.Type != nil {
				body.Icon = existing.Type.Icon
			}
			if file != "" {
				if err := readPayload(cmd, file, &body); err != nil {
					return err
				}
			}

			if name != "" {
				body.Name = &name
			}
			if plural != "" {
				body.PluralName = &plural
			}
			if cmd.Flags().Changed("key") {
				body.Key = &key
			}
			if layout != "" {
				l := api.TypeLayoutKind(layout)
				body.TypeLayoutKind = &l
			}
			if icon != "" {
				ic, err := emojiIcon(icon)
				if err != nil {
					return err
				}
				body.Icon = ic
			}

			result, err := client.UpdateType(cmd.Context(), spaceID, args[0], body)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Updated type:")
			return printTypeDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the type belongs to (required)")
	cmd.Flags().StringVarP(&file, "file", "f", "", `JSON payload file ("-" for stdin)`)
	cmd.Flags().StringVar(&name, "name", "", "New name of the type")
	cmd.Flags().StringVar(&plural, "plural", "", "New plural name of the type")
	cmd.Flags().StringVar(&key, "key", "", "New key of the type (snake_case)")
	cmd.Flags().StringVar(&layout, "layout", "", "New layout: basic, note, profile or action")
	cmd.Flags().StringVar(&icon, "icon", "", "New emoji icon for the type")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newTypesDeleteCmd() *cobra.Command {
	var (
		spaceID string
		yes     bool
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "delete [type-id]",
		Short: "Delete (archive) a type",
		Args:  cobra.ExactArgs(1),
		Long: `Delete a type. This archives the type rather than permanently removing it.

You are asked to confirm before the type is archived; pass --yes to skip the
prompt when scripting.`,
		Example: `  # Delete a type, with confirmation
  anytype-api types delete bafyre... --space bafyre...

  # Delete without prompting
  anytype-api types delete bafyre... --space bafyre... --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			typeID := args[0]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Archive type %s? This can be undone in the Anytype app.", typeID))
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

			result, err := client.DeleteType(cmd.Context(), spaceID, typeID)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Deleted type:")
			return printTypeDetail(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the type belongs to (required)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

// readPayload decodes a JSON payload from the given path into v. A path of "-"
// reads from the command's stdin.
func readPayload(cmd *cobra.Command, path string, v any) error {
	var r io.Reader
	if path == "-" {
		r = cmd.InOrStdin()
	} else {
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("opening payload file: %w", err)
		}
		defer f.Close()
		r = f
	}

	dec := json.NewDecoder(r)
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("parsing JSON payload: %w", err)
	}
	return nil
}

// emojiIcon builds an Icon union value from an emoji string.
func emojiIcon(emoji string) (*api.Icon, error) {
	format := api.IconFormatEmoji
	var ic api.Icon
	if err := ic.FromEmojiIcon(api.EmojiIcon{Emoji: &emoji, Format: &format}); err != nil {
		return nil, fmt.Errorf("building icon: %w", err)
	}
	return &ic, nil
}

// confirm prompts the user for a yes/no answer on the command's input. It
// returns true only when the user explicitly answers yes.
func confirm(cmd *cobra.Command, prompt string) (bool, error) {
	fmt.Fprintf(cmd.OutOrStdout(), "%s [y/N]: ", prompt)

	reader := bufio.NewReader(cmd.InOrStdin())
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("reading confirmation: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
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
