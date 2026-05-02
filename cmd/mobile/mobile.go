package mobile

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/iyear/tdl/app/daemon"
)

var (
	globalCancel context.CancelFunc
)

// StartEngine starts the TDL daemon engine for mobile.
// It will run a local HTTP server and WebSocket hub on the given port.
func StartEngine(port int) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if port == 0 {
		port = 8080
	}

	srv := daemon.NewServer(port, logger)

	ctx, cancel := context.WithCancel(context.Background())
	globalCancel = cancel

	fmt.Printf("Starting tdl mobile daemon on port %d\n", port)
	return srv.Start(ctx)
}

// StopEngine gracefully shuts down the running daemon engine.
func StopEngine() {
	if globalCancel != nil {
		globalCancel()
	}
}
