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

// typeResponseOK is a valid TypeResponse payload: a single type with an emoji
// icon. The icon is a valid Icon union payload so it decodes cleanly.
const typeResponseOK = `{"type":{"id":"ty1","name":"Task","plural_name":"Tasks","key":"task","layout":"basic","archived":false,"icon":{"emoji":"✅","format":"emoji"}}}`

// emojiIcon builds an Icon union from an emoji, mirroring how the command layer
// constructs one. It lets request-body tests supply a realistic icon payload.
func emojiIcon(t *testing.T, emoji string) *api.Icon {
	t.Helper()
	format := api.IconFormatEmoji
	ic := &api.Icon{}
	require.NoError(t, ic.FromEmojiIcon(api.EmojiIcon{Emoji: &emoji, Format: &format}))
	return ic
}

func TestListTypes(t *testing.T) {
	const okBody = `{"data":[{"id":"ty1","name":"Task","layout":"basic","icon":null}],"pagination":{"total":1,"offset":0,"limit":100,"has_more":false}}`

	tests := []struct {
		name    string
		stub    responseStub
		wantErr string
		wantLen int
	}{
		{"success", responseStub{http.StatusOK, "application/json", okBody}, "", 1},
		{"401 unauthorized", responseStub{http.StatusUnauthorized, "application/json", `{"message":"nope"}`}, "unauthorized", 0},
		{"500 server error", responseStub{http.StatusInternalServerError, "application/json", `{"message":"boom"}`}, "server error: boom", 0},
		{"unexpected status", responseStub{http.StatusBadGateway, "text/plain", "nope"}, "unexpected response (HTTP 502)", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured *capturedReq
			c := newStubClient(t, tc.stub, &captured)

			res, err := c.ListTypes(context.Background(), ListTypesOptions{SpaceID: "sp1", Limit: 100, Offset: 0})

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, res)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res.Data)
			assert.Len(t, *res.Data, tc.wantLen)
		})
	}

	t.Run("request is GET /v1/spaces/{space}/types with pagination", func(t *testing.T) {
		var captured *capturedReq
		c := newStubClient(t, responseStub{http.StatusOK, "application/json", okBody}, &captured)

		_, err := c.ListTypes(context.Background(), ListTypesOptions{SpaceID: "sp1", Limit: 50, Offset: 25})
		require.NoError(t, err)
		require.NotNil(t, captured)
		assert.Equal(t, http.MethodGet, captured.Method)
		assert.Equal(t, "/v1/spaces/sp1/types", captured.Path)
		assert.Equal(t, "50", captured.Query.Get("limit"))
		assert.Equal(t, "25", captured.Query.Get("offset"))
	})
}

func TestGetType(t *testing.T) {
	tests := []struct {
		name    string
		stub    responseStub
		wantErr string
		wantNil bool
	}{
		{"success", responseStub{http.StatusOK, "application/json", typeResponseOK}, "", false},
		{"401 unauthorized", responseStub{http.StatusUnauthorized, "application/json", `{}`}, "unauthorized", true},
		{"404 not found", responseStub{http.StatusNotFound, "application/json", `{"message":"gone"}`}, "type not found: gone", true},
		{"410 deleted", responseStub{http.StatusGone, "application/json", `{"message":"archived"}`}, "type deleted: archived", true},
		{"500 server error", responseStub{http.StatusInternalServerError, "application/json", `{"message":"boom"}`}, "server error: boom", true},
		{"unexpected status", responseStub{http.StatusBadGateway, "text/plain", "x"}, "unexpected response (HTTP 502)", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured *capturedReq
			c := newStubClient(t, tc.stub, &captured)

			res, err := c.GetType(context.Background(), "sp1", "ty1")

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, res)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			require.NotNil(t, res.Type)
			assert.Equal(t, "ty1", *res.Type.Id)
			assert.Equal(t, "Task", *res.Type.Name)
		})
	}

	t.Run("request is GET /v1/spaces/{space}/types/{type}", func(t *testing.T) {
		var captured *capturedReq
		c := newStubClient(t, responseStub{http.StatusOK, "application/json", typeResponseOK}, &captured)

		_, err := c.GetType(context.Background(), "sp1", "ty1")
		require.NoError(t, err)
		require.NotNil(t, captured)
		assert.Equal(t, http.MethodGet, captured.Method)
		assert.Equal(t, "/v1/spaces/sp1/types/ty1", captured.Path)
	})
}

