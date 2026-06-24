package anytype

import (
	"context"
	"fmt"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
)

// ListSpacesOptions are the parameters for listing the spaces accessible to the
// authenticated user.
type ListSpacesOptions struct {
	Limit  int
	Offset int
}

// ListSpaces returns the spaces accessible to the authenticated user together
// with pagination metadata.
func (c *Client) ListSpaces(ctx context.Context, opts ListSpacesOptions) (*api.PaginatedResponseSpace, error) {
	params := &api.ListSpacesParams{
		AnytypeVersion: APIVersion,
	}
	if opts.Limit > 0 {
		params.Limit = &opts.Limit
	}
	if opts.Offset > 0 {
		params.Offset = &opts.Offset
	}

	resp, err := c.api.ListSpacesWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list spaces request failed: %w", err)
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

// GetSpace fetches a single space by its id.
func (c *Client) GetSpace(ctx context.Context, spaceID string) (*api.SpaceResponse, error) {
	params := &api.GetSpaceParams{AnytypeVersion: APIVersion}

	resp, err := c.api.GetSpaceWithResponse(ctx, spaceID, params)
	if err != nil {
		return nil, fmt.Errorf("get space request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("space not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// CreateSpace creates a new space from the supplied request body and returns the
// created space.
func (c *Client) CreateSpace(ctx context.Context, body api.CreateSpaceRequest) (*api.SpaceResponse, error) {
	params := &api.CreateSpaceParams{AnytypeVersion: APIVersion}

	resp, err := c.api.CreateSpaceWithResponse(ctx, params, body)
	if err != nil {
		return nil, fmt.Errorf("create space request failed: %w", err)
	}

	switch {
	case resp.JSON201 != nil:
		return resp.JSON201, nil
	case resp.JSON400 != nil:
		return nil, fmt.Errorf("invalid request: %s", derefMessage(resp.JSON400.Message))
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON429 != nil:
		return nil, fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// UpdateSpace updates the name and/or description of an existing space and
// returns the updated space.
func (c *Client) UpdateSpace(ctx context.Context, spaceID string, body api.UpdateSpaceRequest) (*api.SpaceResponse, error) {
	params := &api.UpdateSpaceParams{AnytypeVersion: APIVersion}

	resp, err := c.api.UpdateSpaceWithResponse(ctx, spaceID, params, body)
	if err != nil {
		return nil, fmt.Errorf("update space request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON400 != nil:
		return nil, fmt.Errorf("invalid request: %s", derefMessage(resp.JSON400.Message))
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON403 != nil:
		return nil, fmt.Errorf("forbidden: %s", derefMessage(resp.JSON403.Message))
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("space not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON429 != nil:
		return nil, fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}
