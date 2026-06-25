package anytype

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	viewsOKBody   = `{"data":[{"id":"view1","name":"All","layout":"list"}],"pagination":{"total":1,"offset":0,"limit":100,"has_more":false}}`
	objectsOKBody = `{"data":[{"id":"obj1","name":"Task","space_id":"sp1","icon":null,"type":null}],"pagination":{"total":1,"offset":0,"limit":100,"has_more":false}}`
)

func TestListViews(t *testing.T) {
	tests := []struct {
		name    string
		stub    responseStub
		wantErr string
	}{
		{"success", responseStub{http.StatusOK, "application/json", viewsOKBody}, ""},
		{"401 unauthorized", responseStub{http.StatusUnauthorized, "application/json", `{}`}, "unauthorized"},
		{"404 list not found", responseStub{http.StatusNotFound, "application/json", `{"message":"no list"}`}, "list not found: no list"},
		{"500 server error", responseStub{http.StatusInternalServerError, "application/json", `{"message":"boom"}`}, "server error: boom"},
		{"unexpected status", responseStub{http.StatusBadGateway, "text/plain", "x"}, "unexpected response (HTTP 502)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured *capturedReq
			c := newStubClient(t, tc.stub, &captured)

			res, err := c.ListViews(context.Background(), ListViewsOptions{SpaceID: "sp1", ListID: "list1", Limit: 100, Offset: 0})

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, res)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res.Data)
			assert.Len(t, *res.Data, 1)
		})
	}

	t.Run("request is GET /v1/spaces/{space}/lists/{list}/views", func(t *testing.T) {
		var captured *capturedReq
		c := newStubClient(t, responseStub{http.StatusOK, "application/json", viewsOKBody}, &captured)

		_, err := c.ListViews(context.Background(), ListViewsOptions{SpaceID: "sp1", ListID: "list1", Limit: 50, Offset: 25})
		require.NoError(t, err)
		require.NotNil(t, captured)
		assert.Equal(t, http.MethodGet, captured.Method)
		assert.Equal(t, "/v1/spaces/sp1/lists/list1/views", captured.Path)
		assert.Equal(t, "50", captured.Query.Get("limit"))
		assert.Equal(t, "25", captured.Query.Get("offset"))
	})
}

func TestListObjects(t *testing.T) {
	tests := []struct {
		name    string
		stub    responseStub
		wantErr string
	}{
		{"success", responseStub{http.StatusOK, "application/json", objectsOKBody}, ""},
		{"401 unauthorized", responseStub{http.StatusUnauthorized, "application/json", `{}`}, "unauthorized"},
		{"404 list or view not found", responseStub{http.StatusNotFound, "application/json", `{"message":"no view"}`}, "list or view not found: no view"},
		{"500 server error", responseStub{http.StatusInternalServerError, "application/json", `{"message":"boom"}`}, "server error: boom"},
		{"unexpected status", responseStub{http.StatusBadGateway, "text/plain", "x"}, "unexpected response (HTTP 502)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured *capturedReq
			c := newStubClient(t, tc.stub, &captured)

			res, err := c.ListObjects(context.Background(), ListObjectsOptions{SpaceID: "sp1", ListID: "list1", ViewID: "view1", Limit: 100, Offset: 0})

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, res)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res.Data)
			assert.Len(t, *res.Data, 1)
		})
	}

	t.Run("request is GET /v1/spaces/{space}/lists/{list}/views/{view}/objects", func(t *testing.T) {
		var captured *capturedReq
		c := newStubClient(t, responseStub{http.StatusOK, "application/json", objectsOKBody}, &captured)

		_, err := c.ListObjects(context.Background(), ListObjectsOptions{SpaceID: "sp1", ListID: "list1", ViewID: "view1", Limit: 10, Offset: 5})
		require.NoError(t, err)
		require.NotNil(t, captured)
		assert.Equal(t, http.MethodGet, captured.Method)
		assert.Equal(t, "/v1/spaces/sp1/lists/list1/views/view1/objects", captured.Path)
		assert.Equal(t, "10", captured.Query.Get("limit"))
		assert.Equal(t, "5", captured.Query.Get("offset"))
	})
}

