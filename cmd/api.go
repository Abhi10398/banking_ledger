package cmd

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"

	"awesomeProject/config"
	"awesomeProject/internal/appcontext"
	"awesomeProject/internal/factory"
	"awesomeProject/logger"
)

var apiCommand = &cobra.Command{
	Use:   "api",
	Short: "Start the API server",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		err := appcontext.Initiate()
		if err != nil {
			logger.Log.Error(err)
			return
		}

		controller := factory.InitializeController()
		portStr := os.Getenv("PORT")
		if portStr == "" {
			portStr = strconv.Itoa(int(cfg.Port))
		}
		port := ":" + portStr

		wg := sync.WaitGroup{}
		wg.Add(1)
		go handleGracefulShutdown(controller.App, &wg)

		go func() {
			logger.Log.Info("Starting HTTP server on port", port)
			if err := controller.App.Listen(port); err != nil {
				logger.Log.Errorf("HTTP server error: %v", err)
			}
		}()

		logger.Log.Infof("Server starting service on port %s", port)

		wg.Wait()
	},
}

func handleGracefulShutdown(fiberApp *fiber.App, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGINT,
		syscall.SIGTERM)

	s := <-sig
	logger.Get(ctx).Infof("got signal %v, attempting graceful shutdown", s)
	if err := fiberApp.ShutdownWithTimeout(60000 * time.Millisecond); err != nil {
		logger.Get(ctx).Infof("HTTP server shutdown error: %v", err)
	}
	logger.Get(ctx).Infof("HTTP graceful shutdown successful")
}
