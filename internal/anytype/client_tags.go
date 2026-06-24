package anytype

import (
	"context"
	"fmt"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
)

// ListTags returns the tags defined on the given property within a space.
func (c *Client) ListTags(ctx context.Context, spaceID, propertyID string) (*api.PaginatedResponseTag, error) {
	params := &api.ListTagsParams{AnytypeVersion: APIVersion}

	resp, err := c.api.ListTagsWithResponse(ctx, spaceID, propertyID, params)
	if err != nil {
		return nil, fmt.Errorf("list tags request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("property not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// GetTag fetches a single tag by its id from the given property within a space.
func (c *Client) GetTag(ctx context.Context, spaceID, propertyID, tagID string) (*api.TagResponse, error) {
	params := &api.GetTagParams{AnytypeVersion: APIVersion}

	resp, err := c.api.GetTagWithResponse(ctx, spaceID, propertyID, tagID, params)
	if err != nil {
		return nil, fmt.Errorf("get tag request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("tag not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("tag deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// CreateTag creates a new tag on the given property within a space from the
// supplied request body and returns the created tag.
func (c *Client) CreateTag(ctx context.Context, spaceID, propertyID string, body api.CreateTagRequest) (*api.TagResponse, error) {
	params := &api.CreateTagParams{AnytypeVersion: APIVersion}

	resp, err := c.api.CreateTagWithResponse(ctx, spaceID, propertyID, params, body)
	if err != nil {
		return nil, fmt.Errorf("create tag request failed: %w", err)
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

// UpdateTag updates an existing tag identified by tagID on the given property
// within a space and returns the updated tag.
func (c *Client) UpdateTag(ctx context.Context, spaceID, propertyID, tagID string, body api.UpdateTagRequest) (*api.TagResponse, error) {
	params := &api.UpdateTagParams{AnytypeVersion: APIVersion}

	resp, err := c.api.UpdateTagWithResponse(ctx, spaceID, propertyID, tagID, params, body)
	if err != nil {
		return nil, fmt.Errorf("update tag request failed: %w", err)
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
		return nil, fmt.Errorf("tag not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("tag deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON429 != nil:
		return nil, fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// DeleteTag deletes the tag identified by tagID on the given property within a
// space and returns the deleted tag.
func (c *Client) DeleteTag(ctx context.Context, spaceID, propertyID, tagID string) (*api.TagResponse, error) {
	params := &api.DeleteTagParams{AnytypeVersion: APIVersion}

	resp, err := c.api.DeleteTagWithResponse(ctx, spaceID, propertyID, tagID, params)
	if err != nil {
		return nil, fmt.Errorf("delete tag request failed: %w", err)
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSON401 != nil:
		return nil, fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON403 != nil:
		return nil, fmt.Errorf("forbidden: %s", derefMessage(resp.JSON403.Message))
	case resp.JSON404 != nil:
		return nil, fmt.Errorf("tag not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("tag already deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON429 != nil:
		return nil, fmt.Errorf("rate limited: %s", derefMessage(resp.JSON429.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}
