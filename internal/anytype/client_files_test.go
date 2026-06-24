package anytype

import (
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const uploadOKBody = `{"object_id":"file1","name":"photo.png","media":"image/png","extension":"png","size_in_bytes":11}`

// newUploadClient builds a client whose server drains the multipart upload
// body (so the wrapper's io.Pipe streaming goroutine never blocks) and records
// the "file" part's filename and contents, plus a request snapshot.
func newUploadClient(t *testing.T, stub responseStub, cap **capturedReq, gotFile *string, gotContent *[]byte) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Re-parse the drained multipart bytes to inspect the "file" part. The
		// boundary is carried in the request's Content-Type header.
		_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		require.NoError(t, err)
		mr := multipart.NewReader(bytes.NewReader(body), params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			if p.FormName() == "file" {
				if gotFile != nil {
					*gotFile = p.FileName()
				}
				if gotContent != nil {
					b, err := io.ReadAll(p)
					require.NoError(t, err)
					*gotContent = b
				}
			}
		}
		if cap != nil {
			*cap = &capturedReq{
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
	c, err := NewClient(Config{APIKey: "test-token", BaseURL: srv.URL})
	require.NoError(t, err)
	return c
}

func TestUploadFile(t *testing.T) {
	tests := []struct {
		name    string
		stub    responseStub
		wantErr string
	}{
		{"success 200", responseStub{http.StatusOK, "application/json", uploadOKBody}, ""},
		{"400 invalid request", responseStub{http.StatusBadRequest, "application/json", `{"message":"bad"}`}, "invalid request: bad"},
		{"401 unauthorized", responseStub{http.StatusUnauthorized, "application/json", `{}`}, "unauthorized"},
		{"403 forbidden", responseStub{http.StatusForbidden, "application/json", `{"message":"no"}`}, "forbidden: no"},
		{"404 space not found", responseStub{http.StatusNotFound, "application/json", `{"message":"no space"}`}, "space not found: no space"},
		{"410 space deleted", responseStub{http.StatusGone, "application/json", `{"message":"gone"}`}, "space deleted: gone"},
		{"500 server error", responseStub{http.StatusInternalServerError, "application/json", `{"message":"boom"}`}, "server error: boom"},
		{"unexpected status", responseStub{http.StatusBadGateway, "text/plain", "x"}, "unexpected response (HTTP 502)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newUploadClient(t, tc.stub, nil, nil, nil)

			res, err := c.UploadFile(context.Background(), "sp1", "photo.png", strings.NewReader("hello bytes"))

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, res)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			require.NotNil(t, res.ObjectId)
			assert.Equal(t, "file1", *res.ObjectId)
			require.NotNil(t, res.Name)
			assert.Equal(t, "photo.png", *res.Name)
			require.NotNil(t, res.SizeInBytes)
			assert.Equal(t, 11, *res.SizeInBytes)
		})
	}

	t.Run("streams file part with original filename and bytes", func(t *testing.T) {
		var (
			cap       *capturedReq
			fileName  string
			fileBytes []byte
		)
		c := newUploadClient(t, responseStub{http.StatusOK, "application/json", uploadOKBody}, &cap, &fileName, &fileBytes)

		_, err := c.UploadFile(context.Background(), "sp1", "photo.png", strings.NewReader("hello bytes"))
		require.NoError(t, err)
		require.NotNil(t, cap)
		assert.Equal(t, http.MethodPost, cap.Method)
		assert.Equal(t, "/v1/spaces/sp1/files", cap.Path)
		assert.Contains(t, cap.Header.Get("Content-Type"), "multipart/form-data")
		assert.Equal(t, "photo.png", fileName)
		assert.Equal(t, "hello bytes", string(fileBytes))
	})
}

func TestDownloadFile(t *testing.T) {
	t.Run("success returns body and content type", func(t *testing.T) {
		var captured *capturedReq
		c := newStubClient(t, responseStub{
			status:      http.StatusOK,
			contentType: "image/png",
			body:        "PNG-BYTES",
		}, &captured)

		rc, ct, err := c.DownloadFile(context.Background(), "sp1", "file1")
		require.NoError(t, err)
		require.NotNil(t, rc)
		t.Cleanup(func() { require.NoError(t, rc.Close()) })
		b, err := io.ReadAll(rc)
		require.NoError(t, err)
		assert.Equal(t, "PNG-BYTES", string(b))
		assert.Equal(t, "image/png", ct)
		require.NotNil(t, captured)
		assert.Equal(t, http.MethodGet, captured.Method)
		assert.Equal(t, "/v1/spaces/sp1/files/file1", captured.Path)
	})

	// DownloadFile maps every non-200 to a generic "unexpected response" error;
	// it has no per-status decoding, so we only sample a couple of statuses.
	for _, tc := range []struct {
		name string
		stub responseStub
	}{
		{"404 is unexpected", responseStub{http.StatusNotFound, "application/json", `{"message":"no"}`}},
		{"500 is unexpected", responseStub{http.StatusInternalServerError, "text/plain", "boom"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newStubClient(t, tc.stub, nil)
			rc, _, err := c.DownloadFile(context.Background(), "sp1", "file1")
			require.Error(t, err)
			assert.Nil(t, rc)
			assert.Contains(t, err.Error(), "unexpected response")
		})
	}
}

func TestDeleteFile(t *testing.T) {
	tests := []struct {
		name    string
		stub    responseStub
		wantErr string
	}{
		{"success 204 no body", responseStub{http.StatusNoContent, "", ""}, ""},
		{"400 invalid request", responseStub{http.StatusBadRequest, "application/json", `{"message":"bad"}`}, "invalid request: bad"},
		{"401 unauthorized", responseStub{http.StatusUnauthorized, "application/json", `{}`}, "unauthorized"},
		{"500 server error", responseStub{http.StatusInternalServerError, "application/json", `{"message":"boom"}`}, "server error: boom"},
		{"unexpected status", responseStub{http.StatusBadGateway, "text/plain", "x"}, "unexpected response (HTTP 502)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var captured *capturedReq
			c := newStubClient(t, tc.stub, &captured)

			err := c.DeleteFile(context.Background(), "sp1", "file1", false)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}

	t.Run("request is DELETE with skip_bin query reflecting flag", func(t *testing.T) {
		var captured *capturedReq
		c := newStubClient(t, responseStub{http.StatusNoContent, "", ""}, &captured)

		require.NoError(t, c.DeleteFile(context.Background(), "sp1", "file1", true))
		require.NotNil(t, captured)
		assert.Equal(t, http.MethodDelete, captured.Method)
		assert.Equal(t, "/v1/spaces/sp1/files/file1", captured.Path)
		assert.Equal(t, "true", captured.Query.Get("skip_bin"))
	})
}
