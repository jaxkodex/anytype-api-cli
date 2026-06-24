package anytype

import (
	"context"
	"fmt"
	"os"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
)

// NewUnauthenticatedClient builds a Client for the auth endpoints, which do not
// require an API key. The base URL is read from EnvBaseURL, falling back to
// DefaultBaseURL. Use this to bootstrap an API key via the challenge flow.
func NewUnauthenticatedClient() (*Client, error) {
	baseURL := os.Getenv(EnvBaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return NewClient(Config{BaseURL: baseURL})
}

// CreateChallenge starts an authentication challenge for the given app name.
// The Anytype desktop app then displays a 4-digit code that, together with the
// returned challenge id, is exchanged for an API key via CreateAPIKey.
//
// This endpoint does not require an API key, so it can be called before one has
// been obtained.
func (c *Client) CreateChallenge(ctx context.Context, appName string) (*api.CreateChallengeResponse, error) {
	params := &api.CreateAuthChallengeParams{AnytypeVersion: APIVersion}
	body := api.CreateChallengeRequest{AppName: &appName}

	resp, err := c.api.CreateAuthChallengeWithResponse(ctx, params, body)
	if err != nil {
		return nil, fmt.Errorf("create challenge request failed: %w", err)
	}

	switch {
	case resp.JSON201 != nil:
		return resp.JSON201, nil
	case resp.JSON400 != nil:
		return nil, fmt.Errorf("invalid request: %s", derefMessage(resp.JSON400.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}

// CreateAPIKey exchanges a challenge id and the 4-digit code shown in the
// Anytype desktop app for an API key that can be used to authenticate further
// requests.
//
// This endpoint does not require an existing API key.
func (c *Client) CreateAPIKey(ctx context.Context, challengeID, code string) (*api.CreateApiKeyResponse, error) {
	params := &api.CreateAuthApiKeyParams{AnytypeVersion: APIVersion}
	body := api.CreateApiKeyRequest{ChallengeId: &challengeID, Code: &code}

	resp, err := c.api.CreateAuthApiKeyWithResponse(ctx, params, body)
	if err != nil {
		return nil, fmt.Errorf("create api key request failed: %w", err)
	}

	switch {
	case resp.JSON201 != nil:
		return resp.JSON201, nil
	case resp.JSON400 != nil:
		return nil, fmt.Errorf("invalid request: %s", derefMessage(resp.JSON400.Message))
	case resp.JSON500 != nil:
		return nil, fmt.Errorf("server error: %s", derefMessage(resp.JSON500.Message))
	default:
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode(), string(resp.Body))
	}
}
