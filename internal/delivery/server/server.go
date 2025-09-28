package server

import (
	"betera-tz/internal/config"
	"betera-tz/internal/dto"
	"betera-tz/pkg/monitoring"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	_ "betera-tz/docs"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

type AppServer struct {
	Server *http.Server
	Metric *http.Server
}

func NewAppServer(scfg config.ServerConfig, acfg config.AppConfig, si ServerInterface, ps *monitoring.PrometheusSetup) *AppServer {
	r := chi.NewRouter()
	metricsMux := chi.NewMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metrics := &http.Server{
		Addr:    ":" + scfg.MetricPort,
		Handler: metricsMux,
	}
	logger := httplog.NewLogger("betera-tz-http", httplog.Options{
		JSON: true,
	})
	logFile, err := os.OpenFile(acfg.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(fmt.Errorf("failed to open log file: %w", err))
	}
	writer := io.MultiWriter(logFile, os.Stdout)
	logger = logger.Output(writer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.RedirectSlashes)
	r.Use(func(h http.Handler) http.Handler {
		return httplog.RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			httplog.LogEntrySetField(ctx, "type", "http")
			h.ServeHTTP(w, r)
		}))
	})
	r.Use(middleware.Recoverer)
	r.Use(MetricsMiddleware(ps))
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.ApiResponse{
			Code:    http.StatusOK,
			Message: "ok",
		})
	})

	h := HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    "/api/v1",
		BaseRouter: r,
	})

	return &AppServer{
		Server: &http.Server{
			Addr:         fmt.Sprintf("%s:%s", scfg.Host, scfg.Port),
			Handler:      h,
			ReadTimeout:  scfg.ReadTimeout,
			IdleTimeout:  scfg.IdleTimeout,
			WriteTimeout: scfg.WriteTimeout,
		},
		Metric: metrics,
	}
}

func (as *AppServer) MustClose(ctx context.Context) {
	if err := as.Metric.Shutdown(ctx); err != nil {
		panic(fmt.Errorf("failed to close metric server: %w", err))
	}
	if err := as.Server.Shutdown(ctx); err != nil {
		panic(fmt.Errorf("failed to close app server: %w", err))
	}
}

func MetricsMiddleware(ps *monitoring.PrometheusSetup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			duration := time.Since(start).Seconds()
			path := r.URL.Path
			method := r.Method
			status := fmt.Sprintf("%d", ww.Status())

			ps.HTTPRequestsTotal.WithLabelValues(path, method).Inc()
			ps.HTTPRequestDuration.WithLabelValues(path, method, status).Observe(duration)
			if ww.Status() >= 400 {
				ps.HTTPErrorTotal.WithLabelValues(path, method, status).Inc()
			}
		})
	}
}
