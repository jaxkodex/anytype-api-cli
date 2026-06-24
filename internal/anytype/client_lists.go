package anytype

import (
	"context"
	"fmt"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
)

// ListViewsOptions are the parameters for listing the views of a list.
type ListViewsOptions struct {
	SpaceID string
	ListID  string
	Limit   int
	Offset  int
}

// ListViews returns the views defined for a list (query or collection) within
// the given space together with pagination metadata.
func (c *Client) ListViews(ctx context.Context, opts ListViewsOptions) (*api.PaginatedResponseView, error) {
	params := &api.GetListViewsParams{
		AnytypeVersion: APIVersion,
		Limit:          &opts.Limit,
		Offset:         &opts.Offset,
	}

	resp, err := c.api.GetListViewsWithResponse(ctx, opts.SpaceID, opts.ListID, params)
	if err != nil {
		return nil, fmt.Errorf("get list views request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("list not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// ListObjectsOptions are the parameters for listing the objects in a list view.
type ListObjectsOptions struct {
	SpaceID string
	ListID  string
	ViewID  string
	Limit   int
	Offset  int
}

// ListObjects returns the objects in a list, filtered and sorted according to
// the given view, together with pagination metadata.
func (c *Client) ListObjects(ctx context.Context, opts ListObjectsOptions) (*api.PaginatedResponseObject, error) {
	params := &api.GetListObjectsParams{
		AnytypeVersion: APIVersion,
		Limit:          &opts.Limit,
		Offset:         &opts.Offset,
	}

	resp, err := c.api.GetListObjectsWithResponse(ctx, opts.SpaceID, opts.ListID, opts.ViewID, params)
	if err != nil {
		return nil, fmt.Errorf("get list objects request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("list or view not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// AddListObjects adds the given object ids to the list (collection) within the
// given space and returns the confirmation message from the API.
func (c *Client) AddListObjects(ctx context.Context, spaceID, listID string, objectIDs []string) (string, error) {
	params := &api.AddListObjectsParams{AnytypeVersion: APIVersion}
	body := api.AddObjectsToListRequest{Objects: &objectIDs}

	resp, err := c.api.AddListObjectsWithResponse(ctx, spaceID, listID, params, body)
	if err != nil {
		return "", fmt.Errorf("add list objects request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return *resp.JSON200, nil
	case resp.JSON400 != nil:
		return "", fmt.Errorf("invalid request: %s", derefMessage(resp.JSON400.Message))
	case resp.JSON401 != nil:
		return "", fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return "", fmt.Errorf("list not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON429 != nil:
		return "", fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return "", fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return "", fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// RemoveListObject removes the given object from the list (collection) within
// the given space and returns the confirmation message from the API.
func (c *Client) RemoveListObject(ctx context.Context, spaceID, listID, objectID string) (string, error) {
	params := &api.RemoveListObjectParams{AnytypeVersion: APIVersion}

	resp, err := c.api.RemoveListObjectWithResponse(ctx, spaceID, listID, objectID, params)
	if err != nil {
		return "", fmt.Errorf("remove list object request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return *resp.JSON200, nil
	case resp.JSON400 != nil:
		return "", fmt.Errorf("invalid request: %s", derefMessage(resp.JSON400.Message))
	case resp.JSON401 != nil:
		return "", fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return "", fmt.Errorf("list or object not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON429 != nil:
		return "", fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return "", fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return "", fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}
