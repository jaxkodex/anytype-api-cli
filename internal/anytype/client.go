// Package anytype provides a thin, CLI-friendly wrapper around the generated
// Anytype API client. It centralises configuration (read from the environment)
// and request authentication so command implementations stay small.
package anytype

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
)

const (
	// APIVersion is the Anytype API version this CLI targets. It is sent in the
	// required `Anytype-Version` header on every request.
	APIVersion = "2025-11-08"

	// DefaultBaseURL is the address of the local Anytype API server, which the
	// desktop app exposes by default.
	DefaultBaseURL = "http://127.0.0.1:31009"

	// EnvAPIKey is the environment variable holding the bearer token.
	EnvAPIKey = "ANYTYPE_API_KEY"

	// EnvBaseURL optionally overrides the API base URL.
	EnvBaseURL = "ANYTYPE_API_URL"
)

// Config holds the resolved settings used to talk to the API.
type Config struct {
	APIKey  string
	BaseURL string
}

// ConfigFromEnv reads configuration from the environment, applying defaults.
// It returns an error if the required API key is missing so callers can print
// an actionable message.
func ConfigFromEnv() (Config, error) {
	key := os.Getenv(EnvAPIKey)
	if key == "" {
		return Config{}, fmt.Errorf("missing API token: set the %s environment variable to your Anytype API key", EnvAPIKey)
	}

	baseURL := os.Getenv(EnvBaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return Config{APIKey: key, BaseURL: baseURL}, nil
}

// Client wraps the generated typed client.
type Client struct {
	api *api.ClientWithResponses
}

// NewClient builds a Client that authenticates every request with the bearer
// token from the given config.
func NewClient(cfg Config) (*Client, error) {
	authEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		return nil
	}

	c, err := api.NewClientWithResponses(cfg.BaseURL, api.WithRequestEditorFn(authEditor))
	if err != nil {
		return nil, fmt.Errorf("creating API client: %w", err)
	}

	return &Client{api: c}, nil
}

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
	params := &api.SearchGlobalParams{
		AnytypeVersion: APIVersion,
		Limit:          &opts.Limit,
		Offset:         &opts.Offset,
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

// UploadFile uploads the bytes from r as a multipart/form-data "file" field to
// the given space and returns the created file object. The filename is sent in
// the multipart part so the server can preserve the original name.
func (c *Client) UploadFile(ctx context.Context, spaceID, filename string, r io.Reader) (*api.FileUploadResponse, error) {
	params := &api.UploadFileParams{AnytypeVersion: APIVersion}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("building multipart body: %w", err)
	}
	if _, err := io.Copy(part, r); err != nil {
		return nil, fmt.Errorf("reading file contents: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("finalizing multipart body: %w", err)
	}

	resp, err := c.api.UploadFileWithBodyWithResponse(ctx, spaceID, params, writer.FormDataContentType(), &body)
	if err != nil {
		return nil, fmt.Errorf("upload file request failed: %w", err)
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
	case resp.JSON410 != nil:
		return nil, fmt.Errorf("space deleted: %s", derefMessage(resp.JSON410.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// DownloadFile streams the raw bytes of the file identified by fileID within the
// given space. It returns the contents together with the response Content-Type,
// which callers may use to pick a filename or extension.
func (c *Client) DownloadFile(ctx context.Context, spaceID, fileID string) ([]byte, string, error) {
	params := &api.DownloadFileParams{AnytypeVersion: APIVersion}

	resp, err := c.api.DownloadFileWithResponse(ctx, spaceID, fileID, params)
	if err != nil {
		return nil, "", fmt.Errorf("download file request failed: %w", err)
	}

	switch {
	case resp.JSON400 != nil:
		return nil, "", fmt.Errorf("invalid request: %s", derefMessage(resp.JSON400.Message))
	case resp.JSON401 != nil:
		return nil, "", fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON404 != nil:
		return nil, "", fmt.Errorf("file not found: %s", derefMessage(resp.JSON404.Message))
	case resp.JSON500 != nil:
		return nil, "", fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	case resp.StatusCode() == http.StatusOK:
		contentType := ""
		if resp.HTTPResponse != nil {
			contentType = resp.HTTPResponse.Header.Get("Content-Type")
		}
		return resp.Body, contentType, nil
	default:
		return nil, "", fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// DeleteFile removes the file identified by fileID within the given space. By
// default the file is moved to the bin; pass skipBin to permanently delete it.
func (c *Client) DeleteFile(ctx context.Context, spaceID, fileID string, skipBin bool) error {
	params := &api.DeleteFileParams{
		AnytypeVersion: APIVersion,
		SkipBin:        &skipBin,
	}

	resp, err := c.api.DeleteFileWithResponse(ctx, spaceID, fileID, params)
	if err != nil {
		return fmt.Errorf("delete file request failed: %w", err)
	}

	switch {
	case resp.StatusCode() == http.StatusNoContent || resp.JSON204 != nil:
		return nil
	case resp.JSON400 != nil:
		return fmt.Errorf("invalid request: %s", derefMessage(resp.JSON400.Message))
	case resp.JSON401 != nil:
		return fmt.Errorf("unauthorized: check that %s holds a valid API token", EnvAPIKey)
	case resp.JSON500 != nil:
		return fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

func derefMessage(s *string) string {
	if s == nil {
		return "unknown error"
	}
	return *s
}
