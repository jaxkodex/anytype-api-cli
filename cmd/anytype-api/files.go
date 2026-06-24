package main

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/jaxkodex/anytype-api-cli/internal/api"
	"github.com/spf13/cobra"
)

func newFilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "Upload, download and delete files in a space",
		Long: `Work with the files stored within a space.

Files are scoped to a space, so every subcommand requires a --space id. List the
spaces accessible to you in the Anytype app, or via the API, to find one.`,
	}

	cmd.AddCommand(newFilesUploadCmd())
	cmd.AddCommand(newFilesDownloadCmd())
	cmd.AddCommand(newFilesDeleteCmd())

	return cmd
}

func newFilesUploadCmd() *cobra.Command {
	var (
		spaceID string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "upload [path]",
		Short: "Upload a local file to a space",
		Args:  cobra.ExactArgs(1),
		Example: `  # Upload a file
  anytype-api files upload ./photo.png --space bafyre...

  # Emit raw JSON for scripting
  anytype-api files upload ./photo.png --space bafyre... --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			f, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			defer f.Close()

			client, err := newClient()
			if err != nil {
				return err
			}

			result, err := client.UploadFile(cmd.Context(), spaceID, filepath.Base(path), f)
			if err != nil {
				return err
			}

			if jsonOut {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Uploaded file:")
			return printFileUpload(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id to upload the file to (required)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output raw JSON")
	cmd.MarkFlagRequired("space")

	return cmd
}

func newFilesDownloadCmd() *cobra.Command {
	var (
		spaceID string
		output  string
	)

	cmd := &cobra.Command{
		Use:   "download [file-id]",
		Short: "Download a file's contents",
		Args:  cobra.ExactArgs(1),
		Long: `Download the raw bytes of a file.

By default the contents are written to a file named after the file id with an
extension inferred from the response media type. Override the destination with
--output (use "-" to write to stdout). When stdout is not a terminal and no
--output is given, the bytes are streamed to stdout so the command pipes safely.`,
		Example: `  # Download to a file in the current directory
  anytype-api files download bafyre... --space bafyre...

  # Download to a specific path
  anytype-api files download bafyre... --space bafyre... --output photo.png

  # Stream to stdout for piping
  anytype-api files download bafyre... --space bafyre... --output - > photo.png`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fileID := args[0]

			client, err := newClient()
			if err != nil {
				return err
			}

			body, contentType, err := client.DownloadFile(cmd.Context(), spaceID, fileID)
			if err != nil {
				return err
			}
			defer body.Close()

			// Decide where the bytes go. An explicit --output wins; "-" means
			// stdout. With no --output, stream to stdout when it is not a TTY
			// (so piping works), otherwise write to a sensible local filename.
			dest := output
			if dest == "" && !isTerminal(os.Stdout) {
				dest = "-"
			}

			if dest == "-" {
				_, err := io.Copy(cmd.OutOrStdout(), body)
				return err
			}

			if dest == "" {
				dest = fileID + extensionFor(contentType)
			}
			out, err := os.Create(dest)
			if err != nil {
				return fmt.Errorf("creating file: %w", err)
			}
			defer out.Close()

			written, err := io.Copy(out, body)
			if err != nil {
				return fmt.Errorf("writing file: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote %d bytes to %s\n", written, dest)
			return nil
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the file belongs to (required)")
	cmd.Flags().StringVarP(&output, "output", "o", "", `Destination path ("-" for stdout)`)
	cmd.MarkFlagRequired("space")

	return cmd
}

func newFilesDeleteCmd() *cobra.Command {
	var (
		spaceID string
		skipBin bool
		yes     bool
	)

	cmd := &cobra.Command{
		Use:   "delete [file-id]",
		Short: "Delete a file",
		Args:  cobra.ExactArgs(1),
		Long: `Delete a file. By default the file is moved to the bin and can be restored;
pass --skip-bin to delete it permanently.

You are asked to confirm before the file is deleted; pass --yes to skip the
prompt when scripting.`,
		Example: `  # Delete a file, with confirmation
  anytype-api files delete bafyre... --space bafyre...

  # Delete permanently, without prompting
  anytype-api files delete bafyre... --space bafyre... --skip-bin --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fileID := args[0]

			if !yes {
				prompt := fmt.Sprintf("Delete file %s? It is moved to the bin and can be restored.", fileID)
				if skipBin {
					prompt = fmt.Sprintf("Permanently delete file %s? This cannot be undone.", fileID)
				}
				confirmed, err := confirm(cmd, prompt)
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			if err := client.DeleteFile(cmd.Context(), spaceID, fileID, skipBin); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted file %s.\n", fileID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space id the file belongs to (required)")
	cmd.Flags().BoolVar(&skipBin, "skip-bin", false, "Permanently delete instead of moving to the bin")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip the confirmation prompt")
	cmd.MarkFlagRequired("space")

	return cmd
}

func printFileUpload(cmd *cobra.Command, result *api.FileUploadResponse) error {
	out := cmd.OutOrStdout()

	if result == nil {
		fmt.Fprintln(out, "No file returned.")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", orDash(deref(result.Name)))
	fmt.Fprintf(w, "MIME:\t%s\n", orDash(deref(result.Media)))
	fmt.Fprintf(w, "Size:\t%d bytes\n", derefInt(result.SizeInBytes))
	fmt.Fprintf(w, "ID:\t%s\n", orDash(deref(result.ObjectId)))
	return w.Flush()
}

// isTerminal reports whether f is connected to a terminal, used to decide
// whether downloaded bytes should be streamed to stdout for piping.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// extensionFor returns a leading-dot file extension for the given media type,
// or an empty string when none can be determined.
func extensionFor(contentType string) string {
	if contentType == "" {
		return ""
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = contentType
	}
	exts, err := mime.ExtensionsByType(mediaType)
	if err != nil || len(exts) == 0 {
		return ""
	}
	return exts[0]
}
