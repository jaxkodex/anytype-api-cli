// Command anytype-api is a command line interface for the Anytype API.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	// Cancel the command context on Ctrl+C (and SIGTERM) so in-flight HTTP
	// requests are aborted promptly.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
