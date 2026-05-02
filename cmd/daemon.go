package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/iyear/tdl/app/daemon"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/pkg/consts"
)

func NewDaemon() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Start tdl as a local background daemon/web server",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logctx.From(cmd.Context()).Named("daemon")
			
			port := viper.GetInt("daemon-port")
			if port == 0 {
				port = 8080 // Default port
			}

			srv := daemon.NewServer(port, logger)

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			// Setup graceful shutdown
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigCh
				logger.Info("Received shutdown signal")
				cancel()
			}()

			fmt.Printf("Starting tdl daemon on http://localhost:%d\n", port)
			fmt.Printf("WebSocket endpoint: ws://localhost:%d/api/v1/ws\n", port)

			return srv.Start(ctx)
		},
	}

	cmd.Flags().Int("port", 8080, "Port for the daemon to listen on")
	_ = viper.BindPFlag("daemon-port", cmd.Flags().Lookup("port"))

	return cmd
}
