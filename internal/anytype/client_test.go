package anytype

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// responseStub describes a canned HTTP response a test server returns. The
// generated client's response parsers only populate the typed JSONnnn fields
// when the Content-Type contains "json" and the status code matches, so tests
// that want a decoded body must set both status and contentType.
type responseStub struct {
	status      int
	contentType string
	body        string
}

// capturedReq is a snapshot of an inbound request. The body is drained inside
// the handler (the live *http.Request.Body is closed by the test server once
// the handler returns, so it cannot be read reliably after the fact).
type capturedReq struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   []byte
}

// newStubServer starts an httptest.Server that returns stub for every request
// and records a snapshot of the most recent request into captured (when
// non-nil). It drains the request body so streaming uploaders (e.g. multipart
// over an io.Pipe) never block waiting for the server to consume bytes.
func newStubServer(t *testing.T, stub responseStub, captured **capturedReq) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
		}
		if captured != nil {
			*captured = &capturedReq{
				Method: r.Method,
				Path:   r.URL.Path,
				Query:  r.URL.Query(),
				Header: r.Header.Clone(),
				Body:   body,
			}
		}
		if stub.contentType != "" {
			w.Header().Set("Content-Type", stub.contentType)
		}
		w.WriteHeader(stub.status)
		_, _ = io.WriteString(w, stub.body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// newStubClient builds an authenticated Client pointing at a stub server. It is
// the shared seam for all wrapper-level tests: the real generated client makes
// real HTTP calls to the test server, so these tests exercise request
// serialization and response parsing end to end without touching the network.
func newStubClient(t *testing.T, stub responseStub, captured **capturedReq) *Client {
	t.Helper()
	srv := newStubServer(t, stub, captured)
	c, err := NewClient(Config{APIKey: "test-token", BaseURL: srv.URL})
	require.NoError(t, err)
	return c
}

// ptr is a small generic helper for building pointer literals in tests.
func ptr[T any](v T) *T { return &v }

func TestConfigFromEnv(t *testing.T) {
	t.Run("missing API key is an error", func(t *testing.T) {
		t.Setenv(EnvAPIKey, "")
		_, err := ConfigFromEnv()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing API token")
		assert.Contains(t, err.Error(), EnvAPIKey)
	})

	t.Run("key present uses default base URL", func(t *testing.T) {
		t.Setenv(EnvAPIKey, "abc123")
		t.Setenv(EnvBaseURL, "")
		cfg, err := ConfigFromEnv()
		require.NoError(t, err)
		assert.Equal(t, "abc123", cfg.APIKey)
		assert.Equal(t, DefaultBaseURL, cfg.BaseURL)
	})

	t.Run("base URL can be overridden", func(t *testing.T) {
		t.Setenv(EnvAPIKey, "abc123")
		t.Setenv(EnvBaseURL, "http://example.test:1234")
		cfg, err := ConfigFromEnv()
		require.NoError(t, err)
		assert.Equal(t, "http://example.test:1234", cfg.BaseURL)
	})
}

func TestNewClient_BadURLSurfacesOnFirstRequest(t *testing.T) {
	// The generated client does not parse the base URL until a request is made,
	// so construction never fails on a bad URL. The wrapper must still surface
	// the per-request failure, wrapped so callers see which call failed.
	c, err := NewClient(Config{APIKey: "x", BaseURL: "://not-a-url"})
	require.NoError(t, err)

	_, err = c.Search(context.Background(), SearchOptions{Query: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "search request failed")
}

// TestRequestWiring asserts the auth and version headers, method and path are
// applied to every outbound request. It uses Search as the vehicle because it
// is the simplest operation; the same wiring is shared by every wrapper method
// via NewClient and the APIVersion constant.
func TestRequestWiring(t *testing.T) {
	var captured *capturedReq
	c := newStubClient(t, responseStub{
		status:      http.StatusOK,
		contentType: "application/json",
		body:        `{"data":[],"pagination":{}}`,
	}, &captured)

	_, err := c.Search(context.Background(), SearchOptions{
		Query:  "roadmap",
		Types:  []string{"page"},
		Limit:  5,
		Offset: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, captured)

	assert.Equal(t, http.MethodPost, captured.Method)
	assert.Equal(t, "/v1/search", captured.Path)
	assert.Equal(t, "Bearer test-token", captured.Header.Get("Authorization"))
	assert.Equal(t, APIVersion, captured.Header.Get("Anytype-Version"))
	assert.Equal(t, "5", captured.Query.Get("limit"))
	assert.Equal(t, "10", captured.Query.Get("offset"))
}
