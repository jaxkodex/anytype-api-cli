package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/jaxkodex/anytype-api-cli/internal/anytype"
	"github.com/jaxkodex/anytype-api-cli/internal/api"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var (
		types   []string
		limit   int
		offset  int
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search objects across all spaces",
		Long: `Run a global search over every space accessible to the authenticated user.

The query matches against object names and snippets. Restrict results to
specific object types with one or more --type flags (e.g. page, task, bookmark).`,
		Args: cobra.MaximumNArgs(1),
		Example: `  # Search every space for "roadmap"
  anytype-api search roadmap

  # Only return tasks and pages, limited to 10 results
  anytype-api search "launch" --type task --type page --limit 10

  # Emit raw JSON for scripting
  anytype-api search roadmap --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var query string
			if len(args) == 1 {
				query = args[0]
			}

			cfg, err := anytype.ConfigFromEnv()
			if err != nil {
				return err
			}

			client, err := anytype.NewClient(cfg)
			if err != nil {
				return err
			}

			result, err := client.Search(cmd.Context(), anytype.SearchOptions{
				Query:  query,
				Types:  types,
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				return err
			}

			if jsonOut {
				return printJSON(cmd, result)
			}
			return printTable(cmd, result)
		},
	}

	cmd.Flags().StringSliceVarP(&types, "type", "t", nil, "Object type to include (repeatable), e.g. page, task, bookmark")
	cmd.Flags().IntVarP(&limit, "limit", "L", 100, "Maximum number of results to return (max 1000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")

	return cmd
}

func printJSON(cmd *cobra.Command, result *api.PaginatedResponseObject) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func printTable(cmd *cobra.Command, result *api.PaginatedResponseObject) error {
	out := cmd.OutOrStdout()

	if result == nil {
		fmt.Fprintln(out, "No objects matched your search.")
		return nil
	}

	objects := deref(result.Data)
	if len(objects) == 0 {
		fmt.Fprintln(out, "No objects matched your search.")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tSPACE\tID")
	for _, obj := range objects {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			displayName(obj),
			typeName(obj.Type),
			shortID(obj.SpaceId),
			shortID(obj.Id),
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

func displayName(obj api.Object) string {
	if name := deref(obj.Name); name != "" {
		return name
	}
	if snip := strings.TrimSpace(deref(obj.Snippet)); snip != "" {
		return truncate(snip, 50)
	}
	return "(untitled)"
}

func typeName(t *api.Type) string {
	if t == nil {
		return "-"
	}
	if name := deref(t.Name); name != "" {
		return name
	}
	if key := deref(t.Key); key != "" {
		return key
	}
	return "-"
}

// shortID renders only the leading segment of an Anytype id, which is enough to
// identify a row without overwhelming the table.
func shortID(id *string) string {
	s := deref(id)
	if s == "" {
		return "-"
	}
	return truncate(s, 16)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return string(runes[:n])
	}
	return string(runes[:n-1]) + "…"
}

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func derefInt(p *int) int    { return deref(p) }
func derefBool(p *bool) bool { return deref(p) }
