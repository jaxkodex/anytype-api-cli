package anytype

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
)

// UploadFile uploads the bytes from r as a multipart/form-data "file" field to
// the given space and returns the created file object. The filename is sent in
// the multipart part so the server can preserve the original name.
func (c *Client) UploadFile(ctx context.Context, spaceID, filename string, r io.Reader) (*api.FileUploadResponse, error) {
	params := &api.UploadFileParams{AnytypeVersion: APIVersion}

	// Stream the file into the request body via an io.Pipe so the whole file is
	// never buffered in memory: a goroutine writes the multipart form data while
	// the HTTP client reads from the other end of the pipe.
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		var err error
		defer func() {
			if err != nil {
				pw.CloseWithError(err)
			} else {
				pw.Close()
			}
		}()
		// writer.Close() writes the terminating multipart boundary; a failure
		// there would silently truncate the upload, so capture it. Deferred
		// functions run LIFO, so this closes the writer (setting err if it
		// fails) before the closure above propagates err to the pipe.
		defer func() {
			if cerr := writer.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			return
		}
		_, err = io.Copy(part, r)
	}()

	resp, err := c.api.UploadFileWithBodyWithResponse(ctx, spaceID, params, writer.FormDataContentType(), pr)
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
// given space. It returns the response body as an io.ReadCloser together with the
// Content-Type, which callers may use to pick a filename or extension. The caller
// is responsible for closing the returned reader. Streaming the body avoids
// buffering the entire file in memory.
func (c *Client) DownloadFile(ctx context.Context, spaceID, fileID string) (io.ReadCloser, string, error) {
	params := &api.DownloadFileParams{AnytypeVersion: APIVersion}

	resp, err := c.api.DownloadFile(ctx, spaceID, fileID, params)
	if err != nil {
		return nil, "", fmt.Errorf("download file request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return resp.Body, resp.Header.Get("Content-Type"), nil
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