func TestAddListObjects(t *testing.T) {
	// The success response body is a JSON-quoted string; the wrapper returns the
	// decoded string verbatim.
	tests := []struct {
		name    string
		stub    responseStub
		wantErr string
		wantMsg string
	}{
		{"success", responseStub{http.StatusOK, "application/json", `"added"`}, "", "added"},
		{"400 invalid request", responseStub{http.StatusBadRequest, "application/json", `{"message":"bad"}`}, "invalid request: bad", ""},
		{"401 unauthorized", responseStub{http.StatusUnauthorized, "application/json", `{}`}, "unauthorized", ""},
		{"404 list not found", responseStub{http.StatusNotFound, "application/json", `{"message":"no list"}`}, "list not found: no list", ""},
		{"429 rate limited", responseStub{http.StatusTooManyRequests, "application/json", `{"message":"slow"}`}, "rate limited: slow", ""},
		{"500 server error", responseStub{http.StatusInternalServerError, "application/json", `{"message":"boom"}`}, "server error: boom", ""},
		{"unexpected status", responseStub{http.StatusBadGateway, "text/plain", "x"}, "unexpected response (HTTP 502)", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured *capturedReq
			c := newStubClient(t, tc.stub, &captured)

			msg, err := c.AddListObjects(context.Background(), "sp1", "list1", []string{"a", "b"})

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Empty(t, msg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantMsg, msg)
		})
	}

	t.Run("request is POST with objects array body", func(t *testing.T) {
		var captured *capturedReq
		c := newStubClient(t, responseStub{http.StatusOK, "application/json", `"added"`}, &captured)

		_, err := c.AddListObjects(context.Background(), "sp1", "list1", []string{"a", "b"})
		require.NoError(t, err)
		require.NotNil(t, captured)
		assert.Equal(t, http.MethodPost, captured.Method)
		assert.Equal(t, "/v1/spaces/sp1/lists/list1/objects", captured.Path)

		var got api.AddObjectsToListRequest
		require.NoError(t, json.Unmarshal(captured.Body, &got))
		require.NotNil(t, got.Objects)
		assert.Equal(t, []string{"a", "b"}, *got.Objects)
	})
}

func TestRemoveListObject(t *testing.T) {
	tests := []struct {
		name    string
		stub    responseStub
		wantErr string
		wantMsg string
	}{
		{"success", responseStub{http.StatusOK, "application/json", `"removed"`}, "", "removed"},
		{"400 invalid request", responseStub{http.StatusBadRequest, "application/json", `{"message":"bad"}`}, "invalid request: bad", ""},
		{"401 unauthorized", responseStub{http.StatusUnauthorized, "application/json", `{}`}, "unauthorized", ""},
		{"404 list or object not found", responseStub{http.StatusNotFound, "application/json", `{"message":"no"}`}, "list or object not found: no", ""},
		{"429 rate limited", responseStub{http.StatusTooManyRequests, "application/json", `{"message":"slow"}`}, "rate limited: slow", ""},
		{"500 server error", responseStub{http.StatusInternalServerError, "application/json", `{"message":"boom"}`}, "server error: boom", ""},
		{"unexpected status", responseStub{http.StatusBadGateway, "text/plain", "x"}, "unexpected response (HTTP 502)", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured *capturedReq
			c := newStubClient(t, tc.stub, &captured)

			msg, err := c.RemoveListObject(context.Background(), "sp1", "list1", "obj1")

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Empty(t, msg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantMsg, msg)
		})
	}

	t.Run("request is DELETE /v1/spaces/{space}/lists/{list}/objects/{object}", func(t *testing.T) {
		var captured *capturedReq
		c := newStubClient(t, responseStub{http.StatusOK, "application/json", `"removed"`}, &captured)

		_, err := c.RemoveListObject(context.Background(), "sp1", "list1", "obj1")
		require.NoError(t, err)
		require.NotNil(t, captured)
		assert.Equal(t, http.MethodDelete, captured.Method)
		assert.Equal(t, "/v1/spaces/sp1/lists/list1/objects/obj1", captured.Path)
	})
}
