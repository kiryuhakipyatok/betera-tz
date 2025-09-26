package app

import (
	"betera-tz/internal/config"
	"betera-tz/internal/delivery/handlers"
	"betera-tz/internal/delivery/server"
	"betera-tz/internal/domain/repositories"
	"betera-tz/internal/domain/services"
	"betera-tz/pkg/logger"
	"betera-tz/pkg/storage"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	cfg := config.MustLoadConfig(os.Getenv("CONFIG_PATH"))
	logger := logger.NewLogger(cfg.App)
	logger.Info("config loaded")

	storage := storage.MustConnect(cfg.Storage)
	logger.Info("connected to postgres")
	defer func() {
		storage.Close()
		logger.Info("storage closed")
	}()

	taskRepository := repositories.NewTaskRepository(storage)

	taskService := services.NewTaskService(taskRepository, logger)

	taskHandler := handlers.NewTaskHandler(taskService)

	appServer := server.NewAppServer(cfg.Server, taskHandler)
	logger.Info("server created")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.CloseTimeout)
		defer cancel()
		appServer.MustClose(ctx)
		logger.Info("server closed")
	}()

	ln, err := net.Listen("tcp", appServer.Server.Addr)
	if err != nil {
		panic(fmt.Errorf("failed to bind: %w", err))
	}

	logger.Info("server started", "addr", ln.Addr().String())

	go func() {
		if err := appServer.Server.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("failed to start server: %w", err))
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("app is shutting down...")
}
