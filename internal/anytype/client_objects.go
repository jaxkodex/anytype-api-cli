package anytype

import (
	"context"
	"fmt"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
)

// ListSpaceObjectsOptions are the parameters for listing the objects within a
// space.
type ListSpaceObjectsOptions struct {
	SpaceID string
	Limit   int
	Offset  int
}

// ListSpaceObjects returns the objects in the given space together with
// pagination metadata.
func (c *Client) ListSpaceObjects(ctx context.Context, opts ListSpaceObjectsOptions) (*api.PaginatedResponseObject, error) {
	params := &api.ListObjectsParams{
		AnytypeVersion: APIVersion,
		Limit:          &opts.Limit,
		Offset:         &opts.Offset,
	}

	resp, err := c.api.ListObjectsWithResponse(ctx, opts.SpaceID, params)
	if err != nil {
		return nil, fmt.Errorf("list objects request failed: %w", err)
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

// GetObject fetches a single object by its id within the given space.
func (c *Client) GetObject(ctx context.Context, spaceID, objectID string) (*api.ObjectResponse, error) {
	params := &api.GetObjectParams{AnytypeVersion: APIVersion}

	resp, err := c.api.GetObjectWithResponse(ctx, spaceID, objectID, params)
	if err != nil {
		return nil, fmt.Errorf("get object request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("object not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("object deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// CreateObject creates a new object in the given space from the supplied request
// body and returns the created object.
func (c *Client) CreateObject(ctx context.Context, spaceID string, body api.CreateObjectRequest) (*api.ObjectResponse, error) {
	params := &api.CreateObjectParams{AnytypeVersion: APIVersion}

	resp, err := c.api.CreateObjectWithResponse(ctx, spaceID, params, body)
	if err != nil {
		return nil, fmt.Errorf("create object request failed: %w", err)
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

// UpdateObject updates an existing object identified by objectID within the
// given space and returns the updated object.
func (c *Client) UpdateObject(ctx context.Context, spaceID, objectID string, body api.UpdateObjectRequest) (*api.ObjectResponse, error) {
	params := &api.UpdateObjectParams{AnytypeVersion: APIVersion}

	resp, err := c.api.UpdateObjectWithResponse(ctx, spaceID, objectID, params, body)
	if err != nil {
		return nil, fmt.Errorf("update object request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON400 != nil:
		return nil, fmt.Errorf("invalid request: %s", derefMessage(resp.JSON400.Message))
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("object not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("object deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON429 != nil:
		return nil, fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// DeleteObject archives the object identified by objectID within the given space
// and returns the archived object.
func (c *Client) DeleteObject(ctx context.Context, spaceID, objectID string) (*api.ObjectResponse, error) {
	params := &api.DeleteObjectParams{AnytypeVersion: APIVersion}

	resp, err := c.api.DeleteObjectWithResponse(ctx, spaceID, objectID, params)
	if err != nil {
		return nil, fmt.Errorf("delete object request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON403 != nil:
		return nil, fmt.Errorf("forbidden: %s", derefMessage(resp.JSON403.Message))
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("object not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("object already deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON429 != nil:
		return nil, fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}
