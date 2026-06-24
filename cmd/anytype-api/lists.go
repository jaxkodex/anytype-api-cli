package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/jaxkodex/anytype-api-cli/internal/anytype"
	"github.com/jaxkodex/anytype-api-cli/internal/api"
	"github.com/spf13/cobra"
)

func newListsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lists",
		Short: "Work with the lists (collections and sets) in a space",
		Long: `Work with lists (collections and sets) within a space.

A list has one or more views (grid, list, gallery, ...) that filter and sort
its objects. Use these subcommands to inspect a list's views and objects, and to
add or remove objects from a collection.

Lists are scoped to a space, so every subcommand requires a --space id. Find a
list id by searching the space for objects of type "collection" or "set".`,
	}

	cmd.AddCommand(newListsViewsCmd())
	cmd.AddCommand(newListsObjectsCmd())
	cmd.AddCommand(newListsAddCmd())
	cmd.AddCommand(newListsRemoveCmd())

	return cmd
}

func newListsViewsCmd() *cobra.Command {
	var (
		spaceID string
		limit   int
		offset  int
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "views [list-id]",
		Short: "List the views defined for a list",
		Args:  cobra.ExactArgs(1),
		Example: `  # List every view in a list
  anytype-api lists views bafyre... --space bafyre...

  # Emit raw JSON for scripting
  anytype-api lists views bafyre... --space bafyre... --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.ListViews(cmd.Context(), anytype.ListViewsOptions{
				SpaceID: spaceID,
				ListID:  args[0],
				Limit:   limit,
				Offset:  offset,
			})
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			return printViewsTable(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the list belongs to (required)")
	cmd.Flags().IntVarP(&limit, "limit", "L", 100, "Maximum number of results to return (max 1000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newListsObjectsCmd() *cobra.Command {
	var (
		spaceID string
		viewID  string
		limit   int
		offset  int
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "objects [list-id]",
		Short: "List the objects in a list view",
		Args:  cobra.ExactArgs(1),
		Long: `List the objects in a list, filtered and sorted according to a view.

Find view ids with "anytype-api lists views <list-id> --space <id>".`,
		Example: `  # List the objects in a view
  anytype-api lists objects bafyre... --space bafyre... --view 67bf3f21...

  # Emit raw JSON for scripting
  anytype-api lists objects bafyre... --space bafyre... --view 67bf3f21... --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.ListObjects(cmd.Context(), anytype.ListObjectsOptions{
				SpaceID: spaceID,
				ListID:  args[0],
				ViewID:  viewID,
				Limit:   limit,
				Offset:  offset,
			})
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			return printTable(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the list belongs to (required)")
	cmd.Flags().StringVar(&viewID, "view", "", "View id to filter and sort by (required)")
	cmd.Flags().IntVarP(&limit, "limit", "L", 100, "Maximum number of results to return (max 1000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")
	cmd.MarkFlagRequired("view")

	return cmd
}

func newListsAddCmd() *cobra.Command {
	var (
		spaceID string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "add [list-id] [object-id...]",
		Short: "Add objects to a list (collection)",
		Args:  cobra.MinimumNArgs(2),
		Long: `Add one or more objects to a list. Only collections can be modified this
way; the objects of a set are determined by its query.`,
		Example: `  # Add two objects to a collection
  anytype-api lists add bafyre... --space bafyre... bafyreA... bafyreB...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			listID := args[0]
			objectIDs := args[1:]

			client, err := newClient()
			if err != nil {
				return err
			}

			msg, err := client.AddListObjects(cmd.Context(), spaceID, listID, objectIDs)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, map[string]string{"message": msg})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added %d object(s) to list %s: %s\n", len(objectIDs), listID, msg)
			return nil
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the list belongs to (required)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newListsRemoveCmd() *cobra.Command {
	var (
		spaceID string
		yes     bool
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "remove [list-id] [object-id]",
		Short: "Remove an object from a list (collection)",
		Args:  cobra.ExactArgs(2),
		Long: `Remove an object from a list. This only removes the object from the
collection; the underlying object is not deleted.

You are asked to confirm before the object is removed; pass --yes to skip the
prompt when scripting.`,
		Example: `  # Remove an object from a collection, with confirmation
  anytype-api lists remove bafyre... bafyreA... --space bafyre...

  # Remove without prompting
  anytype-api lists remove bafyre... bafyreA... --space bafyre... --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			listID := args[0]
			objectID := args[1]

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Remove object %s from list %s?", objectID, listID))
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

			msg, err := client.RemoveListObject(cmd.Context(), spaceID, listID, objectID)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, map[string]string{"message": msg})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed object %s from list %s: %s\n", objectID, listID, msg)
			return nil
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the list belongs to (required)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func printViewsTable(cmd *cobra.Command, result *api.PaginatedResponseView) error {
	out := cmd.OutOrStdout()

	if result == nil || len(deref(result.Data)) == 0 {
		fmt.Fprintln(out, "No views found for this list.")
		return nil
	}

	views := deref(result.Data)
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tLAYOUT\tFILTERS\tSORTS\tID")
	for _, v := range views {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
			orDash(deref(v.Name)),
			orDash(string(deref(v.Layout))),
			len(deref(v.Filters)),
			len(deref(v.Sorts)),
			shortID(v.Id),
		)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if p := result.Pagination; p != nil {
		shown := len(views)
		total := derefInt(p.Total)
		fmt.Fprintf(out, "\nShowing %d of %d", shown, total)
		if derefBool(p.HasMore) {
			fmt.Fprintf(out, " (use --offset %d for more)", derefInt(p.Offset)+shown)
		}
		fmt.Fprintln(out)
	}

	return nil
}
