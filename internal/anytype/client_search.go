package anytype

import (
	"context"
	"fmt"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
)

// SearchOptions are the parameters for a global search.
type SearchOptions struct {
	Query  string
	Types  []string
	Limit  int
	Offset int
}

// Search runs a global search across all spaces accessible to the authenticated
// user and returns the matched objects together with pagination metadata.
func (c *Client) Search(ctx context.Context, opts SearchOptions) (*api.PaginatedResponseObject, error) {
	params := &api.SearchGlobalParams{AnytypeVersion: APIVersion}
	// Only send limit/offset when set: a pointer to 0 still marshals (omitempty
	// does not drop pointers-to-zero), which would override the server defaults
	// and could return zero items.
	if opts.Limit > 0 {
		params.Limit = &opts.Limit
	}
	if opts.Offset > 0 {
		params.Offset = &opts.Offset
	}

	body := api.SearchRequest{}
	if opts.Query != "" {
		body.Query = &opts.Query
	}
	if len(opts.Types) > 0 {
		body.Types = &opts.Types
	}

	resp, err := c.api.SearchGlobalWithResponse(ctx, params, body)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}
