package anytype

import (
	"context"
	"fmt"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
)

// ListTypesOptions are the parameters for listing the types within a space.
type ListTypesOptions struct {
	SpaceID string
	Limit   int
	Offset  int
}

// ListTypes returns the object types defined in the given space together with
// pagination metadata.
func (c *Client) ListTypes(ctx context.Context, opts ListTypesOptions) (*api.PaginatedResponseType, error) {
	params := &api.ListTypesParams{
		AnytypeVersion: APIVersion,
		Limit:          &opts.Limit,
		Offset:         &opts.Offset,
	}

	resp, err := c.api.ListTypesWithResponse(ctx, opts.SpaceID, params)
	if err != nil {
		return nil, fmt.Errorf("list types request failed: %w", err)
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

// GetType fetches a single type by its id within the given space.
func (c *Client) GetType(ctx context.Context, spaceID, typeID string) (*api.TypeResponse, error) {
	params := &api.GetTypeParams{AnytypeVersion: APIVersion}

	resp, err := c.api.GetTypeWithResponse(ctx, spaceID, typeID, params)
	if err != nil {
		return nil, fmt.Errorf("get type request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("type not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("type deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// CreateType creates a new object type in the given space from the supplied
// request body and returns the created type.
func (c *Client) CreateType(ctx context.Context, spaceID string, body api.CreateTypeRequest) (*api.TypeResponse, error) {
	params := &api.CreateTypeParams{AnytypeVersion: APIVersion}

	resp, err := c.api.CreateTypeWithResponse(ctx, spaceID, params, body)
	if err != nil {
		return nil, fmt.Errorf("create type request failed: %w", err)
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

// UpdateType updates an existing type identified by typeID within the given
// space and returns the updated type.
func (c *Client) UpdateType(ctx context.Context, spaceID, typeID string, body api.UpdateTypeRequest) (*api.TypeResponse, error) {
	params := &api.UpdateTypeParams{AnytypeVersion: APIVersion}

	resp, err := c.api.UpdateTypeWithResponse(ctx, spaceID, typeID, params, body)
	if err != nil {
		return nil, fmt.Errorf("update type request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON400 != nil:
		return nil, fmt.Errorf("invalid request: %s", derefMessage(resp.JSON400.Message))
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("type not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("type deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON429 != nil:
		return nil, fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// DeleteType archives the type identified by typeID within the given space and
// returns the archived type.
func (c *Client) DeleteType(ctx context.Context, spaceID, typeID string) (*api.TypeResponse, error) {
	params := &api.DeleteTypeParams{AnytypeVersion: APIVersion}

	resp, err := c.api.DeleteTypeWithResponse(ctx, spaceID, typeID, params)
	if err != nil {
		return nil, fmt.Errorf("delete type request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON403 != nil:
		return nil, fmt.Errorf("forbidden: %s", derefMessage(resp.JSON403.Message))
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("type not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("type already deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON429 != nil:
		return nil, fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}