func TestCreateType(t *testing.T) {
	tests := []struct {
		name    string
		stub    responseStub
		wantErr string
	}{
		{"success 201", responseStub{http.StatusCreated, "application/json", typeResponseOK}, ""},
		{"400 invalid request", responseStub{http.StatusBadRequest, "application/json", `{"message":"bad name"}`}, "invalid request: bad name"},
		{"401 unauthorized", responseStub{http.StatusUnauthorized, "application/json", `{}`}, "unauthorized"},
		{"429 rate limited", responseStub{http.StatusTooManyRequests, "application/json", `{"message":"slow down"}`}, "rate limited: slow down"},
		{"500 server error", responseStub{http.StatusInternalServerError, "application/json", `{"message":"boom"}`}, "server error: boom"},
		{"unexpected status", responseStub{http.StatusBadGateway, "text/plain", "x"}, "unexpected response (HTTP 502)"},
	}

	body := api.CreateTypeRequest{
		Name:            "Task",
		PluralName:      "Tasks",
		TypeLayoutKind:  api.TypeLayoutBasic,
		Key:             ptr("task"),
		Icon:            emojiIcon(t, "✅"),
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured *capturedReq
			c := newStubClient(t, tc.stub, &captured)

			res, err := c.CreateType(context.Background(), "sp1", body)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, res)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
		})
	}

	t.Run("request is POST with name/plural/layout/icon body", func(t *testing.T) {
		var captured *capturedReq
		c := newStubClient(t, responseStub{http.StatusCreated, "application/json", typeResponseOK}, &captured)

		_, err := c.CreateType(context.Background(), "sp1", body)
		require.NoError(t, err)
		require.NotNil(t, captured)
		assert.Equal(t, http.MethodPost, captured.Method)
		assert.Equal(t, "/v1/spaces/sp1/types", captured.Path)

		var got api.CreateTypeRequest
		require.NoError(t, json.Unmarshal(captured.Body, &got))
		assert.Equal(t, "Task", got.Name)
		assert.Equal(t, "Tasks", got.PluralName)
		assert.Equal(t, api.TypeLayoutBasic, got.TypeLayoutKind)
		require.NotNil(t, got.Key)
		assert.Equal(t, "task", *got.Key)
		require.NotNil(t, got.Icon)
		ei, err := got.Icon.AsEmojiIcon()
		require.NoError(t, err)
		require.NotNil(t, ei.Emoji)
		assert.Equal(t, "✅", *ei.Emoji)
		require.NotNil(t, ei.Format)
		assert.Equal(t, api.IconFormatEmoji, *ei.Format)
	})
}

func TestUpdateType(t *testing.T) {
	tests := []struct {
		name    string
		stub    responseStub
		wantErr string
	}{
		{"success 200", responseStub{http.StatusOK, "application/json", typeResponseOK}, ""},
		{"400 invalid request", responseStub{http.StatusBadRequest, "application/json", `{"message":"bad"}`}, "invalid request: bad"},
		{"401 unauthorized", responseStub{http.StatusUnauthorized, "application/json", `{}`}, "unauthorized"},
		{"404 not found", responseStub{http.StatusNotFound, "application/json", `{"message":"no"}`}, "type not found: no"},
		{"410 deleted", responseStub{http.StatusGone, "application/json", `{"message":"archived"}`}, "type deleted: archived"},
		{"429 rate limited", responseStub{http.StatusTooManyRequests, "application/json", `{"message":"slow"}`}, "rate limited: slow"},
		{"500 server error", responseStub{http.StatusInternalServerError, "application/json", `{"message":"boom"}`}, "server error: boom"},
		{"unexpected status", responseStub{http.StatusBadGateway, "text/plain", "x"}, "unexpected response (HTTP 502)"},
	}

	body := api.UpdateTypeRequest{Name: ptr("New name")}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured *capturedReq
			c := newStubClient(t, tc.stub, &captured)

			res, err := c.UpdateType(context.Background(), "sp1", "ty1", body)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, res)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
		})
	}

	t.Run("request is PATCH /v1/spaces/{space}/types/{type} with body", func(t *testing.T) {
		var captured *capturedReq
		c := newStubClient(t, responseStub{http.StatusOK, "application/json", typeResponseOK}, &captured)

		_, err := c.UpdateType(context.Background(), "sp1", "ty1", body)
		require.NoError(t, err)
		require.NotNil(t, captured)
		assert.Equal(t, http.MethodPatch, captured.Method)
		assert.Equal(t, "/v1/spaces/sp1/types/ty1", captured.Path)

		var got api.UpdateTypeRequest
		require.NoError(t, json.Unmarshal(captured.Body, &got))
		require.NotNil(t, got.Name)
		assert.Equal(t, "New name", *got.Name)
	})
}

func TestDeleteType(t *testing.T) {
	tests := []struct {
		name    string
		stub    responseStub
		wantErr string
	}{
		{"success 200", responseStub{http.StatusOK, "application/json", typeResponseOK}, ""},
		{"401 unauthorized", responseStub{http.StatusUnauthorized, "application/json", `{}`}, "unauthorized"},
		{"403 forbidden", responseStub{http.StatusForbidden, "application/json", `{"message":"no"}`}, "forbidden: no"},
		{"404 not found", responseStub{http.StatusNotFound, "application/json", `{"message":"no"}`}, "type not found: no"},
		{"410 already deleted", responseStub{http.StatusGone, "application/json", `{"message":"archived"}`}, "type already deleted: archived"},
		{"429 rate limited", responseStub{http.StatusTooManyRequests, "application/json", `{"message":"slow"}`}, "rate limited: slow"},
		{"500 server error", responseStub{http.StatusInternalServerError, "application/json", `{"message":"boom"}`}, "server error: boom"},
		{"unexpected status", responseStub{http.StatusBadGateway, "text/plain", "x"}, "unexpected response (HTTP 502)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured *capturedReq
			c := newStubClient(t, tc.stub, &captured)

			res, err := c.DeleteType(context.Background(), "sp1", "ty1")

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, res)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
		})
	}

	t.Run("request is DELETE /v1/spaces/{space}/types/{type}", func(t *testing.T) {
		var captured *capturedReq
		c := newStubClient(t, responseStub{http.StatusOK, "application/json", typeResponseOK}, &captured)

		_, err := c.DeleteType(context.Background(), "sp1", "ty1")
		require.NoError(t, err)
		require.NotNil(t, captured)
		assert.Equal(t, http.MethodDelete, captured.Method)
		assert.Equal(t, "/v1/spaces/sp1/types/ty1", captured.Path)
	})
}
