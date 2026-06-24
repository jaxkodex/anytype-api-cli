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

func TestSearch(t *testing.T) {
	// A minimal but valid success payload. Object's icon/type are required
	// (non-omitempty) on the model, so "null" keeps the pointers nil without
	// needing a valid Icon union payload.
	const okBody = `{"data":[
		{"id":"obj1","name":"Roadmap","snippet":"q2 plan","space_id":"sp1","icon":null,"type":null},
		{"id":"obj2","name":"Launch","snippet":"","space_id":"sp2","icon":null,"type":null}
	],"pagination":{"total":2,"offset":10,"limit":5,"has_more":false}}`

	tests := []struct {
		name      string
		stub      responseStub
		wantErr   string
		wantObjs  int
		wantTotal int
	}{
		{
			name:      "success decodes objects and pagination",
			stub:      responseStub{status: http.StatusOK, contentType: "application/json", body: okBody},
			wantObjs:  2,
			wantTotal: 2,
		},
		{
			name:    "401 maps to unauthorized",
			stub:    responseStub{status: http.StatusUnauthorized, contentType: "application/json", body: `{"message":"nope"}`},
			wantErr: "unauthorized",
		},
		{
			name:    "500 maps to server error with message",
			stub:    responseStub{status: http.StatusInternalServerError, contentType: "application/json", body: `{"message":"boom"}`},
			wantErr: "server error: boom",
		},
		{
			name:    "500 without message falls back to unknown error",
			stub:    responseStub{status: http.StatusInternalServerError, contentType: "application/json", body: `{}`},
			wantErr: "server error: unknown error",
		},
		{
			name:    "unexpected status surfaces body verbatim",
			stub:    responseStub{status: http.StatusBadGateway, contentType: "text/plain", body: "oops"},
			wantErr: "unexpected response (HTTP 502): oops",
		},
		{
			// JSON body returned at a status the parser does not handle lands in
			// the default branch too.
			name:    "403 at search endpoint is unexpected",
			stub:    responseStub{status: http.StatusForbidden, contentType: "application/json", body: `{"message":"no"}`},
			wantErr: "unexpected response (HTTP 403)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured *capturedReq
			c := newStubClient(t, tc.stub, &captured)

			res, err := c.Search(context.Background(), SearchOptions{
				Query: "roadmap", Types: []string{"page"}, Limit: 5, Offset: 10,
			})

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, res)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			require.NotNil(t, res.Data)
			assert.Len(t, *res.Data, tc.wantObjs)
			if res.Pagination != nil && res.Pagination.Total != nil {
				assert.Equal(t, tc.wantTotal, *res.Pagination.Total)
			}
		})
	}
}

func TestSearch_SerializesQueryAndTypes(t *testing.T) {
	var captured *capturedReq
	c := newStubClient(t, responseStub{
		status:      http.StatusOK,
		contentType: "application/json",
		body:        `{"data":[],"pagination":{}}`,
	}, &captured)

	_, err := c.Search(context.Background(), SearchOptions{
		Query: "roadmap", Types: []string{"page", "task"}, Limit: 5, Offset: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, captured)

	var body api.SearchRequest
	require.NoError(t, json.Unmarshal(captured.Body, &body))
	require.NotNil(t, body.Query)
	assert.Equal(t, "roadmap", *body.Query)
	require.NotNil(t, body.Types)
	assert.Equal(t, []string{"page", "task"}, *body.Types)
}

func TestSearch_OmitsEmptyQueryAndTypes(t *testing.T) {
	var captured *capturedReq
	c := newStubClient(t, responseStub{
		status:      http.StatusOK,
		contentType: "application/json",
		body:        `{"data":[],"pagination":{}}`,
	}, &captured)

	_, err := c.Search(context.Background(), SearchOptions{Limit: 1, Offset: 0})
	require.NoError(t, err)
	require.NotNil(t, captured)

	var body api.SearchRequest
	require.NoError(t, json.Unmarshal(captured.Body, &body))
	assert.Nil(t, body.Query)
	assert.Nil(t, body.Types)
}
