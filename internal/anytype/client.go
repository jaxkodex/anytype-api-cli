// Package anytype provides a thin, CLI-friendly wrapper around the generated
// Anytype API client. It centralises configuration (read from the environment)
// and request authentication so command implementations stay small.
//
// The methods on *Client are split across per-resource files (client_search.go,
// client_types.go, client_files.go, client_lists.go) so that concurrent feature
// branches add new files instead of all appending to this one. This file holds
// only the shared Client, its configuration, and helpers.
package anytype

import (
	"context"
	"fmt"
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

func derefMessage(s *string) string {
	if s == nil {
		return "unknown error"
	}
	return *s
}
