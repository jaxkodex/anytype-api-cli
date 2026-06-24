package anytype

import (
	"context"
	"fmt"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
)

// ListPropertiesOptions are the parameters for listing the properties within a
// space.
type ListPropertiesOptions struct {
	SpaceID string
	Limit   int
	Offset  int
}

// ListProperties returns the properties defined in the given space together
// with pagination metadata.
func (c *Client) ListProperties(ctx context.Context, opts ListPropertiesOptions) (*api.PaginatedResponseProperty, error) {
	params := &api.ListPropertiesParams{
		AnytypeVersion: APIVersion,
		Limit:          &opts.Limit,
		Offset:         &opts.Offset,
	}

	resp, err := c.api.ListPropertiesWithResponse(ctx, opts.SpaceID, params)
	if err != nil {
		return nil, fmt.Errorf("list properties request failed: %w", err)
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

// GetProperty fetches a single property by its id within the given space.
func (c *Client) GetProperty(ctx context.Context, spaceID, propertyID string) (*api.PropertyResponse, error) {
	params := &api.GetPropertyParams{AnytypeVersion: APIVersion}

	resp, err := c.api.GetPropertyWithResponse(ctx, spaceID, propertyID, params)
	if err != nil {
		return nil, fmt.Errorf("get property request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("property not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("property deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// CreateProperty creates a new property in the given space from the supplied
// request body and returns the created property.
func (c *Client) CreateProperty(ctx context.Context, spaceID string, body api.CreatePropertyRequest) (*api.PropertyResponse, error) {
	params := &api.CreatePropertyParams{AnytypeVersion: APIVersion}

	resp, err := c.api.CreatePropertyWithResponse(ctx, spaceID, params, body)
	if err != nil {
		return nil, fmt.Errorf("create property request failed: %w", err)
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

// UpdateProperty updates an existing property identified by propertyID within
// the given space and returns the updated property.
func (c *Client) UpdateProperty(ctx context.Context, spaceID, propertyID string, body api.UpdatePropertyRequest) (*api.PropertyResponse, error) {
	params := &api.UpdatePropertyParams{AnytypeVersion: APIVersion}

	resp, err := c.api.UpdatePropertyWithResponse(ctx, spaceID, propertyID, params, body)
	if err != nil {
		return nil, fmt.Errorf("update property request failed: %w", err)
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
		return nil, fmt.Errorf("property not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("property deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON429 != nil:
		return nil, fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// DeleteProperty archives the property identified by propertyID within the
// given space and returns the archived property.
func (c *Client) DeleteProperty(ctx context.Context, spaceID, propertyID string) (*api.PropertyResponse, error) {
	params := &api.DeletePropertyParams{AnytypeVersion: APIVersion}

	resp, err := c.api.DeletePropertyWithResponse(ctx, spaceID, propertyID, params)
	if err != nil {
		return nil, fmt.Errorf("delete property request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON403 != nil:
		return nil, fmt.Errorf("forbidden: %s", derefMessage(resp.JSON403.Message))
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("property not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("property already deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON429 != nil:
		return nil, fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}
