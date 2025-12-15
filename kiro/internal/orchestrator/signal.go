package orchestrator

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler sets up signal handling for graceful shutdown
func SetupSignalHandler(ctx context.Context, cleanup func()) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigChan:
			fmt.Printf("\n\nReceived signal %v, cleaning up...\n", sig)
			if cleanup != nil {
				cleanup()
			}
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx
}
