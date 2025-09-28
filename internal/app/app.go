package app

import (
	"betera-tz/internal/config"
	"betera-tz/internal/delivery/handlers"
	"betera-tz/internal/delivery/server"
	"betera-tz/internal/domain/repositories"
	"betera-tz/internal/domain/services"
	"betera-tz/internal/workers"
	"betera-tz/pkg/logger"
	"betera-tz/pkg/monitoring"
	"betera-tz/pkg/queue"
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
	logger := logger.NewLogger(cfg.App, "app")
	logger.Info("config loaded")

	storage := storage.MustConnect(cfg.Storage)
	logger.Info("connected to postgres")
	defer func() {
		storage.Close()
		logger.Info("storage closed")
	}()

	consumer := queue.NewConsumer(cfg.Queue)

	producer := queue.NewProducer(cfg.Queue)

	taskRepository := repositories.NewTaskRepository(storage)

	taskWorker := workers.NewTaskWorker(consumer, producer, logger, taskRepository)

	taskService := services.NewTaskService(taskRepository, logger, producer)

	taskHandler := handlers.NewTaskHandler(taskService)

	prometheusSetup := monitoring.NewPrometheusSetup(cfg.Monitoring)

	go func() {
		taskWorker.MustStart()
	}()

	logger.Info("task worker started")
	appServer := server.NewAppServer(cfg.Server, cfg.App, taskHandler, prometheusSetup)
	logger.Info("server created")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.CloseTimeout)
		defer cancel()
		appServer.MustClose(ctx)
		logger.Info("servers closed")
	}()

	appListener, err := net.Listen("tcp", appServer.Server.Addr)
	if err != nil {
		panic(fmt.Errorf("failed to bind: %w", err))
	}

	metricsListener, err := net.Listen("tcp", appServer.Metric.Addr)
	if err != nil {
		panic(fmt.Errorf("failed to bind: %w", err))
	}

	logger.Info("server started")

	go func() {
		if err := appServer.Server.Serve(appListener); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("failed to start app server: %w", err))
		}
	}()

	go func() {
		if err := appServer.Metric.Serve(metricsListener); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("failed to start metrics server: %w", err))
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("app is shutting down...")
}
